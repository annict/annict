# Annict Go 実装プロジェクト設計書

## 1. プロジェクト概要

### 目的

既存の Rails 実装の Annict を Go で再実装する。データベースは既存のものをそのまま使用し、段階的に移行する。

### 基本方針

- 既存の PostgreSQL データベースをそのまま使用
- データマイグレーションは行わない（Rails 側で管理）
- Go 初心者にもわかりやすい実装を心がける
- 段階的な移行を前提とした設計

## 2. 既存システムの構成

### 技術スタック

- **フレームワーク**: Rails 7.0.8
- **Ruby**: 3.3.5
- **データベース**: PostgreSQL
- **GraphQL**: graphql-ruby
- **認証**: Devise, Doorkeeper (OAuth2)
- **フロントエンド**: jsbundling-rails, cssbundling-rails
- **バックグラウンドジョブ**: delayed_job_active_record
- **キャッシュ**: Redis (hiredis) - キャッシュストアとしてのみ使用
- **セッションストア**: activerecord-session_store (PostgreSQL を使用)

### 主要機能

- アニメ作品管理
- エピソード管理
- 視聴記録（レコード）
- ユーザー認証・認可
- GraphQL API
- OAuth2 プロバイダー
- アクティビティフィード
- コレクション機能

## 3. データベース構造

### 主要テーブル（既存の structure.sql から確認済み）

- users: ユーザー情報
- works: 作品情報
- episodes: エピソード情報
- episode_records: 視聴記録
- activities: アクティビティフィード
- collections: コレクション
- characters: キャラクター情報
- casts: 声優情報
- channels: 放送局情報

### 使用する拡張機能

- citext: 大文字小文字を区別しない文字列型
- pg_stat_statements: クエリ統計情報

## 4. Go 実装で必要な技術選定

### 技術選定基準

本プロジェクトでは以下の基準で技術を選定します：

1. **Go の思想に近い** - 標準ライブラリとの親和性を重視
2. **シンプルで理解しやすい** - 過度な抽象化を避ける
3. **必要十分な機能** - 多機能より単機能で組み合わせ可能
4. **コミュニティの支持** - メンテナンスが活発で信頼性が高い
5. **学習価値** - Go の標準的な書き方が身につく

### 推奨ライブラリ・フレームワーク

#### Web フレームワーク

- **Chi**: 標準ライブラリに近い設計（採用）
  - Go の標準的な書き方を学べる
  - 将来の保守性が高い
  - 必要な機能だけを追加できる

#### データベース

- **sqlc**: SQL から Go コードを生成（採用）
  - コンパイル時の型安全性が高い
  - 既存 DB スキーマから正確な型を生成
  - SQL ファーストで理解しやすい
  - 生成コードは標準的な Go コード
  - 比較的活発にメンテナンスされている

#### 認証・セッション管理

- **セッション管理**: PostgreSQL の sessions テーブルを直接読み取る実装（Rails 互換性のため）
  - gorilla/sessions は使用せず、Rails の activerecord-session_store との互換性を保つ
- **golang-jwt**: JWT 処理（必要に応じて）
- **OAuth2**: golang.org/x/oauth2

#### テンプレートエンジン

- **templ**: 型安全なテンプレート（採用）
  - Go の型システムを活用した型安全性
  - コンパイル時のエラーチェック
  - IDE サポートが充実
  - XSS 対策が組み込み
  - Go コードとして生成されるためパフォーマンスが良い

#### フロントエンド

- **esbuild**: 高速な JavaScript バンドラー（採用）
  - Go で実装された超高速ビルド
  - TypeScript のトランスパイル不要（JavaScript のみ）
  - シンプルな設定
  - 本番ビルドの最適化（ミニファイ、Tree Shaking）
- **Datastar**: リアクティブな HTML 拡張（採用）
  - サーバーサイドレンダリングと相性良好
  - シンプルな JavaScript で動的 UI を実現
  - templ との統合が容易
  - SPA の複雑さを避ける
- **Tailwind CSS**: ユーティリティファースト CSS（採用）
  - 高速な開発
  - ビルドで未使用クラスを削除
  - レスポンシブ対応が簡単
