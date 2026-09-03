package viewmodel

// PublishingStatus is the Presentation-layer lifecycle state (published / archived /
// deleted) of a resource shown on the Annict DB screens. The domain keeps one status enum
// per resource (model.WorkStatus, model.EpisodeStatus), but every one of them names the
// same three values and the screens render them with the same badge, so the Presentation
// layer projects them all onto this single type. That keeps components.StatusLabel usable
// from any resource's screen instead of one label component per resource.
//
// The constants are written as literals rather than derived from one resource's domain enum,
// since no single domain enum is the natural source for a type every resource projects onto.
// The projections are pinned by tests that convert each domain constant and compare.
//
// [Ja] PublishingStatus は Annict DB 画面に表示するリソースのライフサイクル状態
// (published / archived / deleted) を表す Presentation 層の型。ドメインではリソースごとに
// status の enum を持つ (model.WorkStatus, model.EpisodeStatus) が、いずれも同じ 3 値を
// 持ち、画面も同じバッジで描画するため、Presentation 層ではこの 1 つの型に射影する。
// これにより components.StatusLabel をリソースごとに用意せず、どのリソースの画面からも
// 使える形に保つ。
//
// 定数は特定のリソースのドメイン enum から導出せずリテラルで書いている。すべてのリソースが
// 射影する型の source として、どれか 1 つのドメイン enum を選ぶ自然な理由が無いため。
// 射影が保たれていることは、各ドメイン定数を変換して比較するテストで担保する。
type PublishingStatus string

const (
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
	PublishingStatusDeleted   PublishingStatus = "deleted"
)

// String returns the textual representation of the status.
//
// [Ja] ステータスの文字列表現を返す。
func (s PublishingStatus) String() string { return string(s) }
