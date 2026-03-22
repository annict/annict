# フィーチャーフラグ 作業計画書

<!--
このテンプレートの使い方:
1. このファイルを `docs/plans/2_todo/` ディレクトリにコピー
   例: cp docs/plans/template.md docs/plans/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**作業計画書の性質**:
- 作業計画書は「何をどう変えるか」という変更内容を記述するドキュメントです
- 新しい機能の場合は、概要・要件・設計もこのドキュメントに記述します
- 現在のシステムの状態は `docs/specs/` の仕様書に記述されています
- タスク完了後は、仕様書を新しい状態に更新してください（設計判断や採用しなかった方針も含める）

**仕様書との関係**:
- 新しい機能の場合: タスク完了後に `docs/specs/` に仕様書を作成する
- 既存機能の変更の場合: 「仕様書」セクションに対応する仕様書へのリンクを記載し、タスク完了後に仕様書を更新する

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 仕様書

<!--
- 既存機能を変更する場合: 変更対象の仕様書へのリンクを記載してください
- 新しい機能の場合: タスク完了後に作成予定の仕様書のパスを記載してください
-->

- [フィーチャーフラグ 仕様書](../specs/feature-flag/overview.md)（タスク完了後に作成）

## 概要

<!--
ガイドライン:
- この機能が「何であるか」「なぜ必要か」を簡潔に説明
- 2-3段落程度で簡潔に
- 既存機能の変更の場合は、変更の背景と目的を記述
-->

フィーチャーフラグは、RailsからGoへの段階的移行において、Go版で再実装した機能を特定のユーザーやデバイスにだけ先行公開するための仕組みである。`feature_flags` テーブルで `device_token`（Cookieに保存されるデバイス識別子）または `user_id` と `name` の組み合わせによりフラグの有効/無効を管理する。レコードが存在すればフラグ有効、存在しなければフラグ無効という単純なモデルで動作する。

リバースプロキシミドルウェアにフラグ判定機能を統合し、URLパターンに応じてGoまたはRailsにルーティングする。フラグが有効なユーザー/デバイスはGo版のハンドラーで処理され、フラグが無効な場合やフラグ判定でエラーが発生した場合はRails版にプロキシされる。未ログインユーザーに対しても `device_token` 経由でフラグの制御が可能である。

本設計は Wikino プロジェクトで実装済みのフィーチャーフラグ仕様（`/wikino/docs/specs/feature-flag/overview.md`）をベースとし、Annict のセッション管理方式に合わせて適応する。

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- システムは `feature_flags` テーブルで `device_token` または `user_id` によるフラグの有効/無効を管理する
- レコードが存在 = フラグ有効、レコードが存在しない = フラグ無効
- 開発者はDBを直接操作（psqlやマイグレーション）してフラグを管理できる
- システムはリバースプロキシミドルウェアで、`device_token` Cookieが存在しない場合に自動生成してレスポンスにセットする
- システムはリクエストのURLパターンとフラグ状態に基づいて、GoまたはRailsにルーティングする
- フラグが無効な場合、Cookieが存在しない場合、またはフラグ判定でエラーが発生した場合は、Rails版にフォールバックする

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- **パフォーマンス**: フラグ判定はフィーチャーフラグ付きパスにマッチするリクエストのみで実行される。非対象パスでは追加のDB問い合わせは発生しない
- **信頼性**: フラグ判定でエラーが発生した場合は、Rails版にフォールバックする（Go版が表示されないほうがサービス断よりも安全）
- **セキュリティ**: `device_token` Cookieは安全なランダム値を使用し、`HttpOnly` + `Secure` + `SameSite=Lax` で設定する

## 実装ガイドラインの参照

<!--
**重要**: 作業計画書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
作業計画書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン（**ファイル名は標準の9種類のみ**）
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templテンプレートガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

## 設計

<!--
ガイドライン:
- 技術的な実装の設計を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - UI設計（画面構成、ユーザーフローなど）
  - セキュリティ設計（認証・認可、トークン管理など）
  - コード設計（パッケージ構成、主要な構造体など）

**重要: 設計は実装中に更新する**:
- 作業計画書内の設計は初期の方針であり、完璧ではない
- 実装中により良いアプローチが見つかった場合は、設計を積極的に更新する
- 設計に固執して実装の質を下げるよりも、実装で得た知見を設計に反映する方が重要
- 変更した場合は「採用しなかった方針」セクションに変更前の方針と変更理由を記録する
-->