- **Basecoat**: Tailwind CSS ベースの UI コンポーネント（採用）
  - アクセシブルな実装
  - カスタマイズ可能
  - モダンなデザイン

#### バックグラウンドジョブ

- **riverqueue/river**: PostgreSQL ベースのジョブキュー（推奨）
  - PostgreSQL ネイティブ（追加インフラ不要）
  - トランザクショナルなジョブエンキュー
  - Go のジェネリクスを活用した型安全な API
  - Web UI（River UI）による監視機能
  - リトライ、スケジューリング、キャンセル機能
  - メール送信などの非同期処理に使用

#### メール送信

- **resend-go/v2**: Resend API を使用したメール送信ライブラリ（推奨）
  - シンプルで使いやすい API
  - SMTP サーバーの管理が不要（API 経由）
  - 高い到達率と信頼性
  - 開発者フレンドリー（無料プランで月 100 通まで）
  - HTML メール、テキストメール、マルチパート対応
  - 日本語メール対応

#### Rate Limiting

- **Redis + go-redis/v9**: Rate Limiting の実装（推奨）
  - Rails が既に使用している Redis を活用（追加インフラ不要）
  - 複数プロセス間で共有可能（Dokku での web プロセススケーリングに対応）
  - 高速なインメモリストア
  - TTL による自動クリーンアップ
  - Lua スクリプトによるアトミック操作
  - 将来の Web API でも使用可能

#### その他

- **環境変数**: 標準の`os.Getenv`（推奨）
  - viper より シンプルで理解しやすい
  - 12 Factor App の原則に従う
- **log/slog**: Go 1.21+標準構造化ログ（推奨）
  - 標準ライブラリで十分な機能
  - 構造化ログに対応
- **testing**: 標準テストパッケージ（推奨）
  - testify は便利だが、標準で十分
  - テーブル駆動テストを活用
- **Redis**: Rate Limiting に使用（Rails が既に使用中）
  - キャッシュ機能は当面使用しない（必要に応じて将来的に検討）
  - Rate Limiting と将来の Web API で活用

## 5. 既存 Rails コードの参照方法

### Rails ソースコードの場所

モノレポ化により、既存の Rails アプリケーションのソースコードは `/workspace/rails/` に配置されています。
Go への移行時は、以下のディレクトリ構造を参照してください：

```bash
/workspace/rails/
├── app/
│   ├── controllers/      # コントローラー（HTTPリクエスト処理）
│   ├── models/           # モデル（ビジネスロジック）
│   ├── views/            # ビューテンプレート
│   ├── services/         # サービスクラス
│   ├── queries/          # クエリオブジェクト
│   └── graphql/          # GraphQL関連
├── config/
│   ├── routes.rb         # ルーティング定義
│   └── database.yml      # DB設定
└── db/
    ├── structure.sql     # DBスキーマ
    └── migrate/          # マイグレーション
```

### 参照例

```bash
# コントローラーの確認
cat /workspace/rails/app/controllers/application_controller.rb
cat /workspace/rails/app/controllers/works_controller.rb

# モデルの確認
cat /workspace/rails/app/models/work.rb
cat /workspace/rails/app/models/episode.rb

# ビューテンプレートの確認
cat /workspace/rails/app/views/works/show.html.erb

# ルーティングの確認
cat /workspace/rails/config/routes.rb
```

## 6. 移行戦略

このプロジェクトでは、**認証関連機能（ログイン・パスワードリセット・サインアップ）の移行のみ**を行います。
他の機能は将来的な検討課題とし、このプロジェクトのスコープには含めません。

### フェーズ 1: 基盤構築

1. プロジェクト構造の設定
2. データベース接続
3. セッション共有メカニズムの実装
4. リバースプロキシの実装（Go 版で未実装の機能を Rails 版にプロキシ）

### フェーズ 2: 認証機能の移行

1. ログイン機能
   - メールアドレス・ユーザー名でのログイン
   - パスワード照合（bcrypt）
   - セッション作成
2. パスワードリセット機能
   - パスワードリセット申請
   - トークン生成・管理
   - メール送信（Resend API）
   - パスワード更新
