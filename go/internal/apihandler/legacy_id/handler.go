// Package legacy_idはREST API V2の旧IDから `anime_id` を引くハンドラーを提供する
package legacy_id

import (
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは旧IDの変換のAPIを担当するハンドラー
type Handler struct {
	getAPILegacyIDsUC *usecase.GetAPILegacyIDsUsecase
}

// NewHandlerは新しいHandlerを作成する
func NewHandler(getAPILegacyIDsUC *usecase.GetAPILegacyIDsUsecase) *Handler {
	return &Handler{getAPILegacyIDsUC: getAPILegacyIDsUC}
}
