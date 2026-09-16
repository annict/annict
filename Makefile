.PHONY: dev
dev: ## 全サービスの開発サーバーを起動
	hivemind Procfile.dev

.PHONY: help
help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# 依存関係のインストールを先に済ませてからデータベースを初期化する。
# db-prepare-devは1Password CLI経由でDATABASE_URLを解決するため、解決に失敗しても
# 「依存関係は入ったがDBだけ未初期化」という切り分けやすい状態で止まる。
.PHONY: setup
setup: ## 開発環境をセットアップ (依存関係のインストールとデータベースの初期化)
	pnpm install --frozen-lockfile
	$(MAKE) -C go setup
	$(MAKE) -C rails setup
	$(MAKE) -C go db-prepare-dev
	$(MAKE) -C go seed-master

.PHONY: fmt
fmt: ## コードをフォーマット (Oxfmt)
	pnpm fmt

.PHONY: fmt-check
fmt-check: ## フォーマットチェック (Oxfmt)
	pnpm fmt:check

# koryluslintはkorylus-toolsが提供するKorylus共通リンタ。その `md`
# サブコマンドはMarkdownの句点改行 (semantic line break) を検査する。バージョンは
# go/go.modのtoolディレクティブで固定し、公開モジュールgithub.com/korylus/tools
# から取得する。
# その固定版バイナリをビルドしてリポジトリルートから実行することで、ルート直下の
# Markdown (README・docsなど) を走査する。
#
# mdを `go tool koryluslint` で直接実行できないのは、Goモジュールがgo/ にネスト
# している一方、走査対象のMarkdownはリポジトリルートにあるため。`go tool` は
# モジュールのディレクトリを起点に走査するためgo/ を見てしまう。固定版バイナリを
# ビルドすれば任意の作業ディレクトリから検査を実行できる。
KORYLUSLINT ?= /tmp/koryluslint

.PHONY: koryluslint-build
koryluslint-build:
	@go -C go build -o $(KORYLUSLINT) github.com/korylus/tools/cmd/koryluslint

.PHONY: lint-md
lint-md: koryluslint-build ## Markdownの句点改行をチェック (変更行のみ)
	@$(KORYLUSLINT) md

.PHONY: lint-md-base
lint-md-base: koryluslint-build ## BASE refとの差分行でMarkdownの句点改行をチェック (例: BASE=origin/main)
	@test -n "$(BASE)" || { echo "BASE is required. Example: make lint-md-base BASE=origin/develop"; exit 1; }
	@$(KORYLUSLINT) md -base=$(BASE)

.PHONY: lint-md-fix
lint-md-fix: koryluslint-build ## Markdownの句点改行を自動修正 (変更ファイル)
	@$(KORYLUSLINT) md --write
