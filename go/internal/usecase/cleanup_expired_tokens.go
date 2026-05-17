package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// CleanupExpiredTokensUsecase は有効期限切れトークンのクリーンアップを担当します
type CleanupExpiredTokensUsecase struct {
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
}

// NewCleanupExpiredTokensUsecase は新しい CleanupExpiredTokensUsecase を作成します
func NewCleanupExpiredTokensUsecase(passwordResetTokenRepo *repository.PasswordResetTokenRepository) *CleanupExpiredTokensUsecase {
	return &CleanupExpiredTokensUsecase{
		passwordResetTokenRepo: passwordResetTokenRepo,
	}
}

// Execute は24時間以上前に期限切れまたは使用済みになったトークンを削除します
func (uc *CleanupExpiredTokensUsecase) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "トークンクリーンアップを開始します")

	cutoff := time.Now().Add(-24 * time.Hour)

	if err := uc.passwordResetTokenRepo.DeleteExpired(ctx, cutoff); err != nil {
		slog.ErrorContext(ctx, "トークンの削除に失敗しました",
			"cutoff", cutoff,
			"error", err,
		)
		return fmt.Errorf("トークンの削除に失敗: %w", err)
	}

	slog.InfoContext(ctx, "トークンクリーンアップが完了しました",
		"cutoff", cutoff,
	)

	return nil
}
