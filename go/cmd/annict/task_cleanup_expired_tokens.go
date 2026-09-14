package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupExpiredTokensUsecaseはsqlcクエリからパスワードリセットトークンの
// クリーンアップUseCaseを組み立てる。`serve` (毎日の定期ジョブとして登録) と
// `task cleanup-expired-tokens` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てるため、
// 両エントリポイントが共有する配線が1箇所にまとまり、定期実行と手動実行でずれない。
func newCleanupExpiredTokensUsecase(queries *query.Queries) *usecase.CleanupExpiredTokensUsecase {
	return usecase.NewCleanupExpiredTokensUsecase(repository.NewPasswordResetTokenRepository(queries))
}

// cleanupExpiredTokensはcleanup-expired-tokensタスクの本体で、十分に時間が経った
// 期限切れ・使用済みのパスワードリセットトークンを削除する。`serve` が登録する毎日の定期
// ジョブと違いRiverを介さず、同じUseCaseを組み立てて直接呼ぶ。これにより運用者は
// スケジュールを待たずに、対象となる期限切れ・使用済み行を任意のタイミングで削除できる。
// カットオフも開始・完了のログもUseCase側が持つため、こちらでは何も足さない。
func cleanupExpiredTokens(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupExpiredTokensUsecase(queries).Execute(ctx)
}
