# フィーチャーフラグ 仕様書

<!--
このテンプレートの使い方:
1. 操作対象のモデルに対応するディレクトリを `docs/specs/` 配下に作成（例: `docs/specs/page/`）
2. このファイルをそのディレクトリにコピー（例: cp docs/specs/template.md docs/specs/page/create.md）
3. [機能名] などのプレースホルダーを実際の内容に置き換え
4. 各セクションのガイドラインに従って記述
5. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**ファイルの配置ルール**:
- 仕様書は操作対象のモデル（名詞）ごとにディレクトリを分け、機能（動詞）をファイル名にする
  - 例: `docs/specs/user/sign-up.md`、`docs/specs/page/create.md`
- モデルに分類しにくい横断的な機能は、その機能自体を名詞としてディレクトリにする
  - 例: `docs/specs/search/full-text.md`
- モデルの定義・状態遷移・他モデルとの関係を記述する場合は `overview.md` を作成する
  - `overview.md` はモデルの静的な性質（「これは何か」）を書く場所
  - 操作に紐づく仕様（バリデーション、権限など）は各機能の仕様書に書く
- 詳細は [@docs/README.md](/workspace/docs/README.md) を参照

**仕様書の性質**:
- 仕様書は「現在のシステムの状態」を記述するドキュメントです
- 実装が完了したら、仕様書を最新の状態に更新してください
- 過去の状態はGit履歴で参照できるため、仕様書には常に現在の状態のみを記述します

**作業計画書との関係**:
- 新しい機能の場合: `docs/plans/` の作業計画書に概要・要件・設計を記述し、タスク完了後にこの仕様書を作成します
- 既存機能の変更の場合: `docs/plans/` の作業計画書に変更内容を記述し、タスク完了後にこの仕様書を更新します

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 概要

<!--
ガイドライン:
- この機能が現在「どのように動いているか」を簡潔に説明
- なぜこの仕組みになっているかの背景も記述
- 2-3段落程度で簡潔に
-->

フィーチャーフラグは、Rails から Go への段階的移行において、Go 版で再実装した機能を特定のユーザーやデバイスにだけ先行公開するための仕組みである。`feature_flags` テーブルで `device_token`（Cookie に保存されるデバイス識別子）または `user_id` と `name` の組み合わせによりフラグの有効/無効を管理する。レコードが存在すればフラグ有効、存在しなければフラグ無効という単純なモデルで動作する。

リバースプロキシミドルウェアにフラグ判定機能が統合されており、URL パターンに応じて Go または Rails にルーティングする。フラグが有効なユーザー/デバイスは Go 版のハンドラーで処理され、フラグが無効な場合やフラグ判定でエラーが発生した場合は Rails 版にプロキシされる。未ログインユーザーに対しても `device_token` 経由でフラグの制御が可能である。

**目的**:

- Go 版で再実装した機能を、全ユーザーに公開する前に特定のユーザー/デバイスで先行検証する
- 問題発生時にフラグを無効化するだけで即座に Rails 版にフォールバックできる安全な移行を実現する

**背景**:

- Wikino プロジェクトで実装済みのフィーチャーフラグ仕様をベースとし、Annict のセッション管理方式（JSONB の warden 形式）に合わせて適応している
- フラグ管理は DB 直接操作（psql やマイグレーション）で行う運用を前提としている。変更頻度が低く管理 UI は不要と判断したため

## 仕様

<!--
ガイドライン:
- 現在のシステムの振る舞いを記述
- 「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
- 必要に応じて非機能的な仕様（セキュリティ、パフォーマンスなど）も記述
-->

### フラグ管理

- システムは `feature_flags` テーブルで `device_token` または `user_id` によるフラグの有効/無効を管理する
- レコードが存在 = フラグ有効、レコードが存在しない = フラグ無効
- 同一対象（device_token または user_id）に同一フラグ名を重複登録できない（UNIQUE 制約）
- `device_token` と `user_id` のうち少なくとも一方が設定されている必要がある（CHECK 制約）
- 開発者は DB を直接操作（psql やマイグレーション）してフラグを管理できる

### device_token Cookie

- システムはリバースプロキシミドルウェアで、`device_token` Cookie が存在しない場合に自動生成してレスポンスにセットする
- トークンは 24 バイトのランダムデータを Base64 URL-safe エンコードした 32 文字の文字列
- Cookie 属性: `HttpOnly: true`, `Secure: true`（本番環境）, `SameSite: Lax`, `MaxAge: 10年`

### ルーティング

