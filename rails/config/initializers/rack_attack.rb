# typed: false
# frozen_string_literal: true

require "digest"
require_relative "../../lib/rate_limit_store"

# APIホスト (`ANNICT_API_DOMAIN`) に掛けるレート制限。
# Webサイトのレート制限はGo版で掛けるため、APIホスト以外のリクエストは数えない。
# 数値と数え方は開発者向けドキュメント (Annictのレート制限) と揃える。
module ApiRateLimit
  PERIOD = 1.minute.to_i
  # アクセストークンと接続元IPの組で数える
  TOKEN_LIMIT = 600
  # トークンを付けないリクエストを接続元IPで数える
  IP_LIMIT = 60
  # `rate_limits` の `key` はthrottleの名前と識別子をつないだものになるため、名前にAPIの名前空間を付ける
  TOKEN_THROTTLE = "api:token"
  IP_THROTTLE = "api:ip"
  THROTTLES = [TOKEN_THROTTLE, IP_THROTTLE].freeze

  def self.remote_ip(req)
    (req.env["HTTP_CF_CONNECTING_IP"] || req.env["action_dispatch.remote_ip"] || req.ip).to_s
  end

  def self.api_host?(req)
    api_domain = ENV.fetch("ANNICT_API_DOMAIN", "")
    api_domain.present? && req.env["SERVER_NAME"] == api_domain
  end

  # Doorkeeperがトークンを受け取る `Authorization: Bearer` ヘッダーと `access_token` パラメーターから取り出す。
  # トークンはカウンターのキーに残すため、ハッシュ化した値を返す
  def self.token_digest(req)
    token = req.get_header("HTTP_AUTHORIZATION").to_s[/\ABearer\s+(.+)\z/i, 1].presence ||
      req.params["access_token"].presence
    Digest::SHA256.hexdigest(token) if token.is_a?(String)
  rescue Rack::QueryParser::ParameterTypeError, Rack::QueryParser::InvalidParameterError, Rack::QueryParser::QueryLimitError, EOFError
    nil
  end

  # 窓が終わるまでの秒数 (`Retry-After` と `RateLimit-Reset` に使う)
  def self.seconds_until_reset(data, now = Time.now.to_i)
    data[:period] - (now % data[:period])
  end

  # APIホストのすべての応答に残量のヘッダーを付けるミドルウェア。
  # rack-attackが判定の結果をリクエストに残すため、rack-attackの外側に置いて応答を受け取ってから読む
  class Headers
    def initialize(app)
      @app = app
    end

    def call(env)
      # 残量はrack-attackが数えた時点の窓の値のため、残り秒数もアプリケーションの処理時間を含めずにその時点から求める
      now = Time.now.to_i
      status, headers, body = @app.call(env)

      data = env["rack.attack.throttle_data"]&.values_at(*THROTTLES)&.compact&.first
      if data
        headers["RateLimit-Limit"] = data[:limit].to_s
        headers["RateLimit-Remaining"] = [data[:limit] - data[:count], 0].max.to_s
        headers["RateLimit-Reset"] = ApiRateLimit.seconds_until_reset(data, now).to_s
      end

      [status, headers, body]
    end
  end
end

Rack::Attack.enabled = !Rails.env.test?
Rack::Attack.cache.store = RateLimitStore.new(period: ApiRateLimit::PERIOD)
Rails.application.config.middleware.insert_before(Rack::Attack, ApiRateLimit::Headers)

Rack::Attack.throttle(ApiRateLimit::TOKEN_THROTTLE, limit: ApiRateLimit::TOKEN_LIMIT, period: ApiRateLimit::PERIOD) do |req|
  next unless ApiRateLimit.api_host?(req)

  digest = ApiRateLimit.token_digest(req)
  "#{digest}:#{ApiRateLimit.remote_ip(req)}" if digest
end

Rack::Attack.throttle(ApiRateLimit::IP_THROTTLE, limit: ApiRateLimit::IP_LIMIT, period: ApiRateLimit::PERIOD) do |req|
  next unless ApiRateLimit.api_host?(req)

  ApiRateLimit.remote_ip(req) unless ApiRateLimit.token_digest(req)
end

# REST v1のエラー形式で返す。GraphQLとOAuthのエンドポイントも同じ形式にする
Rack::Attack.throttled_responder = lambda do |req|
  retry_after = ApiRateLimit.seconds_until_reset(req.env["rack.attack.match_data"])
  error = {
    type: "rate_limit_exceeded",
    message: "リクエストが多すぎます。しばらく待ってから再度お試しください",
    developer_message: "429 Too Many Requests. Retry after #{retry_after} seconds"
  }

  [429, {"Content-Type" => "application/json; charset=utf-8", "Retry-After" => retry_after.to_s}, [{errors: [error]}.to_json]]
end

ActiveSupport::Notifications.subscribe(/rack_attack/) do |_name, _start, _finish, _request_id, payload|
  req = payload[:request]

  next unless req.env["rack.attack.match_type"] == :throttle

  # クエリ文字列にアクセストークンが含まれうるため、パスだけを残す
  Rails.logger.info("Rate limit hit (#{req.env["rack.attack.matched"]}): #{ApiRateLimit.remote_ip(req)} #{req.request_method} #{req.path}")
end
