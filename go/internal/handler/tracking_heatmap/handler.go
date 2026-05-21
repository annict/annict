// Package tracking_heatmap provides the HTTP handler for the watch-record
// heatmap fragment shown on the profile page.
//
// [Ja] tracking_heatmap パッケージはプロフィールページに表示する視聴記録ヒートマップ
// フラグメントの HTTP ハンドラーを提供する。
package tracking_heatmap

import (
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies the tracking heatmap fragment endpoint needs.
// [Ja] Handler は視聴記録ヒートマップフラグメントエンドポイントが必要とする依存をまとめる。
type Handler struct {
	getTrackingHeatmapUC *usecase.GetTrackingHeatmapUsecase
}

// NewHandler constructs the Handler.
// [Ja] NewHandler は Handler を生成する。
func NewHandler(getTrackingHeatmapUC *usecase.GetTrackingHeatmapUsecase) *Handler {
	return &Handler{
		getTrackingHeatmapUC: getTrackingHeatmapUC,
	}
}
