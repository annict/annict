package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/worker"
)

// disableTwoFactorAuthはdisable-two-factor-authタスクの本体で、指定したユーザーの二要素認証を本人確認なしで無効にする。
// 本人確認は管理者がCLIの外で済ませる前提で、認証アプリとリカバリーコードの両方を失ったユーザーの救済に使う。
// ユーザーの指定はdelete-userタスクにそろえ、ユーザーIDで受け取る。
func disableTwoFactorAuth(ctx context.Context, _ *config.Config, db *sql.DB, queries *query.Queries, args []string) error {
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || userID <= 0 {
		return fmt.Errorf("ユーザーIDには正の整数を指定してください: %q", args[0])
	}

	// 通知メールのジョブはキューに積むだけで、送信は起動中のサーバーのワーカーが行う。
	txInserter, err := worker.NewTxInserter(db)
	if err != nil {
		return fmt.Errorf("ジョブ投入用のRiverクライアントの作成に失敗しました: %w", err)
	}
	uc := usecase.NewForceDisableTwoFactorAuthUsecase(
		db,
		repository.NewUserTwoFactorAuthRepository(queries),
		repository.NewUserTwoFactorRecoveryCodeRepository(queries),
		repository.NewUserRepository(queries),
		dispatcher.NewDispatcher(txInserter, txInserter),
	)
	output, err := uc.Execute(ctx, usecase.ForceDisableTwoFactorAuthInput{UserID: model.UserID(userID)})
	if err != nil {
		return fmt.Errorf("二要素認証の無効化に失敗しました: %w", err)
	}

	if !output.Disabled {
		slog.WarnContext(ctx, "ユーザーは二要素認証を有効にしていなかったため、無効にする設定はありませんでした", "user_id", userID)
		return nil
	}
	slog.InfoContext(ctx, "管理者がユーザーの二要素認証を無効にしました", "user_id", userID)
	return nil
}
