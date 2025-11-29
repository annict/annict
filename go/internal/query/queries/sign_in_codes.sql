-- name: CreateSignInCode :one
INSERT INTO sign_in_codes (
    user_id,
    code_digest,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetValidSignInCode :one
SELECT * FROM sign_in_codes
WHERE user_id = $1
AND used_at IS NULL
AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementSignInCodeAttempts :exec
UPDATE sign_in_codes
SET attempts = attempts + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkSignInCodeAsUsed :exec
UPDATE sign_in_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredSignInCodes :exec
DELETE FROM sign_in_codes
WHERE expires_at < $1 OR used_at < $1;

-- name: InvalidateUserSignInCodes :exec
UPDATE sign_in_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1
AND used_at IS NULL;
