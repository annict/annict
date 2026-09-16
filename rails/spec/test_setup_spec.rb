# typed: false
# frozen_string_literal: true

require "open3"
require "rake"
require "tmpdir"
require "fileutils"
require_relative "test_setup"

RSpec.describe TestSetup do
  describe ".load_environment!" do
    it "Rails初期化前に固定値を適用し、DB接続とブラウザ設定を引き継ぐこと" do
      # -eのスクリプトはロケールのエンコーディングで解釈されるため、LANG未設定の環境でも日本語を読めるよう明示する。
      script = <<~RUBY_SCRIPT
        # encoding: utf-8
        require_relative "spec/test_setup"
        TestSetup.load_environment!
        expected = Dotenv.parse(TestEnv::ENV_FILE).slice(*TestEnv::KEYS)
        abort "固定値が不一致です" unless ENV.to_h.slice(*TestEnv::KEYS) == expected
        abort "アセットURLが不一致です" unless Rails.application.config.asset_host == expected.fetch("ANNICT_ASSET_URL")
        abort "DB接続が不一致です" unless ActiveRecord::Base.connection_db_config.database == "annict_test_loader_probe"
        abort "ブラウザが不一致です" unless ENV["ANNICT_CAPYBARA_BROWSER"] == "firefox"
      RUBY_SCRIPT
      output, status = Open3.capture2e({
        "ANNICT_CURRENT_SEASON" => "2025-autumn",
        "ANNICT_COOKIE_DOMAIN" => ".example.com",
        "ANNICT_ASSET_URL" => "https://example.com",
        "DATABASE_URL" => "postgresql:///annict_test_loader_probe",
        "ANNICT_CAPYBARA_BROWSER" => "firefox"
      }, "bundle", "exec", "ruby", "-e", script, chdir: Rails.root)

      expect(status.success?).to be(true), output
    end
  end

  describe ".validate_database!" do
    it "通常のテストDBと検証用のテストDBを許可すること" do
      %w[annict_test annict_test_probe_123].each do |database|
        config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary", adapter: "postgresql", database:)
        expect { TestSetup.validate_database!(config) }.not_to raise_error
      end
    end

    it "DATABASE_URLが開発DBを指す場合は拒否すること" do
      config = ActiveRecord::DatabaseConfigurations::UrlConfig.new("test", "primary", "postgresql:///annict_development", adapter: "postgresql", database: "annict_test")

      expect { TestSetup.validate_database!(config) }.to raise_error(RuntimeError, /初期化できるDB/)
    end

    it "test以外の環境はDB名がテスト用でも拒否すること" do
      config = ActiveRecord::DatabaseConfigurations::HashConfig.new("development", "primary", adapter: "postgresql", database: "annict_test")

      expect { TestSetup.validate_database!(config) }.to raise_error(RuntimeError, /初期化できるDB/)
    end
  end

  describe ".prepare_metadata!" do
    self.use_transactional_tests = false

    let(:pool) { ActiveRecord::Base.connection_pool }
    let(:use_metadata_table) { true }

    # 保護処理が共有テストDBの環境名やデータを壊さないよう、使い捨てDBで検証する。
    around do |example|
      config = ActiveRecord::Base.connection_db_config.configuration_hash.merge(database: "annict_test_guard_#{SecureRandom.hex(6)}", use_metadata_table:)
      admin = PG.connect(config.slice(:host, :port, :password).merge(user: config[:username], dbname: "postgres"))
      database = PG::Connection.quote_ident(config.fetch(:database))
      admin.exec("CREATE DATABASE #{database}")
      begin
        db_config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary", config)
        ActiveRecord::Tasks::DatabaseTasks.with_temporary_connection(db_config) do
          example.run
        end
      ensure
        admin.exec("DROP DATABASE #{database}")
        admin.close
      end
    end

    context "環境名の保存が無効な場合" do
      let(:use_metadata_table) { false }

      it "保護チェックが使えないため拒否すること" do
        expect { TestSetup.prepare_metadata!(pool) }.to raise_error(RuntimeError, /保存する設定が無効/)
      end
    end

    it "新規DBに環境名を登録できること" do
      TestSetup.prepare_metadata!(pool)

      expect(pool.internal_metadata[:environment]).to eq("test")
    end

    it "Go版と同様に環境名が未登録のテーブルへ補完できること" do
      pool.internal_metadata.create_table

      TestSetup.prepare_metadata!(pool)

      expect(pool.internal_metadata[:environment]).to eq("test")
    end

    it "登録済みのtest環境名を再実行で書き換えないこと" do
      pool.internal_metadata.create_table_and_set_flags("test", "既存の値")
      before = pool.with_connection { |connection| connection.select_all("SELECT * FROM ar_internal_metadata ORDER BY key").to_a }

      TestSetup.prepare_metadata!(pool)

      after = pool.with_connection { |connection| connection.select_all("SELECT * FROM ar_internal_metadata ORDER BY key").to_a }
      expect(after).to eq(before)
    end

    %w[development production].each do |environment|
      it "#{environment}の環境名を拒否して既存データと環境名を維持すること" do
        pool.internal_metadata.create_table_and_set_flags(environment)
        pool.with_connection do |connection|
          connection.execute("CREATE TABLE guard_records (value text)")
          connection.execute("INSERT INTO guard_records VALUES ('保持する値')")
        end

        expect { TestSetup.prepare_metadata!(pool) }.to raise_error(RuntimeError, /別の環境名/)
        expect(pool.internal_metadata[:environment]).to eq(environment)
        expect(pool.with_connection { |connection| connection.select_value("SELECT value FROM guard_records") }).to eq("保持する値")
      end
    end
  end

  describe ".validate_existing_database!" do
    it "DB未作成以外の接続エラーは握り潰さずに中止すること" do
      config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary",
        ActiveRecord::Base.connection_db_config.configuration_hash.merge(database: "annict_test_unreachable", port: 1))

      expect { TestSetup.validate_existing_database!(config) }.to raise_error(ActiveRecord::ConnectionNotEstablished)
    end
  end

  describe ".install_browser!" do
    around do |example|
      previous = ENV["ANNICT_CAPYBARA_BROWSER"]
      begin
        example.run
      ensure
        ENV["ANNICT_CAPYBARA_BROWSER"] = previous
      end
    end

    it "未指定時はChromiumだけを導入すること" do
      ENV.delete("ANNICT_CAPYBARA_BROWSER")
      allow(TestSetup).to receive(:system)

      TestSetup.install_browser!([])

      expect(TestSetup).to have_received(:system).with("yarn", "playwright", "install", "chromium", exception: true)
    end

    it "選択したブラウザとCI用のフラグをインストーラに渡すこと" do
      ENV["ANNICT_CAPYBARA_BROWSER"] = "firefox"
      allow(TestSetup).to receive(:system)

      TestSetup.install_browser!(["--with-deps"])

      expect(TestSetup).to have_received(:system).with("yarn", "playwright", "install", "--with-deps", "firefox", exception: true)
    end
  end

  describe ".run!" do
    let(:calls) { [] }
    let(:pool) { instance_double(ActiveRecord::ConnectionAdapters::ConnectionPool) }

    # DBの保護は手順の順序で成り立つため、各手順を記録用のスタブに置き換えて呼び出し順を検証する。
    before do
      allow(TestSetup).to receive(:puts)
      allow(TestSetup).to receive(:load_environment!) { calls << :load_environment! }
      allow(Rails.application).to receive(:load_tasks) { calls << :load_tasks }
      allow(ActiveRecord::Base.configurations).to receive(:configs_for).and_call_original
      allow(ActiveRecord::Base.configurations).to receive(:configs_for).with(env_name: "test").and_return(configs)
      allow(Rake::Task).to receive(:[]) do |name|
        instance_double(Rake::Task).tap { |task| allow(task).to receive(:invoke) { calls << name } }
      end
      allow(ActiveRecord::Tasks::DatabaseTasks).to receive(:with_temporary_pool_for_each).with(env: "test").and_yield(pool)
      allow(TestSetup).to receive(:validate_existing_database!) { |config| calls << [:validate_existing_database!, config.database] }
      allow(TestSetup).to receive(:reset_database!) { |config| calls << [:reset_database!, config.database] }
      allow(TestSetup).to receive(:prepare_metadata!) { |target| calls << [:prepare_metadata!, target] }
      allow(TestSetup).to receive(:install_browser!) { |flags| calls << [:install_browser!, flags] }
    end

    context "接続先がすべてテストDBの場合" do
      let(:configs) do
        %w[annict_test annict_test_probe].map do |database|
          ActiveRecord::DatabaseConfigurations::HashConfig.new("test", database, adapter: "postgresql", database:)
        end
      end

      it "環境名を確認してからアセットのビルドとブラウザ導入を行うこと" do
        TestSetup.run!(["--with-deps"])

        expect(calls).to eq([
          :load_environment!,
          :load_tasks,
          [:validate_existing_database!, "annict_test"],
          [:validate_existing_database!, "annict_test_probe"],
          [:reset_database!, "annict_test"],
          [:reset_database!, "annict_test_probe"],
          [:prepare_metadata!, pool],
          "javascript:build",
          "css:build",
          [:install_browser!, ["--with-deps"]]
        ])
      end
    end

    context "テストDB以外の接続先が含まれる場合" do
      let(:configs) do
        %w[annict_test annict_development].map do |database|
          ActiveRecord::DatabaseConfigurations::HashConfig.new("test", database, adapter: "postgresql", database:)
        end
      end

      it "環境名の記録やアセットのビルドを行わずに中止すること" do
        expect { TestSetup.run!([]) }.to raise_error(RuntimeError, /初期化できるDB/)

        expect(calls).to eq([:load_environment!, :load_tasks])
      end
    end
  end

  describe ".database_url" do
    it "解決済みの接続設定を特殊文字やSSL設定も含めて保持すること" do
      config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary",
        adapter: "postgresql", database: "annict_test_probe", host: "/tmp/postgresql socket",
        port: 5433, username: "user+name", password: "p@ss word&/?", sslmode: "require", pool: 5)

      params = PG::Connection.conninfo_parse(TestSetup.database_url(config)).to_h { |item| [item[:keyword], item[:val]] }

      expect(params.slice("dbname", "host", "port", "user", "password", "sslmode")).to eq(
        "dbname" => "annict_test_probe", "host" => "/tmp/postgresql socket", "port" => "5433",
        "user" => "user+name", "password" => "p@ss word&/?", "sslmode" => "require"
      )
    end
  end

  describe "Go側へのDB構築の委譲" do
    self.use_transactional_tests = false

    it "未作成のDBを検証後に構築し、Go側のenvを再読込しないこと" do
      config = ActiveRecord::Base.connection_db_config.configuration_hash.merge(database: "annict_test_new_#{SecureRandom.hex(6)}")
      db_config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary", config)
      admin = PG.connect(config.slice(:host, :port, :password).merge(user: config[:username], dbname: "postgres"))
      database = PG::Connection.quote_ident(config.fetch(:database))
      begin
        Dir.mktmpdir("annict-test-reset-") do |directory|
          # Go側でop runを再実行すると失敗させ、構築先の設定が切り替わる回帰を検出する。
          op = File.join(directory, "op")
          File.write(op, "#!/bin/sh\nexit 99\n")
          FileUtils.chmod(0o755, op)
          # Rails CIにはdbmateが無いため、作成だけをPostgreSQL付属のcreatedbで代行する。
          dbmate = File.join(directory, "dbmate")
          File.write(dbmate, "#!/bin/sh" + "\n" + 'exec createdb --maintenance-db="$ANNICT_TEST_ADMIN_URL" "$ANNICT_TEST_DATABASE"' + "\n")
          FileUtils.chmod(0o755, dbmate)
          admin_config = ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary", config.merge(database: "postgres"))
          allow(TestSetup).to receive(:system).and_wrap_original do |original, env, *args, **options|
            expect(env).to include("MAKEFLAGS" => nil, "MAKEOVERRIDES" => nil, "MFLAGS" => nil)
            original.call(env.merge("PATH" => "#{directory}:#{ENV.fetch("PATH")}",
              "ANNICT_TEST_ADMIN_URL" => TestSetup.database_url(admin_config), "ANNICT_TEST_DATABASE" => config.fetch(:database)), *args, **options)
          end

          expect { TestSetup.validate_existing_database!(db_config) }.not_to raise_error
          TestSetup.reset_database!(db_config)
        end

        PG.connect(TestSetup.database_url(db_config)) do |connection|
          expect(connection.exec("SELECT current_database()").getvalue(0, 0)).to eq(config.fetch(:database))
          expect(connection.exec("SELECT to_regclass('animes')::text").getvalue(0, 0)).to eq("animes")
        end
      ensure
        admin.exec("DROP DATABASE IF EXISTS #{database}")
        admin.close
      end
    end
  end

  describe "MakefileからのDB保護" do
    self.use_transactional_tests = false

    let(:database_prefix) { "annict_test_entry" }
    let(:stored_environment) { "production" }
    let(:entry_config) do
      ActiveRecord::DatabaseConfigurations::HashConfig.new("test", "primary",
        ActiveRecord::Base.connection_db_config.configuration_hash.merge(database: "#{database_prefix}_#{SecureRandom.hex(6)}"))
    end

    around do |example|
      config = entry_config.configuration_hash
      admin = PG.connect(config.slice(:host, :port, :password).merge(user: config[:username], dbname: "postgres"))
      database = PG::Connection.quote_ident(config.fetch(:database))
      admin.exec("CREATE DATABASE #{database}")
      begin
        PG.connect(TestSetup.database_url(entry_config)) do |connection|
          connection.exec("CREATE TABLE ar_internal_metadata (key text PRIMARY KEY, value text)")
          connection.exec_params("INSERT INTO ar_internal_metadata VALUES ('environment', $1)", [stored_environment])
          connection.exec("CREATE TABLE guard_records (value text)")
          connection.exec("INSERT INTO guard_records VALUES ('保持する値')")
        end
        example.run
      ensure
        admin.exec("DROP DATABASE #{database}")
        admin.close
      end
    end

    def run_setup
      Open3.capture2e({"DATABASE_URL" => TestSetup.database_url(entry_config)},
        "make", "--no-print-directory", "test-setup", "OP_RUN_TEST=APP_ENV=test", chdir: Rails.root)
    end

    def expect_records_preserved
      PG.connect(TestSetup.database_url(entry_config)) do |connection|
        expect(connection.exec("SELECT value FROM guard_records").getvalue(0, 0)).to eq("保持する値")
        expect(connection.exec("SELECT value FROM ar_internal_metadata WHERE key = 'environment'").getvalue(0, 0)).to eq(stored_environment)
      end
    end

    %w[development production].each do |environment|
      context "既存の環境名が#{environment}の場合" do
        let(:stored_environment) { environment }

        it "構築を拒否し、既存データと環境名を保持すること" do
          output, status = run_setup

          expect(status.success?).to be(false), output
          expect(output).to include("別の環境名")
          expect_records_preserved
        end
      end
    end

    context "テスト用以外のDB名の場合" do
      let(:database_prefix) { "annict_review_guard" }
      let(:stored_environment) { "test" }

      it "構築を拒否し、既存データを保持すること" do
        output, status = run_setup

        expect(status.success?).to be(false), output
        expect(output).to include("初期化できるDB")
        expect_records_preserved
      end
    end
  end

  describe "Makefileのテスト実行順" do
    %w[test test-file].each do |target|
      [0, 1].each do |setup_status|
        it "#{target}で準備処理の終了コードが#{setup_status}の場合に実行順を守ること" do
          Dir.mktmpdir("annict-test-entry-") do |directory|
            FileUtils.cp(Rails.root.join("Makefile"), File.join(directory, "Makefile"))
            FileUtils.mkdir_p(File.join(directory, "bin"))
            log = File.join(directory, "calls")
            File.write(File.join(directory, "bin/bundle"), "#!/bin/sh\necho setup >> \"$ANNICT_TEST_ENTRY_LOG\"\nexit #{setup_status}\n")
            File.write(File.join(directory, "bin/rspec"), "#!/bin/sh\necho rspec >> \"$ANNICT_TEST_ENTRY_LOG\"\n")
            FileUtils.chmod(0o755, Dir[File.join(directory, "bin/*")])

            output, status = Open3.capture2e({"PATH" => "#{directory}/bin:#{ENV.fetch("PATH")}", "ANNICT_TEST_ENTRY_LOG" => log},
              "make", "--no-print-directory", target, "OP_RUN_TEST=APP_ENV=test", "FILE=spec/test_setup_spec.rb", chdir: directory)

            expect(status.success?).to eq(setup_status.zero?), output
            expect(File.readlines(log, chomp: true)).to eq(setup_status.zero? ? %w[setup rspec] : %w[setup])
          end
        end
      end
    end
  end

  # TestSetupがスキーマを構築しないのは、Railsがスキーマを所有しないことが前提になっている。
  # config/database.ymlのschema_dump: falseが外れると、Railsが自身のスキーマファイルを正本と
  # して扱い直し、Go所有のテーブルを含まないスキーマでテストDBを作り直してしまう。
  describe "スキーマの所有境界" do
    let(:configs) { ActiveRecord::Base.configurations.configs_for(env_name: "test") }

    it "Railsがスキーマファイルを持たないこと" do
      expect(configs).not_to be_empty
      configs.each do |config|
        expect(ActiveRecord::Tasks::DatabaseTasks.schema_dump_path(config)).to be_nil
      end
    end

    it "テストDBの作り直しを要求しないこと" do
      configs.each do |config|
        expect(ActiveRecord::Tasks::DatabaseTasks.schema_up_to_date?(config)).to be(true)
      end
    end
  end
end
