# animes テーブル導入による作品・エピソード統合 設計書

<!--
このテンプレートの使い方:
1. このファイルを `.claude/designs/2_todo/` ディレクトリにコピー
   例: cp .claude/designs/template.md .claude/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

作品（Work）とエピソード（Episode）を統一的に識別する `animes` テーブルを導入し、作品とエピソードの相互変換を容易にする。

現在、`works` テーブルと `episodes` テーブルは別々に存在し、それぞれの ID が多数のテーブルから参照されている。このため、エピソードを単独作品に変換したり、逆に作品をエピソード化する際に、関連するすべてのテーブルを更新する必要があり、非常に複雑な処理が必要となっている。

**目的**:

- Annict DB でのデータ整理を効率化する（エピソードを作品に、作品をエピソードに変換）
- 関連テーブルが `anime_id` を参照することで、変換時の更新対象を最小化する

**背景**:

- 現状の `works.id` は約 25 テーブルから参照されており、変換時にすべて更新が必要
- エピソードとして登録したものを単独の作品にしてシリーズで紐付けるといった運用ニーズがある
- `animes` テーブル導入により、関連テーブルは `anime_id` を参照するようになり、変換時の影響範囲が大幅に縮小される

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

- システムは `animes` テーブルで作品とエピソードを統一的に識別できる
  - 作品: `parent_id` が NULL の `animes` レコード（詳細情報は `works` テーブルに格納）
  - エピソード: `parent_id` に親作品の `animes.id` が設定された `animes` レコード（詳細情報は `episodes` テーブルに格納）
- 管理者は Annict DB で作品をエピソードに変換できる（`parent_id` を設定 + `episodes` レコードを作成）
- 管理者は Annict DB でエピソードを作品に変換できる（`parent_id` を NULL に設定 + `works` レコードを作成）
- システムは階層を 2 段階に制限する（作品 → エピソードのみ、エピソードの子は不可）
- 既存の `works` テーブルと `episodes` テーブルは存続し、それぞれ作品固有・エピソード固有の情報を保持する

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
- **パフォーマンス**: 既存の作品・エピソード取得クエリのパフォーマンスを維持すること

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

### 設計方針: 分離設計

本設計では「分離設計」を採用する。`animes` テーブルは共通カラムのみを持ち、作品固有・エピソード固有の情報は既存の `works`・`episodes` テーブルに保持する。

```
animes テーブル（共通情報）
├── works テーブル（作品固有情報）: animes has_one works
└── episodes テーブル（エピソード固有情報）: animes has_one episodes
```

**この設計を採用した理由**:

1. **カラムの責務が明確**: 作品固有/エピソード固有/共通が分離されており、仕様を忘れてしまっても理解しやすい
2. **既存テーブル構造の変更が最小限**: `works` と `episodes` は存続するため、既存コードへの影響が比較的小さい
3. **NULL カラムの削減**: `animes` テーブルに作品/エピソード両方のカラムを持たせる必要がない

**トレードオフ**:

- 変換時は `parent_id` の変更に加えて、`works` または `episodes` レコードの作成が必要
- 作品情報の取得時は `animes JOIN works`、エピソード情報の取得時は `animes JOIN episodes` が必要

### データベース設計

#### animes テーブル定義（共通情報のみ）

```sql
CREATE TABLE public.animes (
    id bigint NOT NULL PRIMARY KEY,
    parent_id bigint REFERENCES public.animes(id),

    -- 共通属性
    title character varying(510) NOT NULL,
    title_kana character varying,   -- NULL許容（未設定を表現）
    title_alter character varying,  -- NULL許容（未設定を表現）

    -- 状態管理
    deleted_at timestamp without time zone,
    hidden_at timestamp without time zone,  -- 非公開日時（旧 unpublished_at）

    -- タイムスタンプ
    created_at timestamp with time zone,
    updated_at timestamp with time zone,

    -- 階層制限の制約（parent_id を持つレコードの parent_id は NULL であること）
    CONSTRAINT animes_max_depth_check CHECK (
        parent_id IS NULL OR NOT EXISTS (
            SELECT 1 FROM animes parent WHERE parent.id = animes.parent_id AND parent.parent_id IS NOT NULL
        )
    )
);

-- インデックス
CREATE INDEX index_animes_on_parent_id ON public.animes(parent_id);
```

#### 既存テーブルへの変更

