# op run ラッパー: 1Password CLI 経由で .env の環境変数を解決してコマンドを実行する
# CI環境ではopコマンドが利用できないため、環境変数はGitHub Actionsで直接設定する
OP_RUN_DEV = APP_ENV=dev op run --env-file=".env" --
ifeq ($(CI),true)
OP_RUN_TEST = APP_ENV=test
else
OP_RUN_TEST = APP_ENV=test op run --env-file=".env" --
endif

.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## サーバーを起動
	$(OP_RUN_DEV) go run ./cmd/annict serve

.PHONY: browse-login
browse-login: ## dev サイトへブラウザでログイン (別ユーザーは make browse-login BROWSE_USER=2)
	$(OP_RUN_DEV) bash scripts/browse.sh login $(BROWSE_USER)

.PHONY: browse-shot
browse-shot: ## ログイン済みセッションでページを開きスクショ (例: make browse-shot URL=/db)
	bash scripts/browse.sh shot "$(URL)"

.PHONY: browse-close
browse-close: ## ブラウザセッションを閉じ一時ファイルを削除
	bash scripts/browse.sh close

.PHONY: sync-animes
sync-animes: ## animes 同期バッチを 1 回実行 (works/episodes → animes)
	$(OP_RUN_DEV) go run ./cmd/annict sync-animes

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
	go build -o bin/annict ./cmd/annict

.PHONY: test
test: db-setup-test ## テストを実行
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse)
	@bash -c 'set -o pipefail && $(OP_RUN_TEST) go test -json -race ./... | tparse -all'

.PHONY: test-pkg
test-pkg: db-setup-test ## 特定のパッケージのテストを実行（使用例: make test-pkg PKG=internal/handler/password_reset）
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse)
	@if [ -z "$(PKG)" ]; then \
		echo "エラー: PKG変数を指定してください"; \
		echo "使用例: make test-pkg PKG=internal/handler/password_reset"; \
		exit 1; \
	fi
	@echo "パッケージ $(PKG) のテストを実行"
	@bash -c 'set -o pipefail && $(OP_RUN_TEST) go test -json -race ./$(PKG) | tparse -all'

.PHONY: test-run
test-run: db-setup-test ## 特定のテストを実行（使用例: make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification）
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse)
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
	@bash -c 'set -o pipefail && $(OP_RUN_TEST) go test -json -race ./$(PKG) -run $(RUN) | tparse -all'

.PHONY: test-verbose
test-verbose: db-setup-test ## 詳細ログ付きで全テストを実行（デバッグ用）
	@which tparse > /dev/null || (echo "Installing tparse from go.mod..." && go install github.com/mfridman/tparse)
	@echo "ローカル環境: 詳細ログ付きで全テストを実行"
	@bash -c 'set -o pipefail && $(OP_RUN_TEST) go test -json -race -count=1 ./... | tparse -all'

.PHONY: test-js
test-js: ## Run frontend unit tests (Vitest). [Ja] フロントエンドのユニットテストを実行 (Vitest)
	pnpm test

.PHONY: typecheck-js
typecheck-js: ## Type-check frontend TypeScript (tsc --noEmit). [Ja] フロントエンドの TypeScript を型検査 (tsc --noEmit)
	pnpm typecheck

.PHONY: fmt
fmt: ## コードをフォーマット
	go fmt ./...
	@$(MAKE) goimports

.PHONY: lint-golangci
lint-golangci: ## golangci-lintを実行
	@echo "🔍 golangci-lintを実行中..."
	@golangci-lint run --config=.golangci.yml ./...

.PHONY: lint-comments
lint-comments: ## Check [Ja]-marker misuse in code comments (koryluslint). [Ja] コードコメントの [Ja] マーカー誤用をチェック (koryluslint)
	@echo "🔍 コメントの英日併記をチェック中..."
	@go tool koryluslint comment

