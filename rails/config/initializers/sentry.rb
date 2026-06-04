# typed: false
# frozen_string_literal: true

require_relative "../../lib/sentry_config"

Sentry.init do |config|
  # Prefer the ANNICT_-prefixed variable and fall back to the legacy
  # SENTRY_DSN until the production environment variables are migrated
  # (the fallback will be removed afterwards). When neither is set,
  # Sentry stays disabled.
  #
  # [Ja] ANNICT_ プレフィックス付きの環境変数を優先し、本番環境変数の移行が
  # 完了するまでは旧 SENTRY_DSN にフォールバックする (フォールバックは移行
  # 完了後に削除する)。どちらも未設定の場合、Sentry は無効のままになる。
  config.dsn = ENV["ANNICT_SENTRY_DSN"] || ENV["SENTRY_DSN"]

  config.breadcrumbs_logger = %i[active_support_logger http_logger]

  # Restrict Sentry to production so that a stray DSN in dev or test never
  # leaks events outside of the deployed environment.
  #
  # [Ja] DSN が誤って dev / test に設定されても Sentry に送信されないよう、
  # production のみ送信を有効化する。
  config.enabled_environments = %w[production]

  # Fall back to Rails.env when the operator does not override the tag.
  # [Ja] ANNICT_SENTRY_ENVIRONMENT 未指定時は Rails.env を environment タグとして使う。
  config.environment = ENV["ANNICT_SENTRY_ENVIRONMENT"].presence || Rails.env

  # Tag events with the deployed release so error grouping respects release
  # boundaries. Skip when the value is missing to keep Sentry's
  # auto-detection from being overridden with an empty string.
  #
  # [Ja] デプロイ単位でエラーを分離できるよう release タグを設定する。
  # 値が空の場合は Sentry の自動検出を空文字で上書きしないよう設定自体を行わない。
  release = ENV["ANNICT_SENTRY_RELEASE"].presence
  config.release = release if release

  config.traces_sample_rate = SentryConfig.resolve_traces_sample_rate(ENV["ANNICT_SENTRY_TRACES_SAMPLE_RATE"])

  # Pair with the before_send filter below: never auto-attach PII so the
  # scrub only has to defend against accidentally-captured request payloads.
  #
  # [Ja] 後段の before_send と合わせて PII を自動添付しない方針を明示する。
  # 想定外のリクエストペイロード捕捉に対する多層防御として動作する。
  config.send_default_pii = false

  # Drop client-disconnect and malformed-query noise on top of the SDK
  # defaults (e.g. ActionController::RoutingError is already excluded).
  #
  # [Ja] SDK のデフォルト除外 (ActionController::RoutingError など) に加え、
  # クライアント切断起因のノイズと不正クエリ起因のエラーを除外する。
  config.excluded_exceptions += %w[
    Errno::EPIPE
    Errno::ECONNRESET
    Rack::QueryParser::ParameterTypeError
  ]

  # Scrub sensitive values with Rails' parameter filter before events leave
  # the application. Since sentry-ruby 6.0, before_send must return a
  # Sentry::ErrorEvent (returning a filtered Hash makes the SDK discard the
  # event), so filter each hash-valued section in place and return the event.
  #
  # [Ja] イベントが送信される前に Rails の parameter filter でセンシティブ情報を
  # マスクする。sentry-ruby 6.0 以降、before_send は Sentry::ErrorEvent を
  # 返す必要がある (フィルタ済み Hash を返すと SDK がイベントを破棄する) ため、
  # ハッシュ値を持つ各セクションを直接マスクしてイベント自体を返す。
  filter = ActiveSupport::ParameterFilter.new(Rails.application.config.filter_parameters)
  config.before_send = lambda do |event, _hint|
    if (request = event.request)
      request.data = filter.filter(request.data) if request.data.is_a?(Hash)
      request.headers = filter.filter(request.headers) if request.headers.is_a?(Hash)
    end

    event.extra = filter.filter(event.extra) if event.extra.is_a?(Hash)

    # Breadcrumbs recorded by active_support_logger can carry controller
    # payloads (including request parameters), so scrub them as well.
    #
    # [Ja] active_support_logger が記録する breadcrumbs にはコントローラーの
    # ペイロード (リクエストパラメータを含む) が乗るため、こちらもマスクする。
    event.breadcrumbs&.buffer&.each do |crumb|
      crumb.data = filter.filter(crumb.data) if crumb.data.is_a?(Hash)
    end

    event
  end
end