3. サインアップ機能
   - ユーザー登録フォーム
   - バリデーション
   - アカウント作成
   - 確認メール送信（将来的な検討）

### フェーズ 3: 本番デプロイ

1. 本番環境へのデプロイ検証
2. パフォーマンステスト
3. セキュリティチェック
4. 段階的な本番移行

## 7. プロジェクト構造

モノレポ化により、Go 版のコードは `/workspace/go/` に配置されています：

```
/workspace/
├── go/                      # Go版の実装
│   ├── cmd/
│   │   └── server/
│   │       └── main.go      # エントリーポイント
│   ├── internal/
│   │   ├── config/          # 設定管理
│   │   ├── handler/         # HTTPハンドラー（Railsのコントローラー相当）
│   │   ├── middleware/      # ミドルウェア
│   │   │   ├── auth.go      # 認証ミドルウェア
│   │   │   ├── csrf.go      # CSRF保護
│   │   │   ├── reverse_proxy.go  # Railsへのリバースプロキシ
│   │   │   └── method_override.go
│   │   ├── repository/      # データアクセス層（sqlc生成コード）
│   │   │   ├── queries/     # SQLクエリファイル
│   │   │   └── sqlc/        # sqlc生成コード（自動生成）
│   │   ├── usecase/         # ビジネスロジック層（フラット配置）
│   │   │   ├── create_session.go
│   │   │   ├── create_password_reset_token.go
│   │   │   └── update_password_reset.go
│   │   ├── viewmodel/       # プレゼンテーション層のデータ変換
│   │   ├── i18n/            # 国際化
│   │   │   └── locales/     # 翻訳ファイル（ja.toml, en.toml）
│   │   └── testutil/        # テストヘルパー
│   ├── views/               # templテンプレート
│   │   ├── layouts/
│   │   ├── components/
│   │   └── pages/
│   ├── static/              # 静的ファイル
│   │   ├── css/             # Tailwind CSSビルド結果
│   │   └── js/              # esbuildビルド結果
│   ├── db/                  # データベース
│   │   ├── migrations/      # DBマイグレーション
│   │   └── schema.sql       # DBスキーマ
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
├── rails/                   # Rails版の実装
│   ├── app/
│   ├── config/
│   └── ...
└── .github/                 # 共通のCI/CD設定

```

## 8. 実装上の注意点

### データベース関連

- タイムスタンプは `timestamp with time zone` と `timestamp without time zone` が混在
- citext カラムは大文字小文字を区別しない
- 既存のインデックスを活用する

### Rails 互換性

- ActiveRecord の `created_at`, `updated_at` の扱い
- Enum 値の扱い（Rails の enum との互換性）
- 多対多関連の中間テーブルの命名規則

### パフォーマンス

- N+1 問題の回避（DataLoader パターンの活用）
- 適切なキャッシュ戦略
- コネクションプーリングの設定

## 9. テスト戦略

### テストの種類

- ユニットテスト: 各パッケージの機能テスト
- 統合テスト: API エンドポイントのテスト
- データベーステスト: 実際の DB を使用したテスト

### テストデータ

- 既存の Rails の fixture や factory を参考に作成
- テスト用 DB は別途用意（annict_test）

## 10. セッション共有の実装方針

### Rails セッションの互換性確保

1. **セッションストア**: PostgreSQL（Rails 側が activerecord-session_store で使用する sessions テーブル）
2. **セッションキー**: `_annict_session_v201904`（Rails 側と同一）
3. **Cookie 設定**:
   - Domain: `.annict.com`（本番）、`.example.dev`（開発）
   - Secure: true（Cloudflare Tunnel 経由で HTTPS）
   - HttpOnly: true
   - SameSite: Lax

### 開発環境でのドメイン共有

リバースプロキシの実装により、以下の構成になりました：

- **メインドメイン**: `example.dev` → Go 版（リバースプロキシ）
  - Go 版で実装済みの機能は Go 版で処理
  - 未実装の機能は自動的に Rails 版にプロキシ
