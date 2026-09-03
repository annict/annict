package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// cleanupExpiredSessionsBatchSize is how many sessions one DELETE removes at most.
// The deletion runs as a loop of bounded statements rather than a single one so that
// a large backlog does not hold row locks on the whole matching set or produce one
// oversized burst of WAL.
//
// [Ja] cleanupExpiredSessionsBatchSize は 1 回の DELETE が削除するセッションの上限。
// 削除を 1 文でまとめず区切った文のループにするのは、滞留が大きいときに一致する行全体の
// ロックを保持したり、WAL を一度に大量生成したりしないようにするため。
const cleanupExpiredSessionsBatchSize = 1000

// CleanupExpiredSessionsUsecase is responsible for cleaning up expired sessions.
//
// [Ja] CleanupExpiredSessionsUsecase は期限切れセッションのクリーンアップを担当する。
type CleanupExpiredSessionsUsecase struct {
	sessionRepo *repository.SessionRepository
}

// NewCleanupExpiredSessionsUsecase creates a new CleanupExpiredSessionsUsecase.
//
// [Ja] NewCleanupExpiredSessionsUsecase は新しい CleanupExpiredSessionsUsecase を作成する。
func NewCleanupExpiredSessionsUsecase(sessionRepo *repository.SessionRepository) *CleanupExpiredSessionsUsecase {
	return &CleanupExpiredSessionsUsecase{
		sessionRepo: sessionRepo,
	}
}

// Execute deletes the sessions that have not been accessed for model.SessionMaxAge,
// in batches, until no eligible session is left. It sets no upper bound on one run:
// stopping early would leave the surplus for the next run and a backlog larger than
// the cap would never be worked off.
//
// [Ja] Execute は model.SessionMaxAge の間アクセスされていないセッションを、対象が無く
// なるまでバッチで削除する。1 回の実行に上限は設けない。途中で打ち切ると超過分が次回に
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
