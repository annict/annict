# Go版 本番運用設定 設計書

## 概要

Go版Annictアプリケーションを本番環境で安定稼働させるために必要な設定を追加します。

**目的**:

- 高負荷時でもアプリケーションが安定して動作するようにする
- リソース枯渇（コネクション、メモリ）を防止する
- セキュリティリスク（Slow Loris攻撃など）を軽減する

**背景**:

- 現在、PostgreSQLのコネクションプール設定が本番コードに設定されていない（デフォルト値＝無制限）
- HTTPサーバーのタイムアウトが設定されていない
- 外部API呼び出しのタイムアウトが設定されていない箇所がある

## 要件

### 機能要件

- システムはPostgreSQLへの接続数を適切に制限する
- システムはRedisへの接続数を適切に制限する
- システムはHTTPリクエスト/レスポンスに適切なタイムアウトを設定する
- システムは外部API呼び出しに適切なタイムアウトを設定する
- システムはリクエストボディサイズを制限する

### 非機能要件

#### パフォーマンス

- コネクションプールの設定により、接続の再利用を最適化する
- アイドル接続の適切な管理により、リソース使用量を最小化する

#### 可用性・信頼性

- タイムアウト設定により、ハング状態の接続を防止する
- Graceful shutdownにより、処理中のリクエストを安全に完了させる（実装済み）

#### セキュリティ

- リクエストボディサイズ制限により、大容量データによる攻撃を防止する
- タイムアウト設定により、Slow Loris攻撃を軽減する

## 設計

### コネクションプール設定

#### PostgreSQL（メインDB接続）

`cmd/server/main.go` で `sql.Open()` 直後に設定を追加：

```go
db.SetMaxOpenConns(25)              // 同時接続の最大数
db.SetMaxIdleConns(5)               // アイドル状態の接続プール
db.SetConnMaxLifetime(5 * time.Minute)  // 接続の最大有効期限
db.SetConnMaxIdleTime(2 * time.Minute)  // アイドル接続のタイムアウト
```

**設定値の根拠**:

- `MaxOpenConns: 25`: PostgreSQLのデフォルト`max_connections`は100。Railsアプリとの共有を考慮し、余裕を持たせた値
- `MaxIdleConns: 5`: 通常時の負荷に対応できる最小限のアイドル接続
- `ConnMaxLifetime: 5分`: DBの再起動やフェイルオーバーに備え、古い接続を定期的にリフレッシュ
- `ConnMaxIdleTime: 2分`: 使われていない接続を早めに解放

#### River（ジョブキュー用pgxpool）

`internal/worker/client.go` で `pgxpool.ParseConfig()` 後に設定を追加：

```go
poolConfig.MaxConns = 10             // River専用プールの最大接続数
poolConfig.MinConns = 2              // 最小接続数（ウォームスタート）
poolConfig.MaxConnLifetime = 5 * time.Minute
poolConfig.MaxConnIdleTime = 2 * time.Minute
```

**設定値の根拠**:

- `MaxConns: 10`: 現在のMaxWorkers(10)に対応。1ワーカー1接続を想定
- `MinConns: 2`: ジョブ実行開始時の接続遅延を軽減

#### Redis

`cmd/server/main.go` で `redis.ParseURL()` 後に設定を追加：

```go
opt.PoolSize = 10                    // 最大接続数
opt.MinIdleConns = 2                 // 最小アイドル接続数
opt.ConnMaxIdleTime = 5 * time.Minute // アイドル接続のタイムアウト
```

**設定値の根拠**:

- `PoolSize: 10`: Rate Limiting + セッションストアの用途に十分な数
- `MinIdleConns: 2`: 通常時の負荷に対応

### HTTPサーバータイムアウト設定

`cmd/server/main.go` の `http.Server` 定義を更新：

