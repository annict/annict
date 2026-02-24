# リバースプロキシ機能の実装 設計書

## 概要

`annict.com` を Go 版アプリのメインドメインとし、Go 版で未実装の機能については Rails 版にリバースプロキシする仕組みを実装します。これにより、`go.annict.com` のような別ドメインを使用せず、段階的に Rails から Go へ移行できるようにします。

**目的**:

- **SEO 対策**: `annict.com` と `go.annict.com` の 2 つのドメインで同じコンテンツを配信することによる SEO の問題を回避
- **リダイレクト不要**: ユーザーが `go.annict.com` にアクセスする必要をなくし、常に `annict.com` でアクセス可能に
- **段階的移行**: Go 版で実装済みの機能から順次切り替え、未実装の機能は Rails 版で継続提供
- **開発体験の向上**: 開発環境と本番環境で同じプロキシロジックを使用し、環境差分を最小化

**背景**:

- 現在 `go.annict.com` というサブドメインで Go 版を運用しようとしているが、以下の問題がある：
  - SEO 的に `annict.com` と `go.annict.com` で同じコンテンツが重複する
  - ユーザーが 2 つのドメインを意識する必要がある
  - Rails 版から Go 版へのリダイレクトが必要になる
- `annict.com` を Go 版のメインドメインにすることで、これらの問題を解決する

## 要件

### 機能要件

- **デフォルトフォールバック方式**: Go 版で処理できないリクエストは自動的に Rails 版にプロキシする
- **Go 版で処理するパス**（プロキシしない）:
  - `/static/*` - 静的ファイル（CSS、JS、画像など）
  - `/health` - ヘルスチェックエンドポイント
  - `/manifest.json` - Web App Manifest
  - `/sign_in` - ログインページ（GET）・ログイン処理（POST）
  - `/password/reset` - パスワードリセット申請ページ（GET）・申請処理（POST）
  - `/password/edit` - パスワードリセット実行ページ（GET）
  - `/password` - パスワード更新処理（PUT）
  - その他、Go 版で実装完了した機能を順次追加（例: `/works/popular` など）
- **Rails 版にプロキシするパス**: 上記以外のすべてのパス（`/`, `/works`, `/works/popular`, `/@username` など）
- **API サブドメイン（`api.annict.com`）**: すべてのリクエストを Rails 版にプロキシ
  - GraphQL API（`/graphql`）
  - REST API（`/api/*`）
  - OAuth エンドポイント（`/oauth/*`）
- **セッション共有**: 既存の PostgreSQL sessions テーブルを共有し、Cookie ドメインで認証状態を共有
- **リクエストの透過的転送**: HTTP ヘッダー、Cookie、リクエストボディをそのまま Rails 版に転送
- **レスポンスの透過的返却**: Rails 版からのレスポンスをそのままクライアントに返却
- **エラーハンドリング**: Rails 版が応答しない場合は適切なエラーページを表示

### 非機能要件

#### セキュリティ

- **SSRF 対策**: プロキシ先は環境変数で指定された Rails 版の URL のみに制限
- **ヘッダーの適切な処理**: `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Real-IP` などを適切に設定
- **Cookie の適切な転送**: セッション Cookie を含むすべての Cookie を転送

#### パフォーマンス

- **タイムアウト設定**: Rails 版への接続タイムアウト・読み取りタイムアウトを適切に設定（30 秒程度）
- **接続プーリング**: Rails 版への HTTP クライアントは接続プーリングを使用
- **ストリーミング**: 大きなレスポンスはストリーミングで転送

#### ユーザビリティ（UX）

- **透過的な動作**: ユーザーはプロキシされていることを意識しない
- **一貫した URL**: すべてのリクエストは `annict.com`（本番）または `example.dev`（開発）で処理

#### 可用性・信頼性

- **エラーページ**: Rails 版が応答しない場合は、Go 版のエラーページを表示
- **ログ出力**: プロキシ処理のログを出力し、トラブルシューティングを容易に

#### 保守性

- **環境変数で設定**: Rails 版の URL は環境変数で設定し、環境ごとに切り替え可能
- **ホワイトリスト方式**: Go 版で処理するパスをホワイトリストで管理し、追加・削除が容易
- **テスト可能**: プロキシロジックを単体テストで検証可能

## 設計

### 技術スタック

- **Go 標準ライブラリ**: `net/http`, `net/http/httputil`
- **httputil.ReverseProxy**: Go 標準の リバースプロキシ実装を使用

