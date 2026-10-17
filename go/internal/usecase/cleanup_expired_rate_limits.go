package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
)

// CleanupExpiredRateLimitsUsecaseは終わった窓のレート制限のカウンターのクリーンアップを担当する。
type CleanupExpiredRateLimitsUsecase struct {
	rateLimitRepo *repository.RateLimitRepository
}

// NewCleanupExpiredRateLimitsUsecaseは新しいCleanupExpiredRateLimitsUsecaseを作成する。
func NewCleanupExpiredRateLimitsUsecase(rateLimitRepo *repository.RateLimitRepository) *CleanupExpiredRateLimitsUsecase {
	return &CleanupExpiredRateLimitsUsecase{
		rateLimitRepo: rateLimitRepo,
	}
}

// Executeは窓の長さの上限 (ratelimit.MaxWindow) より前に始まった窓のカウンターを削除する。
// 窓はMaxWindow以下の長さに限られるため、これらの窓はすべて終わっている。
func (uc *CleanupExpiredRateLimitsUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "レート制限のカウンターのクリーンアップを開始します")

	cutoff := time.Now().Add(-ratelimit.MaxWindow)

	if err := uc.rateLimitRepo.DeleteExpired(ctx, cutoff); err != nil {
		slog.ErrorContext(ctx, "レート制限のカウンターの削除に失敗しました",
			"cutoff", cutoff,
			"error", err,
		)
		return fmt.Errorf("レート制限のカウンターの削除に失敗: %w", err)
	}

	slog.InfoContext(ctx, "レート制限のカウンターのクリーンアップが完了しました",
		"cutoff", cutoff,
	)

	return nil
}