- **Rails 版**: `localhost:3000`（Cloudflare Tunnel は使用せず、プロキシ経由でアクセス）
- **Cookie 共有**: `.example.dev`で Cookie を共有
- **利点**:
  - SEO 対策（同一ドメインで運用）
  - 段階的な機能移行が可能
  - ユーザーは常に`example.dev`でアクセス

### セッションデータの形式

- Rails 側のセッションデータ形式を解析
  - PostgreSQL の sessions テーブルに格納されているデータを読み取る
  - セッションデータは Base64 エンコードされた Marshal 形式
- 必要な情報のみを読み取り（user_id、CSRF token など）
- Go 側での書き込みは最小限に留める

### 実装手順

1. Rails のセッションクッキー（`_annict_session_v201904`）を読み取る
2. セッション ID を使って PostgreSQL の sessions テーブルからセッションデータを取得
3. data カラムの Base64 + Marshal 形式のデータをデコード
4. ユーザー認証状態の確認

## 11. 今後の検討事項

### 技術選定で決定が必要な項目

1. ORM を使うか、生 SQL を使うか
2. テンプレートエンジンの選択（html/template vs templ）
3. ログ出力形式（Rails 互換にするか）
4. エラーハンドリング戦略
5. デプロイ方法（コンテナ化など）

### 移行時の課題

1. Rails の Marshal 形式セッションデータの扱い（PostgreSQL の sessions テーブル内）
2. CSRF 対策の互換性
3. バックグラウンドジョブの共存方法
4. 段階的なトラフィック切り替え方法（リバースプロキシ設定）

## 12. コード例

### 3 層アーキテクチャの構成

このプロジェクトでは、**3 層アーキテクチャ**を採用しています：

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層（プレゼンテーション層）                    │
│ - Handler                                              │
│ - ViewModel                                            │
│ - Template                                             │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Application層（アプリケーション層）                        │
│ - UseCase（ビジネスフロー、トランザクション管理）           │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc)                                         │
│ - Repository                                           │
│ - Model                                                │
│ （同じ層なので相互に依存できる）                          │
└─────────────────────────────────────────────────────────┘
```

### Domain/Infrastructure 層を統合する理由

- **実用的**: データベース変更（PostgreSQL → MySQL など）はほぼ起こらない
- **シンプル**: 層をまたぐ変換コストを削減し、シンプルさを保つ
- **Go らしい**: Go のプラグマティックな哲学に合致
- **YAGNI 原則**: 必要になってから層を分ければ良い

### Model と Repository の 1:1 関係

各ドメインエンティティに対して対応する Repository を作成します：

- `model.Work` ↔ `repository.WorkRepository`
- `model.User` ↔ `repository.UserRepository`
- `model.Episode` ↔ `repository.EpisodeRepository`

### Model 層の例

#### internal/model/work.go

```go
package model

import "time"

// Work は作品のドメインエンティティです（ページに依存しない汎用的な構造）
// Domain/Infrastructure層に属し、Presentation層に依存しない
type Work struct {
    ID                 int64
    Title              string
    TitleEn            string
    TitleKana          *string
    RecommendedImageURL string
    ImageData          string  // work_imagesテーブルのimage_data (JSON)
    WatchersCount      int32
    SeasonYear         *int32
    SeasonName         *int32  // シーズン番号（0=冬、1=春、2=夏、3=秋）
    CreatedAt          time.Time
}

// WorkWithDetails は作品の詳細情報を含むデータ構造です
type WorkWithDetails struct {
    Work   Work
    Casts  []Cast
    Staffs []Staff
}
```

#### internal/model/cast.go

```go
package model

// Cast はキャスト情報を表します
type Cast struct {
    ID              int64
    WorkID          int64
    Name            string
    NameEn          string
    CharacterName   string
    CharacterNameEn string
    PersonName      string
    PersonNameEn    string
}
```

#### internal/model/staff.go

```go
package model

// Staff はスタッフ情報を表します
type Staff struct {
    ID          int64
    WorkID      int64
    Name        string
    NameEn      string
    Role        string
    RoleOther   string
    RoleOtherEn string
}
```

**重要**: Model 層は Presentation 層に依存しない（`image.Helper`などに依存しない）

### Repository 層の例

#### internal/repository/work.go

```go
package repository

