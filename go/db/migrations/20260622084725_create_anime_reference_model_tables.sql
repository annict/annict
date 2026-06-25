-- migrate:up

-- anime_status: the content-identity lifecycle (all soft; rows are never
-- physically deleted). 'archived' freezes new records / activity but keeps the
-- page directly viewable, so users retain access to records they already made
-- (editors may archive); 'merged' tombstones a duplicate merge and resolves to
-- the canonical via anime_redirects; 'deleted' hides the anime from the site
-- entirely (admin-only). archive_message carries optional guidance for an
-- archived page.
--
-- [Ja] anime_status: コンテンツ同一性のライフサイクル (いずれもソフトで、行は
-- 物理削除しない)。'archived' は新規の記録/活動を凍結するがページは直接閲覧でき、
-- 利用者は既に付けた記録に引き続きアクセスできる (編集者も設定可)。'merged' は
-- 重複統合の墓標で anime_redirects 経由で canonical へ解決する。'deleted' は
-- サイト上から完全に非表示にする (管理者のみ)。archive_message は archived の
-- ページに出す任意の案内文を持つ。
CREATE TYPE public.anime_status AS ENUM ('published', 'archived', 'merged', 'deleted');

CREATE TYPE public.anime_classification_kind AS ENUM ('work', 'episode');

-- anime_media carries the legacy Work#media values (Rails enumerize:
-- {tv, ova, movie, web, other}), renaming only 'web' to 'ona' to match the
-- "Original Net Animation" naming common across anime databases. The phase 2
-- sync maps web (4) to 'ona' and the other codes by name; 'other' (0) keeps
-- every legacy row representable. "Special" is deliberately not a value here:
-- being a special episode is orthogonal to the medium (a TV-aired special is
-- both TV and special), so it belongs to a separate attribute, not this enum.
-- season_name carries the Season name codes (Season::NAME_HASH), renaming
-- 'autumn' to 'fall' to match the season naming common across anime databases
-- (AniList / MyAnimeList / Kitsu); the phase 2 sync maps autumn (4) to 'fall'
-- and the rest by name.
--
-- [Ja] anime_media は旧 Work#media の値 (Rails enumerize の
-- {tv, ova, movie, web, other}) を、'web' のみ 'ona' に改名して引き継ぐ (各
-- アニメ DB で一般的な "Original Net Animation" の呼称に合わせる)。フェーズ 2 の
-- 同期は web (4) を 'ona' に、残りのコードは同名で写像する。'other' (0) は
-- すべての旧行を表現できるよう残す。"Special" はここに値として持たない: 特別編で
-- あることは媒体と直交し (TV 放送の特別編は TV でも special でもある)、媒体 enum
-- ではなく別属性で表すべきものだから。season_name は Season の名前コード
-- (Season::NAME_HASH) を引き継ぎ、'autumn' を 'fall' に改名する (各アニメ DB
-- (AniList / MyAnimeList / Kitsu) で一般的な季節表記に合わせる)。フェーズ 2 の
-- 同期は autumn (4) を 'fall' に、残りは同名で写像する。
CREATE TYPE public.anime_media AS ENUM ('tv', 'ova', 'movie', 'ona', 'other');
CREATE TYPE public.season_name AS ENUM ('winter', 'spring', 'summer', 'fall');

-- release_status is the work's broadcast/release lifecycle. Only three coarse
-- states are stored: 'not_yet_released' (before airing), 'released' (airing has
-- started), and 'cancelled' (production cancelled; editor-set, no backfill
-- source). The finer "airing / finished / rebroadcasting" distinctions are not
-- stored but derived at read time from slots, so the stored transition stays
-- one-way (not_yet_released -> released). Combined with a missing season it
-- yields the "tbd" (not_yet_released) vs "unknown" (released) display. It is
-- used mainly for works; an episode may also carry it (e.g. ahead of a possible
-- promotion to a work) but rarely does. There is no work-only constraint; it is
-- a plain nullable column. See ADR 0010.
--
-- [Ja] release_status は作品の放送/公開ライフサイクル。格納するのは粗い 3 状態
-- だけで、'not_yet_released' (放送前)・'released' (放送開始済み)・'cancelled'
-- (製作中止。編集者が設定し、バックフィル元は無い)。「放送中 / 完結 / 再放送中」の
-- 細かい区別は格納せず読み取り時に slot から導出するため、格納遷移は
-- not_yet_released -> released の一方向で済む。季節欠落と組み合わせて「未定」
-- (not_yet_released) と「不明」(released) の表示を導出する。主に作品で使うが、
-- エピソードに設定してもよい (将来作品に昇格する場合に備えるなど)。ただし
-- エピソードでは積極的には活用しない。work 限定の制約は設けず、単なる NULL 許容の
-- 任意カラムとする。ADR 0010 を参照。
CREATE TYPE public.release_status AS ENUM ('not_yet_released', 'released', 'cancelled');

