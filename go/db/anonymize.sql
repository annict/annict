-- anonymize.sqlは本番スナップショット中の個人情報 (PII) と秘密情報を
-- マスクし、生PIIを持ち込まずにローカル開発DBへ取り込めるようにする。
--
-- 使い捨てのannict_anonymize DBに対してのみ実行する想定 (scripts/load_snapshot_db.sh
-- を参照)。カタログ (works / episodes) と既に公開済みのコンテンツ (レビュー・
-- プロフィール文) は意図的に触らず、認証・課金・連携系のデータのみをマスク /
-- 削除する。

BEGIN;

-- 認証系の秘密情報を消し、課金系の外部キーを外して後段で課金テーブルを
-- 削除できるようにする。パスワードは既知の単一値 (bcrypt("annictdev"), cost 10) に
-- 揃え、どのアカウントでもローカルでログインできるようにする一方、confirmed_atは
-- 残してアカウントを使える状態に保つ。emailとusernameは一意列であり、一時的な
-- 一意制約違反を避けるため後段で2段階更新により別途書き換える。
UPDATE users
SET unconfirmed_email = NULL,
    encrypted_password = '$2a$10$ugGzJcxmltdkVShHMzgeDOUl3vEM2CjM67jj0KzgvC2IPu4To/nxa',
    current_sign_in_ip = NULL,
    last_sign_in_ip = NULL,
    confirmation_token = NULL,
    reset_password_token = NULL,
    gumroad_subscriber_id = NULL,
    stripe_subscriber_id = NULL;

-- 一意列 (email, username) をuser{id} 形へ2段階で書き換え、一時的な一意
-- 制約違反を避ける。PostgreSQLは非遅延のusers_email_key / users_username_keyを
-- 単一UPDATE内で行ごとに検査するため、最終状態が一意でも直接の一括更新は途中で
-- 衝突しうる。例えばハンドルが文字どおり "user66" の実アカウントや、メールが
-- "user66@example.com" の実アカウントは、別の行がその値に改名された瞬間に衝突する。
--
-- 第1段では両列を、実データが取り得ない "-{id}" 値に退避する。usernameは
-- [A-Za-z0-9_] に限られ (USERNAME_FORMAT) "-" を含まず、実メールは必ず "@" を含むが
-- "-{id}" は含まないためである。idで各退避値が一意になる。第2段で最終のuser{id}
-- 形を割り当てる。これは "-{id}" の退避先の値と重ならないため、衝突する値を持つ
-- 行は生じない。
UPDATE users SET email = '-' || id, username = '-' || id;
UPDATE users SET email = 'user' || id || '@example.com', username = 'user' || id;

-- 外部認証プロバイダの資格情報 (OmniAuthのuid / トークン) を消す。
-- providersを参照するテーブルは無いため、素のTRUNCATEで安全に消せる。
TRUNCATE providers;

-- OAuthのアクセストークンと認可を消す。どちらも他テーブルを参照するだけ
-- なので、TRUNCATEしても流入する外部キー制約に違反しない。
TRUNCATE oauth_access_tokens, oauth_access_grants;

-- OAuthアプリのクライアントシークレットをその場で消す。statuses /
-- episode_records / work_recordsがoauth_applicationsを参照しており、その公開
-- コンテンツを残したいため行自体は保持する (uidは再生成)。TRUNCATEもDELETEも
-- その参照を壊す必要が出るため採らない。
UPDATE oauth_applications
SET secret = '',
    uid = md5(random()::text || id::text);

-- 課金データを削除する。usersが外部キーを持つためTRUNCATEではなくDELETE
-- を使う (リンクは上でNULL済みなので削除できる)。stripe_webhook_eventsは独立
-- しているためTRUNCATEしてよい。
DELETE FROM stripe_subscribers;
DELETE FROM gumroad_subscribers;
TRUNCATE stripe_webhook_events;

-- 一時的な認証データ (ワンタイムコードやトークン) を削除する。
TRUNCATE email_confirmations, sign_up_codes, sign_in_codes, password_reset_tokens;

COMMIT;