### Wikinoとの主な違い

Wikino プロジェクトの仕様をベースにするが、Annict のセッション管理方式が異なるため以下を適応する。

| 項目               | Wikino                                              | Annict                                                              |
| ------------------ | --------------------------------------------------- | ------------------------------------------------------------------- |
| セッションCookie名 | `user_session_tokens`                               | `_annict_session_v201904`                                           |
| セッションテーブル | `user_sessions`（token → user_id の直接マッピング） | `sessions`（JSONB の `data` カラムに warden 形式で user_id を格納） |
| セッションID変換   | 不要（トークンをそのまま使用）                      | public ID → private ID（`"2::" + SHA256(publicID)`）への変換が必要  |
| user_id の型       | UUID                                                | BIGINT（int64）                                                     |
| 主キー生成         | `generate_ulid()` による UUID                       | `BIGSERIAL`（自動インクリメント）                                   |
| フラグ判定のSQL    | 1クエリで `user_sessions` をサブクエリでJOIN        | Go コードでセッションから user_id を解決し、シンプルなクエリで判定  |

### データベース設計

```sql
CREATE TABLE feature_flags (
    id BIGSERIAL PRIMARY KEY,
    device_token VARCHAR,
    user_id BIGINT REFERENCES users(id),
    name VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CHECK (device_token IS NOT NULL OR user_id IS NOT NULL),
    UNIQUE(device_token, name),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_feature_flags_device_token ON feature_flags(device_token);
CREATE INDEX idx_feature_flags_user_id ON feature_flags(user_id);
CREATE INDEX idx_feature_flags_name ON feature_flags(name);
```

- `device_token` と `user_id` はどちらも nullable だが、CHECK制約で少なくとも一方が NOT NULL であることを保証する
- `(device_token, name)` と `(user_id, name)` の2つのユニーク制約で、同一対象に同一フラグが重複登録されることを防ぐ
- PostgreSQL では UNIQUE制約の NULL値は重複として扱われないため、`device_token` が NULL のレコードは `UNIQUE(device_token, name)` 制約に影響しない
- 主キーは `BIGSERIAL` を使用（Annict の既存マイグレーションパターンに準拠）

### ルーティングの流れ

```
リクエスト到着
  ↓
[リバースプロキシミドルウェア]
  ├─ device_token Cookie なし？ → 自動生成してレスポンスにセット
  ↓
  ├─ APIサブドメイン？ → Rails版にプロキシ
  ├─ 常にGoのパス（goHandledPaths）？ → Go側ミドルウェアチェーン → ハンドラー
  ├─ フィーチャーフラグ付きパス？
  │   ├─ Yes + device_tokenのフラグ有効 → Go側ミドルウェアチェーン → ハンドラー
  │   ├─ Yes + user_idのフラグ有効（ログインユーザー全デバイス）→ Go側ミドルウェアチェーン → ハンドラー
  │   └─ Yes + フラグ無効/トークンなし/エラー → Railsにプロキシ
  └─ その他 → Railsにプロキシ
```

### コード設計

#### ドメインID型

`internal/model/id.go` に FeatureFlag 関連のドメインID型を定義する。今後、他のモデルのID型もこのファイルに追加していく:

```go
package model

// FeatureFlagID はフィーチャーフラグのID型
type FeatureFlagID int64

// String は文字列表現を返す
func (id FeatureFlagID) String() string { return fmt.Sprintf("%d", id) }

// FeatureFlagName はフィーチャーフラグ名の型
type FeatureFlagName string

// String は文字列表現を返す
func (n FeatureFlagName) String() string { return string(n) }
```

#### Model

`internal/model/feature_flag.go`:

```go
package model

import "time"

// FeatureFlag はフィーチャーフラグを表すドメインモデル
type FeatureFlag struct {
    ID          FeatureFlagID
    DeviceToken *string
    UserID      *int64
    Name        FeatureFlagName
    CreatedAt   time.Time
}
```

フラグ名の定数も同ファイルに定義する:

```go
// フラグ名の定数
// Go版への移行で使用するフラグには go_ プレフィックスを付ける
const (
    // 今後のGo移行タスクで追加される
    // 例: FeatureFlagGoPageEdit FeatureFlagName = "go_page_edit"
)
```

