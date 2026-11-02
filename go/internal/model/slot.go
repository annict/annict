package model

import "time"

// Slotは番組の1回の放送 (放送枠) を表す。
type Slot struct {
	ID        SlotID
	WorkID    WorkID
	Program   *Program
	Number    int32
	StartedAt time.Time
	EpisodeID *EpisodeID

	// PrecedingIrregularCountは、同じ番組の不定期放送のうち、この放送枠より回数が小さく、
	// 番組のMinimumEpisodeGeneratableNumber以上のものの数。値が入るのは自動生成の対象を
	// 読むローダー (GetForEpisodeGenerationByID) のみで、他のロード経路では0のまま残る。
	PrecedingIrregularCount int32
}

// EpisodeOrderは、不定期放送を除いて数えた、作品内でこの放送枠が何番目の話にあたるかを返す。
// 予定エピソード数との比較と、話数の計算に使う。
func (s *Slot) EpisodeOrder() int32 {
	return s.Number - s.PrecedingIrregularCount
}

// EpisodeRawNumberは、作品の開始話数を起点に、この放送枠が求めるエピソードの話数を返す。
func (s *Slot) EpisodeRawNumber(startEpisodeRawNumber float64) float64 {
	return float64(s.EpisodeOrder()) + startEpisodeRawNumber - 1
}
