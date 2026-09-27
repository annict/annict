package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupExpiredSessionsUsecaseはsqlcクエリから期限切れセッションのクリーン
// アップUseCaseを組み立てる。`serve` (毎時の定期ジョブとして登録) と
// `task cleanup-expired-sessions` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てる
// ため、両エントリポイントが共有する配線と1回の実行時間の上限が1箇所にまとまり、定期実行
// と手動実行でずれない。
func newCleanupExpiredSessionsUsecase(queries *query.Queries) *usecase.CleanupExpiredSessionsUsecase {
	return usecase.NewCleanupExpiredSessionsUsecase(
		repository.NewSessionRepository(queries),
		usecase.CleanupExpiredSessionsTimeLimit,
	)
}

// cleanupExpiredSessionsはcleanup-expired-sessionsタスクの本体で、期限切れとなる
// まで長期間アクセスされていないセッションを削除する。`serve` が登録する毎時の定期ジョブと
// 違いRiverを介さず、同じUseCaseを組み立てて直接呼ぶ。これにより運用者はスケジュールを
// 待たずに、定期実行と同じ上限の1回分を任意のタイミングで実行できる。上限で打ち切った
// 残りを続けて消したいときは、繰り返し実行する。カットオフもバッチ分割も上限時間も
// 開始・完了のログもUseCase側が持つため、こちらでは何も足さない。
func cleanupExpiredSessions(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupExpiredSessionsUsecase(queries).Execute(ctx)
}
