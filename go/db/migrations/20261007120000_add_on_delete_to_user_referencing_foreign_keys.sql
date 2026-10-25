-- migrate:up

-- ユーザーの削除を妨げている外部キーに、ON DELETEの挙動を明示する。
-- 制約の検証 (既存行の走査) は書き込みを長くブロックしないよう、NOT VALIDで追加したうえで
-- 次のマイグレーションのVALIDATE CONSTRAINTで行う。

-- 作品画像・キャラクター画像は作品データの一部として残し、投稿者との紐づきだけを外す。
ALTER TABLE public.work_images ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE public.work_images DROP CONSTRAINT fk_rails_3bad625b28;
ALTER TABLE public.work_images
    ADD CONSTRAINT work_images_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL NOT VALID;

ALTER TABLE public.character_images ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE public.character_images DROP CONSTRAINT fk_rails_057f9798e7;
ALTER TABLE public.character_images
    ADD CONSTRAINT character_images_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL NOT VALID;

-- ユーザー単位のフィーチャーフラグは本人に従属するデータのため、本人と一緒に消す。
ALTER TABLE public.feature_flags DROP CONSTRAINT feature_flags_user_id_fkey;
ALTER TABLE public.feature_flags
    ADD CONSTRAINT feature_flags_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

-- reactionsは参照先のユーザー・コレクション項目に従属するデータのため、参照先と一緒に消す。
ALTER TABLE public.reactions DROP CONSTRAINT fk_rails_9f02fc96a0;
ALTER TABLE public.reactions
    ADD CONSTRAINT reactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.reactions DROP CONSTRAINT fk_rails_b76961a906;
ALTER TABLE public.reactions
    ADD CONSTRAINT reactions_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public.users(id) ON DELETE CASCADE NOT VALID;

ALTER TABLE public.reactions DROP CONSTRAINT fk_rails_53df387e77;
ALTER TABLE public.reactions
    ADD CONSTRAINT reactions_collection_item_id_fkey FOREIGN KEY (collection_item_id) REFERENCES public.collection_items(id) ON DELETE CASCADE NOT VALID;

-- migrate:down

ALTER TABLE public.reactions DROP CONSTRAINT reactions_collection_item_id_fkey;
ALTER TABLE public.reactions
    ADD CONSTRAINT fk_rails_53df387e77 FOREIGN KEY (collection_item_id) REFERENCES public.collection_items(id);

ALTER TABLE public.reactions DROP CONSTRAINT reactions_target_user_id_fkey;
ALTER TABLE public.reactions
    ADD CONSTRAINT fk_rails_b76961a906 FOREIGN KEY (target_user_id) REFERENCES public.users(id);

ALTER TABLE public.reactions DROP CONSTRAINT reactions_user_id_fkey;
ALTER TABLE public.reactions
    ADD CONSTRAINT fk_rails_9f02fc96a0 FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.feature_flags DROP CONSTRAINT feature_flags_user_id_fkey;
ALTER TABLE public.feature_flags
    ADD CONSTRAINT feature_flags_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

-- 投稿者が削除されてuser_idがNULLになった行があると、NOT NULLを戻せずに失敗する。
ALTER TABLE public.character_images DROP CONSTRAINT character_images_user_id_fkey;
ALTER TABLE public.character_images
    ADD CONSTRAINT fk_rails_057f9798e7 FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE public.character_images ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE public.work_images DROP CONSTRAINT work_images_user_id_fkey;
ALTER TABLE public.work_images
    ADD CONSTRAINT fk_rails_3bad625b28 FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE public.work_images ALTER COLUMN user_id SET NOT NULL;
