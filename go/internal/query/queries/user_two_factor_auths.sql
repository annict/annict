-- name: UpsertPendingUserTwoFactorAuth :one
-- 登録の途中の行を作るか、秘密鍵を差し替える。有効にした行は書き換えず、そのときは行を返さない。
INSERT INTO user_two_factor_auths (user_id, secret_ciphertext)
VALUES (sqlc.arg('user_id'), sqlc.arg('secret_ciphertext'))
ON CONFLICT (user_id) DO UPDATE
SET secret_ciphertext = EXCLUDED.secret_ciphertext, last_used_step = 0, updated_at = NOW()
WHERE user_two_factor_auths.enabled_at IS NULL
RETURNING *;

-- name: GetUserTwoFactorAuthByUserID :one
SELECT * FROM user_two_factor_auths
WHERE user_id = sqlc.arg('user_id')
LIMIT 1;

-- name: EnableUserTwoFactorAuth :execrows
-- 登録の途中の行だけを有効にし、照合したコードのステップを使用済みとして記録する。
-- 照合後に秘密鍵が差し替わっていた場合は、新しい秘密鍵の行を有効にしない。
UPDATE user_two_factor_auths
SET enabled_at = NOW(), last_used_step = sqlc.arg('step'), updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')
AND secret_ciphertext = sqlc.arg('expected_secret_ciphertext')
AND enabled_at IS NULL
AND last_used_step < sqlc.arg('step');

-- name: UseMatchingUserTwoFactorAuthStep :execrows
-- 照合した有効な設定に限ってステップを記録し、設定の入れ替わりとコードの使い回しを拒む。
UPDATE user_two_factor_auths
SET last_used_step = sqlc.arg('step'), updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')
AND id = sqlc.arg('expected_id')
AND secret_ciphertext = sqlc.arg('expected_secret_ciphertext')
AND enabled_at IS NOT NULL
AND last_used_step < sqlc.arg('step');

-- name: LockMatchingUserTwoFactorAuth :execrows
-- 照合した有効な設定の行をロックする。リカバリーコードでのログインが、無効化と同じく設定の行からリカバリーコードの行の順にロックし、
-- 照合のあとに設定が入れ替わっていればロックできた行の数で見分ける。
SELECT 1 FROM user_two_factor_auths
WHERE user_id = sqlc.arg('user_id')
AND id = sqlc.arg('expected_id')
AND secret_ciphertext = sqlc.arg('expected_secret_ciphertext')
AND enabled_at IS NOT NULL
FOR UPDATE;

-- name: DeleteMatchingUserTwoFactorAuth :execrows
-- 本人確認で照合した設定だけを削除し、別の設定に入れ替わっていたら削除しない。
DELETE FROM user_two_factor_auths
WHERE user_id = sqlc.arg('user_id')
AND id = sqlc.arg('expected_id')
AND secret_ciphertext = sqlc.arg('expected_secret_ciphertext')
AND enabled_at IS NOT NULL;

-- name: DeleteUserTwoFactorAuthByUserID :one
-- 実際に削除した設定を返し、削除の直前に有効化されていてもその状態を取得する。
DELETE FROM user_two_factor_auths
WHERE user_id = sqlc.arg('user_id')
RETURNING *;
