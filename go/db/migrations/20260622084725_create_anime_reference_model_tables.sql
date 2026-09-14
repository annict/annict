-- migrate:up

-- anime_status: コンテンツ同一性のライフサイクル (いずれもソフトで、行は
-- 物理削除しない)。'archived' は新規の記録/活動を凍結するがページは直接閲覧でき、
-- 利用者は既に付けた記録に引き続きアクセスできる (編集者も設定可)。'merged' は
-- 重複統合の墓標でanime_redirects経由で統合先へ解決する。'deleted' は
-- サイト上から完全に非表示にする (管理者のみ)。archive_messageはarchivedの
-- ページに出す任意の案内文を持つ。
CREATE TYPE public.anime_status AS ENUM ('published', 'archived', 'merged', 'deleted');

CREATE TYPE public.anime_classification_kind AS ENUM ('work', 'episode');

-- anime_mediaは旧Work#mediaの値 (Rails enumerizeの
-- {tv, ova, movie, web, other}) を、'web' のみ 'ona' に改名して引き継ぐ (各
-- アニメDBで一般的な "Original Net Animation" の呼称に合わせる)。フェーズ2の
-- 同期はweb (4) を 'ona' に、残りのコードは同名で写像する。'other' (0) は
-- すべての旧行を表現できるよう残す。"Special" はここに値として持たない: 特別編で
-- あることは媒体と直交し (TV放送の特別編はTVでもspecialでもある)、媒体enum
-- ではなく別属性で表すべきものだから。season_nameはSeasonの名前コード
-- (Season::NAME_HASH) を引き継ぎ、'autumn' を 'fall' に改名する (各アニメDB
-- (AniList / MyAnimeList / Kitsu) で一般的な季節表記に合わせる)。フェーズ2の
-- 同期はautumn (4) を 'fall' に、残りは同名で写像する。
CREATE TYPE public.anime_media AS ENUM ('tv', 'ova', 'movie', 'ona', 'other');
CREATE TYPE public.season_name AS ENUM ('winter', 'spring', 'summer', 'fall');

-- release_statusは作品の放送/公開ライフサイクル。格納するのは粗い3状態
-- だけで、'not_yet_released' (放送前)・'released' (放送開始済み)・'cancelled'
-- (製作中止。編集者が設定し、バックフィル元は無い)。「放送中 / 完結 / 再放送中」の
-- 細かい区別は格納せず読み取り時にslotから導出するため、格納遷移は
-- not_yet_released -> releasedの一方向で済む。季節欠落と組み合わせて「未定」
-- (not_yet_released) と「不明」(released) の表示を導出する。主に作品で使うが、
-- エピソードに設定してもよい (将来作品に昇格する場合に備えるなど)。ただし
-- エピソードでは積極的には活用しない。work限定の制約は設けず、単なるNULL許容の
-- 任意カラムとする。ADR 0010を参照。
CREATE TYPE public.release_status AS ENUM ('not_yet_released', 'released', 'cancelled');

-- anime_external_serviceはanime_external_idsの各行が指す外部アニメDBを
-- 表す (現状はSyobocalのtitle DBとMyAnimeList。以降は値追加で増やす)。
-- サービスの追加はALTER TYPE ADD VALUE 1行 + 行追加で済み、スキーマ列を増やさない。
CREATE TYPE public.anime_external_service AS ENUM ('syobocal', 'mal');

-- anime_link_kindはanime_linksの表示用URLの種別。'official_site' /
-- 'wikipedia' / 'other' で開始し、Xのポスト・ニュース記事などは当面 'other' に入れ、
-- 分類できてきたら個別の値に切り出す。動画リンクはここに入れずtrailersテーブルが
-- 担う (サムネイル/プロバイダー処理があるため)。ADR 0013を参照。
CREATE TYPE public.anime_link_kind AS ENUM ('official_site', 'wikipedia', 'other');

-- languageはコンテンツが何語かを表す: ja / enに加え、それ以外の言語や言語中立は
-- 'other'。RailsのApplicationRecord::LOCALESと同じ集合で、行がコンテンツの言語を持つ
-- 箇所で再利用できる。利用者の表示 / UI言語 (常にjaかenで 'other' は取らない) は
-- ここでなく別enumに置く。
CREATE TYPE public.language AS ENUM ('ja', 'en', 'other');

-- anime_event_kindはanime_eventsの日付イベントの種別で、作品のカレンダー /
-- タイムライン用。'broadcast' (放送期間)・'revival_screening'・'other' で開始し、
-- 必要に応じて種別を増やす。これは作品単位の粗い日付で、slots (各話×チャンネルの
-- 放送枠) とは別物。ADR 0014を参照。
CREATE TYPE public.anime_event_kind AS ENUM ('broadcast', 'revival_screening', 'other');

