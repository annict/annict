.PHONY: dev
dev: ## 全サービスの開発サーバーを起動
	hivemind Procfile.dev

.PHONY: help
help: ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## コードをフォーマット（Oxfmt）
	pnpm fmt

.PHONY: fmt-check
fmt-check: ## フォーマットチェック（Oxfmt）
	pnpm fmt:check
