// Package ratelimit はレート制限機能を提供します
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter はRate Limitingを実装する構造体
type Limiter struct {
	client *redis.Client
}

// NewLimiter は新しいLimiterを作成します
func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{client: client}
}

// Check はRate Limitingをチェックします
// key: Rate Limitingのキー（例: "password_reset:ip:192.168.1.1"）
// limit: 制限回数
// window: 時間窓（例: 1時間）
// 戻り値: 許可される場合はtrue、制限超過の場合はfalse
func (l *Limiter) Check(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Luaスクリプトでアトミックに操作
	// 1. キーの値をインクリメント
	// 2. 初回の場合（値が1の場合）、TTLを設定
	// 3. 現在のカウントを返す
	script := `
		local current = redis.call('INCR', KEYS[1])
		if current == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[1])
		end
		return current
	`

	fullKey := fmt.Sprintf("rate_limit:%s", key)
	result, err := l.client.Eval(ctx, script, []string{fullKey}, int(window.Seconds())).Int()
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// 制限回数以下の場合は許可
	return result <= limit, nil
}

// Reset はRate Limitingのカウンタをリセットします（主にテスト用）
func (l *Limiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("rate_limit:%s", key)
	return l.client.Del(ctx, fullKey).Err()
}
