# typed: false
# frozen_string_literal: true

# Lightweight stand-ins for the Sentry event object graph. The before_send
# hook mutates `request.data` / `request.headers` / `extra` / breadcrumb
# `data` in place, so the doubles need both readers and writers; Struct
# (instead of OpenStruct) avoids the Ruby 3.5 ostruct default-gem warning
# and keeps the spec free of SDK initialization coupling.
#
# [Ja] Sentry イベントのオブジェクトグラフの軽量ダブル。before_send フックは
# `request.data` / `request.headers` / `extra` / breadcrumb の `data` を
# 直接書き換えるため、ダブルには reader と writer の両方が必要になる。
# Struct を採用することで Ruby 3.5 で警告対象になる ostruct への依存を避け、
# SDK の初期化に依存せずフィルタの挙動だけを検証できるようにしている。
SentrySpecEventDouble = Struct.new(:request, :extra, :breadcrumbs, keyword_init: true)
SentrySpecRequestDouble = Struct.new(:data, :headers, keyword_init: true)
SentrySpecBreadcrumbDouble = Struct.new(:data, keyword_init: true)
SentrySpecBreadcrumbBufferDouble = Struct.new(:buffer, keyword_init: true)

RSpec.describe "config/initializers/sentry.rb" do # rubocop:disable RSpec/DescribeClass
  let(:config) { Sentry.configuration }

  describe "enabled_environments" do
    it "本番以外では Sentry を有効化しないこと" do
      expect(config.enabled_environments).to eq(%w[production])
    end
  end

  describe "send_default_pii" do
    it "リクエストボディや Cookie の自動添付を無効化していること" do
      expect(config.send_default_pii).to be(false)
    end
  end

  describe "environment" do
    it "ANNICT_SENTRY_ENVIRONMENT 未指定時は Rails.env が使われること" do
      if ENV["ANNICT_SENTRY_ENVIRONMENT"].present?
        skip "ANNICT_SENTRY_ENVIRONMENT がセットされている環境ではこのケースを検証できない"
      end

      expect(config.environment).to eq(Rails.env)
    end
  end

  describe "release" do
    it "ANNICT_SENTRY_RELEASE 未指定時は release を設定せず SDK の自動検出に委ねること" do
      # When ANNICT_SENTRY_RELEASE is blank the initializer never assigns
      # config.release. The SDK's auto-detection only runs when sending is
      # allowed, so in the test environment the value stays blank.
      #
      # [Ja] ANNICT_SENTRY_RELEASE が空のとき release を空文字で上書きしない
      # という実装判断を回帰防止する。SDK の自動検出は送信が許可された環境で
      # のみ動作するため、テスト環境では空のままになる。
      if ENV["ANNICT_SENTRY_RELEASE"].present?
        skip "ANNICT_SENTRY_RELEASE がセットされている環境ではこのケースを検証できない"
      end

      expect(config.release).to be_blank
    end
  end

  describe "traces_sample_rate" do
    it "0.0〜1.0 の範囲に収まること" do
      expect(config.traces_sample_rate).to be_between(0.0, 1.0)
    end

    it "ANNICT_SENTRY_TRACES_SAMPLE_RATE 未指定時は 0.5 (既定値) になること" do
      if ENV["ANNICT_SENTRY_TRACES_SAMPLE_RATE"].present?
        skip "ANNICT_SENTRY_TRACES_SAMPLE_RATE がセットされている環境ではこのケースを検証できない"
      end

      expect(config.traces_sample_rate).to eq(0.5)
    end
  end

  describe "excluded_exceptions" do
    it "クライアント切断ノイズ (Errno::EPIPE) を除外していること" do
      expect(config.excluded_exceptions).to include("Errno::EPIPE")
    end

    it "クライアント切断ノイズ (Errno::ECONNRESET) を除外していること" do
      expect(config.excluded_exceptions).to include("Errno::ECONNRESET")
    end

    it "不正クエリ起因のノイズ (Rack::QueryParser::ParameterTypeError) を除外していること" do
      expect(config.excluded_exceptions).to include("Rack::QueryParser::ParameterTypeError")
    end

    it "SDK の既定除外 (ActionController::RoutingError) を保持していること" do
      expect(config.excluded_exceptions).to include("ActionController::RoutingError")
    end
  end

  describe "before_send" do
    let(:before_send) { config.before_send }

    it "lambda として登録されていること" do
      expect(before_send).to respond_to(:call)
    end

    it "イベント自体を返すこと (sentry-ruby 6.x では Hash を返すとイベントが破棄される)" do
      event = build_event(data: {"password" => "super-secret"})

      result = before_send.call(event, {})

      expect(result).to be(event)
    end

    it "password を [FILTERED] に置き換え、非センシティブなキーは保持すること" do
      event = build_event(data: {"password" => "super-secret", "username" => "annict-user"})

      result = before_send.call(event, {})

      expect(result.request.data["password"]).to eq("[FILTERED]")
      expect(result.request.data["username"]).to eq("annict-user")
    end

    it "authenticity_token を [FILTERED] に置き換えること" do
      event = build_event(data: {"authenticity_token" => "csrf-abc"})

      result = before_send.call(event, {})

      expect(result.request.data["authenticity_token"]).to eq("[FILTERED]")
    end

    it "api_key を [FILTERED] に置き換えること" do
      event = build_event(data: {"api_key" => "key-xyz"})

      result = before_send.call(event, {})

      expect(result.request.data["api_key"]).to eq("[FILTERED]")
    end

    it "email を [FILTERED] に置き換えること" do
      event = build_event(data: {"email" => "user@example.com"})

      result = before_send.call(event, {})

      expect(result.request.data["email"]).to eq("[FILTERED]")
    end

    it "ネストしたハッシュ内のセンシティブキーも置き換えること" do
      event = build_event(data: {"user" => {"password" => "secret"}})

      result = before_send.call(event, {})

      expect(result.request.data["user"]["password"]).to eq("[FILTERED]")
    end

    it "配列の中のハッシュに含まれるセンシティブキーも置き換えること" do
      event = build_event(data: {"items" => [{"password" => "secret-a"}, {"password" => "secret-b"}]})

      result = before_send.call(event, {})

      expect(result.request.data["items"][0]["password"]).to eq("[FILTERED]")
      expect(result.request.data["items"][1]["password"]).to eq("[FILTERED]")
    end

    it "リクエストヘッダーのセンシティブキー (X-CSRF-Token) も置き換えること" do
      event = build_event(headers: {"X-CSRF-Token" => "csrf-abc", "User-Agent" => "test-agent"})

      result = before_send.call(event, {})

      expect(result.request.headers["X-CSRF-Token"]).to eq("[FILTERED]")
      expect(result.request.headers["User-Agent"]).to eq("test-agent")
    end

    it "extra のセンシティブキーも置き換えること" do
      event = build_event(extra: {"password" => "secret", "work_id" => 42})

      result = before_send.call(event, {})

      expect(result.extra["password"]).to eq("[FILTERED]")
      expect(result.extra["work_id"]).to eq(42)
    end

    it "breadcrumbs の data に含まれるセンシティブキーも置き換えること" do
      crumb = SentrySpecBreadcrumbDouble.new(data: {"params" => {"password" => "secret"}})
      event = build_event(breadcrumbs: SentrySpecBreadcrumbBufferDouble.new(buffer: [crumb]))

      result = before_send.call(event, {})

      expect(result.breadcrumbs.buffer[0].data["params"]["password"]).to eq("[FILTERED]")
    end

    it "request.data が nil でも例外にならないこと" do
      event = build_event(data: nil)

      expect { before_send.call(event, {}) }.not_to raise_error
    end

    it "request 自体がない event でも例外にならないこと" do
      event = SentrySpecEventDouble.new(request: nil, extra: nil, breadcrumbs: nil)

      expect { before_send.call(event, {}) }.not_to raise_error
    end
  end

  def build_event(data: nil, headers: nil, extra: nil, breadcrumbs: nil)
    SentrySpecEventDouble.new(
      request: SentrySpecRequestDouble.new(data: data, headers: headers),
      extra: extra,
      breadcrumbs: breadcrumbs
    )
  end
end
