package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// CleanupExpiredSignInCodesUsecaseは期限切れログインコードのクリーンアップを担当する。
type CleanupExpiredSignInCodesUsecase struct {
	signInCodeRepo *repository.SignInCodeRepository
}

// NewCleanupExpiredSignInCodesUsecaseは新しいCleanupExpiredSignInCodesUsecaseを作成する。
func NewCleanupExpiredSignInCodesUsecase(signInCodeRepo *repository.SignInCodeRepository) *CleanupExpiredSignInCodesUsecase {
	return &CleanupExpiredSignInCodesUsecase{
		signInCodeRepo: signInCodeRepo,
	}
}

// Executeは24時間以上前に期限切れまたは使用済みになったログインコードを削除する。
func (uc *CleanupExpiredSignInCodesUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "ログインコードクリーンアップを開始します")

	cutoff := time.Now().Add(-24 * time.Hour)

	if err := uc.signInCodeRepo.DeleteExpired(ctx, cutoff); err != nil {
		slog.ErrorContext(ctx, "ログインコードの削除に失敗しました",
			"cutoff", cutoff,
			"error", err,
		)
		return fmt.Errorf("ログインコードの削除に失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログインコードクリーンアップが完了しました",
		"cutoff", cutoff,
	)

	return nil
}
