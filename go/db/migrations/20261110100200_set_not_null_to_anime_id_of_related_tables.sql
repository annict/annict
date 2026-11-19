-- migrate:up

-- 関連テーブルのanime_idにNOT NULLを付け、検証に使ったCHECK制約を外す。
-- 検証済みのCHECK (anime_id IS NOT NULL) があるため、SET NOT NULLは全行を走査しない。
ALTER TABLE public.casts ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.casts DROP CONSTRAINT casts_anime_id_not_null;
ALTER TABLE public.staffs ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.staffs DROP CONSTRAINT staffs_anime_id_not_null;
ALTER TABLE public.trailers ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.trailers DROP CONSTRAINT trailers_anime_id_not_null;
ALTER TABLE public.work_images ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.work_images DROP CONSTRAINT work_images_anime_id_not_null;
ALTER TABLE public.slots ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.slots DROP CONSTRAINT slots_anime_id_not_null;
ALTER TABLE public.programs ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.programs DROP CONSTRAINT programs_anime_id_not_null;
ALTER TABLE public.series_works ALTER COLUMN anime_id SET NOT NULL;
ALTER TABLE public.series_works DROP CONSTRAINT series_works_anime_id_not_null;

-- migrate:down

-- validate_anime_id_constraints_of_related_tablesを適用したあとの状態 (検証済みのCHECK制約があり、
-- 列はNULLを許す) に戻す。CHECK制約は検証しながら足すが、NOT NULLの列には空の行が無いため失敗しない。
ALTER TABLE public.series_works
    ADD CONSTRAINT series_works_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.series_works ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.programs
    ADD CONSTRAINT programs_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.programs ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.slots
    ADD CONSTRAINT slots_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.slots ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.work_images
    ADD CONSTRAINT work_images_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.work_images ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.trailers
    ADD CONSTRAINT trailers_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.trailers ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.staffs
    ADD CONSTRAINT staffs_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.staffs ALTER COLUMN anime_id DROP NOT NULL;
ALTER TABLE public.casts
    ADD CONSTRAINT casts_anime_id_not_null CHECK (anime_id IS NOT NULL);
ALTER TABLE public.casts ALTER COLUMN anime_id DROP NOT NULL;
