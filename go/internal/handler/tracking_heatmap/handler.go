// Package tracking_heatmapはプロフィールページに表示する視聴記録ヒートマップ
// フラグメントのHTTPハンドラーを提供する。
package tracking_heatmap

import (
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは視聴記録ヒートマップフラグメントエンドポイントが必要とする依存をまとめる。
type Handler struct {
	getTrackingHeatmapUC *usecase.GetTrackingHeatmapUsecase
}

// NewHandlerはHandlerを生成する。
func NewHandler(getTrackingHeatmapUC *usecase.GetTrackingHeatmapUsecase) *Handler {
	return &Handler{
		getTrackingHeatmapUC: getTrackingHeatmapUC,
	}
}
