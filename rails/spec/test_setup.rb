# typed: false
# frozen_string_literal: true

require "uri"

# RSpecと同じ接続設定で破棄前の検証を行い、Go版へスキーマ構築を委譲する。
# 構築後は環境名の記録とアセットの準備を行う。
module TestSetup
  def self.load_environment!
    ENV["RAILS_ENV"] = "test"
    require_relative "test_env"
    TestEnv.apply!
    require_relative "../config/environment"
  end

  def self.validate_database!(config)
    # 検証用の一時DBも、用途が分かるテストDB名だけを許可する。
    unless config.env_name == "test" && /\Aannict_test(?:_[a-z0-9_]+)?\z/.match?(config.database.to_s)
      raise "初期化できるDBはannict_testまたはannict_test_で始まるテストDBだけです"
    end
  end

  def self.validate_metadata!(pool)
    metadata = pool.internal_metadata
    raise "テストDBの環境名を保存する設定が無効です" unless metadata.enabled?

    stored = metadata[:environment] if metadata.table_exists?
    if stored && stored != "test"
      raise "テストDBに別の環境名が記録されているため、初期化を中止します: #{stored}"
    end

    stored
  end

  def self.prepare_metadata!(pool)
    # Go版のテスト後は環境名が未登録になる。既存の環境名を上書きして保護を迂回しない。
    pool.internal_metadata.create_table_and_set_flags("test") if validate_metadata!(pool).nil?
  end

  def self.validate_existing_database!(config)
    ActiveRecord::Tasks::DatabaseTasks.with_temporary_connection(config) do |connection|
      validate_metadata!(connection.pool)
    end
  rescue ActiveRecord::NoDatabaseError
    # 初回のDB作成もGo側に任せる。認証失敗など、DB未作成以外のエラーは中止する。
    nil
  end

  def self.database_url(config)
    # Active RecordのPostgreSQLアダプタと同じキー変換で、解決済み設定をlibpqへ渡す。
    # URLに含まれなかったdatabase.ymlの接続設定も保持し、検証先と構築先を揃える。
    params = config.configuration_hash.compact
    params[:user] = params.delete(:username) if params[:username]
    params[:dbname] = params.delete(:database) if params[:database]
    params = params.slice(*(PG::Connection.conndefaults_hash.keys + [:requiressl]))
    # dbmateの既定はrequireなので、Rails/libpqのSSL設定を明示して初回作成にも揃える。
    params[:sslmode] ||= PG::Connection.conndefaults_hash[:sslmode]
    database = params.delete(:dbname)
    # libpqのURIではフォーム形式の+を空白へ戻さないため、空白は%20で表す。
    "postgresql:///#{URI.encode_www_form_component(database)}?#{URI.encode_www_form(params)}".gsub("+", "%20")
  end

  def self.reset_database!(config)
    # Go側のenvを再読込せず、検証済みの接続先だけを使用する。URLをコマンド引数やログへ出さない。
    # 親makeのコマンドライン変数が接続設定を上書きすることも防ぐ。
    system({"DATABASE_URL" => database_url(config), "MAKEFLAGS" => nil, "MAKEOVERRIDES" => nil, "MFLAGS" => nil}, "make", "--no-print-directory", "-C", Rails.root.join("../go").to_s,
      "db-setup-test", "OP_RUN_TEST=APP_ENV=test", exception: true)
  end

  def self.install_browser!(flags)
    system("yarn", "playwright", "install", *flags, ENV.fetch("ANNICT_CAPYBARA_BROWSER", "chromium"), exception: true)
  end

  def self.run!(flags)
    load_environment!
    Rails.application.load_tasks
    configs = ActiveRecord::Base.configurations.configs_for(env_name: "test")
    configs.each { |config| validate_database!(config) }

    puts "==> テストDBの接続先と環境名を確認..."
    configs.each { |config| validate_existing_database!(config) }
    configs.each { |config| reset_database!(config) }
    ActiveRecord::Tasks::DatabaseTasks.with_temporary_pool_for_each(env: "test") do |pool|
      prepare_metadata!(pool)
    end

    # test:prepareはActive Recordがdb:test:prepareを紐付けるため、ビルドの2タスクを直接呼ぶ。
    puts "==> アセットをビルド..."
    Rake::Task["javascript:build"].invoke
    Rake::Task["css:build"].invoke
    puts "==> Playwrightブラウザをインストール..."
    install_browser!(flags)
    puts "==> テスト環境のセットアップが完了しました"
  end
end

TestSetup.run!(ARGV) if $PROGRAM_NAME == __FILE__
