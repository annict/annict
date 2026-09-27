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

// CleanupExpiredSessionsTimeLimitは期限切れセッションのクリーンアップ1回の実行時間の
// 上限。滞留が大きいときも1回の実行がDBにかける負荷を一定に抑えるため、I/Oの速さで
// 所要時間が変わる件数ではなく時間で区切る。`serve` の毎時の定期ジョブと
// `task cleanup-expired-sessions` の双方がこの値でUseCaseを組み立てる。
const CleanupExpiredSessionsTimeLimit = 10 * time.Minute

// CleanupExpiredSessionsUsecaseは期限切れセッションのクリーンアップを担当する。
type CleanupExpiredSessionsUsecase struct {
	sessionRepo *repository.SessionRepository
	timeLimit   time.Duration
}

// NewCleanupExpiredSessionsUsecaseは新しいCleanupExpiredSessionsUsecaseを作成する。
// timeLimitは1回の実行時間の上限で、超えたら次のバッチに入らずに終了する。
func NewCleanupExpiredSessionsUsecase(sessionRepo *repository.SessionRepository, timeLimit time.Duration) *CleanupExpiredSessionsUsecase {
	return &CleanupExpiredSessionsUsecase{
		sessionRepo: sessionRepo,
		timeLimit:   timeLimit,
	}
}

// Executeはmodel.SessionMaxAgeの間アクセスされていないセッションを、対象が無く
// なるか実行時間の上限に達するまでバッチで削除する。上限はバッチの合間に確認するため、
// 最後のバッチのぶんだけ上限をわずかに超えうる。上限で打ち切った残りは次回の実行に持ち
// 越し、エラーにはしない。毎時の実行で流入を上回る件数を削除できれば、滞留があっても
// 回を重ねるごとに減っていくため。
func (uc *CleanupExpiredSessionsUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "セッションクリーンアップを開始します", "time_limit", uc.timeLimit)

	start := time.Now()
	cutoff := start.Add(-model.SessionMaxAge)

	// lowerBoundは次のバッチが読み始めるupdated_at。直前のバッチで削除した行の最大値を
	// 引き継ぎ、削除済みでVACUUMを待つインデックスエントリを先頭から読み直さないようにする。
	// 実行をまたいでは引き継がないため、各実行の最初のバッチはゼロ値 (下限なし) から読む。
	// SKIP LOCKEDで飛ばした行が下限より古くなっても、次回の実行の最初のバッチで拾われる。
	var lowerBound time.Time
	var deleted int64
	var batches int
	timeLimitReached := false
	for {
		count, maxUpdatedAt, err := uc.sessionRepo.DeleteExpired(ctx, cutoff, lowerBound, cleanupExpiredSessionsBatchSize)
		if err != nil {
			slog.ErrorContext(ctx, "セッションの削除に失敗しました",
				"cutoff", cutoff,
				"lower_bound", lowerBound,
				"deleted", deleted,
				"batches", batches,
				"elapsed", time.Since(start),
				"error", err,
			)
			return fmt.Errorf("セッションの削除に失敗: %w", err)
		}

		deleted += count
		batches++
		lowerBound = maxUpdatedAt
		if count < cleanupExpiredSessionsBatchSize {
			break
		}

		if time.Since(start) >= uc.timeLimit {
			timeLimitReached = true
			break
		}
	}

	slog.InfoContext(ctx, "セッションクリーンアップが完了しました",
		"cutoff", cutoff,
		"deleted", deleted,
		"batches", batches,
		"elapsed", time.Since(start),
		"time_limit_reached", timeLimitReached,
	)

	return nil
}