-- anime_account_serviceはanime_official_accountsの各行が指すアニメ公式
-- アカウントのプラットフォームを表す (多くはソーシャルメディアだが、ソーシャルに
-- 限らない)。プラットフォームの追加はALTER TYPE ADD VALUE 1行 + 行追加で済む。
-- 多くのプラットフォームはaccount列に単一のハンドルを入れるが (x / youtube / line
-- など)、ActivityPub系の分散型 (mastodon) はインスタンスのドメインまで含めて初めて
-- 一意になるため、フルのwebfingerハンドルuser@instanceを入れ、URLはそれを
-- パースして導出する。ADR 0015を参照。
CREATE TYPE public.anime_account_service AS ENUM ('bluesky', 'instagram', 'line', 'mastodon', 'mixi2', 'threads', 'tiktok', 'x', 'youtube');

-- 第1層: コンテンツ同一性。公開ID (新URL / API v2) はここに張り、再分類・
-- 統合・分離を通じて永続する。
CREATE TABLE public.animes (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR,
    title_kana VARCHAR,
    -- title_roはローマ字タイトル、title_enは英語タイトル。別物であり、
    -- works / episodesと同様に分けて持つ。
    title_ro VARCHAR,
    title_en VARCHAR,
    title_alter VARCHAR,
    title_alter_ro VARCHAR,
    title_alter_en VARCHAR,
    -- ja / en以外の言語のタイトル (中国語など) の受け皿。「、」などで区切って
    -- 羅列するフリーテキストで、区切りは厳密でない。サイトの表示はja / enのみのため
    -- 構造化表示はせず、検索・識別用に留める。ADR 0013を参照。
    title_alter_other VARCHAR,
    media public.anime_media,
    -- 主に作品で使う (エピソードにも設定可だが通常は使わない)。詳細は
    -- release_status型のコメントを参照。
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

-- animesのシーケンスは、現在のworks.id / episodes.idの最大値より大きい
-- 「次の10の冪」のすぐ上から開始する。つまり新しいanime IDは、旧IDの最大値を
-- 1桁上の位に切り上げた値に1を足したものになる (例: 旧ID最大が180,372なら
-- 最初のanime IDは1,000,001)。これによりv2のURL / APIで公開される公開ID
-- (animes.id) を短く、かつ10の冪というきれいな境界に揃えつつ、すべてのanime IDを
-- 旧IDの範囲より上に置いて両ID空間を重ならせない。length(max::text) は最大値の
-- 桁数、power(10, 桁数) がその次の10の冪で、setval(..., true) により最初のnextval
-- が冪 + 1を返す。
SELECT setval(
    pg_get_serial_sequence('public.animes', 'id'),
    power(10, length(GREATEST(
        (SELECT COALESCE(MAX(id), 0) FROM public.works),
        (SELECT COALESCE(MAX(id), 0) FROM public.episodes)
    )::text))::bigint,
    true
);

-- 第2層: 分類。現在の分類のみを持つ。分類の履歴はAnnict DBの編集履歴に
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
    -- workは親を持たず、episodeは必ず親を持つ。
    CONSTRAINT anime_classifications_parent_check CHECK ((kind = 'work') = (parent_anime_id IS NULL)),
    -- standaloneを立てられるのはworkだけ (映画などそれ自体が視聴単位の
    -- 作品)。episodeはこのフラグによらず常に視聴単位。
    CONSTRAINT anime_classifications_standalone_check CHECK (kind = 'work' OR NOT standalone),
    -- numberはエピソードの数値の話数 (3、第3話と第4話の間の総集編なら3.5)、
    -- 番号を持たないspecialはNULL。number_textは表示文字列 (「第3話」「総集編」)。
    -- どちらもepisodeだけが持ち、workは持たない。ADR 0018を参照。
    CONSTRAINT anime_classifications_number_check CHECK (kind = 'episode' OR number IS NULL),
    CONSTRAINT anime_classifications_number_text_check CHECK (kind = 'episode' OR number_text IS NULL),
    -- episodeは親内で必ず並び順を持ち (sort_numberあり)、workは持たない
    -- (sort_number NULL)。parent_anime_idのCHECK制約と同じ双方向の形。
    CONSTRAINT anime_classifications_sort_number_check CHECK ((kind = 'episode') = (sort_number IS NOT NULL)),
    -- number_format_idはwork限定の設定 (この作品の生成エピソードをどの
    -- number_formats行で採番するか) で、episodeは持ってはいけない。standaloneと
    -- 同じくwork限定なので片方向の形: workは持ってもよいがNULL許容のまま。
    -- ADR 0016を参照。
    CONSTRAINT anime_classifications_number_format_id_check CHECK (kind = 'work' OR number_format_id IS NULL),
    -- episode_start_numberはwork限定の生成設定で、先頭の生成エピソードに与える
    -- 話数の起点 (例: 2期で13から振るなら13)。旧works.start_episode_raw_numberを
    -- 改名。episodeは持たない。standaloneと同じwork限定の片方向。ADR 0018を参照。
    CONSTRAINT anime_classifications_episode_start_number_check CHECK (kind = 'work' OR episode_start_number IS NULL),
    -- expected_episodes_countは編集者が宣言する作品の予定総話数 (旧
    -- works.manual_episodes_countを改名)。エピソード生成の打ち切りと全話完了の判定に
    -- 使い、episodeは持たない。standaloneと同じくwork限定で片方向の形 (workでも
    -- NULL許容)。ADR 0016を参照。
    CONSTRAINT anime_classifications_expected_episodes_count_check CHECK (kind = 'work' OR expected_episodes_count IS NULL)
);

