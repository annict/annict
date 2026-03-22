.PHONY: dev
dev: ## 全サービスの開発サーバーを起動
	hivemind Procfile.dev

.PHONY: fmt
fmt: ## コードをフォーマット（Oxfmt）
	pnpm run fmt

.PHONY: fmt-check
fmt-check: ## フォーマットチェック（Oxfmt）
	pnpm run fmt:check