- システムはリクエストの URL パターンとフラグ状態に基づいて、Go または Rails にルーティングする
- フラグ判定は `featureFlaggedPatterns` に登録された URL パターンにマッチするリクエストのみで実行される
- フラグ判定では `device_token` によるチェックと `user_id` によるチェックを OR 条件で行う（いずれかが有効であれば Go 版で処理する）
- `user_id` はセッション Cookie から解決する。ログインユーザーの場合、全デバイスでフラグが有効になる
- フラグが無効な場合、Cookie が存在しない場合、またはフラグ判定でエラーが発生した場合は、Rails 版にフォールバックする

### 非機能仕様

- **パフォーマンス**: フラグ判定はフィーチャーフラグ付きパスにマッチするリクエストのみで実行される。非対象パスでは追加の DB 問い合わせは発生しない
- **信頼性**: フラグ判定でエラーが発生した場合は Rails 版にフォールバックする（Go 版が表示されないほうがサービス断よりも安全）
- **セキュリティ**: `device_token` Cookie は `crypto/rand` による安全なランダム値を使用し、`HttpOnly` + `Secure` + `SameSite=Lax` で設定する

## 設計

<!--
ガイドライン:
- 現在の技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
- 該当がない場合も、セクション自体は残しておく（後から追加しやすくするため）
-->

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

- `device_token` と `user_id` はどちらも nullable だが、CHECK 制約で少なくとも一方が NOT NULL であることを保証する
- `(device_token, name)` と `(user_id, name)` の 2 つのユニーク制約で、同一対象に同一フラグが重複登録されることを防ぐ
- PostgreSQL では UNIQUE 制約の NULL 値は重複として扱われないため、`device_token` が NULL のレコードは `UNIQUE(device_token, name)` 制約に影響しない
- 主キーは `BIGSERIAL` を使用（Annict の既存マイグレーションパターンに準拠）

### ルーティングの流れ

```
リクエスト到着
  |
[リバースプロキシミドルウェア]
  |-- APIサブドメイン? --> Rails版にプロキシ
  |-- device_token Cookie なし? --> 自動生成してレスポンスにセット
  |-- 常にGoのパス（goHandledPaths）? --> Go側ミドルウェアチェーン --> ハンドラー
  |-- フィーチャーフラグ付きパス?
  |     |-- Yes + device_tokenのフラグ有効 --> Go側ミドルウェアチェーン --> ハンドラー
  |     |-- Yes + user_idのフラグ有効 --> Go側ミドルウェアチェーン --> ハンドラー
  |     +-- Yes + フラグ無効/トークンなし/エラー --> Railsにプロキシ
  +-- その他 --> Railsにプロキシ
```

### コード設計

#### ファイル構成

```
go/
├── db/
│   └── migrations/
│       └── 20260322083140_create_feature_flags.sql
├── internal/
│   ├── model/
│   │   ├── id.go                       # FeatureFlagID, FeatureFlagName 型
│   │   └── feature_flag.go             # FeatureFlag モデル
│   ├── query/
│   │   └── queries/
│   │       └── feature_flags.sql       # sqlc クエリ定義
│   ├── repository/
│   │   ├── feature_flag.go             # FeatureFlagRepository
│   │   └── feature_flag_test.go        # リポジトリテスト
│   ├── middleware/
│   │   ├── reverse_proxy.go            # フラグ判定統合済みリバースプロキシ
│   │   └── reverse_proxy_test.go       # ミドルウェアテスト
│   ├── session/
│   │   └── token.go                    # GenerateSecureToken 関数
│   └── testutil/
│       └── feature_flag_builder.go     # テスト用ビルダー
└── cmd/
    └── server/
        └── main.go                     # 依存性注入
```

#### ドメインモデル

```go
// internal/model/id.go
type FeatureFlagID int64
type FeatureFlagName string

// internal/model/feature_flag.go
type FeatureFlag struct {
    ID          FeatureFlagID
    DeviceToken *string
    UserID      *int64
    Name        FeatureFlagName
    CreatedAt   time.Time
}
```

フラグ名の定数は `feature_flag.go` に定義する。Go 版への移行で使用するフラグには `go_` プレフィックスを付ける（例: `FeatureFlagGoPageEdit FeatureFlagName = "go_page_edit"`）。現時点では具体的なフラグ名定数は未定義。

#### リポジトリ

```go
// internal/repository/feature_flag.go
type FeatureFlagRepository struct { ... }

func (r *FeatureFlagRepository) IsEnabledByDeviceOrUser(
    ctx context.Context, deviceToken string, userID int64, name model.FeatureFlagName,
) (bool, error)

func (r *FeatureFlagRepository) IsEnabled(
    ctx context.Context, userID int64, name model.FeatureFlagName,
) (bool, error)
```