-- animeごとに現在の分類は1つ (animesとの1:1関係)。
CREATE UNIQUE INDEX index_anime_classifications_on_anime_id ON public.anime_classifications(anime_id);

-- 親作品のエピソードを順番に一覧するため。(parent_anime_id IS NOT NULL) の
-- 部分インデックスとし、work行は索引しない。CHECK制約によりworkは常に
-- parent_anime_idがNULL・sort_numberがNULLであり、索引しても
-- parent_anime_idの検索で決して引かれない (NULL, NULL) の死蔵エントリにしか
-- ならないため。
CREATE INDEX index_anime_classifications_on_parent_anime_id_and_sort_number ON public.anime_classifications(parent_anime_id, sort_number) WHERE parent_anime_id IS NOT NULL;

-- 系譜 (重複統合専用): 旧anime IDを統合先のIDへ解決する。再分類では
-- IDを維持しリダイレクトを作らず、重複統合のときだけ作る。チェーンは1ホップに
-- 畳む (MusicBrainzのgid_redirect方式)。畳む際に既存行のcanonical_anime_idを
-- 付け替えるため、その変更をupdated_atで記録する。
CREATE TABLE public.anime_redirects (
    old_anime_id BIGINT PRIMARY KEY REFERENCES public.animes(id),
    canonical_anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- リダイレクトは常に自分以外を指す: old_anime_idとcanonical_anime_idが
    -- 一致することはない。自己リダイレクトは情報を持たない退化行で、リダイレクト解決を
    -- 無限ループさせるためDBレベルで弾く (複数行にまたがる循環は単一ホップへの
    -- 畳み込みでアプリ側が引き続き担う)。
    CONSTRAINT anime_redirects_no_self_redirect_check CHECK (old_anime_id <> canonical_anime_id)
);

-- 逆引き: ある統合先のanimeに解決される旧IDをすべて引く。チェーンを畳む際は
-- canonical_anime_idが統合元animeに一致する行をまとめて付け替えるため、その更新が
-- この列で走査する。
CREATE INDEX index_anime_redirects_on_canonical_anime_id ON public.anime_redirects(canonical_anime_id);