### アーキテクチャ

```
[クライアント]
    ↓
[Cloudflare / Nginx (Dokku)]
    ↓
[Go版アプリ (annict.com / api.annict.com)]
    ↓
    ├─ annict.com:
    │   ├─ /sign_in → Go版ハンドラー
    │   ├─ /password/* → Go版ハンドラー
    │   └─ その他 → [Rails版アプリ (内部通信)] ← リバースプロキシ
    │
    └─ api.annict.com:
        └─ すべて → [Rails版アプリ (内部通信)] ← リバースプロキシ
```

### ミドルウェア設計

`internal/middleware/reverse_proxy.go` に実装：

```go
type ReverseProxyMiddleware struct {
    railsURL *url.URL
    proxy    *httputil.ReverseProxy
}

func NewReverseProxyMiddleware(railsURL string) *ReverseProxyMiddleware
func (m *ReverseProxyMiddleware) Middleware(next http.Handler) http.Handler
```

**処理フロー**:

1. リクエストのホスト名を確認
   - `api.annict.com`（または開発環境の相当するホスト）の場合 → Rails 版にリバースプロキシ
2. リクエストのパスを確認
3. Go 版で処理するパス（ホワイトリスト）にマッチする場合 → `next.ServeHTTP()` で次のハンドラーに処理を渡す
4. マッチしない場合 → Rails 版にリバースプロキシ

**ホワイトリスト**:

```go
var goHandledPaths = []string{
    "/static",         // 静的ファイル（CSS、JS、画像など）
    "/health",         // ヘルスチェックエンドポイント
    "/manifest.json",  // Web App Manifest
    "/sign_in",        // ログインページ・処理
    "/password/reset", // パスワードリセット申請
    "/password/edit",  // パスワードリセット実行
    "/password",       // パスワード更新
    // 将来追加: "/works/popular"（人気アニメページ）など
}
```

### 環境変数

`.env` / `.env.example` で管理（環境別ファイルは使用していない）：

```bash
# Rails版アプリのURL（内部通信用）
ANNICT_RAILS_APP_URL=http://rails-app:3000  # 開発環境（Docker内部通信）
```

**開発環境**:
- Rails 版: Docker Compose で `rails-app` というサービス名で起動
- Go 版からのプロキシ: `http://rails-app:3000`（Docker 内部ネットワーク経由）
- ホスト名:
  - メインドメイン: `example.dev`
  - API サブドメイン: `api.example.dev`（開発環境では未使用の可能性あり）

**本番環境**:
- Rails 版: Dokku の内部ネットワーク（例: `http://rails.web:3000`）
- Go 版からのプロキシ: 同じ VPS 内の Dokku アプリ間通信
- 環境変数は `dokku config:set` コマンドで設定
- ホスト名:
  - メインドメイン: `annict.com`
  - API サブドメイン: `api.annict.com`

### プロキシ時のヘッダー設定

Rails 版に転送する際、以下のヘッダーを設定・転送する必要があります：

**Go側で設定するヘッダー**:
```go
X-Forwarded-For: <client-ip>
X-Forwarded-Proto: https (本番環境) / http (開発環境)
X-Real-IP: <client-ip>
X-Forwarded-Host: annict.com (本番) / example.dev (開発)
```

**クライアントから受け取ったヘッダーをそのまま転送**:
```go
CF-Connecting-IP: <cloudflare-client-ip>  // メンテナンスモードの管理者IP判定に使用
Origin: <original-origin>                  // CSRF保護のOriginチェックに使用
Referer: <original-referer>                // CSRF保護に使用
Authorization: <basic-auth-header>         // Basic認証に使用（開発環境など）
Cookie: <all-cookies>                      // セッション管理に使用
```

**Rails側での使用箇所**:
- `CF-Connecting-IP`: メンテナンスモード時の管理者IP判定（`config/application.rb:78-85`）
- `Origin` / `Referer`: CSRF保護のOriginチェック（本番環境で有効）
- `Authorization`: Basic認証（`ANNICT_BASIC_AUTH=on`の場合）
- `X-Forwarded-Proto`: SSL強制リダイレクトの判定（`config.force_ssl`）
- `X-Forwarded-Host`: ドメイン正規化リダイレクトの判定（`Rack::Rewrite`）

### エラーハンドリング

Rails 版への接続に失敗した場合：

