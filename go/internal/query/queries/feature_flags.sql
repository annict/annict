-- name: IsFeatureFlagEnabled :one
SELECT EXISTS(
    SELECT 1 FROM feature_flags ff
    WHERE ff.name = $3
    AND (
        (ff.device_token IS NOT NULL AND ff.device_token = $1)
        OR (ff.user_id IS NOT NULL AND ff.user_id = $2)
    )
);
