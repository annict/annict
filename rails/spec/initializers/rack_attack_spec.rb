# typed: false
# frozen_string_literal: true

RSpec.describe "config/initializers/rack_attack.rb" do # rubocop:disable RSpec/DescribeClass
  include ActiveSupport::Testing::TimeHelpers

  let(:inner_app) { ->(_env) { [200, {"Content-Type" => "text/plain"}, ["OK"]] } }
  let(:app) { ApiRateLimit::Headers.new(Rack::Attack.new(inner_app)) }

  around do |example|
    enabled = Rack::Attack.enabled
    api_domain = ENV["ANNICT_API_DOMAIN"]
    Rack::Attack.enabled = true
    ENV["ANNICT_API_DOMAIN"] = "api.annict.test"
    # 窓の途中で数え直しにならないよう、窓の始まりから15秒の時刻に固定する
    travel_to(Time.utc(2026, 10, 14, 12, 34, 15)) { example.run }
  ensure
    Rack::Attack.enabled = enabled
    ENV["ANNICT_API_DOMAIN"] = api_domain
  end

  def request(url, ip:, headers: {})
    Rack::MockRequest.new(app).get(url, {"HTTP_CF_CONNECTING_IP" => ip}.merge(headers))
  end

  it "APIホストより外側に残量のヘッダーを付けるミドルウェアを置いていること" do
    middlewares = Rails.application.middleware.middlewares

    expect(middlewares.index(ApiRateLimit::Headers)).to be < middlewares.index(Rack::Attack)
  end

  it "APIホスト以外のリクエストは数えないこと" do
    response = request("http://annict.test/works", ip: "192.0.2.10")

    expect(response.status).to eq(200)
    expect(response.headers).not_to include("RateLimit-Limit")
  end

  context "トークンを付けないとき" do
    it "接続元IP単位で1分あたり60件まで許可し、残量のヘッダーを付けること" do
      response = request("http://api.annict.test/v1/works", ip: "192.0.2.20")

      expect(response.status).to eq(200)
      expect(response.headers["RateLimit-Limit"]).to eq("60")
      expect(response.headers["RateLimit-Remaining"]).to eq("59")
      expect(response.headers["RateLimit-Reset"]).to eq("45")
    end

    it "上限を超えたら429とREST v1のエラー形式のJSONを返すこと" do
      59.times { request("http://api.annict.test/v1/works", ip: "192.0.2.21") }
      last_allowed = request("http://api.annict.test/v1/works", ip: "192.0.2.21")
      response = request("http://api.annict.test/v1/works", ip: "192.0.2.21")

      expect(last_allowed.status).to eq(200)
      expect(response.status).to eq(429)
      expect(response.headers["Content-Type"]).to eq("application/json; charset=utf-8")
      expect(response.headers["Retry-After"]).to eq("45")
      expect(response.headers["RateLimit-Remaining"]).to eq("0")
      expect(JSON.parse(response.body).dig("errors", 0, "type")).to eq("rate_limit_exceeded")
    end

    it "別のリクエストが窓を進めても応答時刻から待ち時間を計算すること" do
      60.times { request("http://api.annict.test/v1/works", ip: "192.0.2.22") }
      allow(Rack::Attack.cache).to receive(:last_epoch_time).and_return(Time.utc(2026, 10, 14, 12, 35).to_i)

      response = request("http://api.annict.test/v1/works", ip: "192.0.2.22")

      expect(response.status).to eq(429)
      expect(response.headers["Retry-After"]).to eq("45")
      expect(response.headers["RateLimit-Reset"]).to eq("45")
    end

    context "アプリケーションの処理中に窓が切り替わったとき" do
      let(:inner_app) do
        lambda do |_env|
          travel 50.seconds
          [200, {"Content-Type" => "text/plain"}, ["OK"]]
        end
      end

      it "数えた時点の窓の残り秒数を返すこと" do
        response = request("http://api.annict.test/v1/works", ip: "192.0.2.23")

        expect(response.headers["RateLimit-Remaining"]).to eq("59")
        expect(response.headers["RateLimit-Reset"]).to eq("45")
      end
    end
  end

  context "トークンを付けたとき" do
    it "トークンと接続元IPの組で1分あたり600件まで数えること" do
      first = request("http://api.annict.test/v1/works", ip: "192.0.2.30", headers: {"HTTP_AUTHORIZATION" => "Bearer token-a"})
      # access_tokenパラメーターで渡しても同じトークンとして数える
      second = request("http://api.annict.test/v1/works?access_token=token-a", ip: "192.0.2.30")
      # 接続元IPが異なれば別に数える
      other_ip = request("http://api.annict.test/v1/works", ip: "192.0.2.31", headers: {"HTTP_AUTHORIZATION" => "Bearer token-a"})

      expect(first.headers["RateLimit-Limit"]).to eq("600")
      expect(second.headers["RateLimit-Remaining"]).to eq("598")
      expect(other_ip.headers["RateLimit-Remaining"]).to eq("599")
    end

    it "カウンターのキーにトークンをそのまま残さないこと" do
      request("http://api.annict.test/v1/works", ip: "192.0.2.32", headers: {"HTTP_AUTHORIZATION" => "Bearer token-b"})

      keys = ActiveRecord::Base.connection.select_values("SELECT key FROM rate_limits WHERE key LIKE '%192.0.2.32'")
      expect(keys).to eq(["api:token:#{Digest::SHA256.hexdigest("token-b")}:192.0.2.32"])
    end
  end

  describe "CORS", type: :request do
    it "ブラウザから超過時と残量のヘッダーを読めるよう公開していること" do
      get "/v1/works", headers: {"Origin" => "https://example.com"}

      exposed = response.headers["Access-Control-Expose-Headers"].split(/,\s*/)
      expect(exposed).to include("Retry-After", "RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset")
    end
  end
end