- **HTTP 502 Bad Gateway** を返す
- Go 版のエラーページを表示（`internal/templates/errors/502.html`）
- ログに詳細なエラー情報を出力

### テスト戦略

- **単体テスト**: プロキシロジックのテスト（`httptest.Server` でモック Rails サーバーを作成）
- **統合テスト**: 実際の Rails 版との連携テスト（開発環境）
- **エッジケース**:
  - Rails 版が応答しない場合
  - タイムアウトが発生した場合
  - リクエストボディが大きい場合

## タスクリスト

### フェーズ 1: 基本実装

- [x] リバースプロキシミドルウェアの実装

  - `internal/middleware/reverse_proxy.go` の作成
  - `httputil.ReverseProxy` を使用した基本実装
  - Go 版で処理するパスのホワイトリスト管理
  - Rails 版への透過的なリクエスト転送
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

- [x] 環境変数設定の追加

  - `internal/config/config.go` に `RailsAppURL` フィールドを追加
  - `.env.example` に `ANNICT_RAILS_APP_URL` を追加（開発環境用のテンプレート）
  - **実装済み**: `RailsAppURL` フィールドは `internal/config/config.go:52` に追加済み
  - **実装済み**: `.env.example:11` に `ANNICT_RAILS_APP_URL=http://rails-app:3000` を追加済み
  - **注**: 現在は `.env` と `.env.example` のみを使用（環境別ファイルは使用していない）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + 設定ファイル 1）
  - **想定行数**: 約 15 行（実装 10 行 + 設定ファイル 5 行）

- [x] ミドルウェアの登録

  - `cmd/server/main.go` でミドルウェアをルーターに登録
  - ミドルウェアの順序を適切に設定（セッションミドルウェアの後、ルーティングの前）
  - **実装済み**: `cmd/server/main.go:133-157` でリバースプロキシミドルウェアを初期化・登録
  - **実装内容**: `ANNICT_RAILS_APP_URL` が設定されている場合のみミドルウェアを有効化
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 10 行（実装 10 行）

### フェーズ 2: エラーハンドリングとログ出力

- [x] エラーページの実装

  - `internal/templates/errors/502.html` の作成
  - Rails 版への接続エラー時の適切なエラーメッセージ表示
  - **実装済み**: `internal/templates/errors/502.html` を作成
  - **実装済み**: `internal/middleware/reverse_proxy.go` のエラーハンドラーを更新
  - **実装済み**: i18n翻訳ファイル（ja.toml, en.toml）に502エラー用のメッセージを追加
  - **実装済み**: 詳細なエラーログを出力（パス、メソッド、リモートアドレス）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テンプレート 1）
  - **想定行数**: 約 80 行（実装 30 行 + テンプレート 50 行）

- [x] ログ出力の実装

  - プロキシ処理のログ出力（リクエストパス、プロキシ先 URL、レスポンスステータスなど）
  - エラー時の詳細なログ出力
  - **実装済み**: `ModifyResponse` を設定してレスポンスステータスのログを追加
  - **実装済み**: リクエスト転送時とエラー時の詳細なログ出力を実装
  - **想定ファイル数**: 約 1 ファイル（実装 1）
  - **想定行数**: 約 30 行（実装 30 行）

### フェーズ 3: セキュリティとパフォーマンス対策