```sql
-- works テーブルに anime_id を追加（作品固有情報は既存カラムに保持）
ALTER TABLE public.works ADD COLUMN anime_id bigint REFERENCES public.animes(id);
CREATE UNIQUE INDEX index_works_on_anime_id ON public.works(anime_id) WHERE anime_id IS NOT NULL;

-- episodes テーブルに anime_id を追加（エピソード固有情報は既存カラムに保持）
ALTER TABLE public.episodes ADD COLUMN anime_id bigint REFERENCES public.animes(id);
CREATE UNIQUE INDEX index_episodes_on_anime_id ON public.episodes(anime_id) WHERE anime_id IS NOT NULL;
```

#### カラムの分類

**共通カラム（animes テーブル）**:

- `title`, `title_kana`, `title_alter`
- `deleted_at`, `hidden_at`（旧 `unpublished_at`）
- `created_at`, `updated_at`

**作品固有カラム（works テーブルに保持）**:

- `season_id`, `season_year`, `season_name`, `media`
- `sc_tid`, `mal_anime_id`
- `official_site_url`, `official_site_url_en`
- `wikipedia_url`, `wikipedia_url_en`
- `synopsis`, `synopsis_en`, `synopsis_source`, `synopsis_source_en`
- `released_at`, `released_at_about`, `started_on`, `ended_on`
- `twitter_username`, `twitter_hashtag`
- `number_format_id`, `key_pv_id`
- `no_episodes`, `start_episode_raw_number`
- `facebook_og_image_url`, `twitter_image_url`, `recommended_image_url`
- 集計カラム: `episodes_count`, `watchers_count`, `manual_episodes_count`, `work_records_count`, `work_records_with_body_count`, `score`, `ratings_count`, `satisfaction_rate`, `records_count`

**エピソード固有カラム（episodes テーブルに保持）**:

- `work_id`（既存の親作品への参照、移行期間中は維持）
- `number`, `number_en`, `sort_number`, `raw_number`
- `sc_count`, `fetch_syobocal`
- `prev_episode_id`
- 集計カラム: `episode_records_count`, `episode_record_bodies_count`, `score`, `ratings_count`, `satisfaction_rate`

#### series_animes テーブル（既存 series_works の後継）

```sql
CREATE TABLE public.series_animes (
    id bigint NOT NULL PRIMARY KEY,
    series_id bigint NOT NULL REFERENCES public.series(id),
    anime_id bigint NOT NULL REFERENCES public.animes(id),
    summary character varying DEFAULT ''::character varying NOT NULL,
    summary_en character varying DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    hidden_at timestamp without time zone,
    CONSTRAINT series_animes_anime_must_be_work CHECK (
        NOT EXISTS (
            SELECT 1 FROM animes WHERE animes.id = series_animes.anime_id AND animes.parent_id IS NOT NULL
        )
    )
);

CREATE UNIQUE INDEX index_series_animes_on_series_id_and_anime_id ON public.series_animes(series_id, anime_id);
CREATE INDEX index_series_animes_on_anime_id ON public.series_animes(anime_id);
```

### 移行方針

#### フェーズ 1: animes テーブル作成とデータ同期

1. `animes` テーブルを作成（共通カラムのみ）
2. 既存の `works` から `animes` に共通カラムをコピー（作品として）
3. 既存の `episodes` から `animes` に共通カラムをコピー（エピソードとして、`parent_id` に親作品の `anime_id` を設定）
4. `works.anime_id` と `episodes.anime_id` を設定
5. 作品・エピソードの CRUD 時に `animes` も同期更新するトリガーまたはアプリケーションコードを追加

#### フェーズ 2: 関連テーブルへの anime_id 追加

以下のテーブルに `anime_id` カラムを追加し、既存の `work_id` / `episode_id` から値を埋める：

- `activities` (work_id, episode_id → anime_id)
- `casts` (work_id → anime_id)
- `channel_works` (work_id → anime_id)
- `collection_items` (work_id → anime_id)
- `comments` (work_id → anime_id)
- `episode_records` (work_id, episode_id → anime_id, parent_anime_id)
- `library_entries` (work_id, next_episode_id → anime_id, next_anime_id)
- `multiple_episode_records` (work_id → anime_id)
- `programs` (work_id → anime_id)
- `records` (work_id → anime_id)
- `slots` (work_id, episode_id → anime_id, episode_anime_id)
- `staffs` (work_id → anime_id)
- `statuses` (work_id → anime_id)
- `syobocal_alerts` (work_id → anime_id)
- `trailers` (work_id → anime_id)
- `vod_titles` (work_id → anime_id)
- `work_comments` (work_id → anime_id)
- `work_images` (work_id → anime_id)
- `work_records` (work_id → anime_id)
- `work_taggings` (work_id → anime_id)
- `series_works` → `series_animes` に移行