```go
srv := &http.Server{
    Addr:           addr,
    Handler:        r,
    ReadTimeout:    15 * time.Second,  // リクエスト読み込み完了までの時間
    WriteTimeout:   15 * time.Second,  // レスポンス書き込み完了までの時間
    IdleTimeout:    60 * time.Second,  // Keep-Alive接続のタイムアウト
    MaxHeaderBytes: 1 << 20,           // 1MBのヘッダーサイズ制限
}
```

**設定値の根拠**:

- `ReadTimeout: 15秒`: 通常のHTMLフォーム送信に十分な時間
- `WriteTimeout: 15秒`: テンプレートレンダリングを含む応答時間
- `IdleTimeout: 60秒`: Keep-Alive接続の標準的な値
- `MaxHeaderBytes: 1MB`: 大きすぎるヘッダーによる攻撃を防止

### 外部API呼び出しのタイムアウト設定

#### Turnstile（Bot対策）

`internal/turnstile/client.go` でHTTPクライアントにタイムアウトを設定：

```go
httpClient := &http.Client{
    Timeout: 10 * time.Second,
}
```

#### Resend（メール送信）

`internal/worker/send_email.go` でHTTPクライアントにタイムアウトを設定：

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,  // メール送信は時間がかかる場合がある
}
client := resend.NewClient(cfg.ResendAPIKey, resend.WithHTTPClient(httpClient))
```

### リクエストボディサイズ制限

ミドルウェアでリクエストボディサイズを制限：

```go
// 10MBの制限（画像アップロードなどを考慮）
r.Use(middleware.RequestBodyLimit(10 * 1024 * 1024))
```

**注意**: chi/v5には`RequestBodyLimit`ミドルウェアがないため、カスタム実装が必要。

### 環境変数による設定のカスタマイズ（将来対応）

今回の実装では固定値を使用しますが、将来的に環境変数で上書き可能にすることを検討：

- `ANNICT_DB_MAX_OPEN_CONNS`
- `ANNICT_DB_MAX_IDLE_CONNS`
- `ANNICT_HTTP_READ_TIMEOUT`
- `ANNICT_HTTP_WRITE_TIMEOUT`

## タスクリスト

### フェーズ 1: コネクションプール設定

- [x] **1-1**: PostgreSQLコネクションプール設定の追加

  - `cmd/server/main.go` に設定を追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [x] **1-2**: Riverコネクションプール設定の追加

  - `internal/worker/client.go` に設定を追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 15 行（実装 15 行 + テスト 0 行）

- [x] **1-3**: Redisコネクションプール設定の追加
  - `cmd/server/main.go` に設定を追加
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

### フェーズ 2: タイムアウト設定

- [x] **2-1**: HTTPサーバータイムアウト設定の追加

  - `cmd/server/main.go` の `http.Server` 定義を更新
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [x] **2-2**: Turnstileクライアントのタイムアウト設定

  - `internal/turnstile/verify.go` のタイムアウトを10秒に設定
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [x] **2-3**: Resendクライアントのタイムアウト設定
  - `internal/worker/client.go` にタイムアウトを設定
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

### フェーズ 3: リクエスト制限

- [x] **3-1**: リクエストボディサイズ制限ミドルウェアの追加
  - カスタムミドルウェアを作成
  - `cmd/server/main.go` にミドルウェアを追加
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **環境変数による設定のカスタマイズ**: 固定値で十分であり、複雑性を避けるため
- **ヘルスチェックのRedis/River対応拡張**: 別途対応を検討
- **TLS設定**: リバースプロキシ（Cloudflare/nginx）経由のため不要
- **リソース監視（メトリクス収集）**: Sentryで基本的なエラー追跡は実装済み

## 参考資料

- [Go database/sql パッケージ - 接続プール設定](https://pkg.go.dev/database/sql#DB.SetMaxOpenConns)
- [pgxpool 設定ドキュメント](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
- [go-redis 設定ドキュメント](https://pkg.go.dev/github.com/redis/go-redis/v9)
- [net/http Server タイムアウト設定](https://pkg.go.dev/net/http#Server)
- [Cloudflare Blog: So you want to expose Go on the Internet](https://blog.cloudflare.com/exposing-go-on-the-internet/)
