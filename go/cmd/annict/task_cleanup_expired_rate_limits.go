package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupExpiredRateLimitsUsecaseはsqlcクエリからレート制限のカウンターのクリーンアップ
// UseCaseを組み立てる。`serve` (毎時の定期ジョブとして登録) と
// `task cleanup-expired-rate-limits` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てる
// ため、両エントリポイントが共有する配線が1箇所にまとまり、定期実行と手動実行でずれない。
func newCleanupExpiredRateLimitsUsecase(queries *query.Queries) *usecase.CleanupExpiredRateLimitsUsecase {
	return usecase.NewCleanupExpiredRateLimitsUsecase(repository.NewRateLimitRepository(queries))
}

// cleanupExpiredRateLimitsはcleanup-expired-rate-limitsタスクの本体で、終わった窓の
// レート制限のカウンターを削除する。`serve` が登録する毎時の定期ジョブと違いRiverを介さず、
// 同じUseCaseを組み立てて直接呼ぶ。カットオフも開始・完了のログもUseCase側が持つため、
// こちらでは何も足さない。
func cleanupExpiredRateLimits(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries, _ []string) error {
	return newCleanupExpiredRateLimitsUsecase(queries).Execute(ctx)
}
