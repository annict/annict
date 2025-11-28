-- name: CreateProfile :one
INSERT INTO profiles (user_id, name, description, created_at, updated_at, background_image_animated)
VALUES ($1, $2, '', NOW(), NOW(), false)
RETURNING id, user_id, name, description, created_at, updated_at;

-- name: GetProfileByUserID :one
SELECT id, user_id, name, description, url, created_at, updated_at
FROM profiles
WHERE user_id = $1
LIMIT 1;

-- name: UpdateProfileImageData :exec
UPDATE profiles
SET image_data = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: ListAllProfiles :many
SELECT id, user_id
FROM profiles
ORDER BY id;
