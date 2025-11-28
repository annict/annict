-- name: GetUserByID :one
SELECT id, username, email, role, encrypted_password, locale, created_at, updated_at
FROM users
WHERE id = $1
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