#### sqlc クエリ

`internal/query/queries/feature_flags.sql`:

```sql
-- name: IsFeatureFlagEnabled :one
SELECT EXISTS(
    SELECT 1 FROM feature_flags ff
    WHERE ff.name = $3
    AND (
        (ff.device_token IS NOT NULL AND ff.device_token = $1)
        OR (ff.user_id IS NOT NULL AND ff.user_id = $2)
    )
);
```

- `$1` = `device_token` Cookie の値（空文字列は何にもマッチしない）
- `$2` = user_id（セッションから解決済み、0は何にもマッチしない）
- `$3` = フラグ名

#### Repository

`internal/repository/feature_flag.go`:

```go
type FeatureFlagRepository struct {
    q *query.Queries
}

// IsEnabled は指定ユーザーに対してフラグが有効かどうかを返す（内部利用・テスト用）
func (r *FeatureFlagRepository) IsEnabled(ctx context.Context, userID int64, name model.FeatureFlagName) (bool, error)

// IsEnabledByDeviceOrUser はデバイストークンまたはユーザーIDでフラグが有効かどうかを返す
func (r *FeatureFlagRepository) IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID int64, name model.FeatureFlagName) (bool, error)
```

#### リバースプロキシミドルウェア

`internal/middleware/reverse_proxy.go` にフィーチャーフラグ判定を統合:

```go
// DeviceTokenCookieName はデバイス（ブラウザ）識別用のCookieキー名
const DeviceTokenCookieName = "device_token"

// featureFlagChecker はフィーチャーフラグの有効判定を行うインターフェース
type featureFlagChecker interface {
    IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID int64, name model.FeatureFlagName) (bool, error)
}

// featureFlaggedPattern はフィーチャーフラグで制御するURLパターンを定義
type featureFlaggedPattern struct {
    pattern *regexp.Regexp
    flag    model.FeatureFlagName
}

// フィーチャーフラグで制御するURLパターンのリスト
// 具体的なパターンは各Go移行タスクで追加される
var featureFlaggedPatterns []featureFlaggedPattern
```

`ReverseProxyMiddleware` 構造体に以下のフィールドを追加:

```go
type ReverseProxyMiddleware struct {
    railsURL       *url.URL
    proxy          *httputil.ReverseProxy
    cfg            *config.Config
    featureFlagRepo featureFlagChecker  // nil許容（テスト時やフラグ機能不要時）
    sessionMgr      sessionUserResolver // nil許容
}
```

セッションからuser_idを解決するためのインターフェース:

```go
// sessionUserResolver はセッションCookieからユーザーIDを解決するインターフェース
type sessionUserResolver interface {
    GetSessionID(r *http.Request) (string, error)
    GetSession(ctx context.Context, sessionID string) (*session.SessionData, error)
}
```

`Middleware` メソッドの処理フロー:

1. `device_token` Cookieが存在しない場合、`ensureDeviceToken` で自動生成してレスポンスにセットする
2. APIサブドメインかどうかを判定する
3. 常にGoのパスかどうかを判定する
4. フィーチャーフラグ付きパスの場合、`isFeatureFlagEnabled` で判定する
5. その他は Rails にプロキシする

`isFeatureFlagEnabled` の処理:

1. `featureFlagRepo` が `nil` なら `false` を返す
2. リクエストから `device_token` Cookie を取得
3. `sessionMgr` が `nil` でなければ、セッションCookieから user_id を解決する
4. `FeatureFlagRepository.IsEnabledByDeviceOrUser` で判定
5. エラー時は `false` を返す（Rails版にフォールバック）

#### device_token Cookie

`internal/session/token.go` にセキュアトークン生成関数を追加:

```go
// GenerateSecureToken は安全なランダムトークンを生成する
// 24バイトのランダムデータをBase64 URL-safeエンコードした32文字の文字列を返す
func GenerateSecureToken() (string, error) {
    b := make([]byte, 24)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
}
```

Cookie属性: `HttpOnly: true`, `Secure: true`（本番環境）, `SameSite: Lax`, `MaxAge: 10年`

#### テストユーティリティ

`internal/testutil/feature_flag_builder.go`:

