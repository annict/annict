package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupExpiredSessionsUsecase wires the expired session cleanup usecase from the
// sqlc queries. Both `serve` (which registers it as a daily periodic job) and
// `task cleanup-expired-sessions` (which runs it once, synchronously) build it through
// this helper, so the wiring the two entry points share lives in one place and cannot
// drift between the scheduled and the manual run.
//
// [Ja] newCleanupExpiredSessionsUsecase は sqlc クエリから期限切れセッションのクリーン
// アップ UseCase を組み立てる。`serve` (毎日の定期ジョブとして登録) と
// `task cleanup-expired-sessions` (1 回だけ同期実行) の双方が本ヘルパー経由で組み立てる
// ため、両エントリポイントが共有する配線が 1 箇所にまとまり、定期実行と手動実行でずれない。
func newCleanupExpiredSessionsUsecase(queries *query.Queries) *usecase.CleanupExpiredSessionsUsecase {
	return usecase.NewCleanupExpiredSessionsUsecase(repository.NewSessionRepository(queries))
}

// cleanupExpiredSessions is the body of the cleanup-expired-sessions task: it deletes
// the sessions that have not been accessed for long enough to have expired. Unlike the
// daily periodic job that `serve` registers, it does not go through River: it builds the
// same usecase and calls it directly, so that an operator can drain a backlog of expired
// sessions on demand without waiting for the schedule. The cutoff, the batching and the
// outcome all belong to the usecase, so nothing is added on this side.
//
// [Ja] cleanupExpiredSessions は cleanup-expired-sessions タスクの本体で、期限切れとなる
// まで長期間アクセスされていないセッションを削除する。`serve` が登録する毎日の定期ジョブと
// 違い River を介さず、同じ UseCase を組み立てて直接呼ぶ。これにより運用者はスケジュールを
// 待たずに、滞留した期限切れセッションを任意のタイミングで解消できる。カットオフもバッチ
// 分割も開始・完了のログも UseCase 側が持つため、こちらでは何も足さない。
func cleanupExpiredSessions(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupExpiredSessionsUsecase(queries).Execute(ctx)
}
