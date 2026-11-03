-- name: CreateUserTwoFactorRecoveryCodes :exec
-- ユーザーのリカバリーコードを、渡したダイジェストの数だけ1つの文で作る。
INSERT INTO user_two_factor_recovery_codes (user_id, code_digest)
SELECT sqlc.arg('user_id'), unnest(sqlc.arg('code_digests')::varchar[]);

-- name: CountUnusedUserTwoFactorRecoveryCodes :one
SELECT COUNT(*) FROM user_two_factor_recovery_codes
WHERE user_id = sqlc.arg('user_id')
AND used_at IS NULL;

-- name: UseUserTwoFactorRecoveryCode :execrows
-- 未使用のコードだけを使用済みにし、更新できたかどうかでコードを使えたかを判定する。
UPDATE user_two_factor_recovery_codes
SET used_at = NOW(), updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')
AND code_digest = sqlc.arg('code_digest')
AND used_at IS NULL;

-- name: DeleteUserTwoFactorRecoveryCodesByUserID :exec
DELETE FROM user_two_factor_recovery_codes
WHERE user_id = sqlc.arg('user_id');
