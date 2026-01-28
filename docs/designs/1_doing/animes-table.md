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

作品（Work）とエピソード（Episode）を統一的に管理する `animes` テーブルを導入し、作品とエピソードの相互変換を容易にする。

現在、`works` テーブルと `episodes` テーブルは別々に存在し、それぞれの ID が多数のテーブルから参照されている。このため、エピソードを単独作品に変換したり、逆に作品をエピソード化する際に、関連するすべてのテーブルを更新する必要があり、非常に複雑な処理が必要となっている。

**目的**:

- Annict DB でのデータ整理を効率化する（エピソードを作品に、作品をエピソードに変換）
- 作品とエピソードを統一モデルで管理することで、将来的なデータ構造の柔軟性を確保する

**背景**:

- 現状の `works.id` は約 25 テーブルから参照されており、変換時にすべて更新が必要
- エピソードとして登録したものを単独の作品にしてシリーズで紐付けるといった運用ニーズがある
- `animes` テーブル導入により、変換は `parent_id` の更新だけで済むようになる

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

- システムは `animes` テーブルで作品とエピソードを統一的に管理できる
  - 作品: `parent_id` が NULL の `animes` レコード
  - エピソード: `parent_id` に親作品の `animes.id` が設定された `animes` レコード
- 管理者は Annict DB で作品をエピソードに変換できる（`parent_id` を設定）
- 管理者は Annict DB でエピソードを作品に変換できる（`parent_id` を NULL に設定）
- システムは階層を 2 段階に制限する（作品 → エピソードのみ、エピソードの子は不可）
- 既存の `works` テーブルと `episodes` テーブルは段階的に廃止する

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

### データベース設計

#### animes テーブル定義

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
    number character varying(510),           -- 話数表示用（"1", "OVA" など）
    sort_number integer DEFAULT 0 NOT NULL,  -- 並び順用
    raw_number double precision,             -- 計算用の数値

    -- 作品用属性（エピソードの場合は NULL）
    season_year integer,
    season_name integer,
    media integer,
    official_site_url character varying(510) DEFAULT ''::character varying NOT NULL,
    official_site_url_en character varying DEFAULT ''::character varying NOT NULL,
    wikipedia_url character varying(510) DEFAULT ''::character varying NOT NULL,
    wikipedia_url_en character varying DEFAULT ''::character varying NOT NULL,
    twitter_username character varying(510),
    twitter_hashtag character varying(510),
    synopsis text DEFAULT ''::text NOT NULL,
    synopsis_en text DEFAULT ''::text NOT NULL,
    synopsis_source character varying DEFAULT ''::character varying NOT NULL,
    synopsis_source_en character varying DEFAULT ''::character varying NOT NULL,
    released_at date,
    released_at_about character varying,
    started_on date,
    ended_on date,
    mal_anime_id integer,
    sc_tid integer,                          -- しょぼいカレンダー TID
    no_episodes boolean DEFAULT false NOT NULL,
    start_episode_raw_number double precision DEFAULT 1.0 NOT NULL,

    -- エピソード用属性（作品の場合は NULL）
    sc_count integer,                        -- しょぼいカレンダー count
    fetch_syobocal boolean DEFAULT false NOT NULL,

    -- 集計カラム
    episodes_count integer DEFAULT 0 NOT NULL,
    manual_episodes_count integer,
    watchers_count integer DEFAULT 0 NOT NULL,
    records_count integer DEFAULT 0 NOT NULL,
    episode_records_count integer DEFAULT 0 NOT NULL,
    episode_record_bodies_count integer DEFAULT 0 NOT NULL,
    work_records_count integer DEFAULT 0 NOT NULL,
    work_records_with_body_count integer DEFAULT 0 NOT NULL,
    ratings_count integer DEFAULT 0 NOT NULL,
    score double precision,
    satisfaction_rate double precision,

    -- 画像関連
    facebook_og_image_url character varying DEFAULT ''::character varying NOT NULL,
    twitter_image_url character varying DEFAULT ''::character varying NOT NULL,
    recommended_image_url character varying DEFAULT ''::character varying NOT NULL,
    key_pv_id bigint,
    number_format_id bigint,

    -- 状態管理
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone,

    -- 旧テーブルとの紐付け（移行期間中のみ使用）
    legacy_work_id bigint,
    legacy_episode_id bigint,

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
CREATE INDEX index_animes_on_season_year_and_season_name ON public.animes(season_year, season_name);
CREATE INDEX index_animes_on_title ON public.animes(title);
CREATE INDEX index_animes_on_aasm_state ON public.animes(aasm_state);
CREATE UNIQUE INDEX index_animes_on_legacy_work_id ON public.animes(legacy_work_id) WHERE legacy_work_id IS NOT NULL;
CREATE UNIQUE INDEX index_animes_on_legacy_episode_id ON public.animes(legacy_episode_id) WHERE legacy_episode_id IS NOT NULL;
```

#### 既存テーブルへの変更

```sql
-- works テーブルに anime_id を追加
ALTER TABLE public.works ADD COLUMN anime_id bigint REFERENCES public.animes(id);
CREATE INDEX index_works_on_anime_id ON public.works(anime_id);

