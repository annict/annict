-- name: IncrementRateLimit :one
-- キーと窓の組のカウンターを1つ増やし、増やした後の値を返す。
-- 窓の最初のリクエストでは行を作り、以降は同じ行のcountを増やす。
INSERT INTO rate_limits (key, window_start, count)
VALUES ($1, $2, 1)
ON CONFLICT (key, window_start)
DO UPDATE SET count = rate_limits.count + 1
RETURNING count;

-- name: DeleteExpiredRateLimits :exec
DELETE FROM rate_limits
WHERE window_start < $1;