#### フェーズ 3: アプリケーションコードの移行

1. Go 版で `animes` テーブルを参照するように変更
2. Rails 版で `animes` テーブルを参照するように変更
3. 古い `work_id` / `episode_id` カラムを非推奨化

#### フェーズ 4: 旧カラムの廃止（将来）

1. 関連テーブルの `work_id` / `episode_id` カラムを削除
2. `works` と `episodes` テーブルは存続（作品固有・エピソード固有の情報を保持）

### 階層制限の実装

アプリケーション層で階層制限を実装する：

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

### 変換処理の実装

```go
// エピソードを作品に変換
func (u *ConvertAnimeUsecase) ConvertEpisodeToWork(ctx context.Context, animeID int64, workInput WorkInput) error {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        anime, err := u.animeRepo.FindByID(ctx, tx, animeID)
        if err != nil {
            return err
        }

        if anime.ParentID == nil {
            return errors.New("既に作品です")
        }

        // 1. parent_id を NULL に設定
        anime.ParentID = nil
        if err := u.animeRepo.Update(ctx, tx, anime); err != nil {
            return err
        }

        // 2. works レコードを作成（作品固有情報を保存）
        work := &Work{
            AnimeID:    animeID,
            SeasonYear: workInput.SeasonYear,
            Media:      workInput.Media,
            // ... その他の作品固有カラム
        }
        return u.workRepo.Create(ctx, tx, work)

        // 注: 既存の episodes レコードは残しておく（参照されなくなるだけ）
    })
}

// 作品をエピソードに変換
func (u *ConvertAnimeUsecase) ConvertWorkToEpisode(ctx context.Context, animeID, parentID int64, episodeInput EpisodeInput) error {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        anime, err := u.animeRepo.FindByID(ctx, tx, animeID)
        if err != nil {
            return err
        }

        if anime.ParentID != nil {
            return errors.New("既にエピソードです")
        }

        // 1. parent_id を設定
        anime.ParentID = &parentID
        if err := u.animeRepo.Update(ctx, tx, anime); err != nil {
            return err
        }

        // 2. episodes レコードを作成（エピソード固有情報を保存）
        episode := &Episode{
            AnimeID:    animeID,
            Number:     episodeInput.Number,
            SortNumber: episodeInput.SortNumber,
            // ... その他のエピソード固有カラム
        }
        return u.episodeRepo.Create(ctx, tx, episode)

        // 注: 既存の works レコードは残しておく（参照されなくなるだけ）
    })
}
```

### データ取得パターン

