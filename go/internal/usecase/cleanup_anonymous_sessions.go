package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// cleanupAnonymousSessionsBatchSizeは1回のDELETEが削除する未ログインのセッションの上限。
// 期限切れセッションのクリーンアップと同じく、ロックの保持とWALの生成を1文あたり一定に
// 抑えるために区切る。
const cleanupAnonymousSessionsBatchSize = 1000

// CleanupAnonymousSessionsTimeLimitは未ログインのセッションのクリーンアップ1回の実行
// 時間の上限。期限切れセッションのクリーンアップ (CleanupExpiredSessionsTimeLimit) とは
// 別に調整できるよう、独立した値として持つ。`serve` の毎時の定期ジョブと
// `task cleanup-anonymous-sessions` の双方がこの値でUseCaseを組み立てる。
const CleanupAnonymousSessionsTimeLimit = 5 * time.Minute

// CleanupAnonymousSessionsUsecaseは未ログインのセッションのクリーンアップを担当する。
type CleanupAnonymousSessionsUsecase struct {
	sessionRepo *repository.SessionRepository
	timeLimit   time.Duration
}

// NewCleanupAnonymousSessionsUsecaseは新しいCleanupAnonymousSessionsUsecaseを作成する。
// timeLimitは1回の実行時間の上限で、超えたら次のバッチに入らずに終了する。
func NewCleanupAnonymousSessionsUsecase(sessionRepo *repository.SessionRepository, timeLimit time.Duration) *CleanupAnonymousSessionsUsecase {
	return &CleanupAnonymousSessionsUsecase{
		sessionRepo: sessionRepo,
		timeLimit:   timeLimit,
	}
}

// Executeはログインしていないセッションのうち、model.AnonymousSessionMaxAgeの間アクセス
// されていないものを、対象が無くなるか実行時間の上限に達するまでバッチで削除する。
// model.SessionMaxAgeより古い行は期限切れセッションのクリーンアップに任せ、ここでは
// 読まない。上限で打ち切った残りは次回の実行に持ち越し、エラーにはしない。
func (uc *CleanupAnonymousSessionsUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "未ログインセッションのクリーンアップを開始します", "time_limit", uc.timeLimit)

	start := time.Now()
	cutoff := start.Add(-model.AnonymousSessionMaxAge)

	// lowerBoundは次のバッチが読み始めるupdated_at。各実行の最初のバッチは期限切れの
	// 境目から読み、以降は直前のバッチで削除した行の最大値を引き継ぐ。これにより、1回の
	// 実行の中では、読み飛ばしたログイン済みの行や削除済みのインデックスエントリを読み直さない。
	lowerBound := start.Add(-model.SessionMaxAge)
	var deleted int64
	var batches int
	timeLimitReached := false
	for {
		count, maxUpdatedAt, err := uc.sessionRepo.DeleteAnonymous(ctx, cutoff, lowerBound, cleanupAnonymousSessionsBatchSize)
		if err != nil {
			slog.ErrorContext(ctx, "未ログインセッションの削除に失敗しました",
				"cutoff", cutoff,
				"lower_bound", lowerBound,
				"deleted", deleted,
				"batches", batches,
				"elapsed", time.Since(start),
				"error", err,
			)
			return fmt.Errorf("未ログインセッションの削除に失敗: %w", err)
		}

		deleted += count
		batches++
		lowerBound = maxUpdatedAt
		if count < cleanupAnonymousSessionsBatchSize {
			break
		}

		if time.Since(start) >= uc.timeLimit {
			timeLimitReached = true
			break
		}
	}

	slog.InfoContext(ctx, "未ログインセッションのクリーンアップが完了しました",
		"cutoff", cutoff,
		"deleted", deleted,
		"batches", batches,
		"elapsed", time.Since(start),
		"time_limit_reached", timeLimitReached,
	)

	return nil
}
