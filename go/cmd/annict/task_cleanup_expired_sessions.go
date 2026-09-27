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
// アップUseCaseを組み立てる。`serve` (毎日の定期ジョブとして登録) と
// `task cleanup-expired-sessions` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てる
// ため、両エントリポイントが共有する配線が1箇所にまとまり、定期実行と手動実行でずれない。
func newCleanupExpiredSessionsUsecase(queries *query.Queries) *usecase.CleanupExpiredSessionsUsecase {
	return usecase.NewCleanupExpiredSessionsUsecase(repository.NewSessionRepository(queries))
}

// cleanupExpiredSessionsはcleanup-expired-sessionsタスクの本体で、期限切れとなる
// まで長期間アクセスされていないセッションを削除する。`serve` が登録する毎日の定期ジョブと
// 違いRiverを介さず、同じUseCaseを組み立てて直接呼ぶ。これにより運用者はスケジュールを
// 待たずに、滞留した期限切れセッションを任意のタイミングで解消できる。カットオフもバッチ
// 分割も開始・完了のログもUseCase側が持つため、こちらでは何も足さない。
func cleanupExpiredSessions(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupExpiredSessionsUsecase(queries).Execute(ctx)
}
