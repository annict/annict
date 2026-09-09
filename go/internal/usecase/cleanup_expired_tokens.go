package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// CleanupExpiredTokensUsecase is responsible for cleaning up expired password reset tokens.
//
// [Ja] CleanupExpiredTokensUsecase は有効期限切れトークンのクリーンアップを担当する。
type CleanupExpiredTokensUsecase struct {
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
}

// NewCleanupExpiredTokensUsecase creates a new CleanupExpiredTokensUsecase.
//
// [Ja] NewCleanupExpiredTokensUsecase は新しい CleanupExpiredTokensUsecase を作成する。
func NewCleanupExpiredTokensUsecase(passwordResetTokenRepo *repository.PasswordResetTokenRepository) *CleanupExpiredTokensUsecase {
	return &CleanupExpiredTokensUsecase{
		passwordResetTokenRepo: passwordResetTokenRepo,
	}
}

// Execute deletes tokens that expired or were used more than 24 hours ago.
//
// [Ja] Execute は 24 時間以上前に期限切れまたは使用済みになったトークンを削除する。
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
