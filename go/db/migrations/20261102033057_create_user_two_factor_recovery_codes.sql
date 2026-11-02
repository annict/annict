-- migrate:up

-- user_two_factor_recovery_codesは、認証アプリを使えないときにTOTPの代わりに使うリカバリーコードを持つ。
-- 1行が1つのコードで、平文のコードは有効にした直後の画面にしか出さない。
--
-- code_digestはコードのHMAC-SHA-256のダイジェスト。鍵はANNICT_TOTP_ENCRYPTION_KEYから導く。
-- コードは40ビットほどのエントロピーしか無く、鍵の無いSHA-256ではデータベースが漏れたときに
-- 総当たりで平文へ戻せてしまうため、鍵付きのダイジェストにする。
--
-- used_atはコードを使った日時で、NULLはまだ使えることを表す。
-- 消費はused_at IS NULLを条件にしたUPDATEで行い、同じコードを同時に使っても一方しか通らないようにする。
-- 使ったコードの行を残すのは、残りの数を数えるときに使用済みと区別するため。
--
-- user_idの外部キーをON DELETE CASCADEにするのは、コードがユーザーに従属する認証情報のため。
-- (user_id, code_digest) の一意制約は、ユーザー単位でコードを照合・削除するときと、
-- 外部キーのインデックスを兼ねる。
CREATE TABLE user_two_factor_recovery_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code_digest VARCHAR NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, code_digest)
);

-- migrate:down

DROP TABLE IF EXISTS user_two_factor_recovery_codes;
