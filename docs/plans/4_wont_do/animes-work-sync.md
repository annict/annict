# 作品に紐付く animes レコードを作成する 設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 実装ガイドラインの参照

<!--
**重要**: 設計書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
設計書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン（**ファイル名は標準の8種類のみ**）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templテンプレートガイド

### Rails版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - 全体的なコーディング規約
- [@rails/docs/architecture-guide.md](/workspace/rails/docs/architecture-guide.md) - アーキテクチャガイド
- [@rails/docs/testing-guide.md](/workspace/rails/docs/testing-guide.md) - テスト戦略ガイド
- [@rails/docs/security-guide.md](/workspace/rails/docs/security-guide.md) - セキュリティガイドライン

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

作品（Work）に紐付く `animes` レコードを作成・同期する機能を実装する。これは [animes テーブル導入設計書](../1_doing/animes-table.md) の一部として、作品データと `animes` テーブルの連携を実現する。

**目的**:

- 既存の作品データを `animes` テーブルに移行する
- Annict DB での作品 CRUD 時に `animes` レコードも同期作成・更新する
- Go 版への段階的移行を進める

**背景**:

- `animes` テーブルと `works` テーブルへの `anime_id` カラム追加は完了済み
- 作品作成・更新時に `animes` レコードも同時に管理する必要がある
- 既存の作品データを `animes` テーブルに移行するスクリプトが必要

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

- システムは既存の `works` データから `animes` レコードを作成できる
- 管理者は Annict DB で作品を作成すると、対応する `animes` レコードも自動作成される
- 管理者は Annict DB で作品を編集すると、対応する `animes` レコードも自動更新される
- 管理者は Annict DB で作品を非公開にすると、対応する `animes` レコードも非公開になる
- 管理者は Annict DB で作品を削除すると、対応する `animes` レコードもソフト削除される

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

- **データ整合性**: 移行中も既存機能が正常に動作すること
- **段階的移行**: Rails 版と Go 版が共存する期間中、両方から `animes` テーブルを参照できること

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

詳細な設計は [animes テーブル導入設計書](../1_doing/animes-table.md) を参照してください。

### データ移行

既存の `works` レコードから `animes` レコードを作成するスクリプトを実装する。

```sql
-- works → animes の共通カラムコピー
INSERT INTO animes (id, parent_id, title, title_kana, title_alter, ratings_count, satisfaction_rate, score, deleted_at, hidden_at, created_at, updated_at)
SELECT
    w.id,
    NULL,  -- parent_id は NULL（作品は親を持たない）
    w.title,
    w.title_kana,
    NULLIF(w.title_alter, ''),
    w.ratings_count,
    w.satisfaction_rate,
    w.score,
    w.deleted_at,
    w.unpublished_at,
    w.created_at,
    w.updated_at
FROM works w
WHERE w.anime_id IS NULL;

-- works.anime_id の設定
UPDATE works SET anime_id = id WHERE anime_id IS NULL;
```

### 作品 CRUD 時の animes 同期

Go 版で作品を作成・更新・削除する際に、対応する `animes` レコードも同期する。

```go
// 作品作成時
func (u *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*Work, error) {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        // 1. animes レコードを作成
        anime, err := u.animeRepo.Create(ctx, tx, &Anime{
            Title:     input.Title,
            TitleKana: input.TitleKana,
            // ...
        })
        if err != nil {
            return err
        }

        // 2. works レコードを作成（anime_id を設定）
        work, err := u.workRepo.Create(ctx, tx, &Work{
            AnimeID:    anime.ID,
            SeasonYear: input.SeasonYear,
            Media:      input.Media,
            // ...
        })
        if err != nil {
            return err
        }

        return work, nil
    })
}
```

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
-->

### フェーズ 1: 既存データの移行

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [ ] **1-1**: [Go] 既存 works データの animes テーブルへの移行スクリプト作成
  - works → animes の共通カラムコピー
  - works.anime_id の設定
  - ratings_count, satisfaction_rate の集計・設定（既存の works から移行）
  - score は別設計書で定義する計算式に基づき設定
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 70 行 + テスト 30 行）

### フェーズ 2: Annict DB - 作品作成機能の実装

- [ ] **2-1**: [Go] Anime リポジトリの実装
  - sqlc クエリの定義（作成、取得）
  - リポジトリ層の実装
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **2-2**: [Go] Work リポジトリの実装
  - sqlc クエリの定義（作成、取得、タイトル重複チェック）
  - リポジトリ層の実装
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **2-3**: [Go] Work 作成ユースケースの実装
  - Work 作成ユースケースの実装
    - animes レコードの同時作成
  - バリデーション: title 必須・ユニーク、media 必須、URL 形式チェック
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **2-4**: [Go] Work 作成ハンドラーの実装
  - Work 作成ハンドラーの実装（`/db/works/new` GET, `/db/works` POST）
  - テンプレートの実装
  - 認可: committer ロールのチェック
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）
  - **参考**: Rails 版 `app/controllers/db/works_controller.rb`

- [ ] **2-5**: [Rails] Work 作成機能の削除
  - コントローラーの削除（`db/works#new`, `db/works#create`）
  - ビューの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 3 ファイル
  - **想定行数**: 約 100 行（削除）

### フェーズ 3: Annict DB - 作品編集機能の実装

- [ ] **3-1**: [Go] Work 更新ユースケースの実装
  - Work 更新ユースケースの実装（animes 同期更新含む）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **3-2**: [Go] Work 編集ハンドラーの実装
  - Work 編集ハンドラーの実装（`/db/works/:id/edit` GET, `/db/works/:id` PATCH）
  - テンプレートの実装
  - 認可: committer ロールのチェック
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）
  - **参考**: Rails 版 `app/controllers/db/works_controller.rb`

- [ ] **3-3**: [Rails] Work 編集機能の削除
  - コントローラーの削除（`db/works#edit`, `db/works#update`）
  - ビューの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 3 ファイル
  - **想定行数**: 約 100 行（削除）

### フェーズ 4: Annict DB - 作品非公開機能の実装

- [ ] **4-1**: [Go] Work 非公開ユースケースの実装
  - Work 非公開ユースケースの実装（animes.hidden_at 同期更新含む）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）
  - **参考**: Rails 版 `app/models/concerns/unpublishable.rb`

- [ ] **4-2**: [Go] Work 非公開ハンドラーの実装
  - Work 非公開ハンドラーの実装（`/db/works/:id/hide` POST）
  - 認可: admin ロールのチェック
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [ ] **4-3**: [Rails] Work 非公開機能の削除
  - コントローラーの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 2 ファイル
  - **想定行数**: 約 50 行（削除）

### フェーズ 5: Annict DB - 作品削除機能の実装

- [ ] **5-1**: [Go] Work 削除ユースケースの実装
  - Work ソフト削除ユースケースの実装（animes.deleted_at 同期更新含む）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **5-2**: [Go] Work 削除ハンドラーの実装
  - Work 削除ハンドラーの実装（`/db/works/:id` DELETE）
  - 認可: admin ロールのチェック
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [ ] **5-3**: [Rails] Work 削除機能の削除
  - コントローラーの削除（`db/works#destroy`）
  - ルーティングの削除
  - **想定ファイル数**: 約 2 ファイル
  - **想定行数**: 約 50 行（削除）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **DbActivity による活動履歴作成**: ポリモーフィック対応は別途検討。Work/Episode を保存する形で対応
- **エピソードに紐付く animes レコードの作成**: 別設計書で対応

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [animes テーブル導入設計書](../1_doing/animes-table.md)
- [現在の works テーブル定義](/workspace/go/db/schema.sql:3181)
