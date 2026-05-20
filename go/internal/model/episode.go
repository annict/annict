package model

// EpisodeStatus represents the lifecycle state of an episode and mirrors the episode_status PostgreSQL enum.
// [Ja] EpisodeStatus はエピソードのライフサイクル状態を表し、PostgreSQL の episode_status enum と対応する。
type EpisodeStatus string

const (
	EpisodeStatusPublished EpisodeStatus = "published"
	EpisodeStatusArchived  EpisodeStatus = "archived"
	EpisodeStatusDeleted   EpisodeStatus = "deleted"
)

// String returns the textual representation of the status.
// [Ja] ステータスの文字列表現を返す。
func (s EpisodeStatus) String() string { return string(s) }