import (
    "context"
    "github.com/annict/annict/internal/model"
    "github.com/annict/annict/internal/query"
)

// WorkRepository はWork関連のデータアクセスを担当します
type WorkRepository struct {
    queries *query.Queries
}

// NewWorkRepository はWorkRepositoryを作成します
func NewWorkRepository(queries *query.Queries) *WorkRepository {
    return &WorkRepository{queries: queries}
}

// GetPopularWorksWithDetails は人気作品をキャスト・スタッフ情報と共に取得します
func (r *WorkRepository) GetPopularWorksWithDetails(ctx context.Context) ([]model.WorkWithDetails, error) {
    // 1. クエリ実行
    worksRows, err := r.queries.GetPopularWorks(ctx)
    if err != nil {
        return nil, err
    }

    if len(worksRows) == 0 {
        return []model.WorkWithDetails{}, nil
    }

    // 2. query.GetPopularWorksRow → model.Work に変換
    works := make([]model.Work, len(worksRows))
    workIDs := make([]int64, len(worksRows))
    for i, row := range worksRows {
        works[i] = r.workFromPopularRow(row)
        workIDs[i] = row.ID
    }

    // 3. キャストとスタッフを取得
    castsRows, err := r.queries.GetCastsByWorkIDs(ctx, workIDs)
    if err != nil {
        return nil, err
    }

    staffsRows, err := r.queries.GetStaffsByWorkIDs(ctx, workIDs)
    if err != nil {
        return nil, err
    }

    // 4. query結果をmodelに変換
    casts := r.castsFromRows(castsRows)
    staffs := r.staffsFromRows(staffsRows)

    // 5. 組み合わせる
    return r.combineWorkData(works, casts, staffs), nil
}

// workFromPopularRow は query.GetPopularWorksRow を model.Work に変換します
func (r *WorkRepository) workFromPopularRow(row query.GetPopularWorksRow) model.Work {
    work := model.Work{
        ID:                 row.ID,
        Title:              row.Title,
        TitleEn:            row.TitleEn,
        RecommendedImageURL: row.RecommendedImageUrl,
        ImageData:          row.ImageData.String,
        WatchersCount:      row.WatchersCount,
        CreatedAt:          row.CreatedAt,
    }

    if row.SeasonYear.Valid {
        work.SeasonYear = &row.SeasonYear.Int32
    }
    if row.SeasonName.Valid {
        work.SeasonName = &row.SeasonName.Int32
    }

    return work
}

// castsFromRows は query結果を model.Cast に変換します
func (r *WorkRepository) castsFromRows(rows []query.GetCastsByWorkIDsRow) []model.Cast {
    casts := make([]model.Cast, len(rows))
    for i, row := range rows {
        casts[i] = model.Cast{
            ID:              row.ID,
            WorkID:          row.WorkID,
            Name:            row.Name,
            NameEn:          row.NameEn,
            CharacterName:   row.CharacterName.String,
            CharacterNameEn: row.CharacterNameEn.String,
            PersonName:      row.PersonName.String,
            PersonNameEn:    row.PersonNameEn.String,
        }
    }
    return casts
}

// staffsFromRows は query結果を model.Staff に変換します
func (r *WorkRepository) staffsFromRows(rows []query.GetStaffsByWorkIDsRow) []model.Staff {
    staffs := make([]model.Staff, len(rows))
    for i, row := range rows {
        staffs[i] = model.Staff{
            ID:          row.ID,
            WorkID:      row.WorkID,
            Name:        row.Name,
            NameEn:      row.NameEn,
            Role:        row.Role,
            RoleOther:   row.RoleOther.String,
            RoleOtherEn: row.RoleOtherEn,
        }
    }
    return staffs
}