```go
// 作品情報を取得（animes JOIN works）
func (r *AnimeRepository) FindWorkByID(ctx context.Context, animeID int64) (*AnimeWithWork, error) {
    query := `
        SELECT a.*, w.*
        FROM animes a
        JOIN works w ON w.anime_id = a.id
        WHERE a.id = $1 AND a.parent_id IS NULL
    `
    // ...
}

// エピソード情報を取得（animes JOIN episodes）
func (r *AnimeRepository) FindEpisodeByID(ctx context.Context, animeID int64) (*AnimeWithEpisode, error) {
    query := `
        SELECT a.*, e.*
        FROM animes a
        JOIN episodes e ON e.anime_id = a.id
        WHERE a.id = $1 AND a.parent_id IS NOT NULL
    `
    // ...
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
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: animes テーブル作成とデータ同期

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [ ] **1-1**: animes テーブルのマイグレーション作成

  - マイグレーションファイルの作成（共通カラムのみ）
  - インデックスの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 70 行 + テスト 30 行）

- [ ] **1-2**: works/episodes テーブルへの anime_id カラム追加

  - マイグレーションファイルの作成
  - ユニークインデックスの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

- [ ] **1-3**: 既存データの animes テーブルへの移行スクリプト作成

  - works → animes の共通カラムコピー
  - episodes → animes の共通カラムコピー（parent_id 設定含む）
  - works.anime_id, episodes.anime_id の設定
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [ ] **1-4**: Go 版 Anime リポジトリの実装

  - sqlc クエリの定義
  - リポジトリ層の実装
  - 階層制限バリデーションの実装
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 400 行（実装 200 行 + テスト 200 行）

- [ ] **1-5**: Go 版 Work/Episode 作成時の animes 同期処理

  - Work 作成時に animes レコードも作成
  - Episode 作成時に animes レコードも作成
  - 更新・削除時の同期処理
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **1-6**: Rails 版 Work/Episode 作成時の animes 同期処理

  - Work モデルのコールバック追加
  - Episode モデルのコールバック追加
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 250 行（実装 100 行 + テスト 150 行）

### フェーズ 2: 関連テーブルへの anime_id 追加

- [ ] **2-1**: activities, casts, channel_works への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-2**: collection_items, comments, episode_records への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-3**: library_entries, multiple_episode_records, programs への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-4**: records, slots, staffs への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-5**: statuses, syobocal_alerts, trailers への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-6**: vod_titles, work_comments, work_images への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-7**: work_records, work_taggings への anime_id 追加

  - マイグレーションの作成
  - データ移行スクリプト
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [ ] **2-8**: series_animes テーブルの作成と series_works からの移行

  - series_animes テーブルのマイグレーション
  - series_works からのデータ移行
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### フェーズ 3: アプリケーションコードの移行

- [ ] **3-1**: Go 版作品ページの animes テーブル参照への切り替え

  - ハンドラーの修正（animes JOIN works）
  - テンプレートの修正
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **3-2**: Go 版エピソードページの animes テーブル参照への切り替え

  - ハンドラーの修正（animes JOIN episodes）
  - テンプレートの修正
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **3-3**: Annict DB 変換機能の実装（Go 版）

  - 変換ユースケースの実装
  - 管理画面 UI の実装
  - **想定ファイル数**: 約 8 ファイル（実装 4 + テスト 4）
  - **想定行数**: 約 400 行（実装 200 行 + テスト 200 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **works テーブルと episodes テーブルの廃止**: 分離設計を採用したため、これらのテーブルは存続する
- **GraphQL API の animes 対応**: 既存 API の互換性を維持するため、当面は works/episodes を返す
- **3 階層以上のネスト対応**: 現時点では作品 → エピソードの 2 階層のみをサポート
- **古い works/episodes レコードのクリーンアップ**: 変換後に残る古いレコードの削除は将来的に検討

## ボツ案: 統合設計

<!--
検討したが採用しなかった設計案を記録
-->

以下の「統合設計」は検討したが、最終的に採用しなかった。

### 概要

`animes` テーブルに作品・エピソード両方のカラムを統合し、`works` と `episodes` テーブルを将来的に廃止する設計。

### テーブル定義

```sql
CREATE TABLE public.animes (
    id bigint NOT NULL PRIMARY KEY,
    parent_id bigint REFERENCES public.animes(id),

    -- 共通属性
    title character varying(510) NOT NULL,
    title_kana character varying DEFAULT ''::character varying NOT NULL,
    title_ro character varying DEFAULT ''::character varying NOT NULL,
    title_en character varying DEFAULT ''::character varying NOT NULL,
    title_alter character varying DEFAULT ''::character varying NOT NULL,
    title_alter_en character varying DEFAULT ''::character varying NOT NULL,

    -- エピソード用属性（作品の場合は NULL）
    number character varying(510),
    sort_number integer DEFAULT 0 NOT NULL,
    raw_number double precision,

    -- 作品用属性（エピソードの場合は NULL）
    season_year integer,
    season_name integer,
    media integer,
    official_site_url character varying(510) DEFAULT ''::character varying NOT NULL,
    -- ... その他の作品固有カラム

    -- 集計カラム
    episodes_count integer DEFAULT 0 NOT NULL,
    -- ... その他の集計カラム

    -- 状態管理
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone,

    -- 旧テーブルとの紐付け（移行期間中のみ使用）
    legacy_work_id bigint,
    legacy_episode_id bigint,

    -- タイムスタンプ
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
```

### メリット

1. **変換処理が最もシンプル**: `parent_id` の変更だけで完結
2. **JOINが不要**: 単一テーブルで完結するためクエリがシンプル
3. **将来的に works/episodes テーブルを廃止可能**

### 採用しなかった理由

1. **カラムの責務が不明確**: 作品用カラムとエピソード用カラムが混在し、どのカラムがどちらで使われるか判別が困難
2. **NULL カラムが多い**: 作品レコードにはエピソード用カラムが NULL、エピソードレコードには作品用カラムが NULL になる
3. **既存コードへの影響が大きい**: すべてのカラムを `animes` に移行するため、変更範囲が広い

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [現在の works テーブル定義](/workspace/go/db/schema.sql:3181)
- [現在の episodes テーブル定義](/workspace/go/db/schema.sql:805)
- [現在の series_works テーブル定義](/workspace/go/db/schema.sql:2149)
