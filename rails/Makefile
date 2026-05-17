.PHONY: test test-file test-setup console server lint lint-fix fmt zeitwerk sorbet sorbet-update db-migrate db-rollback db-setup graphql-dump help

# APP_ENVのデフォルト値を設定（外部から指定された場合はその値を使用）
# 例: APP_ENV=test make db-migrate
APP_ENV ?= dev

# 1Password CLI経由で環境変数を読み込んでコマンドを実行
OP_RUN := APP_ENV=$(APP_ENV) op run --env-file=".env.local" --
# テスト専用（常にtest環境）
OP_RUN_TEST := APP_ENV=test op run --env-file=".env.local" --

## テスト
test: ## 全テストを実行
	$(OP_RUN_TEST) bin/rspec

test-file: ## 特定のテストファイルを実行 (例: make test-file FILE=spec/models/user_spec.rb)
	$(OP_RUN_TEST) bin/rspec $(FILE)

test-setup: ## テスト環境をセットアップ（シード投入、Playwrightインストール）
	@echo "==> テストDBにシードデータを投入..."
	RAILS_ENV=test bin/rails db:seed
	@echo "==> Playwrightブラウザをインストール..."
	yarn playwright install
	@echo "==> テスト環境のセットアップが完了しました"

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

## 型チェック
zeitwerk: ## Zeitwerk（オートロード）チェック
	$(OP_RUN) bin/rails zeitwerk:check

sorbet: ## Sorbet型チェック
	bin/srb tc

sorbet-update: ## Sorbet型定義を更新
	$(OP_RUN) bin/rails sorbet:update

## データベース
db-migrate: ## マイグレーションを実行 (例: APP_ENV=test make db-migrate)
	$(OP_RUN) bin/rails db:migrate

db-rollback: ## 最後のマイグレーションをロールバック (例: APP_ENV=test make db-rollback)
	$(OP_RUN) bin/rails db:rollback

db-setup: ## データベースのセットアップ (例: APP_ENV=test make db-setup)
	$(OP_RUN) bin/rails db:setup

## GraphQL
graphql-dump: ## GraphQL APIスキーマをダンプ
	$(OP_RUN) bin/rails graphql:dump_schema

## ヘルプ
help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
