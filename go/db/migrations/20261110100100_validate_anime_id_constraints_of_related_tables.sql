-- migrate:up

-- add_anime_id_constraints_to_related_tablesでNOT VALIDとして足した制約を検証する。
-- 追加とは別のトランザクションで行い、ロックをSHARE UPDATE EXCLUSIVEにとどめて、既存の行の走査中も
-- 行の読み書きを止めない。
ALTER TABLE public.casts VALIDATE CONSTRAINT casts_anime_id_fkey;
ALTER TABLE public.staffs VALIDATE CONSTRAINT staffs_anime_id_fkey;
ALTER TABLE public.trailers VALIDATE CONSTRAINT trailers_anime_id_fkey;
ALTER TABLE public.work_images VALIDATE CONSTRAINT work_images_anime_id_fkey;
ALTER TABLE public.vod_titles VALIDATE CONSTRAINT vod_titles_anime_id_fkey;
ALTER TABLE public.slots VALIDATE CONSTRAINT slots_anime_id_fkey;
ALTER TABLE public.slots VALIDATE CONSTRAINT slots_episode_anime_id_fkey;
ALTER TABLE public.programs VALIDATE CONSTRAINT programs_anime_id_fkey;
ALTER TABLE public.series_works VALIDATE CONSTRAINT series_works_anime_id_fkey;
ALTER TABLE public.casts VALIDATE CONSTRAINT casts_anime_id_not_null;
ALTER TABLE public.staffs VALIDATE CONSTRAINT staffs_anime_id_not_null;
ALTER TABLE public.trailers VALIDATE CONSTRAINT trailers_anime_id_not_null;
ALTER TABLE public.work_images VALIDATE CONSTRAINT work_images_anime_id_not_null;
ALTER TABLE public.slots VALIDATE CONSTRAINT slots_anime_id_not_null;
ALTER TABLE public.programs VALIDATE CONSTRAINT programs_anime_id_not_null;
ALTER TABLE public.series_works VALIDATE CONSTRAINT series_works_anime_id_not_null;

-- migrate:down

-- 検証済みの制約をNOT VALIDに戻す手段は無く、戻す必要も無いため何もしない。
