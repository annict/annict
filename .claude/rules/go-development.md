---
paths:
  - "go/**/*.{go,templ,sql}"
---

# 開発環境ガイド

このドキュメントは、Go 版 Wikino の開発環境のセットアップと運用に関するガイドを提供します。

## 環境変数の設定

### 命名規則

- Wikino で定義する環境変数は、外部ライブラリなどが指定してくるものを除き、**必ずプレフィックス `WIKINO_` を付ける**
- 例:
  - `WIKINO_PORT`, `WIKINO_DOMAIN`, `WIKINO_RAILS_APP_URL`
  - 外部ライブラリが要求する環境変数はそのまま使用（例: `DATABASE_URL`, `REDIS_URL`）
- **例外**: `APP_ENV` は `WIKINO_` プレフィックスなしで使用する

### .env ファイルを使用する

`.env.example` をコピーして `.env` ファイルを作成し、実際の値を設定します：

```sh
cp .env.example .env
```

`.env` ファイルは `.gitignore` に含まれているため、秘密情報を安全に管理できます。

### 環境変数の優先順位（上から順に優先）

1. **事前に設定された環境変数** - シークレット管理ツールやホスティング環境が提供（メモリ上に存在）
2. **`.env`** - 環境変数設定ファイル（gitignore 対象）

## 開発ツールの依存関係管理

Go プロジェクトで使用する CLI ツール（`golangci-lint`, `templ`, `goimports` など）は、`tools.go` ファイルで依存関係を管理します。

### tools.go ファイルの目的

- **バージョン管理**: `go.mod` でツールのバージョンを固定
- **チーム全体で統一**: 全開発者が同じバージョンのツールを使用
- **CI との整合性**: ローカル環境と CI 環境で同じバージョンを使用
- **自動インストール**: `go mod download` でツールの依存関係も自動ダウンロード

### 新しい CLI ツールを追加する手順

**1. tools.go にツールを追加**

```go
//go:build tools

package tools

import (
	_ "github.com/a-h/templ/cmd/templ"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "github.com/新しいツール/cmd/ツール名"  // 新しいツールを追加
)
```

**2. 依存関係を go.mod に追加**

```sh
go get github.com/新しいツール/cmd/ツール名@vX.Y.Z
go mod tidy
```

**3. ツールをインストール**

```sh
go install github.com/新しいツール/cmd/ツール名
```

**4. Makefile にターゲットを追加（オプション）**

```makefile
.PHONY: 新しいコマンド
新しいコマンド:
	@which ツール名 > /dev/null || (echo "Installing ツール名 from go.mod..." && go install github.com/新しいツール/cmd/ツール名)
	ツール名 [オプション]
```

## データベースマイグレーション

[dbmate](https://github.com/amacneil/dbmate) を使用してデータベースマイグレーションを管理しています。

### 基本コマンド

```sh
make db-new name=create_users   # 新しいマイグレーションファイルを作成
make db-migrate                 # マイグレーションを実行（開発環境）
make db-setup-test              # テスト用データベースのセットアップ
make db-rollback                # 最後のマイグレーションをロールバック
make db-dump                    # DBスキーマをダンプ
make sqlc-generate              # sqlcでGoコードを生成
```

### 注意事項

- マイグレーションファイルは `db/migrations/` ディレクトリに作成される
- スキーマは `db/schema.sql` に出力される
- `make test` 実行時は自動的に `make db-setup-test` が実行される
- テスト用 DB (`wikino_test`) は完全にリセット（`DROP SCHEMA public CASCADE`）してから `db/schema.sql` を適用
- **マイグレーション単体のテストは不要**: リポジトリやユースケースのテストで間接的に検証される

### カラム定義のガイドライン

**文字列型**:

- `VARCHAR`（長さ指定なし）を使用し、長さ制限はアプリケーションコードでバリデーション
- `VARCHAR(n)` は既存テーブルとの互換性が必要な場合を除き使用しない

```sql
-- 良い例
title VARCHAR NOT NULL,
-- 悪い例
title VARCHAR(510) NOT NULL,
```

**タイムスタンプ型**:

- `TIMESTAMP WITH TIME ZONE` を使用（すべての時刻は UTC で保存）
- `TIMESTAMP WITHOUT TIME ZONE` は使用しない

```sql
-- 良い例
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
-- 悪い例
deleted_at TIMESTAMP WITHOUT TIME ZONE,
```

## golangci-lint

golangci-lint は複数の静的解析ツールを統合して実行するリンターです。

### 基本コマンド

```sh
make lint                                                            # 全パッケージ
golangci-lint run --config=.golangci.yml ./internal/handler/...      # 特定パッケージ
golangci-lint run --config=.golangci.yml --disable-all --enable=gosec ./...  # 特定リンター
```

### 設定ファイル

`.golangci.yml` で管理。主な設定内容：

- **有効化されているリンター**: gofmt, govet, staticcheck, gosec, depguard, errcheck, ineffassign, unused
- **アーキテクチャルール（depguard）**: 3 層アーキテクチャの依存関係を強制
  - Presentation 層（Handler, Middleware, ViewModel, Templates）は Query に直接依存できない
  - Application 層（UseCase）は Query に直接依存できない
  - Domain/Infrastructure 層（Query, Repository, Model）は上位層に依存できない
