# メンテナンスモード 設計書

## 概要

Go 版 Annict にメンテナンスモード機能を実装します。Rails 版と同様の仕組みで、環境変数によってメンテナンスモードを有効化し、管理者 IP 以外のユーザーにメンテナンスページを表示します。

**目的**:

- サーバーメンテナンス時にユーザーへ適切な案内を表示する
- 管理者はメンテナンス中でもサイトにアクセスして動作確認ができる

**背景**:

- Rails 版には既にメンテナンスモードが存在する
- Go 版でも同様の機能が必要

## 要件

### 機能要件

- 環境変数 `ANNICT_MAINTENANCE_MODE` が `"on"` のとき、メンテナンスモードを有効化する
- 環境変数 `ANNICT_ADMIN_IP` で指定した IP アドレスからのアクセスは通常通り処理する
- 管理者 IP 以外からのアクセスにはメンテナンスページ（HTTP 503）を返す
- メンテナンスページのデザインは Rails 版（`rails/public/maintenance.html`）と同じにする

### 非機能要件

- **パフォーマンス**: ミドルウェアとして実装し、すべてのリクエストで最初にチェックする
- **セキュリティ**: 管理者 IP のチェックは `X-Forwarded-For` ヘッダーも考慮する（Cloudflare 経由のアクセス）

## 設計

### 技術スタック

- Go 標準ライブラリのみ使用（追加ライブラリ不要）
- 既存の `internal/clientip` パッケージを再利用（IP アドレス取得）

### アーキテクチャ

ミドルウェアとして実装し、ルーターの最初に適用します。

```
リクエスト → MaintenanceMiddleware → 他のミドルウェア → ハンドラー
                    ↓
              メンテナンス中 && 管理者IP以外
                    ↓
              503 メンテナンスページ
```

### コード設計

#### 環境変数

| 環境変数 | 説明 | 例 |
|----------|------|-----|
| `ANNICT_MAINTENANCE_MODE` | メンテナンスモードの有効/無効 | `"on"` で有効 |
| `ANNICT_ADMIN_IP` | 管理者 IP アドレス（カンマ区切りで複数指定可） | `"192.168.1.1"` または `"192.168.1.1,10.0.0.1"` |

#### パッケージ構成

```
internal/
├── clientip/
│   └── clientip.go         # 既存のIPアドレス取得（再利用）
├── middleware/
│   ├── maintenance.go      # メンテナンスミドルウェア
│   └── maintenance_test.go # テスト
├── templates/
│   └── pages/
│       └── maintenance.templ  # メンテナンスページテンプレート
└── config/
    └── config.go           # 環境変数の追加
```

#### 主要な構造体・関数

```go
// internal/middleware/maintenance.go

// MaintenanceMiddleware はメンテナンスモードをチェックするミドルウェアを返します
func MaintenanceMiddleware(cfg *config.Config) func(next http.Handler) http.Handler

// isAdminIP は指定された IP が管理者 IP かどうかをチェックします
// 内部で clientip.GetClientIP() を使用してクライアント IP を取得
func isAdminIP(r *http.Request, adminIPs []string) bool
```

#### IP アドレス取得（既存の clientip パッケージを使用）

`internal/clientip/clientip.go` の `GetClientIP()` 関数を使用します。
優先順位:
1. `CF-Connecting-IP`（Cloudflare が設定する実際のクライアント IP）
2. `X-Forwarded-For`（プロキシチェーンの最初の IP）
3. `X-Real-IP`
4. `RemoteAddr`（直接接続の場合）

### テスト戦略

- ミドルウェアの単体テスト
  - メンテナンスモード OFF の場合は通常処理
  - メンテナンスモード ON + 管理者 IP の場合は通常処理
  - メンテナンスモード ON + 一般 IP の場合は 503 を返す
  - 複数の管理者 IP に対応
  - X-Forwarded-For ヘッダーからの IP 取得

## タスクリスト

### フェーズ 1: 基盤実装

- [x] **1-1**: Config への環境変数追加

  - `internal/config/config.go` に `MaintenanceMode` と `AdminIPs` を追加
  - 環境変数 `ANNICT_MAINTENANCE_MODE` と `ANNICT_ADMIN_IP` を読み込む
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

- [x] **1-2**: メンテナンスページテンプレート作成

  - `internal/templates/pages/maintenance/maintenance.templ` を作成
  - Rails 版と同じデザインを実装
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 60 行（実装 60 行）

### フェーズ 2: ミドルウェア実装

- [x] **2-1**: メンテナンスミドルウェア実装

  - `internal/middleware/maintenance.go` を作成
  - IP アドレス取得ロジック実装
  - 管理者 IP チェックロジック実装
  - テストを実装
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 80 行 + テスト 120 行）

### フェーズ 3: 統合

- [x] **3-1**: ルーターへのミドルウェア適用

  - `cmd/server/main.go` でミドルウェアを適用
  - `.env.example` に環境変数を追加
  - **想定ファイル数**: 約 2 ファイル（実装 2）
  - **想定行数**: 約 20 行（実装 20 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **メンテナンス終了時刻の表示**: 現時点では不要（Rails 版にもない）
- **メンテナンスページの国際化**: Rails 版が日本語のみのため、同様に日本語のみ
- **IP アドレスの範囲指定（CIDR）**: 現時点では単一 IP またはカンマ区切りで十分

## 参考資料

- [Rails 版メンテナンスページ](../../../rails/public/maintenance.html)
- [Chi Middleware Examples](https://github.com/go-chi/chi#middleware-handlers)