-- アニメが掲載される季節。増える軸は「作品を出す季節」で、1作品が複数季節に
-- 属しうる (分割放送で季節をまたいで再開、海外先行→日本放送など) ため、季節を行で
-- 持ち第1層 (同一性) のanime_idをキーにする (エピソードは行を持たない)。yearは
-- 必ず入り、name (season_name enum) はNULL許容で、NULLは「年は判明・季節名は未定」
-- (旧Season(year, "all")) を表す。完全未定の作品は単に行を持たない。is_primaryは
-- 単一の代表季節が要る場面 (凍結v1 APIのseason_nameスカラ・ソート・コンパクト
-- 表示) 用で、最古でも最新でも一律には正しくないため編集者が設定する。並びは
-- (year, name) 昇順なのでsort_numberは持たない。ADR 0017を参照。
CREATE TABLE public.anime_seasons (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    year INTEGER NOT NULL,
    name public.season_name,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- (anime, year, name) ごとに高々1行。NULLS NOT DISTINCTにより季節名未設定の
-- 年 (anime_id, year, NULL) も1作品1行に保ち、重複登録を防ぐ。
CREATE UNIQUE INDEX index_anime_seasons_on_anime_id_and_year_and_name ON public.anime_seasons(anime_id, year, name) NULLS NOT DISTINCT;

-- 1作品の主季節 (単一季節を返す場面の代表) は高々1つ。必ず1つはアプリで担保。
CREATE UNIQUE INDEX index_anime_seasons_on_anime_id_primary ON public.anime_seasons(anime_id) WHERE is_primary;

-- ある季節の作品一覧 (「2026春」) を引くため。
CREATE INDEX index_anime_seasons_on_year_and_name ON public.anime_seasons(year, name);

-- animeを他のアニメDBの同一作品へ対応づける外部ID (Syobocalのtid、
-- MyAnimeListのidなど)。第1層 (同一性) のanime_idをキーにするため、再分類を
-- またいでも対応が安定し、エピソードは単に行を持たない。external_idはサービス横断で
-- 統一するためVARCHARとする (integerのSyobocal / MyAnimeListのidも文字列で
-- 保持し、必要時にアプリ側でパースする)。ADR 0012を参照。
CREATE TABLE public.anime_external_ids (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    service public.anime_external_service NOT NULL,
    external_id VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- animeごと・サービスごとに外部IDは高々1件。
CREATE UNIQUE INDEX index_anime_external_ids_on_anime_id_and_service ON public.anime_external_ids(anime_id, service);

-- 逆引き: サービスの外部IDからanimeを解決する (取り込み時にSyobocalの
-- tidから引くなど)。複合インデックス1本で全サービスに対応する。あえて非ユニーク:
-- 1つのSyobocal titleを複数のAnnict作品へ対応づけられるようにするため。
CREATE INDEX index_anime_external_ids_on_service_and_external_id ON public.anime_external_ids(service, external_id);

-- animeの表示用URL (公式サイト・Wikipediaと、Xのポストやニュース記事などの
-- 'other')。増える軸は言語でなくリンクの種類なので、種類を行で持つ。languageはリンク先
-- コンテンツの言語 (ja / en、言語に紐づかない/その他言語は 'other') で、既定はja。
-- label / label_enは閲覧者のUI言語に応じて出す任意の表示ラベルで、コンテンツの言語とは
-- 独立する -- 例: 日本語のニュースリンクに、英語UIの利用者向けに英語ラベル
-- "Production announcement (in Japanese)" を付ける。ユニーク制約は無く、1作品が同種別の
-- リンクを複数持てる (Wikipedia複数など)。並びはsort_number。動画リンクはここでなく
-- trailersテーブルが担う。ADR 0013を参照。
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

-- animeのリンクを種類ごとに一覧するため。
CREATE INDEX index_anime_links_on_anime_id_and_kind ON public.anime_links(anime_id, kind);

-- 作品単位のカレンダー / タイムライン用の日付 (放送期間・復刻上映・'other')。
-- 増える軸はイベント種別なので種別を行で持つ。started_onは点イベントの日付または
-- 期間の開始日、ended_onは期間の終了日 (点イベントはNULL)。titleはカレンダーの
-- タイトル、descriptionはその概要。title_en / description_enはen表示用の英語版
-- (animesのsynopsis / synopsis_enと同様)。
-- ユニーク制約は無く、1作品が同種別の日付を複数持てる (復刻上映が複数回など)。
-- 並びはsort_number。これは作品単位の粗い日付で、各話×チャンネルのスケジュールは
-- slotsが持つ。ADR 0014を参照。
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

-- animeのイベントを種類ごとに一覧するため。
CREATE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);

-- 作品横断の日付範囲カレンダークエリ用 (今月のイベントなど)。
CREATE INDEX index_anime_events_on_started_on ON public.anime_events(started_on);

-- アニメの各種サービスの公式アカウント (X / YouTube / LINEなど。多くは
-- ソーシャルメディアだが、それに限らない)。増える軸はプラットフォームなので、
-- プラットフォームを行で持つ (service enum)。accountはそのサービスでアカウントを
-- 一意に識別する素のハンドルを持ち、URLや @メンションはサービスごとにアプリで導出
-- する (編集UIのハンドル入力をURL入力に変える副作用を避ける)。label / label_enは
-- 同一サービスに複数アカウントがあるときの区別用の任意の補足 (「アニメ公式」「原作
-- 公式」など)。ユニーク制約は無く、1作品が同サービスに複数アカウントを持てる。並びは
-- sort_number。ハッシュタグはプラットフォーム非依存なのでここでなくanime_hashtags
-- が持つ。ADR 0015を参照。
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

-- animeのアカウントをサービスごとに一覧する用、および1サービスを引く用
-- (v1 API互換でX行を選ぶなど)。
CREATE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);

-- アニメのハッシュタグ (#リゼロ / #rezeroなど)。ハッシュタグはXやBluesky
-- などで横断的に使われるプラットフォーム非依存の素の文字列で、アニメがそれらに
-- アカウントを持つかと独立しているためservice列は持たない (持つとハッシュタグの無い
-- プラットフォームでNULLになる)。hashtagは先頭 '#' なしで保存し (表示時に付与)、旧
-- works.twitter_hashtagを踏襲する。1作品が複数のハッシュタグを持て、主タグは
-- sort_number 0。ADR 0015を参照。
CREATE TABLE public.anime_hashtags (
    id BIGSERIAL PRIMARY KEY,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id),
    hashtag VARCHAR NOT NULL,
    sort_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- animeごとに同じハッシュタグは高々1件。anime_idが先頭列なのでanimeの
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
