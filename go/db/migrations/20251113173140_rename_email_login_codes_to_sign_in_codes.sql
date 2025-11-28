-- migrate:up

-- テーブルをリネーム
ALTER TABLE email_login_codes RENAME TO sign_in_codes;

-- シーケンスをリネーム
ALTER SEQUENCE email_login_codes_id_seq RENAME TO sign_in_codes_id_seq;

-- 主キー制約をリネーム
ALTER TABLE sign_in_codes RENAME CONSTRAINT email_login_codes_pkey TO sign_in_codes_pkey;

-- インデックスをリネーム
ALTER INDEX idx_email_login_codes_expires_at RENAME TO idx_sign_in_codes_expires_at;
ALTER INDEX idx_email_login_codes_used_at RENAME TO idx_sign_in_codes_used_at;
ALTER INDEX idx_email_login_codes_user_id RENAME TO idx_sign_in_codes_user_id;

-- 外部キー制約をリネーム
ALTER TABLE sign_in_codes RENAME CONSTRAINT email_login_codes_user_id_fkey TO sign_in_codes_user_id_fkey;

-- migrate:down

-- 外部キー制約を元に戻す
ALTER TABLE sign_in_codes RENAME CONSTRAINT sign_in_codes_user_id_fkey TO email_login_codes_user_id_fkey;

-- インデックスを元に戻す
ALTER INDEX idx_sign_in_codes_user_id RENAME TO idx_email_login_codes_user_id;
ALTER INDEX idx_sign_in_codes_used_at RENAME TO idx_email_login_codes_used_at;
ALTER INDEX idx_sign_in_codes_expires_at RENAME TO idx_email_login_codes_expires_at;

-- 主キー制約を元に戻す
ALTER TABLE sign_in_codes RENAME CONSTRAINT sign_in_codes_pkey TO email_login_codes_pkey;

-- シーケンスを元に戻す
ALTER SEQUENCE sign_in_codes_id_seq RENAME TO email_login_codes_id_seq;

-- テーブルを元に戻す
ALTER TABLE sign_in_codes RENAME TO email_login_codes;
