-- name: CreateSignUpCode :one
INSERT INTO sign_up_codes (
    email,
    code_digest,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetValidSignUpCode :one
SELECT * FROM sign_up_codes
WHERE email = $1
AND used_at IS NULL
AND expires_at > NOW()
AND attempts < 5
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementSignUpCodeAttempts :exec
UPDATE sign_up_codes
SET attempts = attempts + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkSignUpCodeAsUsed :exec
UPDATE sign_up_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: InvalidateSignUpCodesByEmail :exec
UPDATE sign_up_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE email = $1
AND used_at IS NULL;
