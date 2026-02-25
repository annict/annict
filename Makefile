.PHONY: fmt
fmt: ## コードをフォーマット（Oxfmt）
	pnpm run fmt

.PHONY: fmt-check
fmt-check: ## フォーマットチェック（Oxfmt）
	pnpm run fmt:check
