package main

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// newCleanupAnonymousSessionsUsecaseはsqlcクエリから未ログインセッションのクリーン
// アップUseCaseを組み立てる。`serve` (毎時の定期ジョブとして登録) と
// `task cleanup-anonymous-sessions` (1回だけ同期実行) の双方が本ヘルパー経由で組み立てる
// ため、定期実行と手動実行で配線と1回の実行時間の上限がずれない。
func newCleanupAnonymousSessionsUsecase(queries *query.Queries) *usecase.CleanupAnonymousSessionsUsecase {
	return usecase.NewCleanupAnonymousSessionsUsecase(
		repository.NewSessionRepository(queries),
		usecase.CleanupAnonymousSessionsTimeLimit,
	)
}

// cleanupAnonymousSessionsはcleanup-anonymous-sessionsタスクの本体で、最終アクセスから
// 一定期間を過ぎた未ログインのセッションを削除する。`serve` が登録する毎時の定期ジョブと
// 違いRiverを介さず、同じUseCaseを組み立てて直接呼ぶ。上限で打ち切った残りを続けて
// 消したいときは、繰り返し実行する。
func cleanupAnonymousSessions(ctx context.Context, _ *config.Config, _ *sql.DB, queries *query.Queries) error {
	return newCleanupAnonymousSessionsUsecase(queries).Execute(ctx)
}