- `IsEnabledByDeviceOrUser`: `device_token` または `user_id` のいずれかでフラグが有効かを判定する。空文字列の `deviceToken` や 0 の `userID` は何にもマッチしない
- `IsEnabled`: `user_id` のみでフラグを判定する便利メソッド

#### リバースプロキシミドルウェア

```go
// internal/middleware/reverse_proxy.go

// featureFlagChecker はフィーチャーフラグの有効判定を行うインターフェース
type featureFlagChecker interface {
    IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID int64, name model.FeatureFlagName) (bool, error)
}

// featureFlaggedPattern はフィーチャーフラグで制御するURLパターンを定義
type featureFlaggedPattern struct {
    pattern *regexp.Regexp
    flag    model.FeatureFlagName
}
```

- `ReverseProxyMiddleware` 構造体は `featureFlagRepo`（`featureFlagChecker` インターフェース）と `sessionMgr`（`*session.Manager`）を保持する
- `featureFlagRepo` が `nil` の場合、フラグ判定は常に `false` を返す（テスト時やフラグ機能不要時）
- セッションから `user_id` の解決には `*session.Manager` を直接使用する

### フラグ管理の運用例

```sql
-- 未ログインユーザーの特定デバイスに対してフラグを設定
INSERT INTO feature_flags (device_token, name) VALUES ('cookie値', 'go_page_edit');

-- ログインユーザーの全デバイスに対してフラグを設定
INSERT INTO feature_flags (user_id, name) VALUES (123, 'go_page_edit');
```

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録として活用する
- 後から実装された場合は、該当項目を削除する
- 該当がない場合も、セクション自体は残しておく（後から追加しやすくするため）
-->

### セッション JOIN による 1 クエリでのフラグ判定（Wikino 方式）

Wikino では `user_session_tokens` テーブルを `IsEnabledForDevice` のサブクエリで JOIN し、1 クエリで `device_token` と `user_id` の両方をチェックしている。Annict でも同様に sessions テーブルの JSONB データからサブクエリで `user_id` を抽出する方式を検討した。

**不採用の理由**: Annict のセッションテーブルは JSONB の `data` カラムに warden 形式（`{"warden.user.user.key": [[user_id], "salt"]}`）で `user_id` を格納しており、SQL 内での JSONB パス抽出は Wikino の単純な `token -> user_id` マッピングに比べて脆弱性が高い。Go コードでセッションから `user_id` を解決してからシンプルな SQL で判定するほうが、保守性と可読性が優れる。パフォーマンスへの影響はフィーチャーフラグ付きパスのみで発生するため、実用上問題ない。

### UUID 主キーの使用

Wikino では `generate_ulid()` による UUID を主キーに使用している。

**不採用の理由**: Annict の Go 版マイグレーションでは `BIGSERIAL` を主キーに使用する規約が確立されており（`password_reset_tokens`, `sign_up_codes` 等）、UUID 生成関数（`generate_ulid()`）はスキーマに存在しない。既存パターンとの一貫性を優先する。

### sessionUserResolver インターフェースの導入

セッションから `user_id` を解決するための専用インターフェースを定義し、ミドルウェアの依存を抽象化する方式を検討した。

**不採用の理由**: `session.Manager` は `GetSessionID(r)` と `GetSession(ctx, sessionID)` メソッドを持ち、ミドルウェア（Presentation 層）からの依存として適切である。現時点で `session.Manager` の実装を差し替える必要がないため、専用インターフェースの導入は過剰な抽象化と判断した。

### 管理 UI

フラグの管理 UI を実装する方式を検討した。

**不採用の理由**: 現時点ではフラグの変更頻度が低く（開発者が手動で設定する程度）、psql やマイグレーションで十分に管理できる。パフォーマンスが問題になった場合やフラグの数が増えた場合に追加を検討する。

### インメモリキャッシュ

フラグ判定結果をインメモリキャッシュする方式を検討した。

**不採用の理由**: フラグ判定はフィーチャーフラグ付きパスのみで実行されるため、単純な DB クエリで十分な性能が得られる。パフォーマンスが問題になった場合に追加する。

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Wikino フィーチャーフラグ仕様書](/wikino/docs/specs/feature-flag/overview.md) - 設計のベースとなった仕様
- [作業計画書](/workspace/docs/plans/1_doing/feature-flag.md) - 実装時の作業計画書
