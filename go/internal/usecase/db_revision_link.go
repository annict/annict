package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// canShowDBRevisionsは、Annict DBの画面に履歴へのリンクを出すかを返す。履歴の画面はリバースプロキシが
// フィーチャーフラグdb_edit_historyをデバイスかユーザーのどちらかで判定して公開するため、リンクも同じ判定で
// 出す。ログインしていない閲覧者はデバイスだけで判定し、デバイストークンも無いときは出さない。
func canShowDBRevisions(ctx context.Context, featureFlagRepo *repository.FeatureFlagRepository, deviceToken string, user *model.User) (bool, error) {
	var userID model.UserID
	if user != nil {
		userID = user.ID
	}
	if deviceToken == "" && userID == 0 {
		return false, nil
	}
	enabled, err := featureFlagRepo.IsEnabledByDeviceOrUser(ctx, deviceToken, userID, model.FeatureFlagDBEditHistory)
	if err != nil {
		return false, fmt.Errorf("フィーチャーフラグの判定に失敗: %w", err)
	}
	return enabled, nil
}
