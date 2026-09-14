package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// cleanupExpiredSessionsBatchSizeは1回のDELETEが削除するセッションの上限。
// 削除を1文でまとめず区切った文のループにするのは、滞留が大きいときに一致する行全体の
// ロックを保持したり、WALを一度に大量生成したりしないようにするため。
const cleanupExpiredSessionsBatchSize = 1000

// CleanupExpiredSessionsUsecaseは期限切れセッションのクリーンアップを担当する。
type CleanupExpiredSessionsUsecase struct {
	sessionRepo *repository.SessionRepository
}

// NewCleanupExpiredSessionsUsecaseは新しいCleanupExpiredSessionsUsecaseを作成する。
func NewCleanupExpiredSessionsUsecase(sessionRepo *repository.SessionRepository) *CleanupExpiredSessionsUsecase {
	return &CleanupExpiredSessionsUsecase{
		sessionRepo: sessionRepo,
	}
}

// Executeはmodel.SessionMaxAgeの間アクセスされていないセッションを、対象が無く
// なるまでバッチで削除する。1回の実行に上限は設けない。途中で打ち切ると超過分が次回に
// 持ち越され、上限を超える滞留がいつまでも解消しないため。
func (uc *CleanupExpiredSessionsUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "セッションクリーンアップを開始します")

	cutoff := time.Now().Add(-model.SessionMaxAge)

	var deleted int64
	for {
		count, err := uc.sessionRepo.DeleteExpired(ctx, cutoff, cleanupExpiredSessionsBatchSize)
		if err != nil {
			slog.ErrorContext(ctx, "セッションの削除に失敗しました",
				"cutoff", cutoff,
				"deleted", deleted,
				"error", err,
			)
			return fmt.Errorf("セッションの削除に失敗: %w", err)
		}

		deleted += count
		if count < cleanupExpiredSessionsBatchSize {
			break
		}
	}

	slog.InfoContext(ctx, "セッションクリーンアップが完了しました",
		"cutoff", cutoff,
		"deleted", deleted,
	)

	return nil
}
