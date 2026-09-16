.PHONY: test test-file test-setup setup console server lint lint-fix fmt zeitwerk graphql-dump help

# APP_ENVのデフォルト値を設定（外部から指定された場合はその値を使用）
# 例: APP_ENV=test make console
APP_ENV ?= dev

# 1Password CLI経由で環境変数を読み込んでコマンドを実行
OP_RUN := APP_ENV=$(APP_ENV) op run --env-file=".env.local" --
# テスト専用（常にtest環境）
# CI環境ではopコマンドが利用できないため、環境変数はGitHub Actionsで直接設定する
ifeq ($(CI),true)
OP_RUN_TEST := APP_ENV=test
# CIのランナーにはブラウザの実行に必要なOSのライブラリが無いため、ブラウザと一緒に導入する
PLAYWRIGHT_INSTALL_FLAGS := --with-deps
else
OP_RUN_TEST := APP_ENV=test op run --env-file=".env.local" --
PLAYWRIGHT_INSTALL_FLAGS :=
endif

## テスト
test: test-setup ## 全テストを実行
	$(OP_RUN_TEST) bin/rspec

test-file: test-setup ## 特定のテストファイルを実行 (例: make test-file FILE=spec/models/user_spec.rb)
	$(OP_RUN_TEST) bin/rspec $(FILE)

# テストDBはGo版のテストと共有しており、Go版のテスト後は作成データが残るため、
# 再実行のたびに作り直す。
test-setup: ## テスト環境をセットアップ (テストDBの作り直し、アセットビルド、Playwrightブラウザの導入)
	$(OP_RUN_TEST) bundle exec ruby spec/test_setup.rb $(PLAYWRIGHT_INSTALL_FLAGS)

## セットアップ
setup: ## 依存関係をインストールし、log・tmpを掃除
	yarn install --frozen-lockfile
	bin/setup

## 開発サーバー
console: ## Railsコンソールを起動
	$(OP_RUN) bin/rails console

server: ## Railsサーバーを起動
	$(OP_RUN) bin/rails server

worker: ## Railsワーカーを起動
	$(OP_RUN) bin/rails jobs:work

## リント・フォーマット
lint: ## Rubyコードのリントを実行
	$(OP_RUN) bin/standardrb

lint-fix: ## Rubyコードのリントを実行して自動修正
	$(OP_RUN) bin/standardrb --fix

fmt: lint-fix ## lint-fixのエイリアス

## オートロードチェック
zeitwerk: ## Zeitwerk（オートロード）チェック
	$(OP_RUN) bin/rails zeitwerk:check

## GraphQL
graphql-dump: ## GraphQL APIスキーマをダンプ
	$(OP_RUN) bin/rails graphql:dump_schema

## ヘルプ
help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