```go
type FeatureFlagBuilder struct {
    t           *testing.T
    tx          *sql.Tx
    deviceToken *string
    userID      *int64
    name        string
}

func NewFeatureFlagBuilder(t *testing.T, tx *sql.Tx) *FeatureFlagBuilder
func (b *FeatureFlagBuilder) WithDeviceToken(token string) *FeatureFlagBuilder
func (b *FeatureFlagBuilder) WithUserID(userID int64) *FeatureFlagBuilder
func (b *FeatureFlagBuilder) WithName(name string) *FeatureFlagBuilder
func (b *FeatureFlagBuilder) Build() int64
```

### フラグ管理の運用例

```sql
-- 未ログインユーザーの特定デバイスに対してフラグを設定
INSERT INTO feature_flags (device_token, name) VALUES ('cookie値', 'go_page_edit');

-- ログインユーザーの全デバイスに対してフラグを設定
INSERT INTO feature_flags (user_id, name) VALUES (123, 'go_page_edit');
```

### ファイル構成

```
go/
├── db/
│   └── migrations/
│       └── YYYYMMDDHHMMSS_create_feature_flags.sql   # マイグレーション
├── internal/
│   ├── model/
│   │   ├── id.go                                      # ドメインID型（FeatureFlagID, FeatureFlagName）
│   │   └── feature_flag.go                            # FeatureFlagモデル
│   ├── query/
│   │   └── queries/
│   │       └── feature_flags.sql                      # sqlcクエリ定義
│   ├── repository/
│   │   ├── feature_flag.go                            # リポジトリ
│   │   └── feature_flag_test.go                       # リポジトリテスト
│   ├── middleware/
│   │   ├── reverse_proxy.go                           # フラグ判定統合済みリバースプロキシ（変更）
│   │   └── reverse_proxy_test.go                      # ミドルウェアテスト（変更）
│   ├── session/
│   │   └── token.go                                   # セキュアトークン生成
│   └── testutil/
│       └── feature_flag_builder.go                    # テスト用ビルダー
└── cmd/
    └── server/
        └── main.go                                    # 依存性注入の変更
```

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録
- タスク完了後、この内容は `specs/` の仕様書にも転記する
- 該当がない場合は「なし」と記載
-->

### セッションJOINによる1クエリでのフラグ判定（Wikino方式）

Wikino では `user_session_tokens` テーブルを `IsEnabledForDevice` のサブクエリでJOINし、1クエリで device_token と user_id の両方をチェックしている。Annict でも同様に sessions テーブルの JSONB データからサブクエリで user_id を抽出する方式を検討した。

**不採用の理由**: Annict のセッションテーブルは JSONB の `data` カラムに warden 形式（`{"warden.user.user.key": [[user_id], "salt"]}`）で user_id を格納しており、SQL 内での JSONB パス抽出（`data #>> '{warden.user.user.key,0,0}'`）は Wikino の単純な `token → user_id` マッピングに比べて脆弱性が高い。Go コードでセッションから user_id を解決してからシンプルな SQL で判定するほうが、保守性と可読性が優れる。パフォーマンスへの影響はフィーチャーフラグ付きパスのみで発生するため、実用上問題ない。

### UUID 主キーの使用

Wikino では `generate_ulid()` による UUID を主キーに使用している。

**不採用の理由**: Annict の Go 版マイグレーションでは `BIGSERIAL` を主キーに使用する規約が確立されており（`password_reset_tokens`, `sign_up_codes` 等）、UUID 生成関数（`generate_ulid()`）はスキーマに存在しない。既存パターンとの一貫性を優先する。

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- **フェーズ番号は半角英数字とハイフンのみで表記**してください（ブランチ名に使用するため）
  - 例: フェーズ 1, フェーズ 2, フェーズ 5a（フェーズ 5 と 6 の間に追加する場合）
  - NG: フェーズ 5.5（ドットは使用不可）
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）

プラットフォームプレフィックス:
- Go版またはRails版の修正を行うタスクには、タスク名の先頭にプラットフォームを示すプレフィックスを付けてください
- フォーマット: **フェーズ番号-タスク番号**: [Go] タスク名 または **フェーズ番号-タスク番号**: [Rails] タスク名
- Go版とRails版の両方を修正する場合は、別々のタスクに分けてください
- 例:
  - `- [ ] **1-1**: [Go] マイグレーション作成`
  - `- [ ] **1-2**: [Rails] モデルへのコールバック追加`
-->

### フェーズ 1: データベースとリポジトリ

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
Go版/Rails版の両方を修正する場合は別タスクに分けてください
-->

