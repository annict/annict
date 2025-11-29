-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (
    user_id,
    token_digest,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPasswordResetTokenByDigest :one
SELECT * FROM password_reset_tokens
WHERE token_digest = $1
AND used_at IS NULL
AND expires_at > NOW()
LIMIT 1;

-- name: MarkPasswordResetTokenAsUsed :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: DeleteUnusedPasswordResetTokensByUserID :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1
AND used_at IS NULL;

-- name: InvalidateUserPasswordResetTokens :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE user_id = $1
AND used_at IS NULL;

-- name: DeleteExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens
WHERE expires_at < $1 OR used_at < $1;

-- name: GetPasswordResetTokensByUserID :many
SELECT * FROM password_reset_tokens
WHERE user_id = $1
ORDER BY created_at DESC;