-- anime_external_service identifies which external anime database a row in
-- anime_external_ids maps to (the Syobocal title DB and MyAnimeList for now,
-- more added later). Adding a service is one ALTER TYPE ADD VALUE plus rows,
-- never a new schema column.
--
-- [Ja] anime_external_service は anime_external_ids の各行が指す外部アニメ DB を
-- 表す (現状は Syobocal の title DB と MyAnimeList。以降は値追加で増やす)。
-- サービスの追加は ALTER TYPE ADD VALUE 1 行 + 行追加で済み、スキーマ列を増やさない。
CREATE TYPE public.anime_external_service AS ENUM ('syobocal', 'mal');

-- anime_link_kind categorizes a display URL in anime_links. It starts with
-- 'official_site' / 'wikipedia' / 'other'; X posts, news articles, etc. go under
-- 'other' for now and get their own value once the 'other' bucket is worth
-- classifying. Video links are NOT here -- they belong to the trailers table
-- (thumbnail / provider handling). See ADR 0013.
--
-- [Ja] anime_link_kind は anime_links の表示用 URL の種別。'official_site' /
-- 'wikipedia' / 'other' で開始し、X のポスト・ニュース記事などは当面 'other' に入れ、
-- 分類できてきたら個別の値に切り出す。動画リンクはここに入れず trailers テーブルが
-- 担う (サムネイル/プロバイダー処理があるため)。ADR 0013 を参照。
CREATE TYPE public.anime_link_kind AS ENUM ('official_site', 'wikipedia', 'other');

-- language is the language a piece of content is in: ja / en, plus 'other' for
-- any other language or language-neutral content. It mirrors the Rails
-- ApplicationRecord::LOCALES set and is reusable wherever a row carries a content
-- language. A user-facing display / UI language (always ja or en, never 'other')
-- belongs in a separate enum, not here.
--
-- [Ja] language はコンテンツが何語かを表す: ja / en に加え、それ以外の言語や言語中立は
-- 'other'。Rails の ApplicationRecord::LOCALES と同じ集合で、行がコンテンツの言語を持つ
-- 箇所で再利用できる。利用者の表示 / UI 言語 (常に ja か en で 'other' は取らない) は
-- ここでなく別 enum に置く。
CREATE TYPE public.language AS ENUM ('ja', 'en', 'other');

-- anime_event_kind categorizes a dated event in anime_events for an anime's
-- calendar / timeline. It starts with 'broadcast' (the broadcast period),
-- 'revival_screening', and 'other'; more kinds are added as needed. These are
-- coarse anime-level dates, distinct from slots (per-episode, per-channel
-- airings). See ADR 0014.
--
-- [Ja] anime_event_kind は anime_events の日付イベントの種別で、作品のカレンダー /
-- タイムライン用。'broadcast' (放送期間)・'revival_screening'・'other' で開始し、
-- 必要に応じて種別を増やす。これは作品単位の粗い日付で、slots (各話×チャンネルの
-- 放送枠) とは別物。ADR 0014 を参照。
CREATE TYPE public.anime_event_kind AS ENUM ('broadcast', 'revival_screening', 'other');

-- anime_account_service identifies the platform of an anime's official account in
-- anime_official_accounts (mostly social media, but the table is not limited to
-- social services). Adding a platform is one ALTER TYPE ADD VALUE plus rows. For
-- most platforms the account column holds a single handle (x, youtube, line, ...);
-- for ActivityPub-style federated platforms (mastodon) it holds the full
-- webfinger handle user@instance, since the instance domain is part of the
-- identity and the URL is derived by parsing it. See ADR 0015.
--
-- [Ja] anime_account_service は anime_official_accounts の各行が指すアニメ公式
-- アカウントのプラットフォームを表す (多くはソーシャルメディアだが、ソーシャルに
-- 限らない)。プラットフォームの追加は ALTER TYPE ADD VALUE 1 行 + 行追加で済む。
-- 多くのプラットフォームは account 列に単一のハンドルを入れるが (x / youtube / line
-- など)、ActivityPub 系の分散型 (mastodon) はインスタンスのドメインまで含めて初めて
-- 一意になるため、フルの webfinger ハンドル user@instance を入れ、URL はそれを
-- パースして導出する。ADR 0015 を参照。
CREATE TYPE public.anime_account_service AS ENUM ('bluesky', 'instagram', 'line', 'mastodon', 'mixi2', 'threads', 'tiktok', 'x', 'youtube');

