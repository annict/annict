package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupExpiredTokensUsecase wires the password reset token cleanup usecase from
// the sqlc queries. Both `serve` (which registers it as a daily periodic job) and
// `task cleanup-expired-tokens` (which runs it once, synchronously) build it through
// this helper, so the wiring the two entry points share lives in one place and cannot
// drift between the scheduled and the manual run.
//
// [Ja] newCleanupExpiredTokensUsecase は sqlc クエリからパスワードリセットトークンの
// クリーンアップ UseCase を組み立てる。`serve` (毎日の定期ジョブとして登録) と
// `task cleanup-expired-tokens` (1 回だけ同期実行) の双方が本ヘルパー経由で組み立てるため、
// 両エントリポイントが共有する配線が 1 箇所にまとまり、定期実行と手動実行でずれない。
func newCleanupExpiredTokensUsecase(queries *query.Queries) *usecase.CleanupExpiredTokensUsecase {
	return usecase.NewCleanupExpiredTokensUsecase(repository.NewPasswordResetTokenRepository(queries))
}

// cleanupExpiredTokens is the body of the cleanup-expired-tokens task: it deletes the
// password reset tokens that expired or were used long enough ago. Unlike the daily
// periodic job that `serve` registers, it does not go through River: it builds the same
// usecase and calls it directly, so that an operator can delete eligible expired or
// used rows on demand without waiting for the schedule. The cutoff, the start and the
// outcome all belong to the usecase, so nothing is added on this side.
//
// [Ja] cleanupExpiredTokens は cleanup-expired-tokens タスクの本体で、十分に時間が経った
// 期限切れ・使用済みのパスワードリセットトークンを削除する。`serve` が登録する毎日の定期
// ジョブと違い River を介さず、同じ UseCase を組み立てて直接呼ぶ。これにより運用者は
// スケジュールを待たずに、対象となる期限切れ・使用済み行を任意のタイミングで削除できる。
// カットオフも開始・完了のログも UseCase 側が持つため、こちらでは何も足さない。
func cleanupExpiredTokens(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupExpiredTokensUsecase(queries).Execute(ctx)
}
