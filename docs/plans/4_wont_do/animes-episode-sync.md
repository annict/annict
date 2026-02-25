# エピソードに紐付く animes レコードを作成する 設計書

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

エピソード（Episode）に紐付く `animes` レコードを作成・同期する機能を実装する。これは [animes テーブル導入設計書](../1_doing/animes-table.md) の一部として、エピソードデータと `animes` テーブルの連携を実現する。

**目的**:

- 既存のエピソードデータを `animes` テーブルに移行する
- Annict DB でのエピソード CRUD 時に `animes` レコードも同期作成・更新する
- Syobocal からのエピソード自動生成時にも `animes` レコードを作成する
- Go 版への段階的移行を進める

**背景**:

- `animes` テーブルと `episodes` テーブルへの `anime_id` カラム追加は完了済み
- 作品に紐付く `animes` レコードの作成は別設計書で対応済み
- エピソード作成・更新時に `animes` レコードも同時に管理する必要がある
- エピソードは `animes.parent_id` に親作品の `anime_id` を設定する

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

- システムは既存の `episodes` データから `animes` レコードを作成できる
- 管理者は Annict DB でエピソードを作成すると、対応する `animes` レコードも自動作成される（`parent_id` に親作品の `anime_id` を設定）
- 管理者は Annict DB でエピソードを編集すると、対応する `animes` レコードも自動更新される
- 管理者は Annict DB でエピソードを非公開にすると、対応する `animes` レコードも非公開になる
- 管理者は Annict DB でエピソードを削除すると、対応する `animes` レコードもソフト削除される
- システムは Syobocal からのエピソード自動生成時に `animes` レコードも作成する
- システムは階層を 2 段階に制限する（作品 → エピソードのみ、エピソードの子は不可）

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

### 前提条件

この設計書の実装を開始する前に、以下が完了している必要があります：

- [作品に紐付く animes レコードを作成する](./animes-work-sync.md) の フェーズ 1（既存 works データの移行）が完了していること

### データ移行

既存の `episodes` レコードから `animes` レコードを作成するスクリプトを実装する。

```sql
-- episodes → animes の共通カラムコピー
-- 注: episodes.id は animes.id と重複する可能性があるため、新しい ID を採番する
INSERT INTO animes (parent_id, title, title_kana, title_alter, ratings_count, satisfaction_rate, score, deleted_at, hidden_at, created_at, updated_at)
SELECT
    w.anime_id,  -- parent_id は親作品の anime_id
    e.title,
    '',  -- title_kana は episodes にないため空文字
    NULL,  -- title_alter は episodes にないため NULL
    e.ratings_count,
    e.satisfaction_rate,
    e.score,
    e.deleted_at,
    NULL,  -- hidden_at は episodes にないため NULL
    e.created_at,
    e.updated_at
FROM episodes e
JOIN works w ON w.id = e.work_id
WHERE e.anime_id IS NULL;

-- episodes.anime_id の設定
UPDATE episodes e
SET anime_id = a.id
FROM animes a
WHERE a.title = e.title
  AND a.parent_id = (SELECT anime_id FROM works WHERE id = e.work_id)
  AND e.anime_id IS NULL;

-- episodes.prev_anime_id の設定
UPDATE episodes e
SET prev_anime_id = prev_e.anime_id
FROM episodes prev_e
WHERE e.prev_episode_id = prev_e.id
  AND e.prev_anime_id IS NULL;
```

### 階層制限の実装

```go
// anime を作成・更新する際に、parent_id のバリデーションを行う
func (r *AnimeRepository) validateParentID(ctx context.Context, parentID *int64) error {
    if parentID == nil {
        return nil
    }

    parent, err := r.FindByID(ctx, *parentID)
    if err != nil {
        return err
    }

    // 親がエピソード（parent_id を持つ）の場合はエラー
    if parent.ParentID != nil {
        return errors.New("エピソードの下にエピソードを作成することはできません")
    }

    return nil
}
```

### エピソード CRUD 時の animes 同期

Go 版でエピソードを作成・更新・削除する際に、対応する `animes` レコードも同期する。

