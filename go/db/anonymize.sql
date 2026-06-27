-- anonymize.sql masks personally identifiable information (PII) and secrets in
-- a production snapshot so the data can be loaded into a local development
-- database without carrying raw PII.
--
-- It is meant to run only against the throwaway annict_anonymize database
-- (see scripts/load_snapshot_db.sh). The catalog (works / episodes) and
-- already-public content (reviews, profile texts) are intentionally left
-- untouched; only authentication, billing, and integration data is masked or
-- removed.
--
-- [Ja] anonymize.sql は本番スナップショット中の個人情報 (PII) と秘密情報を
-- マスクし、生 PII を持ち込まずにローカル開発 DB へ取り込めるようにする。
--
-- 使い捨ての annict_anonymize DB に対してのみ実行する想定 (scripts/load_snapshot_db.sh
-- を参照)。カタログ (works / episodes) と既に公開済みのコンテンツ (レビュー・
-- プロフィール文) は意図的に触らず、認証・課金・連携系のデータのみをマスク /
-- 削除する。

BEGIN;

-- Wipe authentication secrets and detach the billing foreign keys so the
-- subscriber tables can be removed further down. The password is reset to a
-- single known value (bcrypt of "annictdev", cost 10) so every account can be
-- signed into locally, while confirmed_at is kept so accounts stay usable. email
-- and username are unique columns and are rewritten separately below, where a
-- two-pass update avoids a transient unique-constraint violation.
--
-- [Ja] 認証系の秘密情報を消し、課金系の外部キーを外して後段で課金テーブルを
-- 削除できるようにする。パスワードは既知の単一値 (bcrypt("annictdev"), cost 10) に
-- 揃え、どのアカウントでもローカルでログインできるようにする一方、confirmed_at は
-- 残してアカウントを使える状態に保つ。email と username は一意列であり、一時的な
-- 一意制約違反を避けるため後段で 2 段階更新により別途書き換える。
UPDATE users
SET unconfirmed_email = NULL,
    encrypted_password = '$2a$10$ugGzJcxmltdkVShHMzgeDOUl3vEM2CjM67jj0KzgvC2IPu4To/nxa',
    current_sign_in_ip = NULL,
    last_sign_in_ip = NULL,
    confirmation_token = NULL,
    reset_password_token = NULL,
    gumroad_subscriber_id = NULL,
    stripe_subscriber_id = NULL;

-- Rewrite the unique columns (email, username) to their user{id} forms in two
-- passes to avoid a transient unique-constraint violation. Postgres enforces the
-- non-deferrable users_email_key / users_username_key per row within a single
-- UPDATE, so a direct bulk rename can clash mid-statement even when the final
-- state is unique. For example, a real account whose handle is literally
-- "user66", or whose email is "user66@example.com", collides the moment another
-- row is renamed onto that value.
--
-- Pass 1 parks both columns under a "-{id}" value that no real row can hold:
-- usernames are limited to [A-Za-z0-9_] (USERNAME_FORMAT) so none contain "-",
-- and every real email contains "@" while "-{id}" does not. id keeps each parked
-- value unique. Pass 2 then assigns the final user{id} forms, which are disjoint
-- from the "-{id}" parking namespace, so no row ever holds a conflicting value.
--
-- [Ja] 一意列 (email, username) を user{id} 形へ 2 段階で書き換え、一時的な一意
-- 制約違反を避ける。PostgreSQL は非遅延の users_email_key / users_username_key を
-- 単一 UPDATE 内で行ごとに検査するため、最終状態が一意でも直接の一括更新は途中で
-- 衝突しうる。例えばハンドルが文字どおり "user66" の実アカウントや、メールが
-- "user66@example.com" の実アカウントは、別の行がその値に改名された瞬間に衝突する。
--
-- 第 1 段では両列を、実データが取り得ない "-{id}" 値に退避する。username は
-- [A-Za-z0-9_] に限られ (USERNAME_FORMAT) "-" を含まず、実メールは必ず "@" を含むが
-- "-{id}" は含まないためである。id で各退避値が一意になる。第 2 段で最終の user{id}
-- 形を割り当てる。これは "-{id}" の退避名前空間と disjoint なため、衝突する値を持つ
-- 行は生じない。
UPDATE users SET email = '-' || id, username = '-' || id;
UPDATE users SET email = 'user' || id || '@example.com', username = 'user' || id;

-- Clear external auth provider credentials (OmniAuth uid / tokens). Nothing
-- references providers, so a plain TRUNCATE is safe.
--
-- [Ja] 外部認証プロバイダの資格情報 (OmniAuth の uid / トークン) を消す。
-- providers を参照するテーブルは無いため、素の TRUNCATE で安全に消せる。
TRUNCATE providers;

-- Clear OAuth access tokens and grants. Both only reference other tables, so
-- truncating them does not violate any inbound foreign key.
--
-- [Ja] OAuth のアクセストークンと認可を消す。どちらも他テーブルを参照するだけ
-- なので、TRUNCATE しても流入する外部キー制約に違反しない。
TRUNCATE oauth_access_tokens, oauth_access_grants;

-- Wipe OAuth application client secrets in place. The rows are kept (and uid is
-- regenerated) because statuses / episode_records / work_records reference
-- oauth_applications; we want to preserve that public content, and both
-- TRUNCATE and DELETE would force us to destroy those references.
--
-- [Ja] OAuth アプリのクライアントシークレットをその場で消す。statuses /
-- episode_records / work_records が oauth_applications を参照しており、その公開
-- コンテンツを残したいため行自体は保持する (uid は再生成)。TRUNCATE も DELETE も
-- その参照を壊す必要が出るため採らない。
UPDATE oauth_applications
SET secret = '',
    uid = md5(random()::text || id::text);

-- Remove billing data. These use DELETE instead of TRUNCATE because users has a
-- foreign key to them; the links were set to NULL above, so the rows can now be
-- removed. stripe_webhook_events is standalone and can be truncated.
--
-- [Ja] 課金データを削除する。users が外部キーを持つため TRUNCATE ではなく DELETE
-- を使う (リンクは上で NULL 済みなので削除できる)。stripe_webhook_events は独立
-- しているため TRUNCATE してよい。
DELETE FROM stripe_subscribers;
DELETE FROM gumroad_subscribers;
TRUNCATE stripe_webhook_events;

-- Remove ephemeral authentication artifacts (one-time codes and tokens).
--
-- [Ja] 一時的な認証データ (ワンタイムコードやトークン) を削除する。
TRUNCATE email_confirmations, sign_up_codes, sign_in_codes, password_reset_tokens;

COMMIT;
