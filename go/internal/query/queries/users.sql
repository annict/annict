-- name: GetActiveUserByID :one
-- 論理削除されていないユーザーをIDで取得する。
-- セッションやログインコードからユーザーを引く認証の経路で使い、
-- 退会・利用停止でdeleted_atが立ったユーザーを未ログインとして扱う。
SELECT
    u.id,
    u.username,
    u.email,
    u.role,
    u.encrypted_password,
    u.locale,
    u.time_zone,
    u.stripe_subscriber_id,
    u.gumroad_subscriber_id,
    u.notifications_count,
    u.created_at,
    u.updated_at,
    p.image_data AS profile_image_data
FROM users u
LEFT JOIN profiles p ON p.user_id = u.id
WHERE u.id = $1
  AND u.deleted_at IS NULL
LIMIT 1;

-- name: GetActiveUserByEmailOrUsername :one
SELECT id, username, email, role, encrypted_password, created_at, updated_at
FROM users
WHERE (LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($1))
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1)
LIMIT 1;

-- name: GetActiveUserByEmailForSignIn :one
SELECT id, email, encrypted_password
FROM users
WHERE LOWER(email) = LOWER($1)
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveUserIDByEmail :one
-- 論理削除されていないユーザーのIDをメールアドレスで検索する。
-- 削除済みユーザーのメールアドレスも一意制約の対象のため、
-- 登録済みかの確認には使わず、パスワードリセットなど本人の操作を受け付ける経路で使う。
SELECT id
FROM users
WHERE LOWER(email) = LOWER($1)
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserPassword :exec
UPDATE users
SET encrypted_password = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE LOWER(username) = LOWER($1)
LIMIT 1;

-- name: GetActiveUserIDByUsername :one
-- 論理削除されていないユーザーのIDをusernameで検索する。
-- 削除済みユーザーに対して404を返すべき機能 (例: 視聴記録ヒートマップ
-- フラグメント) で、他のユーザー属性を取得せずに利用する。
SELECT id
FROM users
WHERE LOWER(username) = LOWER($1)
  AND deleted_at IS NULL
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, email, encrypted_password, locale, role, time_zone, created_at, updated_at)
VALUES ($1, $2, $3, $4, 0, 'Asia/Tokyo', NOW(), NOW())
RETURNING id, username, email, role, locale, created_at, updated_at;

-- name: UpdateUserStripeSubscriberID :exec
UPDATE users
SET stripe_subscriber_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetUserByStripeSubscriberID :one
SELECT id, username, email, role, stripe_subscriber_id, created_at, updated_at
FROM users
WHERE stripe_subscriber_id = $1
LIMIT 1;
-- name: GetUserForDeletion :one
-- 退会の削除処理の対象を確かめるため、論理削除済みのユーザーも含めてIDで取得する。
SELECT id, deleted_at, stripe_subscriber_id, gumroad_subscriber_id
FROM users
WHERE id = $1
LIMIT 1;

-- name: LockStripeSubscriberForUserDeletion :one
-- 購読情報を共有するユーザーの削除を直列化する。
SELECT id FROM stripe_subscribers WHERE id = $1 FOR UPDATE;

-- name: LockGumroadSubscriberForUserDeletion :one
-- 購読情報を共有するユーザーの削除を直列化する。
SELECT id FROM gumroad_subscribers WHERE id = $1 FOR UPDATE;

-- name: DeleteUser :one
-- ユーザー行を削除し、削除した件数を返す。外部キーにON DELETEの無い1件ずつの従属データ
-- (設定・メール通知設定・メールアドレスの確認) と、ほかのユーザーから参照されていない
-- 購読情報も同じ文の中で削除する。外部キーの検査は文の終わりに行われるため、ユーザー行と
-- 参照し合う行を1つの文で削除できる。そのほかの従属データは外部キーのON DELETEに任せる。
WITH deleted_settings AS (
    DELETE FROM settings
    WHERE settings.user_id = sqlc.arg('user_id')
),
deleted_email_notifications AS (
    DELETE FROM email_notifications
    WHERE email_notifications.user_id = sqlc.arg('user_id')
),
deleted_email_confirmations AS (
    DELETE FROM email_confirmations
    WHERE email_confirmations.user_id = sqlc.arg('user_id')
),
deleted_user AS (
    DELETE FROM users
    WHERE users.id = sqlc.arg('user_id')
    RETURNING users.stripe_subscriber_id, users.gumroad_subscriber_id
),
deleted_stripe_subscribers AS (
    DELETE FROM stripe_subscribers
    WHERE stripe_subscribers.id IN (SELECT deleted_user.stripe_subscriber_id FROM deleted_user)
        AND NOT EXISTS (
            SELECT 1
            FROM users AS other_users
            WHERE other_users.stripe_subscriber_id = stripe_subscribers.id
                AND other_users.id <> sqlc.arg('user_id')
        )
),
deleted_gumroad_subscribers AS (
    DELETE FROM gumroad_subscribers
    WHERE gumroad_subscribers.id IN (SELECT deleted_user.gumroad_subscriber_id FROM deleted_user)
        AND NOT EXISTS (
            SELECT 1
            FROM users AS other_users
            WHERE other_users.gumroad_subscriber_id = gumroad_subscribers.id
                AND other_users.id <> sqlc.arg('user_id')
        )
)
SELECT COUNT(*) FROM deleted_user;