```go
// エピソード作成時
func (u *CreateEpisodeUsecase) Execute(ctx context.Context, input CreateEpisodeInput) (*Episode, error) {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        // 親作品の anime_id を取得
        work, err := u.workRepo.FindByID(ctx, tx, input.WorkID)
        if err != nil {
            return err
        }

        // 1. animes レコードを作成（parent_id に親作品の anime_id を設定）
        anime, err := u.animeRepo.Create(ctx, tx, &Anime{
            ParentID: &work.AnimeID,
            Title:    input.Title,
            // ...
        })
        if err != nil {
            return err
        }

        // 2. episodes レコードを作成（anime_id を設定）
        episode, err := u.episodeRepo.Create(ctx, tx, &Episode{
            AnimeID:    anime.ID,
            WorkID:     input.WorkID,
            Number:     input.Number,
            SortNumber: input.SortNumber,
            // ...
        })
        if err != nil {
            return err
        }

        return episode, nil
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

- [ ] **1-1**: [Go] 既存 episodes データの animes テーブルへの移行スクリプト作成
  - episodes → animes の共通カラムコピー（parent_id 設定含む）
  - episodes.anime_id の設定
  - episodes.prev_episode_id → episodes.prev_anime_id への変換
  - ratings_count, satisfaction_rate の集計・設定（既存の episodes から移行）
  - score は別設計書で定義する計算式に基づき設定
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### フェーズ 2: Annict DB - エピソード作成機能の実装

- [ ] **2-1**: [Go] Episode リポジトリの実装
  - sqlc クエリの定義（作成、取得、カウント）
  - リポジトリ層の実装
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **2-2**: [Go] Anime リポジトリに階層制限バリデーションを追加
  - 階層制限バリデーションの実装（parent_id を持つレコードの親にはなれない）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

- [ ] **2-3**: [Go] Episode 作成ユースケースの実装
  - Episode 作成ユースケースの実装
    - animes レコードの同時作成（parent_id 設定）
    - sort_number 自動採番（既存数 × 100 + 100）
    - prev_anime_id 自動設定
    - Work.episodes_count の更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 350 行（実装 200 行 + テスト 150 行）

- [ ] **2-4**: [Go] Episode 作成ハンドラーの実装
  - Episode 作成ハンドラーの実装（`/db/works/:work_id/episodes/new` GET, `/db/works/:work_id/episodes` POST）
  - CSV 形式での複数エピソード一括作成対応
  - テンプレートの実装
  - 認可: committer ロールのチェック
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 350 行（実装 200 行 + テスト 150 行）
  - **参考**: Rails 版 `app/controllers/db/episodes_controller.rb`, `app/forms/deprecated/db/episode_rows_form.rb`

- [ ] **2-5**: [Rails] Episode 作成機能の削除
  - コントローラーの削除（`db/episodes#new`, `db/episodes#create`）
  - フォームオブジェクトの削除（`EpisodeRowsForm`）
  - ビューの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 4 ファイル
  - **想定行数**: 約 150 行（削除）

### フェーズ 3: Annict DB - エピソード編集機能の実装

- [ ] **3-1**: [Go] Episode 更新ユースケースの実装
  - Episode 更新ユースケースの実装（animes 同期更新含む）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **3-2**: [Go] Episode 編集ハンドラーの実装
  - Episode 編集ハンドラーの実装（`/db/episodes/:id/edit` GET, `/db/episodes/:id` PATCH）
  - テンプレートの実装
  - 認可: committer ロールのチェック
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）
  - **参考**: Rails 版 `app/controllers/db/episodes_controller.rb`

- [ ] **3-3**: [Rails] Episode 編集機能の削除
  - コントローラーの削除（`db/episodes#edit`, `db/episodes#update`）
  - ビューの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 3 ファイル
  - **想定行数**: 約 100 行（削除）

### フェーズ 4: Annict DB - エピソード非公開機能の実装

- [ ] **4-1**: [Go] Episode 非公開ユースケースの実装
  - Episode 非公開ユースケースの実装（animes.hidden_at 同期更新含む）
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [ ] **4-2**: [Go] Episode 非公開ハンドラーの実装
  - Episode 非公開ハンドラーの実装（`/db/episodes/:id/hide` POST）
  - 認可: admin ロールのチェック
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [ ] **4-3**: [Rails] Episode 非公開機能の削除
  - コントローラーの削除
  - ルーティングの削除
  - **想定ファイル数**: 約 2 ファイル
  - **想定行数**: 約 50 行（削除）

### フェーズ 5: Annict DB - エピソード削除機能の実装

- [ ] **5-1**: [Go] Episode 削除ユースケースの実装
  - Episode ソフト削除ユースケースの実装（animes.deleted_at 同期更新含む）
  - 後続エピソードの prev_anime_id 更新
  - Work.episodes_count の更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

- [ ] **5-2**: [Go] Episode 削除ハンドラーの実装
  - Episode 削除ハンドラーの実装（`/db/episodes/:id` DELETE）
  - 認可: admin ロールのチェック
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [ ] **5-3**: [Rails] Episode 削除機能の削除
  - コントローラーの削除（`db/episodes#destroy`）
  - ルーティングの削除
  - **想定ファイル数**: 約 2 ファイル
  - **想定行数**: 約 50 行（削除）

### フェーズ 6: エピソード自動生成タスクの Go 移行

<!--
Rails版の定期実行タスク（app.json）の中で作品・エピソードを作成・変更しているものをGo版に移行する。
-->

- [ ] **6-1**: [Go] エピソード自動生成タスクの実装
  - Syobocal からのエピソード情報取得機能
  - Slot に対応する Episode の自動作成（animes 同期作成含む）
  - LibraryEntry の next_episode_id 更新
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 400 行（実装 250 行 + テスト 150 行）
  - **参考**: Rails 版 `app/services/deprecated/episode_generator_service.rb`, `app/services/deprecated/syobocal_episode_data_fetcher_service.rb`

- [ ] **6-2**: [Rails] エピソード自動生成タスクの削除
  - Rake タスクの削除（`episode:generate`）
  - 関連サービスの削除
  - **想定ファイル数**: 約 3 ファイル
  - **想定行数**: 約 200 行（削除）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **DbActivity による活動履歴作成**: ポリモーフィック対応は別途検討。Work/Episode を保存する形で対応
- **作品に紐付く animes レコードの作成**: 別設計書で対応済み

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [animes テーブル導入設計書](../1_doing/animes-table.md)
- [作品に紐付く animes レコードを作成する](./animes-work-sync.md)
- [現在の episodes テーブル定義](/workspace/go/db/schema.sql:805)