- [x] セキュリティ対策の実装

  - `X-Forwarded-*` ヘッダーの適切な設定（`X-Forwarded-For`, `X-Forwarded-Proto`, `X-Real-IP`, `X-Forwarded-Host`）
  - Cloudflare ヘッダーの転送（`CF-Connecting-IP`）- `httputil.ReverseProxy`がデフォルトで転送
  - CSRF 保護用ヘッダーの転送（`Origin`, `Referer`）- `httputil.ReverseProxy`がデフォルトで転送
  - Basic 認証ヘッダーの転送（`Authorization`）- `httputil.ReverseProxy`がデフォルトで転送
  - Cookie の適切な転送 - `httputil.ReverseProxy`がデフォルトで転送
  - SSRF 対策（プロキシ先の URL を環境変数で制限）- 既に実装済み
  - ホスト名のバリデーション（`annict.com` と `api.annict.com` のみ許可、開発環境は `example.dev` と `api.example.dev`）
  - **実装完了**: `internal/middleware/reverse_proxy.go` と `internal/middleware/reverse_proxy_test.go`
  - **実装内容**:
    - **クライアントIP取得** (`getClientIP`関数):
      - 優先順位: `CF-Connecting-IP` > `X-Forwarded-For`の最初のIP > `RemoteAddr`
      - Cloudflare経由のアクセスでは、実際のクライアントIP（IPv6含む）を正しく取得
      - 開発環境でもCloudflare Tunnel経由で実際のアクセス元IPを取得可能
    - `X-Forwarded-For`: 既存の値を維持（Cloudflareなどが設定した値を保持）
      - 注: `httputil.ReverseProxy`の標準動作により、内部で`RemoteAddr`が追加される場合がある
      - Rails版は`CF-Connecting-IP`または`X-Forwarded-For`の最初のIPで実際のクライアントIPを取得するため問題なし
    - `X-Real-IP`: クライアントIPを設定（既存の値がある場合は維持）
    - `X-Forwarded-Proto`: "https"（固定値）
    - `X-Forwarded-Host`: `cfg.Domain`から設定
    - `isAPISubdomain`メソッド: APIサブドメイン（`api.annict.com`など）へのリクエストをすべてRails版にプロキシ
  - **テスト追加**:
    - `TestGetClientIP`: クライアントIP取得ロジックのテスト（CF-Connecting-IP、X-Forwarded-For、RemoteAddrの優先順位）
    - `TestIsAPISubdomain`: APIサブドメインの判定テスト
    - `TestReverseProxyMiddleware_APISubdomain`: APIサブドメインへのリクエストがRails版にプロキシされることを確認
    - `TestReverseProxyMiddleware_PreserveExistingHeaders`: 既存のヘッダー（X-Forwarded-For、X-Real-IP）が維持されることを確認
    - `TestReverseProxyMiddleware_CFConnectingIP`: CF-Connecting-IPヘッダーがそのまま転送されることを確認
    - 既存のテストを更新: X-Forwarded-ForとX-Real-IPの設定確認を追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [x] タイムアウト設定の実装

  - Rails 版への接続タイムアウト（10 秒）
  - レスポンス読み取りタイムアウト（30 秒）
  - **実装完了**: `internal/middleware/reverse_proxy.go` と `internal/middleware/reverse_proxy_test.go`
  - **実装内容**:
    - カスタムの`http.Transport`を設定
    - `DialContext`で接続タイムアウト（10秒）
    - `ResponseHeaderTimeout`でレスポンスヘッダー読み取りタイムアウト（30秒）
    - 接続プーリングの設定（`MaxIdleConns`, `IdleConnTimeout`など）
  - **テスト追加**:
    - `TestReverseProxyMiddleware_ResponseHeaderTimeout`: レスポンスヘッダー読み取りタイムアウトのテスト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

### フェーズ 4: テストと検証

- [x] 単体テストの実装

  - `httptest.Server` でモック Rails サーバーを作成
  - ホワイトリストのパスが Go 版で処理されることを確認
  - その他のパスが Rails 版にプロキシされることを確認
  - `api.annict.com` ホスト名のリクエストがすべて Rails 版にプロキシされることを確認
  - ホスト名のバリデーション（許可されていないホスト名の拒否）
  - ヘッダー転送のテスト（`X-Forwarded-*`, `CF-Connecting-IP`, `Origin`, `Referer`, `Authorization`, `Cookie`）
  - エラーケース（Rails 版が応答しない、タイムアウトなど）のテスト
  - **実装済み**: `internal/middleware/reverse_proxy_test.go` に以下のテストを実装
    - `TestReverseProxyMiddleware_GoHandledPaths`: Go版で処理するパスのテスト（7パターン）
    - `TestReverseProxyMiddleware_RailsProxiedPaths`: Rails版にプロキシするパスのテスト（6パターン）
    - `TestReverseProxyMiddleware_HeaderForwarding`: ヘッダー転送のテスト（CF-Connecting-IP, Origin, Referer, Authorization, Cookie）
    - `TestReverseProxyMiddleware_ErrorHandling`: エラーハンドリングのテスト（Rails版が応答しない場合）
    - `TestIsGoHandledPath`: ホワイトリストパスの判定テスト（11パターン）
    - `TestIsAPISubdomain`: APIサブドメインの判定テスト（6パターン）
    - `TestReverseProxyMiddleware_APISubdomain`: APIサブドメインへのリクエストがRails版にプロキシされることを確認（4パターン）
    - `TestReverseProxyMiddleware_PreserveExistingHeaders`: 既存のヘッダー（X-Forwarded-For, X-Real-IP）が維持されることを確認
    - `TestGetClientIP`: クライアントIP取得ロジックのテスト（5パターン）
    - `TestReverseProxyMiddleware_CFConnectingIP`: CF-Connecting-IPヘッダーがそのまま転送されることを確認
    - `TestReverseProxyMiddleware_ResponseHeaderTimeout`: レスポンスヘッダー読み取りタイムアウトのテスト
    - `TestReverseProxyMiddleware_HTTPMethods`: 様々なHTTPメソッド（GET/POST/PUT/PATCH/DELETE）のテスト
    - `TestReverseProxyMiddleware_RequestBodyForwarding`: リクエストボディの転送テスト
    - `TestReverseProxyMiddleware_MultipleHostnames`: 複数のホスト名でのテスト（6パターン）
    - `TestReverseProxyMiddleware_LargeRequestBody`: 大きなリクエストボディの転送テスト
  - **想定ファイル数**: 約 1 ファイル（テスト 1）
  - **想定行数**: 約 200 行（テスト 200 行）→ **実際**: 約 850 行（より網羅的なテストを実装）