- [x] **1-1**: [Go] フィーチャーフラグのマイグレーション・モデル・リポジトリの実装
  - `db/migrations/YYYYMMDDHHMMSS_create_feature_flags.sql` にマイグレーションを作成
  - `internal/model/feature_flag.go` に FeatureFlag モデルとフラグ名定数を定義
  - `internal/query/queries/feature_flags.sql` に sqlc クエリを定義し、`sqlc generate` を実行
  - `internal/repository/feature_flag.go` に FeatureFlagRepository を実装（`IsEnabled`, `IsEnabledByDeviceOrUser`）
  - `internal/testutil/feature_flag_builder.go` にテスト用ビルダーを作成
  - `internal/repository/feature_flag_test.go` にリポジトリのテストを作成
  - **想定ファイル数**: 約 8 ファイル（実装 5 + sqlc 生成 1 + テスト 2）
  - **想定行数**: 約 280 行（実装 160 行 + テスト 120 行）

### フェーズ 2: ミドルウェア統合

- [ ] **2-1**: [Go] リバースプロキシミドルウェアへのフィーチャーフラグ判定の統合
  - `internal/session/token.go` に `GenerateSecureToken()` 関数を追加
  - `internal/middleware/reverse_proxy.go` を変更:
    - `featureFlagChecker` インターフェースを定義
    - `sessionUserResolver` インターフェースを定義
    - `featureFlaggedPattern` 構造体と空の `featureFlaggedPatterns` スライスを定義
    - `DeviceTokenCookieName` 定数を定義
    - `ReverseProxyMiddleware` 構造体に `featureFlagRepo` と `sessionMgr` フィールドを追加
    - `NewReverseProxyMiddleware` のシグネチャを変更（`featureFlagRepo` と `sessionMgr` を引数に追加）
    - `ensureDeviceToken` メソッドを実装
    - `isFeatureFlagEnabled` メソッドを実装
    - `Middleware` メソッドのルーティングロジックにフィーチャーフラグ判定を追加
  - `cmd/server/main.go` を変更:
    - `FeatureFlagRepository` の初期化を追加
    - `NewReverseProxyMiddleware` の呼び出しに `featureFlagRepo` と `sessionMgr` を渡す
  - `internal/middleware/reverse_proxy_test.go` にテストを追加:
    - device_token Cookie 自動生成のテスト
    - フィーチャーフラグ有効時にGo版で処理されるテスト
    - フィーチャーフラグ無効時にRails版にプロキシされるテスト
    - エラー時のフォールバックテスト
    - `featureFlagRepo` が nil の場合のテスト
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 300 行（実装 130 行 + テスト 170 行）

### フェーズ 3: 仕様書への反映

<!--
**重要**: 実装完了後、必ず仕様書を作成・更新してください。
- 新しい機能の場合: `docs/specs/` に仕様書を新規作成する
- 既存機能の変更の場合: 対応する仕様書を最新の状態に更新する
- 概要・仕様・設計・採用しなかった方針を作業計画書から転記・整理する
-->

- [ ] **3-1**: 仕様書の作成
  - `docs/specs/feature-flag/overview.md` に仕様書を作成する
  - 作業計画書の概要・要件・設計・採用しなかった方針を仕様書に反映する

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **具体的なフラグ付きURLパターンの登録**: `featureFlaggedPatterns` は空のまま。具体的なURLパターンは各ページのGo移行タスクで追加する
- **管理UI**: 現時点ではフラグの変更頻度が低く（開発者が手動で設定する程度）、psqlやマイグレーションで十分に管理できる
- **インメモリキャッシュ**: フラグ判定はフィーチャーフラグ付きパスのみで実行されるため、単純なDBクエリで十分な性能が得られる。パフォーマンスが問題になった場合に追加する

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Wikino フィーチャーフラグ仕様書](/wikino/docs/specs/feature-flag/overview.md) - 設計のベースとなった仕様
- [internal/middleware/reverse_proxy.go](/workspace/go/internal/middleware/reverse_proxy.go) - 現在のリバースプロキシミドルウェア
- [internal/session/session.go](/workspace/go/internal/session/session.go) - 現在のセッション管理
- [internal/repository/session.go](/workspace/go/internal/repository/session.go) - セッションリポジトリ（private ID 変換ロジック）
