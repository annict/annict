// Package episodeはREST API V2の作品のエピソードの一覧のハンドラーを提供する
package episode

import (
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは作品のエピソードの一覧のAPIを担当するハンドラー
type Handler struct {
	getAPIEpisodesUC *usecase.GetAPIEpisodesUsecase
}

// NewHandlerは新しいHandlerを作成する
func NewHandler(getAPIEpisodesUC *usecase.GetAPIEpisodesUsecase) *Handler {
	return &Handler{getAPIEpisodesUC: getAPIEpisodesUC}
}