-- episodes テーブルに anime_id を追加
ALTER TABLE public.episodes ADD COLUMN anime_id bigint REFERENCES public.animes(id);
CREATE INDEX index_episodes_on_anime_id ON public.episodes(anime_id);
```

#### series_animes テーブル（既存 series_works の後継）

```sql
CREATE TABLE public.series_animes (
    id bigint NOT NULL PRIMARY KEY,
    series_id bigint NOT NULL REFERENCES public.series(id),
    anime_id bigint NOT NULL REFERENCES public.animes(id),
    summary character varying DEFAULT ''::character varying NOT NULL,
    summary_en character varying DEFAULT ''::character varying NOT NULL,
    aasm_state character varying DEFAULT 'published'::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    deleted_at timestamp without time zone,
    unpublished_at timestamp without time zone,
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

1. `animes` テーブルを作成
2. 既存の `works` と `episodes` から `animes` にデータをコピー
3. `works.anime_id` と `episodes.anime_id` を設定
4. 作品・エピソードの CRUD 時に `animes` も同期更新するトリガーまたはアプリケーションコードを追加

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

#### フェーズ 4: 旧テーブルの廃止（将来）

1. `works` テーブルと `episodes` テーブルを廃止
2. 関連テーブルの `work_id` / `episode_id` カラムを削除
3. `animes` テーブルの `legacy_work_id` / `legacy_episode_id` カラムを削除

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
func (u *ConvertAnimeUsecase) ConvertEpisodeToWork(ctx context.Context, animeID int64) error {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        anime, err := u.animeRepo.FindByID(ctx, tx, animeID)
        if err != nil {
            return err
        }

        if anime.ParentID == nil {
            return errors.New("既に作品です")
        }

        // parent_id を NULL に設定
        anime.ParentID = nil
        // 必要に応じて作品固有の属性を設定（UI から入力）

        return u.animeRepo.Update(ctx, tx, anime)
    })
}

// 作品をエピソードに変換
func (u *ConvertAnimeUsecase) ConvertWorkToEpisode(ctx context.Context, animeID, parentID int64) error {
    return u.db.WithTx(ctx, func(tx *sql.Tx) error {
        anime, err := u.animeRepo.FindByID(ctx, tx, animeID)
        if err != nil {
            return err
        }

        if anime.ParentID != nil {
            return errors.New("既にエピソードです")
        }

        // parent_id を設定
        anime.ParentID = &parentID

        return u.animeRepo.Update(ctx, tx, anime)
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

  - マイグレーションファイルの作成
  - インデックスの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

- [ ] **1-2**: 既存データの animes テーブルへの移行スクリプト作成

  - works → animes のデータコピー
  - episodes → animes のデータコピー
  - works.anime_id, episodes.anime_id の設定
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 250 行（実装 200 行 + テスト 50 行）

- [ ] **1-3**: Go 版 Anime リポジトリの実装

  - sqlc クエリの定義
  - リポジトリ層の実装
  - 階層制限バリデーションの実装
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 400 行（実装 200 行 + テスト 200 行）

- [ ] **1-4**: Go 版 Work/Episode 作成時の animes 同期処理

  - Work 作成時に animes レコードも作成
  - Episode 作成時に animes レコードも作成
  - 更新・削除時の同期処理
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **1-5**: Rails 版 Work/Episode 作成時の animes 同期処理

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

  - ハンドラーの修正
  - テンプレートの修正
  - **想定ファイル数**: 約 6 ファイル（実装 3 + テスト 3）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [ ] **3-2**: Go 版エピソードページの animes テーブル参照への切り替え

  - ハンドラーの修正
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

- **works テーブルと episodes テーブルの廃止**: フェーズ 4 として将来的に実施
- **GraphQL API の animes 対応**: 既存 API の互換性を維持するため、当面は works/episodes を返す
- **3 階層以上のネスト対応**: 現時点では作品 → エピソードの 2 階層のみをサポート

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [現在の works テーブル定義](/workspace/go/db/schema.sql:3105)
- [現在の episodes テーブル定義](/workspace/go/db/schema.sql:805)
- [現在の series_works テーブル定義](/workspace/go/db/schema.sql:2149)
