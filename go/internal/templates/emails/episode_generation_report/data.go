// Package episode_generation_reportはエピソードの自動生成の報告メールのテンプレートを提供します
package episode_generation_report

// Dataは報告メールのテンプレートに渡す値
type Data struct {
	CreatedEpisodes  []CreatedEpisode
	LinkedSlotCount  int
	SkippedSlots     []SkippedSlot
	FilledTitleCount int
	Errors           []Error
}

// CreatedEpisodeは作ったエピソード
type CreatedEpisode struct {
	WorkTitle     string
	EpisodeNumber string
	// EpisodesURLはAnnict DBの作品のエピソード一覧のURL
	EpisodesURL string
}

// SkippedSlotは飛ばした放送枠
type SkippedSlot struct {
	SlotID    string
	WorkTitle string
	Reason    string
}

// Errorは自動生成の途中で起きたエラー。Targetはエラーが起きた放送枠やエピソードで、特定
// できないときは空。
type Error struct {
	Target  string
	Message string
}