# The tree-wide run above only checks the rules with near-zero false positives.
# The rules that would flag untouched legacy comments (a missing English block,
# the inline form, duplicate markers, a missing blank separator) are checked
# against a base ref instead, limited to the lines added since that ref. CI runs
# this on pull requests; run it locally with the same base to see what CI sees.
#
# [Ja] 上のツリー全体の実行は誤検出がほぼ無い規則だけを検査する。
# 触っていない既存コメントを指摘してしまう規則 (英語ブロックの欠落・インライン形式・
# マーカー重複・区切り空行の欠落) は、基準 ref との差分で追加された行に限定して検査する。
# CI は pull request でこれを実行するため、ローカルでも同じ基準を指定すれば CI と
# 同じ結果を確認できる。
.PHONY: lint-comments-base
lint-comments-base: ## Check comment markers against BASE ref, added lines only (e.g. BASE=origin/develop). [Ja] BASE ref との差分の追加行でコメントのマーカーをチェック (例: BASE=origin/develop)
	@test -n "$(BASE)" || { echo "BASE is required. Example: make lint-comments-base BASE=origin/develop"; exit 1; }
	@echo "🔍 コメントの英日併記をチェック中 (追加行のみ)..."
	@go tool koryluslint comment -base=$(BASE)

.PHONY: lint
lint: lint-golangci lint-comments ## リントを実行（golangci-lint + コメント併記チェック）

.PHONY: clean
clean: ## ビルド成果物をクリーン
	rm -rf bin/
	find . -name "*_templ.go" -type f -delete

# データベースマイグレーションコマンド

# Strip the random \restrict / \unrestrict lines that pg_dump >= 17.6 emits into
# db/schema.sql, which dbmate writes on both up and dump. Shared by db-migrate
# and db-dump so the workaround lives in one place.
# https://github.com/amacneil/dbmate/issues/678
#
# [Ja] dbmate が up / dump の双方で db/schema.sql へ書き込む、pg_dump 17.6 以降が
# 出力するランダムな \restrict / \unrestrict 行を削除する。db-migrate と db-dump で
# 共有し、回避策を 1 箇所にまとめる。
STRIP_SCHEMA_RESTRICT = sed -i '/^\\restrict /d;/^\\unrestrict /d' db/schema.sql

.PHONY: db-migrate
db-migrate: ## データベースマイグレーションを実行
	$(OP_RUN_DEV) dbmate up
	@$(STRIP_SCHEMA_RESTRICT)

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
	$(OP_RUN_DEV) dbmate down

.PHONY: db-new
db-new: ## 新しいマイグレーションファイルを作成（使用法: make db-new name=create_users）
	dbmate new $(name)

.PHONY: db-dump
db-dump: ## データベーススキーマをdb/schema.sqlにダンプ
	$(OP_RUN_DEV) dbmate dump
	@$(STRIP_SCHEMA_RESTRICT)

.PHONY: sqlc-generate
sqlc-generate: ## SQLクエリからGoコードを生成
	sqlc generate

.PHONY: seed
seed: ## 開発環境用のシードデータを生成
	@echo "シードデータを生成しています..."
	$(OP_RUN_DEV) go run cmd/seed/main.go

# Production snapshot for local verification: download the production backup,
# mask PII in an isolated database, and load the masked result locally; and copy
# the production images into the development R2 bucket. load-snapshot runs both.
#
# [Ja] ローカル検証用の本番スナップショット: 本番バックアップを取得し、隔離 DB で
# PII をマスクしてマスク結果をローカルに取り込み、本番画像を開発用 R2 バケットへ
# コピーする。load-snapshot は両方を実行する。

.PHONY: load-snapshot
load-snapshot: load-snapshot-db load-snapshot-images ## Load the masked DB snapshot and copy production images in one shot. [Ja] マスク済み DB スナップショットの取り込みと本番画像のコピーを一括実行する

.PHONY: load-snapshot-db
load-snapshot-db: ## Load a masked production snapshot into the local DB. [Ja] マスク済み本番スナップショットをローカル DB に取り込む
	$(OP_RUN_DEV) bash scripts/load_snapshot_db.sh

.PHONY: restore-snapshot-db
restore-snapshot-db: ## Re-restore the local DB from the masked dump in tmp/. [Ja] tmp/ のマスク済みダンプからローカル DB を再復元する
	$(OP_RUN_DEV) bash scripts/load_snapshot_db.sh restore

.PHONY: load-snapshot-images
load-snapshot-images: ## Sync production work/profile images into the development R2 bucket. [Ja] 本番の作品・プロフィール画像を開発用 R2 バケットへ同期する
	$(OP_RUN_DEV) bash scripts/load_snapshot_images.sh
