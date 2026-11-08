package model

// AnimeLegacyIDは、animeに写した旧作品 (works) と旧エピソード (episodes) のIDの組。
// 旧IDで記録を保存しているAPIの利用者が、animeのIDへ移るために使う。
// 1つのanimeは旧作品か旧エピソードのどちらか1つから写すため、両方が入ることは無い
type AnimeLegacyID struct {
	AnimeID   AnimeID
	WorkID    *WorkID
	EpisodeID *EpisodeID
}