// combineWorkData は作品データとキャスト・スタッフデータを組み合わせます
func (r *WorkRepository) combineWorkData(
    works []model.Work,
    casts []model.Cast,
    staffs []model.Staff,
) []model.WorkWithDetails {
    // キャストとスタッフをwork_idでマッピング
    castsMap := make(map[int64][]model.Cast)
    for _, cast := range casts {
        castsMap[cast.WorkID] = append(castsMap[cast.WorkID], cast)
    }

    staffsMap := make(map[int64][]model.Staff)
    for _, staff := range staffs {
        staffsMap[staff.WorkID] = append(staffsMap[staff.WorkID], staff)
    }

    // WorkWithDetailsのスライスを作成
    result := make([]model.WorkWithDetails, len(works))
    for i, work := range works {
        result[i] = model.WorkWithDetails{
            Work:   work,
            Casts:  castsMap[work.ID],
            Staffs: staffsMap[work.ID],
        }
    }

    return result
}
```

**設計のポイント**：

- Query 結果を Model に変換する責務を持つ
- 複数のクエリを組み合わせて Model を構築
- **Model と Repository は同じ層**（Domain/Infrastructure 層）なので、相互に依存できる
- **Presentation 層（ViewModel）には依存しない**

### usecase 層の例

#### internal/usecase/episode_record/create.go

```go
package episode_record

import (
    "context"
    "github.com/annict/annict/internal/domain/episode"
    "github.com/annict/annict/internal/domain/user"
    "github.com/annict/annict/internal/repository"
)

type Create struct {
    db           *sql.DB
    userRepo     repository.UserRepository
    episodeRepo  repository.EpisodeRepository
    recordRepo   repository.EpisodeRecordRepository
    activityRepo repository.ActivityRepository
}

