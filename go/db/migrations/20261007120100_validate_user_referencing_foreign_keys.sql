-- migrate:up

-- add_on_delete_to_user_referencing_foreign_keysでNOT VALIDとして追加した外部キーを検証する。
-- 追加とは別のトランザクションで行うことで、既存行の走査中も参照元テーブルへの書き込みを止めない。
ALTER TABLE public.work_images VALIDATE CONSTRAINT work_images_user_id_fkey;
ALTER TABLE public.character_images VALIDATE CONSTRAINT character_images_user_id_fkey;
ALTER TABLE public.feature_flags VALIDATE CONSTRAINT feature_flags_user_id_fkey;
ALTER TABLE public.reactions VALIDATE CONSTRAINT reactions_user_id_fkey;
ALTER TABLE public.reactions VALIDATE CONSTRAINT reactions_target_user_id_fkey;
ALTER TABLE public.reactions VALIDATE CONSTRAINT reactions_collection_item_id_fkey;

-- migrate:down

-- 検証済みの制約をNOT VALIDに戻す手段は無く、戻す必要も無いため何もしない。