-- Layer 1: content identity. The public ID (new URL / API v2) is anchored here
-- and stays permanent across re-classification, merge, and split.
--
-- [Ja] 第 1 層: コンテンツ同一性。公開 ID (新 URL / API v2) はここに張り、再分類・
-- 統合・分離を通じて永続する。
CREATE TABLE public.animes (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR,
    title_kana VARCHAR,
    -- title_ro is the romanized (Romaji) title; title_en is the English title.
    -- They are distinct fields, kept separate as in works / episodes.
    --
    -- [Ja] title_ro はローマ字タイトル、title_en は英語タイトル。別物であり、
    -- works / episodes と同様に分けて持つ。
    title_ro VARCHAR,
    title_en VARCHAR,
    title_alter VARCHAR,
    title_alter_ro VARCHAR,
    title_alter_en VARCHAR,
    -- Free-text catch-all for titles in languages other than ja / en (Chinese,
    -- etc.), enumerated separated by a delimiter such as "、". Kept only for
    -- search / identification, not structured display, since the site shows
    -- ja / en only; the delimiter is not strict. See ADR 0013.
    --
    -- [Ja] ja / en 以外の言語のタイトル (中国語など) の受け皿。「、」などで区切って
    -- 羅列するフリーテキストで、区切りは厳密でない。サイトの表示は ja / en のみのため
    -- 構造化表示はせず、検索・識別用に留める。ADR 0013 を参照。
    title_alter_other VARCHAR,
    media public.anime_media,
    -- Used mainly for works; an episode may also carry it but rarely does. See
    -- the release_status type comment.
    --
    -- [Ja] 主に作品で使う (エピソードにも設定可だが通常は使わない)。詳細は
    -- release_status 型のコメントを参照。
    release_status public.release_status,
    synopsis TEXT,
    synopsis_en TEXT,
    synopsis_source VARCHAR,
    synopsis_source_en VARCHAR,
    status public.anime_status NOT NULL DEFAULT 'published',
    archive_message VARCHAR,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Start the animes sequence just above the next power of ten greater than the
-- current max of the legacy works.id / episodes.id space. A new anime ID is the
-- legacy max rounded up to the next order of magnitude, plus one (e.g. a legacy
-- max of 180,372 yields a first anime ID of 1,000,001). This keeps the new
-- public ID -- animes.id, exposed in the v2 URL / API -- short and anchored to a
-- clean power-of-ten boundary, while still placing every anime ID above the
-- legacy range so the two ID spaces stay disjoint. length(max::text) is the
-- digit count of the max, power(10, digits) is that next power of ten, and
-- setval(..., true) makes the first nextval return power + 1.
--
-- [Ja] animes のシーケンスは、現在の works.id / episodes.id の最大値より大きい
-- 「次の 10 の冪」のすぐ上から開始する。つまり新しい anime ID は、旧 ID の最大値を
-- 1 桁上の位に切り上げた値に 1 を足したものになる (例: 旧 ID 最大が 180,372 なら
-- 最初の anime ID は 1,000,001)。これにより v2 の URL / API で公開される公開 ID
-- (animes.id) を短く、かつ 10 の冪というきれいな境界に揃えつつ、すべての anime ID を
-- 旧 ID の範囲より上に置いて両 ID 空間を重ならせない。length(max::text) は最大値の
-- 桁数、power(10, 桁数) がその次の 10 の冪で、setval(..., true) により最初の nextval
-- が冪 + 1 を返す。
SELECT setval(
    pg_get_serial_sequence('public.animes', 'id'),
    power(10, length(GREATEST(
        (SELECT COALESCE(MAX(id), 0) FROM public.works),
        (SELECT COALESCE(MAX(id), 0) FROM public.episodes)
    )::text))::bigint,
    true
);

-- Layer 2: classification. Holds only the current classification; its history
-- is delegated to the Annict DB edit log, not stored here.
--
-- [Ja] 第 2 層: 分類。現在の分類のみを持つ。分類の履歴は Annict DB の編集履歴に
-- 委譲し、ここには持たない。
CREATE TABLE public.anime_classifications (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    kind public.anime_classification_kind NOT NULL,
    parent_anime_id BIGINT REFERENCES public.animes(id),
    number NUMERIC,
    number_text VARCHAR,
    sort_number INTEGER,
    standalone BOOLEAN NOT NULL DEFAULT FALSE,
    number_format_id BIGINT REFERENCES public.number_formats(id),
    episode_start_number NUMERIC,
    expected_episodes_count INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- A work has no parent; an episode always has one.
    --
    -- [Ja] work は親を持たず、episode は必ず親を持つ。
    CONSTRAINT anime_classifications_parent_check CHECK ((kind = 'work') = (parent_anime_id IS NULL)),
    -- Only a work can be standalone (a self-contained viewing unit such as a
    -- film). Episodes are always viewing units regardless of this flag.
    --
    -- [Ja] standalone を立てられるのは work だけ (映画などそれ自体が視聴単位の
    -- 作品)。episode はこのフラグによらず常に視聴単位。
    CONSTRAINT anime_classifications_standalone_check CHECK (kind = 'work' OR NOT standalone),
    -- number is the episode's numeric number (3, or 3.5 for a recap slotted
    -- between episodes 3 and 4), NULL for a special with no number; number_text is
    -- the display string ("第3話", "総集編"). Both belong to episodes only; a work
    -- carries neither. See ADR 0018.
    --
    -- [Ja] number はエピソードの数値の話数 (3、第3話と第4話の間の総集編なら 3.5)、
    -- 番号を持たない special は NULL。number_text は表示文字列 (「第3話」「総集編」)。
    -- どちらも episode だけが持ち、work は持たない。ADR 0018 を参照。
    CONSTRAINT anime_classifications_number_check CHECK (kind = 'episode' OR number IS NULL),
    CONSTRAINT anime_classifications_number_text_check CHECK (kind = 'episode' OR number_text IS NULL),
    -- An episode is always ordered within its parent (sort_number set); a work
    -- has no order (sort_number NULL). Same two-way shape as the parent check.
    --
    -- [Ja] episode は親内で必ず並び順を持ち (sort_number あり)、work は持たない
    -- (sort_number NULL)。parent check と同じ双方向の形。
    CONSTRAINT anime_classifications_sort_number_check CHECK ((kind = 'episode') = (sort_number IS NOT NULL)),
    -- number_format_id is a work-only setting (which number_formats row governs
    -- how this work's generated episodes are numbered); an episode must never
    -- carry it. Work-only like standalone, so the same one-way shape: a work MAY
    -- have it but stays nullable. See ADR 0016.
    --
    -- [Ja] number_format_id は work 限定の設定 (この作品の生成エピソードをどの
    -- number_formats 行で採番するか) で、episode は持ってはいけない。standalone と
    -- 同じく work 限定なので片方向の形: work は持ってもよいが nullable のまま。
    -- ADR 0016 を参照。
    CONSTRAINT anime_classifications_number_format_id_check CHECK (kind = 'work' OR number_format_id IS NULL),
    -- episode_start_number is a work-only generation setting: the starting number
    -- given to the first generated episode (e.g. 13 for a second cours numbering
    -- from 13). Renamed from the legacy works.start_episode_raw_number; an episode
    -- never carries it. Work-only like standalone, one-way shape. See ADR 0018.
    --
    -- [Ja] episode_start_number は work 限定の生成設定で、先頭の生成エピソードに与える
    -- 話数の起点 (例: 2 期で 13 から振るなら 13)。旧 works.start_episode_raw_number を
    -- 改名。episode は持たない。standalone と同じ work 限定の片方向。ADR 0018 を参照。
    CONSTRAINT anime_classifications_episode_start_number_check CHECK (kind = 'work' OR episode_start_number IS NULL),
    -- expected_episodes_count is the editor-declared planned total number of
    -- episodes for a work (renamed from the legacy works.manual_episodes_count).
    -- It caps episode generation and marks completeness; an episode never carries
    -- it. Work-only like standalone, same one-way shape (nullable for works too).
    -- See ADR 0016.
    --
    -- [Ja] expected_episodes_count は編集者が宣言する作品の予定総話数 (旧
    -- works.manual_episodes_count を改名)。エピソード生成の打ち切りと全話完了の判定に
    -- 使い、episode は持たない。standalone と同じく work 限定で片方向の形 (work でも
    -- nullable)。ADR 0016 を参照。
    CONSTRAINT anime_classifications_expected_episodes_count_check CHECK (kind = 'work' OR expected_episodes_count IS NULL)
);

-- One current classification per anime (the 1:1 relationship to animes).
--
-- [Ja] anime ごとに現在の分類は 1 つ (animes との 1:1 関係)。
CREATE UNIQUE INDEX index_anime_classifications_on_anime_id ON public.anime_classifications(anime_id);

-- For listing a parent work's episodes in order. Partial on
-- (parent_anime_id IS NOT NULL) so work rows are not indexed: per the CHECK
-- constraints a work always has a NULL parent_anime_id and NULL sort_number, so
-- it would only add a dead (NULL, NULL) entry that the parent_anime_id lookup
-- never matches.
--
-- [Ja] 親作品のエピソードを順番に一覧するため。(parent_anime_id IS NOT NULL) の
-- 部分インデックスとし、work 行は索引しない。CHECK 制約により work は常に
-- parent_anime_id が NULL・sort_number が NULL であり、索引しても
-- parent_anime_id の検索で決して引かれない (NULL, NULL) の死蔵エントリにしか
-- ならないため。
CREATE INDEX index_anime_classifications_on_parent_anime_id_and_sort_number ON public.anime_classifications(parent_anime_id, sort_number) WHERE parent_anime_id IS NOT NULL;

-- Lineage for duplicate merges only: resolves an old anime ID to its canonical
-- one. Re-classification keeps the ID and creates no redirect; only a merge
-- does. Chains are collapsed to a single hop (MusicBrainz gid_redirect style);
-- collapsing re-points an existing row's canonical_anime_id, which updated_at
-- records.
--
-- [Ja] 系譜 (重複統合専用): 旧 anime ID を canonical な ID へ解決する。再分類では
-- ID を維持しリダイレクトを作らず、重複統合のときだけ作る。チェーンは 1 ホップに
-- 畳む (MusicBrainz の gid_redirect 方式)。畳む際に既存行の canonical_anime_id を
-- 付け替えるため、その変更を updated_at で記録する。
CREATE TABLE public.anime_redirects (
    old_anime_id BIGINT PRIMARY KEY REFERENCES public.animes(id),
    canonical_anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- A redirect always points away from itself: old_anime_id and
    -- canonical_anime_id are never equal. A self-redirect is a degenerate row
    -- that carries no information and would make redirect resolution loop, so it
    -- is rejected at the DB level (cycles across multiple rows stay an app-side
    -- concern, collapsed to a single hop).
    --
    -- [Ja] リダイレクトは常に自分以外を指す: old_anime_id と canonical_anime_id が
    -- 一致することはない。自己リダイレクトは情報を持たない退化行で、リダイレクト解決を
    -- 無限ループさせるため DB レベルで弾く (複数行にまたがる循環は単一ホップへの
    -- 畳み込みでアプリ側が引き続き担う)。
    CONSTRAINT anime_redirects_no_self_redirect_check CHECK (old_anime_id <> canonical_anime_id)
);

-- Reverse lookup: find every old ID that resolves to a given canonical anime.
-- Collapsing a chain re-points the rows whose canonical_anime_id equals the
-- merged-away anime, so that bulk update scans by this column.
--
-- [Ja] 逆引き: ある canonical anime に解決される旧 ID をすべて引く。チェーンを畳む際は
-- canonical_anime_id が統合元 anime に一致する行をまとめて付け替えるため、その更新が
-- この列で走査する。
CREATE INDEX index_anime_redirects_on_canonical_anime_id ON public.anime_redirects(canonical_anime_id);

-- The season(s) an anime is listed under. The growing axis is the season the work
-- is shown in -- a single anime can belong to several (a split-cours title
-- resuming in a later season, or a title aired abroad first then in Japan) -- so
-- seasons are rows keyed by anime_id (layer 1 identity); an episode has no rows.
-- year is always set; name (the season_name enum) is nullable, where NULL means
-- "year known, season name TBD" (the legacy Season(year, "all") state). A fully
-- undetermined work simply has no row. is_primary marks the single representative
-- season for contexts that need one (the frozen v1 API season_name scalar,
-- sorting, compact display); it is editor-set because neither the earliest nor
-- the latest season is universally correct. Rows order by (year, name), so there
-- is no sort_number. See ADR 0017.
--
-- [Ja] アニメが掲載される季節。増える軸は「作品を出す季節」で、1 作品が複数季節に
-- 属しうる (分割放送で季節をまたいで再開、海外先行→日本放送など) ため、季節を行で
-- 持ち第 1 層 identity の anime_id をキーにする (エピソードは行を持たない)。year は
-- 必ず入り、name (season_name enum) は NULL 許容で、NULL は「年は判明・季節名は未定」
-- (旧 Season(year, "all")) を表す。完全未定の作品は単に行を持たない。is_primary は
-- 単一の代表季節が要る場面 (凍結 v1 API の season_name スカラ・ソート・コンパクト
-- 表示) 用で、最古でも最新でも一律には正しくないため編集者が設定する。並びは
-- (year, name) 昇順なので sort_number は持たない。ADR 0017 を参照。
CREATE TABLE public.anime_seasons (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    year INTEGER NOT NULL,
    name public.season_name,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- At most one row per (anime, year, name). NULLS NOT DISTINCT so a year with an
-- unset name (anime_id, year, NULL) is also unique per anime and not duplicable.
--
-- [Ja] (anime, year, name) ごとに高々 1 行。NULLS NOT DISTINCT により季節名未設定の
-- 年 (anime_id, year, NULL) も 1 作品 1 行に保ち、重複登録を防ぐ。
CREATE UNIQUE INDEX index_anime_seasons_on_anime_id_and_year_and_name ON public.anime_seasons(anime_id, year, name) NULLS NOT DISTINCT;

-- At most one primary season per anime (the representative for single-season
-- contexts). Exactly-one is ensured in app code.
--
-- [Ja] 1 作品の主季節 (単一季節を返す場面の代表) は高々 1 つ。必ず 1 つはアプリで担保。
CREATE UNIQUE INDEX index_anime_seasons_on_anime_id_primary ON public.anime_seasons(anime_id) WHERE is_primary;

-- For listing all anime in a given season ("2026 spring").
--
-- [Ja] ある季節の作品一覧 (「2026 春」) を引くため。
CREATE INDEX index_anime_seasons_on_year_and_name ON public.anime_seasons(year, name);

-- External IDs that map an anime to the same work in other anime databases
-- (Syobocal tid, MyAnimeList id, ...). Keyed by anime_id (layer 1 identity) so
-- the mapping stays stable across re-classification and an episode simply has no
-- rows. external_id is VARCHAR to stay uniform across services (the integer
-- Syobocal / MyAnimeList ids are stored as text and parsed in app code when
-- needed). See ADR 0012.
--
-- [Ja] anime を他のアニメ DB の同一作品へ対応づける外部 ID (Syobocal の tid、
-- MyAnimeList の id など)。第 1 層 identity の anime_id をキーにするため、再分類を
-- またいでも対応が安定し、エピソードは単に行を持たない。external_id はサービス横断で
-- 統一するため VARCHAR とする (integer の Syobocal / MyAnimeList の id も文字列で
-- 保持し、必要時にアプリ側でパースする)。ADR 0012 を参照。
CREATE TABLE public.anime_external_ids (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    service public.anime_external_service NOT NULL,
    external_id VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- At most one external ID per service per anime.
--
-- [Ja] anime ごと・サービスごとに外部 ID は高々 1 件。
CREATE UNIQUE INDEX index_anime_external_ids_on_anime_id_and_service ON public.anime_external_ids(anime_id, service);

-- Reverse lookup: resolve a service's external ID back to its anime (e.g. a
-- Syobocal tid during import). One composite index serves every service.
-- Intentionally not unique: one Syobocal title can map to several Annict works.
--
-- [Ja] 逆引き: サービスの外部 ID から anime を解決する (取り込み時に Syobocal の
-- tid から引くなど)。複合インデックス 1 本で全サービスに対応する。あえて非ユニーク:
-- 1 つの Syobocal title を複数の Annict 作品へ対応づけられるようにするため。
CREATE INDEX index_anime_external_ids_on_service_and_external_id ON public.anime_external_ids(service, external_id);

-- Display URLs for an anime (official site, Wikipedia, and 'other' links such as
-- X posts or news articles). The growing axis is the link kind, not language, so
-- kinds are rows. language is the language of the linked content (ja / en, or
-- 'other' for language-neutral / other-language links), defaulting to ja.
-- label / label_en are optional display labels shown per the viewer's UI language
-- and are independent of the content's language -- e.g. an English label
-- "Production announcement (in Japanese)" on a Japanese news link, for en-UI
-- viewers. No uniqueness: an anime may have several links of one kind (e.g.
-- multiple Wikipedia pages); order by sort_number. Video links live in the
-- trailers table, not here. See ADR 0013.
--
-- [Ja] anime の表示用 URL (公式サイト・Wikipedia と、X のポストやニュース記事などの
-- 'other')。増える軸は言語でなくリンクの種類なので、種類を行で持つ。language はリンク先
-- コンテンツの言語 (ja / en、言語に紐づかない/その他言語は 'other') で、既定は ja。
-- label / label_en は閲覧者の UI 言語に応じて出す任意の表示ラベルで、コンテンツの言語とは
-- 独立する -- 例: 日本語のニュースリンクに、英語 UI の利用者向けに英語ラベル
-- "Production announcement (in Japanese)" を付ける。ユニーク制約は無く、1 作品が同種別の
-- リンクを複数持てる (Wikipedia 複数など)。並びは sort_number。動画リンクはここでなく
-- trailers テーブルが担う。ADR 0013 を参照。
CREATE TABLE public.anime_links (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    kind public.anime_link_kind NOT NULL,
    language public.language NOT NULL DEFAULT 'ja',
    url VARCHAR NOT NULL,
    label VARCHAR,
    label_en VARCHAR,
    sort_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- For listing an anime's links grouped by kind.
--
-- [Ja] anime のリンクを種類ごとに一覧するため。
CREATE INDEX index_anime_links_on_anime_id_and_kind ON public.anime_links(anime_id, kind);

-- Anime-level dates for a calendar / timeline (broadcast period, revival
-- screenings, and 'other' events). The growing axis is the event kind, so kinds
-- are rows. started_on is a point event's date or a period's start; ended_on is
-- the period end (NULL for point events). title is the entry's calendar title
-- and description its summary; title_en / description_en hold their English
-- versions for en display (like animes synopsis / synopsis_en). No uniqueness:
-- an anime may have several dates of one kind (e.g. multiple revival
-- screenings); order by sort_number. These are coarse, anime-level dates -- the
-- per-episode / per-channel schedule lives in slots. See ADR 0014.
--
-- [Ja] 作品単位のカレンダー / タイムライン用の日付 (放送期間・復刻上映・'other')。
-- 増える軸はイベント種別なので種別を行で持つ。started_on は点イベントの日付または
-- 期間の開始日、ended_on は期間の終了日 (点イベントは NULL)。title はカレンダーの
-- タイトル、description はその概要。title_en / description_en は en 表示用の英語版
-- (animes の synopsis / synopsis_en と同様)。
-- ユニーク制約は無く、1 作品が同種別の日付を複数持てる (復刻上映が複数回など)。
-- 並びは sort_number。これは作品単位の粗い日付で、各話×チャンネルのスケジュールは
-- slots が持つ。ADR 0014 を参照。
CREATE TABLE public.anime_events (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    kind public.anime_event_kind NOT NULL,
    started_on DATE NOT NULL,
    ended_on DATE,
    title VARCHAR,
    title_en VARCHAR,
    description TEXT,
    description_en TEXT,
    sort_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- For listing an anime's events grouped by kind.
--
-- [Ja] anime のイベントを種類ごとに一覧するため。
CREATE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);

-- For date-range calendar queries across animes (e.g. this month's events).
--
-- [Ja] 作品横断の日付範囲カレンダークエリ用 (今月のイベントなど)。
CREATE INDEX index_anime_events_on_started_on ON public.anime_events(started_on);

-- An anime's official accounts on external services (X, YouTube, LINE, ...) --
-- mostly social media, but the table is not limited to social services. The
-- growing axis is the platform, so platforms are rows (the service enum). account
-- holds the bare handle that uniquely identifies the account on that service; the
-- URL / @-mention are derived per service in app code, which avoids turning the
-- editor's handle input into a URL input. label / label_en are optional notes
-- (e.g. "anime official" vs "original-work official") to tell apart multiple
-- accounts on one service. No uniqueness: an anime may have several accounts on
-- one service; order by sort_number. Hashtags are not here -- they are
-- platform-agnostic and live in anime_hashtags. See ADR 0015.
--
-- [Ja] アニメの各種サービスの公式アカウント (X / YouTube / LINE など。多くは
-- ソーシャルメディアだが、それに限らない)。増える軸はプラットフォームなので、
-- プラットフォームを行で持つ (service enum)。account はそのサービスでアカウントを
-- 一意に識別する素のハンドルを持ち、URL や @メンションはサービスごとにアプリで導出
-- する (編集 UI のハンドル入力を URL 入力に変える副作用を避ける)。label / label_en は
-- 同一サービスに複数アカウントがあるときの区別用の任意の補足 (「アニメ公式」「原作
-- 公式」など)。ユニーク制約は無く、1 作品が同サービスに複数アカウントを持てる。並びは
-- sort_number。ハッシュタグはプラットフォーム非依存なのでここでなく anime_hashtags
-- が持つ。ADR 0015 を参照。
CREATE TABLE public.anime_official_accounts (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    service public.anime_account_service NOT NULL,
    account VARCHAR NOT NULL,
    label VARCHAR,
    label_en VARCHAR,
    sort_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- For listing an anime's accounts grouped by service, and for picking one
-- service (e.g. the X row for v1 API compatibility).
--
-- [Ja] anime のアカウントをサービスごとに一覧する用、および 1 サービスを引く用
-- (v1 API 互換で X 行を選ぶなど)。
CREATE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);

-- An anime's hashtags (e.g. #リゼロ, #rezero). A hashtag is a plain,
-- platform-agnostic string used across X, Bluesky, etc., independent of whether
-- the anime has an account on any of them, so there is no service column (which
-- would be NULL for hashtag-less platforms). hashtag is stored without the
-- leading '#' (added on display), matching the legacy works.twitter_hashtag. An
-- anime may have several hashtags; the primary one is sort_number 0. See ADR 0015.
--
-- [Ja] アニメのハッシュタグ (#リゼロ / #rezero など)。ハッシュタグは X や Bluesky
-- などで横断的に使われるプラットフォーム非依存の素の文字列で、アニメがそれらに
-- アカウントを持つかと独立しているため service 列は持たない (持つとハッシュタグの無い
-- プラットフォームで NULL になる)。hashtag は先頭 '#' なしで保存し (表示時に付与)、旧
-- works.twitter_hashtag を踏襲する。1 作品が複数のハッシュタグを持て、主タグは
-- sort_number 0。ADR 0015 を参照。
CREATE TABLE public.anime_hashtags (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    hashtag VARCHAR NOT NULL,
    sort_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- At most one of each hashtag per anime; the anime_id-leading index also serves
-- listing an anime's hashtags.
--
-- [Ja] anime ごとに同じハッシュタグは高々 1 件。anime_id が先頭列なので anime の
-- ハッシュタグ一覧の引きも兼ねる。
CREATE UNIQUE INDEX index_anime_hashtags_on_anime_id_and_hashtag ON public.anime_hashtags(anime_id, hashtag);

-- migrate:down

DROP TABLE IF EXISTS public.anime_hashtags;
DROP TABLE IF EXISTS public.anime_official_accounts;
DROP TABLE IF EXISTS public.anime_events;
DROP TABLE IF EXISTS public.anime_links;
DROP TABLE IF EXISTS public.anime_external_ids;
DROP TABLE IF EXISTS public.anime_seasons;
DROP TABLE IF EXISTS public.anime_redirects;
DROP TABLE IF EXISTS public.anime_classifications;
DROP TABLE IF EXISTS public.animes;

DROP TYPE IF EXISTS public.anime_account_service;
DROP TYPE IF EXISTS public.anime_event_kind;
DROP TYPE IF EXISTS public.language;
DROP TYPE IF EXISTS public.anime_link_kind;
DROP TYPE IF EXISTS public.anime_external_service;
DROP TYPE IF EXISTS public.release_status;
DROP TYPE IF EXISTS public.season_name;
DROP TYPE IF EXISTS public.anime_media;
DROP TYPE IF EXISTS public.anime_classification_kind;
DROP TYPE IF EXISTS public.anime_status;
