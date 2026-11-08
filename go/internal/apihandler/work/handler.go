// Package workはREST API V2の作品の一覧のハンドラーを提供する
package work

import (
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは作品の一覧のAPIを担当するハンドラー
type Handler struct {
	getAPIWorksUC *usecase.GetAPIWorksUsecase
}

// NewHandlerは新しいHandlerを作成する
func NewHandler(getAPIWorksUC *usecase.GetAPIWorksUsecase) *Handler {
	return &Handler{getAPIWorksUC: getAPIWorksUC}
}
