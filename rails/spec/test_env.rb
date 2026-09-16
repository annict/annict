# typed: false
# frozen_string_literal: true

require "dotenv"

# テストの期待値が前提とする設定を `.env.test` の値で固定する。
# dotenvは既存の環境変数を上書きしないため、開発用の設定 (1Password経由で読み込む `.env.local` など) が
# プロセスに渡っていると、季節・ホスト・Cookieのドメインがテストの前提と食い違う。
# DB接続情報や秘密情報は外部から渡された値を引き継げるよう、固定するキーを明示的に列挙する。
module TestEnv
  ENV_FILE = File.expand_path("../.env.test", __dir__)

  KEYS = %w[
    ANNICT_API_DOMAIN
    ANNICT_ASSET_URL
    ANNICT_BASIC_AUTH
    ANNICT_COOKIE_DOMAIN
    ANNICT_CURRENT_SEASON
    ANNICT_DOMAIN
    ANNICT_EN_DOMAIN
    ANNICT_EN_HOST
    ANNICT_EN_URL
    ANNICT_HOST
    ANNICT_MAINTENANCE_MODE
    ANNICT_NEXT_SEASON
    ANNICT_PREVIOUS_SEASON
    ANNICT_URL
    TZ
  ].freeze

  def self.apply!(env: ENV, path: ENV_FILE)
    values = Dotenv.parse(path)
    missing_keys = KEYS - values.keys
    if missing_keys.any?
      raise KeyError, "#{path} にテスト用の固定値が定義されていません: #{missing_keys.join(", ")}"
    end

    KEYS.each { |key| env[key] = values.fetch(key) }
  end
end
