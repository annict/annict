package model

// EpisodeGenerationSkipReasonは、エピソードの自動生成が放送枠を飛ばした理由を表す。
type EpisodeGenerationSkipReason string

const (
	// EpisodeGenerationSkipReasonCandidatesLinkedは、計算した話数のエピソードが、どれも
	// 同じ番組の別の放送枠に紐づいていることを表す。
	EpisodeGenerationSkipReasonCandidatesLinked EpisodeGenerationSkipReason = "candidates_linked"
	// EpisodeGenerationSkipReasonNoSortNumberGapは、新しいエピソードを入れる位置の前後の
	// sort_numberの間が空いていないことを表す。
	EpisodeGenerationSkipReasonNoSortNumberGap EpisodeGenerationSkipReason = "no_sort_number_gap"
)

// EpisodeGenerationReportは、エピソードの自動生成の1回の実行の結果を、管理者への報告に
// 使う形で表す。
type EpisodeGenerationReport struct {
	CreatedEpisodes []EpisodeGenerationCreatedEpisode
	// LinkedSlotCountは、既存のエピソードに紐づけた放送枠の件数。
	LinkedSlotCount int
	SkippedSlots    []EpisodeGenerationSkippedSlot
	// FilledTitleCountは、しょぼいカレンダーのサブタイトルを入れたエピソードの件数。
	FilledTitleCount int
	Errors           []EpisodeGenerationError
}

// EpisodeGenerationCreatedEpisodeは自動生成で作ったエピソードを表す。
type EpisodeGenerationCreatedEpisode struct {
	WorkID        WorkID
	WorkTitle     string
	EpisodeNumber string
}

// EpisodeGenerationSkippedSlotは、紐づけ先を決められずに飛ばした放送枠を表す。
type EpisodeGenerationSkippedSlot struct {
	SlotID    SlotID
	WorkTitle string
	Reason    EpisodeGenerationSkipReason
}

// EpisodeGenerationErrorは自動生成の途中で起きたエラーを表す。SlotID・EpisodeIDは、
// エラーが特定の放送枠・エピソードの処理で起きたときだけ入る。
type EpisodeGenerationError struct {
	SlotID    *SlotID
	EpisodeID *EpisodeID
	Message   string
}

// IsEmptyは報告することが何も無いかを返す。
func (r *EpisodeGenerationReport) IsEmpty() bool {
	return len(r.CreatedEpisodes) == 0 &&
		r.LinkedSlotCount == 0 &&
		len(r.SkippedSlots) == 0 &&
		r.FilledTitleCount == 0 &&
		len(r.Errors) == 0
}
