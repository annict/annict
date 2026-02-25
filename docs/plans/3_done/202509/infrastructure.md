# 基盤構築

## 概要

Go プロジェクトの基本的なインフラストラクチャとツールチェーンのセットアップ。

## 実装内容

### フェーズ 1: 最小限の基盤構築

- [x] **プロジェクト初期化**
  - [x] go.mod 作成（`go mod init github.com/annict/annict`）
  - [x] 最小限のディレクトリ構造（cmd/server、internal/handler）

- [x] **環境変数設定**
  - [x] .env.example ファイル作成（DB 接続情報、ドメイン設定）
  - [x] 環境変数読み込み（os.Getenv 使用）

- [x] **Cloudflare Tunnel 設定**
  - [x] go.example.dev のトンネル設定
  - [x] ローカルポート 18000 へのルーティング

- [x] **最小限の Web サーバー起動**
  - [x] main.go 作成（Hello World が表示できる状態）
  - [x] Chi ルーターの導入と基本設定
  - [x] go.example.dev でアクセス確認

- [x] **データベース接続確認**
  - [x] PostgreSQL 接続（database/sql 使用）
  - [x] 簡単な SELECT クエリ実行確認

## 技術スタック

- **Go**: 1.25.1
- **Web フレームワーク**: chi/v5
- **データベース**: PostgreSQL 17.3（Rails と共有）
- **環境変数管理**: godotenv
- **ホスティング**: Cloudflare Tunnel（開発環境）

## ディレクトリ構造

```
annict-go/
├── cmd/
│   └── server/
│       └── main.go          # エントリーポイント
├── internal/
│   ├── config/              # 設定管理
│   └── handler/             # HTTPハンドラー
├── go.mod
├── go.sum
└── .env.example
```

## 環境変数

開発環境で必要な環境変数：

```bash
# データベース接続
ANNICT_DB_HOST=host.docker.internal
ANNICT_DB_PORT=15432
ANNICT_DB_USER=postgres
ANNICT_DB_NAME=annict_development
ANNICT_DB_SSLMODE=disable

# サーバー設定
ANNICT_PORT=8080
ANNICT_ENV=development

# ドメイン設定（セッション共有用）
ANNICT_DOMAIN=.example.dev
```

## 成果

- Go の Web サーバーが起動し、`go.example.dev` でアクセス可能
- PostgreSQL データベースに接続できる
- 環境変数による設定管理が動作する
- 開発環境の基盤が整った

## 関連ドキュメント

- [プロジェクト全体の設計書](./go.md)
