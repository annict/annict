package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// CleanupExpiredTokensUsecaseは有効期限切れトークンのクリーンアップを担当する。
type CleanupExpiredTokensUsecase struct {
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
}

// NewCleanupExpiredTokensUsecaseは新しいCleanupExpiredTokensUsecaseを作成する。
func NewCleanupExpiredTokensUsecase(passwordResetTokenRepo *repository.PasswordResetTokenRepository) *CleanupExpiredTokensUsecase {
	return &CleanupExpiredTokensUsecase{
		passwordResetTokenRepo: passwordResetTokenRepo,
	}
}

// Executeは24時間以上前に期限切れまたは使用済みになったトークンを削除する。
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
