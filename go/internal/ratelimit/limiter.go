// Package ratelimitはレート制限機能を提供します
package ratelimit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// MaxWindowは限定器に指定できる窓の長さの上限です。
// 期限切れのカウンターの削除は、この長さより前に始まった窓をすべて終わったものとして扱います。
const MaxWindow = time.Hour

// LimiterはPostgreSQLの `rate_limits` テーブルで数える固定窓のレート制限を実装する構造体
type Limiter struct {
	repo *repository.RateLimitRepository
	now  func() time.Time
}

// NewLimiterは新しいLimiterを作成します
func NewLimiter(repo *repository.RateLimitRepository) *Limiter {
	return &Limiter{repo: repo, now: time.Now}
}

// WithTxはトランザクションを使用する新しいLimiterを返します
func (l *Limiter) WithTx(tx *sql.Tx) *Limiter {
	return &Limiter{repo: l.repo.WithTx(tx), now: l.now}
}

// CheckInputはレート制限の判定の入力です
type CheckInput struct {
	// Keyは数える対象を表すキー (例: "password_reset:ip:192.168.1.1")
	Key string
	// Limitは1つの窓で許可するリクエスト数
	Limit int
	// Windowは窓の長さ。MaxWindow以下で指定する
	Window time.Duration
}

// CheckResultはレート制限の判定の結果です
type CheckResult struct {
	// Allowedはリクエストを許可するかどうか
	Allowed bool
	// Countは今回のリクエストを含めた、現在の窓でのリクエスト数
	Count int
	// Remainingは現在の窓で残っている許可数
	Remaining int
	// ResetAtは現在の窓が終わる時刻
	ResetAt time.Time
}

// Checkは現在の窓のカウンターを1つ増やし、上限を超えていないかを判定します。
// 窓は時刻をWindowの長さで切り捨てた固定窓 (例: 1分なら毎分0秒から59秒) です。
func (l *Limiter) Check(ctx context.Context, input CheckInput) (*CheckResult, error) {
	if input.Key == "" {
		return nil, errors.New("key is required")
	}
	if input.Limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	if input.Window <= 0 || input.Window > MaxWindow {
		return nil, fmt.Errorf("window must be positive and at most %s", MaxWindow)
	}

	windowStart := l.now().UTC().Truncate(input.Window)

	count, err := l.repo.Increment(ctx, input.Key, windowStart)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return &CheckResult{
		Allowed:   count <= input.Limit,
		Count:     count,
		Remaining: max(input.Limit-count, 0),
		ResetAt:   windowStart.Add(input.Window),
	}, nil
}
