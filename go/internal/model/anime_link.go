package model

import "time"

// AnimeLinkKindはAnimeLinkが指す対象を表し、PostgreSQLのanime_link_kind
// enumと対応する。official_siteとwikipediaはworksがsourceとし、otherは編集者が
// animeに直接足すリンク向けに予約する。
type AnimeLinkKind string

const (
	AnimeLinkKindOfficialSite AnimeLinkKind = "official_site"
	AnimeLinkKindWikipedia    AnimeLinkKind = "wikipedia"
	AnimeLinkKindOther        AnimeLinkKind = "other"
)

// Stringはkindの文字列表現を返す。
func (k AnimeLinkKind) String() string { return string(k) }

// Languageはローカライズされた行の言語を表し、PostgreSQLのlanguage enumと対応する。
// anime_linksが最初の利用者のためここで定義する。後続のテーブルが必要とした時点で
// 共有の場所へ引き上げてよい。
type Language string

const (
	LanguageJa    Language = "ja"
	LanguageEn    Language = "en"
	LanguageOther Language = "other"
)

// Stringはlanguageの文字列表現を返す。
func (l Language) String() string { return string(l) }

// AnimeLinkはanime_linksのドメインエンティティ。animeの外部リンク (公式サイト・
// Wikipediaなど) を (kind, language) をキーに持つ。第1層のanime同一性をキーにするため、
// 再分類をまたいでもリンクが紐づき続ける。Label / LabelEnは任意の表示ラベルで、無い場合は
// nil (worksはsourceしないため、同期した行ではnilのまま)。
type AnimeLink struct {
	ID         AnimeLinkID
	AnimeID    AnimeID
	Kind       AnimeLinkKind
	Language   Language
	URL        string
	Label      *string
	LabelEn    *string
	SortNumber int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
