-- migrate:up

-- user_two_factor_authsは、ユーザーが認証アプリに登録したTOTPの秘密鍵を持つ。
-- 1人のユーザーが持つ行は高々1行で、user_idの一意制約でそれを保つ。
--
-- secret_ciphertextは秘密鍵をANNICT_TOTP_ENCRYPTION_KEYから導いた鍵でAES-GCM暗号化した値。
-- 秘密鍵はコードの照合に平文へ戻す必要があるため、ダイジェストではなく暗号化で守る。
-- データベースの内容だけが漏れても、コードを作れる秘密鍵が露出しないようにするため。
--
-- last_used_stepは最後に受け付けたコードのタイムステップ (Unix時刻を30秒で割った値)。
-- これ以下のステップのコードを拒否し、同じコードの使い回しを防ぐ。
-- 0は一度も使っていないことを表し、どのステップのコードも受け付ける。
--
-- enabled_atは二要素認証を有効にした日時で、NULLは認証アプリへの登録の途中を表す。
-- 登録の途中の行は、最初のコードを照合できたときに有効になる。
--
-- user_idの外部キーをON DELETE CASCADEにするのは、秘密鍵がユーザーに従属する認証情報のため。
-- user_idの一意制約が、外部キーのインデックスを兼ねる。
CREATE TABLE user_two_factor_auths (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    secret_ciphertext BYTEA NOT NULL,
    last_used_step BIGINT NOT NULL DEFAULT 0,
    enabled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id)
);

-- migrate:down

DROP TABLE IF EXISTS user_two_factor_auths;
