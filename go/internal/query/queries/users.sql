-- name: GetUserByID :one
SELECT
    u.id,
    u.username,
    u.email,
    u.role,
    u.encrypted_password,
    u.locale,
    u.stripe_subscriber_id,
    u.gumroad_subscriber_id,
    u.notifications_count,
    u.created_at,
    u.updated_at,
    p.image_data AS profile_image_data
FROM users u
LEFT JOIN profiles p ON p.user_id = u.id
WHERE u.id = $1
LIMIT 1;

-- name: GetUserByEmailOrUsername :one
SELECT id, username, email, role, encrypted_password, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($1)
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1)
LIMIT 1;

-- name: GetUserByEmailForSignIn :one
SELECT id, email, encrypted_password
FROM users
WHERE LOWER(email) = LOWER($1)
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