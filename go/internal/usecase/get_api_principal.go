package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetAPIPrincipalUsecaseは、REST API V2のリクエストで受け取ったアクセストークンを照合し、
// 呼び出し主体を返すユースケース。トークンの持ち主にフィーチャーフラグ `rest_api_v2` が有効かもあわせて引く
type GetAPIPrincipalUsecase struct {
	oauthAccessTokenRepo *repository.OAuthAccessTokenRepository
	featureFlagRepo      *repository.FeatureFlagRepository
	now                  func() time.Time
}

func NewGetAPIPrincipalUsecase(
	oauthAccessTokenRepo *repository.OAuthAccessTokenRepository,
	featureFlagRepo *repository.FeatureFlagRepository,
) *GetAPIPrincipalUsecase {
	return &GetAPIPrincipalUsecase{
		oauthAccessTokenRepo: oauthAccessTokenRepo,
		featureFlagRepo:      featureFlagRepo,
		now:                  time.Now,
	}
}

// Executeは `Authorization: Bearer` で受け取ったトークンの文字列 (accessToken) を照合し、呼び出し主体を返す。
// トークンが無い・失効している・期限が切れている・持ち主の利用者がいない場合は、受け付けられないトークンとして (nil, nil) を返す。
// 呼び出し側のミドルウェアがUseCaseのパッケージに依存せずに受け取れるよう、入力は構造体にせず文字列で受け取る
func (uc *GetAPIPrincipalUsecase) Execute(ctx context.Context, accessToken string) (*model.APIPrincipal, error) {
	token, err := uc.oauthAccessTokenRepo.FindByToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("アクセストークンの取得に失敗: %w", err)
	}
	if token == nil || !token.IsUsableAt(uc.now()) {
		return nil, nil
	}

	enabled, err := uc.featureFlagRepo.IsEnabled(ctx, token.ResourceOwnerID, model.FeatureFlagRESTAPIV2)
	if err != nil {
		return nil, fmt.Errorf("フィーチャーフラグの判定に失敗: %w", err)
	}

	kind := model.APITokenKindOAuthAccessToken
	if token.Personal {
		kind = model.APITokenKindPersonalAccessToken
	}
	return &model.APIPrincipal{
		UserID:           token.ResourceOwnerID,
		TokenKind:        kind,
		Scopes:           token.Scopes,
		RESTAPIV2Enabled: enabled,
	}, nil
}
