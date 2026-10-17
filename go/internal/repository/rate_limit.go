package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/query"
)

// RateLimitRepositoryはレート制限のカウンターのデータアクセスを担当します
type RateLimitRepository struct {
	queries *query.Queries
}

// NewRateLimitRepositoryはRateLimitRepositoryを作成します
func NewRateLimitRepository(queries *query.Queries) *RateLimitRepository {
	return &RateLimitRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *RateLimitRepository) WithTx(tx *sql.Tx) *RateLimitRepository {
	return &RateLimitRepository{queries: r.queries.WithTx(tx)}
}

// Incrementはキーと窓の開始時刻の組のカウンターを1つ増やし、増やした後の値を返します
func (r *RateLimitRepository) Increment(ctx context.Context, key string, windowStart time.Time) (int, error) {
	count, err := r.queries.IncrementRateLimit(ctx, query.IncrementRateLimitParams{
		Key:         key,
		WindowStart: windowStart,
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// DeleteExpiredは指定日時より前に始まった窓のカウンターを削除します
func (r *RateLimitRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	return r.queries.DeleteExpiredRateLimits(ctx, cutoff)
}