- [x] 開発環境での動作確認

  - `example.dev` で Go 版と Rails 版の連携を確認
  - `/sign_in`, `/password/reset`, `/password/edit`, `/password` が Go 版で処理されることを確認
  - その他のパス（例: `/works`, `/users`, `/` など）が Rails 版で処理されることを確認
  - `api.example.dev` のすべてのリクエストが Rails 版にプロキシされることを確認
    - GraphQL API（`/graphql`）の動作確認
    - REST API（`/api/*`）の動作確認（存在する場合）
    - OAuth エンドポイント（`/oauth/*`）の動作確認
  - Rails 側のリダイレクト（`/about` → `/` など）が正常に動作することを確認
  - セッション共有が正常に動作することを確認
  - CSRF 保護が正常に動作することを確認（フォーム送信テスト）
  - Basic 認証が有効な場合の動作確認（必要に応じて）
  - **実装完了**: `.claude/designs/1_doing/reverse-proxy-manual-testing.md` に手順書を作成
  - **実装内容**:
    - 前提条件（Rails サービスの起動確認）
    - Go サーバーの起動手順
    - Go 版で処理されるパスの確認手順（静的ファイル、ヘルスチェック、ログイン、パスワードリセット）
    - Rails 版にプロキシされるパスの確認手順（ルート、作品一覧、ユーザープロフィール）
    - Rails のリダイレクト処理の確認（`/about` → `/`、`/activities` → `/`）
    - API サブドメインの確認（GraphQL API、OAuth エンドポイント）
    - セッション共有の確認手順（Go 版でログイン → Rails 版でセッション維持）
    - CSRF 保護の確認手順（Rails フォーム送信）
    - ヘッダー転送の確認（X-Forwarded-*、CF-Connecting-IP、Origin、Referer、Cookie）
    - エラーハンドリングの確認（Rails 版停止時の 502 エラーページ）
    - 確認チェックリスト（25 項目）
    - トラブルシューティングガイド
  - **ファイル数**: 1 ファイル（ドキュメント）
  - **行数**: 約 510 行（想定の 80 行を大幅に上回る詳細な手順書）

### フェーズ 5: 本番環境対応

- [ ] Dokku 設定の更新

  - Go 版アプリの環境変数に `ANNICT_RAILS_APP_URL` を設定（Dokku の内部ネットワーク URL）
  - ドメイン設定:
    - `annict.com` を Go 版アプリに向ける
    - `api.annict.com` も Go 版アプリに向ける（Go版がRails版にプロキシ）
    - `image.annict.com` は imgproxy のまま維持
  - **想定ファイル数**: 約 1 ファイル（ドキュメント 1）
  - **想定行数**: 約 80 行（手順書 + コマンド例）

- [ ] 本番環境での動作確認

  - ステージング環境で動作確認
  - Go 版で処理するパスと Rails 版にプロキシするパスの動作確認
  - `api.annict.com` へのリクエストが正しく Rails 版にプロキシされることを確認
    - GraphQL API の動作確認
    - OAuth 認証フローの動作確認
  - セッション共有の動作確認
  - パフォーマンス確認（プロキシのオーバーヘッドが許容範囲内か）
  - **想定ファイル数**: 約 1 ファイル（ドキュメント 1）
  - **想定行数**: 約 60 行（検証手順）

### フェーズ 6: ドキュメント更新

