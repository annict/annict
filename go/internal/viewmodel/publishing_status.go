package viewmodel

// PublishingStatusはAnnict DB画面に表示するリソースのライフサイクル状態
// (published / archived / deleted) を表すPresentation層の型。ドメインではリソースごとに
// statusのenumを持つ (model.WorkStatus, model.EpisodeStatus) が、いずれも同じ3値を
// 持ち、画面も同じバッジで描画するため、Presentation層ではこの1つの型に射影する。
// これによりcomponents.StatusLabelをリソースごとに用意せず、どのリソースの画面からも
// 使える形に保つ。
//
// 定数は特定のリソースのドメインenumから導出せずリテラルで書いている。すべてのリソースが
// 射影する型のsourceとして、どれか1つのドメインenumを選ぶ自然な理由が無いため。
// 射影が保たれていることは、各ドメイン定数を変換して比較するテストで担保する。
type PublishingStatus string

const (
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
	PublishingStatusDeleted   PublishingStatus = "deleted"
)

// Stringはステータスの文字列表現を返す。
func (s PublishingStatus) String() string { return string(s) }