func (uc *Create) Execute(ctx context.Context, userID int64, episodeID int64, rating string) error {
    // トランザクション開始
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. ユーザーとエピソードの存在確認
    user, err := uc.userRepo.GetByID(ctx, tx, userID)
    if err != nil {
        return err
    }

    episode, err := uc.episodeRepo.GetByID(ctx, tx, episodeID)
    if err != nil {
        return err
    }

    // 2. 視聴記録を作成
    record := &EpisodeRecord{
        UserID:    userID,
        EpisodeID: episodeID,
        Rating:    rating,
    }

    if err := uc.recordRepo.Create(ctx, tx, record); err != nil {
        return err
    }

    // 3. アクティビティを作成
    activity := &Activity{
        UserID:     userID,
        Action:     "create_episode_record",
        RecordID:   record.ID,
    }

    if err := uc.activityRepo.Create(ctx, tx, activity); err != nil {
        return err
    }

    // コミット
    return tx.Commit()
}
```

### 設計方針

- **シンプルな構造体**: DB カラムと 1:1 対応
- **NULLable カラム**: ポインタ型（`*string`, `*int`）で表現
- **最小限のメソッド**: 必要なビジネスロジックのみ
- **db タグ**: sqlc との互換性を保つ
- **エラー処理**: Go 標準の error を返す
- **関連の扱い**: 必要最小限に留める（循環参照を避ける）
- **usecase 層**: アプリケーションロジックとトランザクション管理

### パッケージ構成の理由

ドメインモデルをエンティティごとにパッケージ分けしている理由：

1. **明確な名前空間**

   ```go
   // user.User, work.Work と明確に区別できる
   u := &user.User{}
   w := &work.Work{}
   ```

2. **層ごとの責務分離**

   ```
   internal/domain/user/
   └── user.go                           # User構造体とメソッド

   internal/repository/
   └── user.go                          # UserRepository interface（名前の衝突を避けるためRepository suffix維持）

   internal/usecase/user/
   ├── create.go                         # ユーザー作成（user.Create）
   └── update.go                         # ユーザー更新（user.Update）
   ```

   **usecase 層の責務**（Go の標準的な考え方）：

   - アプリケーション固有のビジネスルール
   - 複数のエンティティやサービスの調整
   - トランザクション境界の管理
   - 1 つの操作 = 1 つのユースケースファイル

   **usecase に含めるべきもの**：

   - データの作成・更新・削除
   - 複数 repository を跨ぐ参照
   - ビジネスロジックを含む処理

   **usecase に含めないもの**：

   - 単純な 1 テーブルの参照（handler から直接 repository 呼び出し）

3. **責務の分離**

   - 各パッケージが単一の責務を持つ
   - テストファイルも同じディレクトリに配置可能
   - 依存関係が明確になる

4. **スケーラビリティ**
   - エンティティが増えても管理しやすい
   - チーム開発時の作業分担が容易

## 13. リポジトリ統合（モノレポ化）

### 現在の構成（モノレポ化完了）

モノレポ化により、Go 版と Rails 版が同一リポジトリで管理されるようになりました：

```
/workspace/
├── go/                       # Go版の実装
│   ├── go.mod (module github.com/annict/annict/go)
│   ├── internal/
│   ├── cmd/
│   ├── views/
│   ├── static/
│   └── ...
├── rails/                    # Rails版の実装
│   ├── Gemfile
│   ├── app/
│   ├── config/
│   └── ...
├── .github/                  # 共通のCI/CD設定
├── CLAUDE.md                 # プロジェクト全体のガイド
└── README.md
```

### 将来の構成（Rails 完全置き換え後）

Rails 版の機能がすべて Go 版に移行された後は、以下の構成になる予定：

```
/workspace/
├── go.mod (module github.com/annict/annict)
├── internal/
├── cmd/
├── views/
├── static/
└── README.md
```

**移行完了後の方針**:

- Rails 版のディレクトリを削除
- Go 版をルートに移動
- モジュール名を`github.com/annict/annict`に変更

## 14. 次のステップ

このプロジェクトは現在、認証関連機能の実装を進めています：

- ✅ 基盤構築（データベース接続、セッション共有、リバースプロキシ）
- ✅ ログイン機能
- ✅ パスワードリセット機能
- 🚧 サインアップ機能（進行中）
- ⏳ 本番デプロイ準備

## 15. プロジェクト完了までのタスクリスト

- [x] [最小限の基盤構築](./infrastructure.md)
- [x] [最初のページ実装（人気アニメ）](./popular-works.md)
- [x] [認証とセッション](./authentication-session.md)
- [x] [テストインフラの改善](./testing-infrastructure.md)
- [x] [ログイン機能](./sign-in.md)
- [x] [パスワードリセット機能](../202510/password-reset.md)
- [x] [ビルドシステムの簡素化](./build-system-simplification.md)
- [x] [デザインの調整](../3_done/202510/design-improvement.md)
- [x] [モノレポ化](../3_done/202510/monorepo-migration.md)
- [x] [リバースプロキシの実装](../1_doing/reverse-proxy-to-rails.md)
- [x] [テストデータを充実させる](../1_doing/test-data-seeding.md)
- [x] [環境変数設定の見直し](../3_done/202511/fix-env.md)
- [x] [templ への移行](../3_done/202511/templ-migration.md)
- [x] [`ANNICT_ENV` を `APP_ENV` に統合する](../3_done/202511/env-var-integration.md)
- [x] [ハンドラーのリファクタリング](../3_done/202511/handler-refactoring.md)
- [x] [メールアドレスでログイン機能の実装](../3_done/202511/email-login.md)
- [x] [Cloudflare Turnstile によるボット対策](../3_done/202511/cloudflare-turnstile.md)
- [x] [サインアップページ](../3_done/202511/sign-up.md)
- [x] submit 時にボタンを無効化する
- [x] [login -> sign_in](../3_done/202511/rename-login-to-sign-in.md)
- [x] [Repository 層の導入とアーキテクチャリファクタリング](../3_done/202511/repository-layer-refactoring.md)
- [x] [セッションタイムアウト問題の修正](../3_done/202511/session-timeout-fix.md)
- [x] [session.Manager リファクタリング](../3_done/202511/session-manager-refactoring.md)
- [x] [golangci/golangci-lint を導入する](../3_done/202511/golangci-lint.md)
- [x] [シードデータの修正](../3_done/202511/seed-data-improvements.md)
- [x] `GET /error/502` って不要では？
- [x] 本番環境へのデプロイ検証
- [x] [CSRF トークンの Rails 互換性対応](../3_done/202511/csrf-token-rails-compatibility.md)
- [x] [セッション Cookie 設定処理の統一化](../3_done/202511/session-cookie-refactoring.md)
- [x] [ログ出力方法の統一](../3_done/202511/logging-unification.md)
- [x] ユーザー登録ページの修正
- [x] [Go 版にもメンテナンスモードを設ける](../3_done/202511/maintenance-mode.md)
- [x] [ログイン後リダイレクト機能](../3_done/202511/redirect-after-login.md)
- [x] [パスワードログインページの UX 改善](../3_done/202511/sign-in-password-improvement.md)
- [x] [Sentry 導入](../3_done/202511/sentry.md)
- [x] [Go 版 本番運用設定](../3_done/202511/go-production-settings.md)
- [x] [Resend メールクライアントのリファクタリング](../3_done/202511/resend-mail-client-refactoring.md)
- [x] [CONTRIBUTING.md 作成](../3_done/202511/contributing-md.md)
- [x] [新規登録フォーム 利用規約同意方式の変更](../3_done/202511/sign-up-terms-agreement.md)
- [x] [Rails 認証ページコードの削除](../3_done/202511/remove-rails-auth-pages.md)
- [x] 動作確認

## 今後やることリスト

タスクが大きくなりすぎるので、一旦見送ります。

### 1. ログアウト機能

- [ ] [ログアウト機能](../202510/sign-out.md)

### 2. 書き込み機能の移行

このプロジェクトのスコープ外ですが、将来的に実装する場合：

- [ ] 視聴記録の作成・更新
- [ ] レビュー投稿機能
- [ ] いいね機能
- [ ] フォロー機能
- [ ] [視聴ステータス変更機能](../202510/other-features.md#フェーズ-11-視聴ステータス変更機能書き込み機能の先行実装)
- [ ] [Web 機能実装（書き込み機能）](../202510/other-features.md#フェーズ-18-web-機能実装書き込み機能)

### 3. API 実装

このプロジェクトのスコープ外ですが、将来的に実装する場合：

- [ ] OpenAPI 仕様の定義
- [ ] REST API の実装
- [ ] 認証機能の実装（既存 DB の users テーブルを使用）
- [ ] 外部サービス向けの OAuth2 実装
- [ ] [Web API 実装](../202510/other-features.md#フェーズ-15-web-api-実装)
- [ ] [個人用アクセストークンの管理](../202510/other-features.md#フェーズ-13-個人用アクセストークンの管理)
- [ ] [API 実装（外部連携が必要になったら）](../202510/other-features.md#フェーズ-20-api-実装外部連携が必要になったら)
- [ ] [OAuth プロバイダー実装](../202510/other-features.md#フェーズ-14-oauth-プロバイダー実装)

### 4. その他の機能

- [ ] [二要素認証（2FA）の実装](../202510/other-features.md#フェーズ-8-二要素認証2faの実装)
- [ ] [フロントエンド改善と UX 向上](../202510/other-features.md#フェーズ-16-フロントエンド改善と-ux-向上)
- [ ] [プロフィール設定ページ](../202510/other-features.md#フェーズ-12-プロフィール設定ページファイルアップロード検証)
- [ ] [Web 機能実装（読み取り専用）](../202510/other-features.md#フェーズ-17-web-機能実装読み取り専用)
- [ ] [ライブラリ評価（優先度：低）](../202510/other-features.md#フェーズ-25-ライブラリ評価優先度低)
- [ ] [Rails 完全置き換えとリポジトリ統合](../202510/other-features.md#フェーズ-24-rails-完全置き換えとリポジトリ統合)
- [ ] [パフォーマンス最適化（問題が顕在化したら）](../202510/other-features.md#フェーズ-21-パフォーマンス最適化問題が顕在化したら)

## 17. 実装の進め方のポイント

### 動作確認を優先

- まず動くものを作ってから、徐々に改善する
- 完璧な設計を求めすぎない
- ブラウザで確認できる状態を維持する

### 必要になったら作る（YAGNI 原則）

- repository 層、usecase 層は必要になってから
- インターフェースも必要になってから定義
- 過度な抽象化は避ける

### Go らしいシンプルさを保つ

- 標準ライブラリを優先的に使用
- エラーハンドリングは明示的に
- 魔法のような処理は避ける

### 段階的な学習

- 最初は単純な SQL クエリから
- 徐々に複雑な処理を追加
- リファクタリングを恐れない

---

このドキュメントは開発の進捗に応じて更新していきます。
