package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// isTwoFactorAuthEnabledは、ユーザーが二要素認証を有効にしているかを返す。
// 第1の要素 (パスワードかメールのコード) を確かめたあと、セッションを発行する前に認証アプリのコードを求めるかを決めるのに使う。
// 登録の途中の設定は、まだ認証アプリのコードを確かめていないため無効として扱う。
func isTwoFactorAuthEnabled(ctx context.Context, repo *repository.UserTwoFactorAuthRepository, userID model.UserID) (bool, error) {
	setting, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("二要素認証の設定の取得に失敗: %w", err)
	}

	return setting != nil && setting.IsEnabled(), nil
}
