// Package viewmodelはPresentation層向けに整形済みの型を提供する。
package viewmodel

import (
	"github.com/annict/annict/go/internal/usecase"
)

// TrackingHeatmapCellは視聴記録ヒートマップフラグメントの1日分のセルに表示するデータ。
type TrackingHeatmapCell struct {
	// 指定タイムゾーン上の日付 (YYYY-MM-DD)。
	Date string
	// その日の視聴記録数 (0を含む)。
	Count int
	// CSSクラス選択用の0〜4段階の密度レベル。
	LeveledCount int
}

// NewTrackingHeatmapCellsFromUsecaseはUseCase出力をヒートマップテンプレート
// が受け取るViewModelスライスに変換する。UseCase側で日付の連続性とレベル化済み
// カウントの計算が完了しているため、変換は単純なフィールドコピーになる。
func NewTrackingHeatmapCellsFromUsecase(cells []usecase.TrackingHeatmapCell) []TrackingHeatmapCell {
	result := make([]TrackingHeatmapCell, len(cells))
	for i, c := range cells {
		result[i] = TrackingHeatmapCell{
			Date:         c.Date,
			Count:        c.Count,
			LeveledCount: c.LeveledCount,
		}
	}
	return result
}
