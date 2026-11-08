package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// apiLegacyIDsMaxCountは、旧IDの変換で1回に受け付ける旧IDの数の上限。種類をまたいで数える
const apiLegacyIDsMaxCount = 100

// APILegacyIDIndexValidatorは、REST API V2で旧IDから `anime_id` を引くときのパラメーターを検証する。
// 種類ごとの形式と件数はOpenAPIの記述で検証し、ここでは種類をまたぐ条件だけを検証する
type APILegacyIDIndexValidator struct{}

// NewAPILegacyIDIndexValidatorはAPILegacyIDIndexValidatorを生成する
func NewAPILegacyIDIndexValidator() *APILegacyIDIndexValidator {
	return &APILegacyIDIndexValidator{}
}

// APILegacyIDIndexValidatorInputはバリデーションの入力パラメータ
type APILegacyIDIndexValidatorInput struct {
	WorkIDs    []model.WorkID
	EpisodeIDs []model.EpisodeID
}

// Validateは、旧IDが1つ以上あり、合計が上限以下であることを確かめる。
// 旧IDが1つも無い要求を空の結果で通すと、パラメーター名の書き間違いに利用者が気づけないため、エラーにする
func (v *APILegacyIDIndexValidator) Validate(ctx context.Context, input APILegacyIDIndexValidatorInput) error {
	ve := model.NewValidationError()

	count := len(input.WorkIDs) + len(input.EpisodeIDs)
	switch {
	case count == 0:
		ve.AddGlobal(i18n.T(ctx, "validation_api_legacy_ids_required"))
	case count > apiLegacyIDsMaxCount:
		ve.AddGlobal(i18n.T(ctx, "validation_api_legacy_ids_too_many", map[string]any{"Max": apiLegacyIDsMaxCount}))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
