// Package viewmodel exposes types prepared for the presentation layer.
// [Ja] viewmodel パッケージは Presentation 層向けに整形済みの型を提供する。
package viewmodel

import (
	"github.com/annict/annict/go/internal/usecase"
)

// TrackingHeatmapCell is the per-day display data for one cell of the
// tracking heatmap fragment.
//
// [Ja] TrackingHeatmapCell は視聴記録ヒートマップフラグメントの 1 日分のセルに表示するデータ。
type TrackingHeatmapCell struct {
	// Date is the day formatted as "YYYY-MM-DD" in the requested time zone.
	// [Ja] 指定タイムゾーン上の日付 (YYYY-MM-DD)。
	Date string
	// Count is the number of records on Date (0 included).
	// [Ja] その日の視聴記録数 (0 を含む)。
	Count int
	// LeveledCount is the 0-4 density bucket that picks the CSS density class.
	// [Ja] CSS クラス選択用の 0〜4 段階の密度レベル。
	LeveledCount int
}

// NewTrackingHeatmapCellsFromUsecase converts the use case output into the
// view model slice used by the heatmap template. The use case already returns
// a contiguous list of days with the leveled count computed, so the conversion
// is a simple field copy.
//
// [Ja] NewTrackingHeatmapCellsFromUsecase は UseCase 出力をヒートマップテンプレート
// が受け取る ViewModel スライスに変換する。UseCase 側で日付の連続性とレベル化済み
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
