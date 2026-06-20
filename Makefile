.PHONY: dev
dev: ## 全サービスの開発サーバーを起動
	hivemind Procfile.dev

.PHONY: help
help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## Format code (Oxfmt). [Ja] コードをフォーマット (Oxfmt)
	pnpm fmt

.PHONY: fmt-check
fmt-check: ## Check code formatting (Oxfmt). [Ja] フォーマットチェック (Oxfmt)
	pnpm fmt:check

# koryluslint is the Korylus shared linter from korylus-tools. Its `md`
# subcommand checks semantic line breaks in Markdown. The version is pinned via
# the `tool` directive in go/go.mod and fetched from the public module
# github.com/korylus/tools; we build that pinned binary and run it from the
# repository root so it scans the root-level Markdown (README, docs, ...).
#
# md cannot run through `go tool koryluslint` directly: the Go module is nested
# in go/, but the Markdown to scan lives at the repository root, so `go tool`
# (which walks from the module directory) would scan go/ instead. Building the
# pinned binary lets us run the check from any working directory.
#
# [Ja] koryluslint は korylus-tools が提供する Korylus 共通リンタ。その `md`
# サブコマンドは Markdown の句点改行 (semantic line break) を検査する。バージョンは
# go/go.mod の tool ディレクティブで固定し、公開モジュール github.com/korylus/tools
# から取得する。
# その固定版バイナリをビルドしてリポジトリルートから実行することで、ルート直下の
# Markdown (README・docs など) を走査する。
#
# md を `go tool koryluslint` で直接実行できないのは、Go モジュールが go/ にネスト
# している一方、走査対象の Markdown はリポジトリルートにあるため。`go tool` は
# モジュールのディレクトリを起点に走査するため go/ を見てしまう。固定版バイナリを
# ビルドすれば任意の作業ディレクトリから検査を実行できる。
KORYLUSLINT ?= /tmp/koryluslint

.PHONY: koryluslint-build
koryluslint-build:
	@go -C go build -o $(KORYLUSLINT) github.com/korylus/tools/cmd/koryluslint

.PHONY: lint-md
lint-md: koryluslint-build ## Check semantic line breaks in Markdown (changed lines). [Ja] Markdown の句点改行をチェック (変更行のみ)
	@$(KORYLUSLINT) md

.PHONY: lint-md-base
lint-md-base: koryluslint-build ## Check Markdown semantic line breaks against BASE ref (e.g. BASE=origin/main). [Ja] BASE ref との差分行で Markdown の句点改行をチェック (例: BASE=origin/main)
	@$(KORYLUSLINT) md -base=$(BASE)

.PHONY: lint-md-fix
lint-md-fix: koryluslint-build ## Auto-fix semantic line breaks in Markdown (changed files). [Ja] Markdown の句点改行を自動修正 (変更ファイル)
	@$(KORYLUSLINT) md --write
