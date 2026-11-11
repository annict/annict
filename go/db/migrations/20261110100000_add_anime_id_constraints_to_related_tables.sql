-- migrate:up

-- 関連テーブルのanime_idに、animes.idへの外部キーと、NOT NULLの検証に使うCHECK制約を足す。
-- 既存の行は走査せず新しい行と変えた行だけを検証するようNOT VALIDで足し、既存の行の検証は
-- 次のマイグレーションのVALIDATE CONSTRAINTで行う。外部キーのON DELETEは、写し元の
-- works.anime_id / episodes.anime_idの外部キーと同じく指定しない。

ALTER TABLE public.casts
    ADD CONSTRAINT casts_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.staffs
    ADD CONSTRAINT staffs_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.trailers
    ADD CONSTRAINT trailers_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.work_images
    ADD CONSTRAINT work_images_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.vod_titles
    ADD CONSTRAINT vod_titles_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.slots
    ADD CONSTRAINT slots_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.slots
    ADD CONSTRAINT slots_episode_anime_id_fkey FOREIGN KEY (episode_anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.programs
    ADD CONSTRAINT programs_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;
ALTER TABLE public.series_works
    ADD CONSTRAINT series_works_anime_id_fkey FOREIGN KEY (anime_id) REFERENCES public.animes(id) NOT VALID;

-- vod_titles.anime_id (work_idがNULLの行がある) とslots.episode_anime_id (エピソードが紐づくまで空) は
-- NULLを許すため、CHECK制約を足さない。
ALTER TABLE public.casts
    ADD CONSTRAINT casts_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.staffs
    ADD CONSTRAINT staffs_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.trailers
    ADD CONSTRAINT trailers_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.work_images
    ADD CONSTRAINT work_images_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.slots
    ADD CONSTRAINT slots_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.programs
    ADD CONSTRAINT programs_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;
ALTER TABLE public.series_works
    ADD CONSTRAINT series_works_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID;

-- migrate:down

ALTER TABLE public.series_works DROP CONSTRAINT series_works_anime_id_not_null;
ALTER TABLE public.programs DROP CONSTRAINT programs_anime_id_not_null;
ALTER TABLE public.slots DROP CONSTRAINT slots_anime_id_not_null;
ALTER TABLE public.work_images DROP CONSTRAINT work_images_anime_id_not_null;
ALTER TABLE public.trailers DROP CONSTRAINT trailers_anime_id_not_null;
ALTER TABLE public.staffs DROP CONSTRAINT staffs_anime_id_not_null;
ALTER TABLE public.casts DROP CONSTRAINT casts_anime_id_not_null;
ALTER TABLE public.series_works DROP CONSTRAINT series_works_anime_id_fkey;
ALTER TABLE public.programs DROP CONSTRAINT programs_anime_id_fkey;
ALTER TABLE public.slots DROP CONSTRAINT slots_episode_anime_id_fkey;
ALTER TABLE public.slots DROP CONSTRAINT slots_anime_id_fkey;
ALTER TABLE public.vod_titles DROP CONSTRAINT vod_titles_anime_id_fkey;
ALTER TABLE public.work_images DROP CONSTRAINT work_images_anime_id_fkey;
ALTER TABLE public.trailers DROP CONSTRAINT trailers_anime_id_fkey;
ALTER TABLE public.staffs DROP CONSTRAINT staffs_anime_id_fkey;
ALTER TABLE public.casts DROP CONSTRAINT casts_anime_id_fkey;