- [ ] 設計書の更新

  - `.claude/designs/1_doing/go.md` にプロキシ構成を追記
  - `CLAUDE.md` にプロキシの説明を追加
  - README.md の更新（ドメイン構成の説明）
  - **想定ファイル数**: 約 3 ファイル（ドキュメント 3）
  - **想定行数**: 約 100 行（ドキュメント 100 行）

### Rails 版の既存機能との互換性

#### Rack::Rewrite によるリダイレクト

Rails 版では `Rack::Rewrite` ミドルウェアで以下のリダイレクト処理を行っています（`/rails-app/config/application.rb:61-76`）：

**ドメイン正規化**:
- `www.annict.com`, `ja.annict.com`, `jp.annict.com` → `annict.com` にリダイレクト

**パスリダイレクト**:
- `/about` → `/`
- `/activities` → `/`
- `/programs` → `/track`
- `/users/{username}` → `/@{username}`
- `/faqs` → `/faq`
- `/works` → `/works/{current_season}`
- など

**プロキシ経由での動作**:
- これらのリダイレクトは Rails 側で処理され、クライアントには正しく 301 リダイレクトが返される
- Go 版のプロキシは透過的に転送するだけなので、特別な対応は不要
- ただし、`X-Forwarded-Host` ヘッダーを正しく設定しないと、ドメイン正規化のリダイレクトが誤動作する可能性がある

#### メンテナンスモード

Rails 版はメンテナンスモード機能を持っています（`/rails-app/config/application.rb:78-85`）：

- `ANNICT_MAINTENANCE_MODE=on` の場合、`public/maintenance.html` を表示
- 管理者 IP（`ANNICT_ADMIN_IP`）からのアクセスは除外
- IP 判定に `CF-Connecting-IP` ヘッダーを使用

**プロキシ経由での動作**:
- `CF-Connecting-IP` ヘッダーを正しく転送すれば、メンテナンスモードは正常に動作する
- Go 版のプロキシはこのヘッダーをそのまま転送する必要がある

#### CSRF 保護

Rails 版は本番環境で CSRF の Origin チェックを有効にしています（`/rails-app/config/initializers/request_forgery_protection.rb:8`）：

```ruby
Rails.application.config.action_controller.forgery_protection_origin_check = Rails.env.production?
```

**プロキシ経由での動作**:
- `Origin` と `Referer` ヘッダーを正しく転送すれば、CSRF 保護は正常に動作する
- `httputil.ReverseProxy` はデフォルトでこれらのヘッダーを転送するため、特別な対応は不要

#### SSL 強制リダイレクト

Rails 版は本番環境で SSL 強制リダイレクトを有効にできます（`/rails-app/config/environments/production.rb:38`）：

```ruby
config.force_ssl = ENV["ANNICT_FORCE_SSL"].present?
```

**プロキシ経由での動作**:
- `X-Forwarded-Proto: https` ヘッダーを設定すれば、Rails は「HTTPS でアクセスされた」と認識する
- 内部通信が HTTP でも問題なく動作する

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **キャッシュ**: Rails 版のレスポンスをキャッシュする機能（将来的に検討）
- **負荷分散**: 複数の Rails 版インスタンスへの負荷分散（現時点では 1 台構成のため不要）
- **A/B テスト**: Go 版と Rails 版で A/B テストを行う機能（段階的移行のため不要）
- **リクエストの書き換え**: プロキシ時にリクエストパスや内容を書き換える機能（透過的転送のため不要）
- **Rails 側のリダイレクト処理の Go 版への移行**: `Rack::Rewrite` のリダイレクト処理は Rails 側で継続（将来的に Go 版に移行を検討）

## 参考資料

- [Go httputil.ReverseProxy の公式ドキュメント](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- [Dokku ドキュメント - アプリ間通信](https://dokku.com/docs/networking/network/)
- [Rails と Go でセッションを共有する方法（既存実装を参照）](../../internal/middleware/session/)

---

## 実装の進め方

1. **フェーズ 1**: 基本的なプロキシ機能を実装し、動作確認
2. **フェーズ 2-3**: エラーハンドリング、セキュリティ、パフォーマンス対策を追加
3. **フェーズ 4**: テストで品質を担保
4. **フェーズ 5**: 本番環境での動作確認とデプロイ
5. **フェーズ 6**: ドキュメント整備

各フェーズは 1 つの Pull Request で完結する粒度で実装してください。
