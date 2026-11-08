-- migrate:up

-- 作品・エピソードのメタデータを持つテーブルの行から、指す作品・エピソードのanimeを
-- 直接引けるようにする列。値はwork_id / episode_idから導き、Rails版の保存とanimes同期
-- バッチが埋める。既存の行を埋め終えるまではNULLの行が残るため、NULL許容とし、外部キーも
-- 付けない。NULL許容・既定値なしの列の追加はテーブルを書き換えないため、ロックは短い。
--
-- slotsは作品とエピソードの両方を指すため、作品のanimeをanime_idに、紐づいたエピソードの
-- animeをepisode_anime_idに写す。
ALTER TABLE public.casts ADD COLUMN anime_id BIGINT;
ALTER TABLE public.staffs ADD COLUMN anime_id BIGINT;
ALTER TABLE public.trailers ADD COLUMN anime_id BIGINT;
ALTER TABLE public.work_images ADD COLUMN anime_id BIGINT;
ALTER TABLE public.vod_titles ADD COLUMN anime_id BIGINT;
ALTER TABLE public.slots ADD COLUMN anime_id BIGINT;
ALTER TABLE public.slots ADD COLUMN episode_anime_id BIGINT;
ALTER TABLE public.programs ADD COLUMN anime_id BIGINT;
ALTER TABLE public.series_works ADD COLUMN anime_id BIGINT;

-- migrate:down

ALTER TABLE public.series_works DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.programs DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.slots DROP COLUMN IF EXISTS episode_anime_id;
ALTER TABLE public.slots DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.vod_titles DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.work_images DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.trailers DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.staffs DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.casts DROP COLUMN IF EXISTS anime_id;
