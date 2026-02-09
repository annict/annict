.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## サーバーを起動
	APP_ENV=dev op run --env-file=".env" -- go run cmd/server/main.go

.PHONY: templ-generate
templ-generate: ## templファイルからGoコードを生成し、import文を整理
	@which templ > /dev/null || (echo "Installing templ from go.mod..." && go install github.com/a-h/templ/cmd/templ)
	templ generate
	@$(MAKE) goimports

.PHONY: goimports
goimports: ## goimportsでimport文を整理
	@which goimports > /dev/null || (echo "Installing goimports from go.mod..." && go install golang.org/x/tools/cmd/goimports)
	goimports -w .

.PHONY: goimports-check
goimports-check: ## goimportsでimport文をチェック（差分があればエラー）
	@which goimports > /dev/null || (echo "Installing goimports from go.mod..." && go install golang.org/x/tools/cmd/goimports)
	@unformatted=$$(goimports -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files have incorrect imports:"; \
		echo "$$unformatted"; \
		echo ""; \
		echo "Please run 'make goimports' to fix imports"; \
		exit 1; \
	fi

.PHONY: build
build: templ-generate ## バイナリをビルド（templ-generateがgoimportsも実行）
	go build -o bin/server cmd/server/main.go

.PHONY: test
test: db-setup-test ## テストを実行
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse@latest)
	@bash -c 'set -o pipefail && go test -json -race ./... | tparse -all'

.PHONY: test-pkg
test-pkg: db-setup-test ## 特定のパッケージのテストを実行（使用例: make test-pkg PKG=internal/handler/password_reset）
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse@latest)
	@if [ -z "$(PKG)" ]; then \
		echo "エラー: PKG変数を指定してください"; \
		echo "使用例: make test-pkg PKG=internal/handler/password_reset"; \
		exit 1; \
	fi
	@echo "パッケージ $(PKG) のテストを実行"
	@bash -c 'set -o pipefail && APP_ENV=test op run --env-file=".env" -- go test -json -race ./$(PKG) | tparse -all'

.PHONY: test-run
test-run: db-setup-test ## 特定のテストを実行（使用例: make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification）
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse@latest)
	@if [ -z "$(PKG)" ]; then \
		echo "エラー: PKG変数を指定してください"; \
		echo "使用例: make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification"; \
		exit 1; \
	fi
	@if [ -z "$(RUN)" ]; then \
		echo "エラー: RUN変数を指定してください"; \
		echo "使用例: make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification"; \
		exit 1; \
	fi
	@echo "パッケージ $(PKG) のテスト $(RUN) を実行"
	@bash -c 'set -o pipefail && APP_ENV=test op run --env-file=".env" -- go test -json -race ./$(PKG) -run $(RUN) | tparse -all'

.PHONY: test-verbose
test-verbose: db-setup-test ## 詳細ログ付きで全テストを実行（デバッグ用）
	@echo "ローカル環境: 詳細ログ付きで全テストを実行"
	@APP_ENV=test op run --env-file=".env" -- go test -v -race -count=1 ./...

.PHONY: fmt
fmt: ## コードをフォーマット
	go fmt ./...
	@$(MAKE) goimports

.PHONY: lint-golangci
lint-golangci: ## golangci-lintを実行
	@echo "🔍 golangci-lintを実行中..."
	@golangci-lint run --config=.golangci.yml ./...

.PHONY: lint
lint: lint-golangci ## リントを実行（golangci-lintを使用）

.PHONY: clean
clean: ## ビルド成果物をクリーン
	rm -rf bin/
	find . -name "*_templ.go" -type f -delete

# データベースマイグレーションコマンド
.PHONY: db-migrate
db-migrate: ## データベースマイグレーションを実行
	APP_ENV=dev op run --env-file=".env" -- dbmate up
	# PostgreSQL 17.6のpg_dumpが生成するランダムな\restrict行を削除
	# https://github.com/amacneil/dbmate/issues/678
	@sed -i '/^\\restrict /d;/^\\unrestrict /d' db/schema.sql

.PHONY: db-setup-test
db-setup-test: ## テスト用データベースのスキーマをセットアップ
	@if [ "$$CI" = "true" ]; then \
		echo "CI環境: テストDBをリセットしてスキーマを適用"; \
		psql -h localhost -U postgres -d annict_test -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;' > /dev/null 2>&1; \
		psql -h localhost -U postgres -d annict_test -f db/schema.sql > /dev/null; \
	else \
		echo "ローカル環境: テストDBをリセットしてスキーマを適用"; \
		psql -h postgresql -U postgres -d annict_test -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;' > /dev/null 2>&1; \
		psql -h postgresql -U postgres -d annict_test -f db/schema.sql > /dev/null; \
	fi

.PHONY: db-migrate-test
db-migrate-test: ## テスト環境のマイグレーションを実行（非推奨、db-setup-testを使用）
	@echo "警告: db-migrate-testは非推奨です。代わりにdb-setup-testを使用してください。"
	@make db-setup-test

.PHONY: db-rollback
db-rollback: ## 最後のマイグレーションをロールバック
	APP_ENV=dev op run --env-file=".env" -- dbmate down

.PHONY: db-new
db-new: ## 新しいマイグレーションファイルを作成（使用法: make db-new name=create_users）
	dbmate new $(name)

.PHONY: db-dump
db-dump: ## データベーススキーマをdb/schema.sqlにダンプ
	APP_ENV=dev op run --env-file=".env" -- dbmate dump
	# PostgreSQL 17.6のpg_dumpが生成するランダムな\restrict行を削除
	# https://github.com/amacneil/dbmate/issues/678
	@sed -i '/^\\restrict /d;/^\\unrestrict /d' db/schema.sql

.PHONY: sqlc-generate
sqlc-generate: ## SQLクエリからGoコードを生成
	sqlc generate

.PHONY: seed
seed: ## 開発環境用のシードデータを生成
	@echo "シードデータを生成しています..."
	APP_ENV=dev op run --env-file=".env" -- go run cmd/seed/main.go
