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
    it "ANNICT_SENTRY_ENVIRONMENT未指定時はRails.envが使われること" do
      configuration = build_configuration("ANNICT_SENTRY_ENVIRONMENT" => nil)

      expect(configuration.environment).to eq(Rails.env)
    end

    it "ANNICT_SENTRY_ENVIRONMENT指定時はその値が使われること" do
      configuration = build_configuration("ANNICT_SENTRY_ENVIRONMENT" => "staging")

      expect(configuration.environment).to eq("staging")
    end
  end

  describe "release" do
    it "ANNICT_SENTRY_RELEASE未指定時はreleaseを設定せずSDKの自動検出に委ねること" do
      # ANNICT_SENTRY_RELEASEが空のときreleaseを空文字で上書きしないという実装判断を回帰防止する。
      # SDKの自動検出は送信が許可された環境でのみ動作するため、未設定のままにしておく必要がある。
      configuration = build_configuration("ANNICT_SENTRY_RELEASE" => nil)

      expect(configuration.release).to be_nil
    end

    it "ANNICT_SENTRY_RELEASE指定時はその値がreleaseタグになること" do
      configuration = build_configuration("ANNICT_SENTRY_RELEASE" => "2026.09.16-1")

      expect(configuration.release).to eq("2026.09.16-1")
    end
  end

  describe "traces_sample_rate" do
    it "ANNICT_SENTRY_TRACES_SAMPLE_RATE未指定時は0.5 (既定値) になること" do
      configuration = build_configuration("ANNICT_SENTRY_TRACES_SAMPLE_RATE" => nil)

      expect(configuration.traces_sample_rate).to eq(0.5)
    end

    it "ANNICT_SENTRY_TRACES_SAMPLE_RATE指定時はその値が使われること" do
      configuration = build_configuration("ANNICT_SENTRY_TRACES_SAMPLE_RATE" => "0.2")

      expect(configuration.traces_sample_rate).to eq(0.2)
    end

    it "ANNICT_SENTRY_TRACES_SAMPLE_RATEが範囲外のときは既定値へフォールバックすること" do
      configuration = build_configuration("ANNICT_SENTRY_TRACES_SAMPLE_RATE" => "1.5")

      expect(configuration.traces_sample_rate).to eq(0.5)
    end
  end

  describe "再評価の分離" do
    it "評価に使った環境変数がexampleの終了後に復元されること" do
      before_value = ENV["ANNICT_SENTRY_ENVIRONMENT"]

      build_configuration("ANNICT_SENTRY_ENVIRONMENT" => "staging")

      expect(ENV["ANNICT_SENTRY_ENVIRONMENT"]).to eq(before_value)
    end

    it "再評価がSentryのグローバル設定を変更しないこと" do
      before_environment = config.environment
      before_release = config.release

      build_configuration("ANNICT_SENTRY_ENVIRONMENT" => "staging", "ANNICT_SENTRY_RELEASE" => "2026.09.16-1")

      expect(config.environment).to eq(before_environment)
      expect(config.release).to eq(before_release)
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

  # 環境変数に応じて変わる設定を、プロセスの環境変数に左右されずに検証するためのヘルパー。
  # initializerが`Sentry.init`へ渡すブロックだけを取り出し、使い捨ての`Sentry::Configuration`へ適用する。
  # `Sentry.init`は呼ばないため、起動時に構築されたグローバル設定 (`Sentry.configuration`) は変化しない。
  def build_configuration(env_overrides)
    initializer_block = nil
    allow(Sentry).to receive(:init) { |&block| initializer_block = block }
    load Rails.root.join("config/initializers/sentry.rb").to_s

    raise "initializerがSentry.initを呼び出していません" if initializer_block.nil?

    configuration = Sentry::Configuration.new
    with_env(env_overrides) { initializer_block.call(configuration) }
    configuration
  end

  def with_env(overrides)
    previous_values = overrides.keys.index_with { |key| ENV[key] }

    apply_env(overrides)

    begin
      yield
    ensure
      apply_env(previous_values)
    end
  end

  def apply_env(values)
    values.each do |key, value|
      if value.nil?
        ENV.delete(key)
      else
        ENV[key] = value
      end
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
