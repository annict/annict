package model

import "time"

// AnimeLinkKind identifies what an AnimeLink points to and mirrors the
// anime_link_kind PostgreSQL enum. Works are the source of official_site and
// wikipedia links; other is reserved for links an editor adds to an anime directly.
//
// [Ja] AnimeLinkKind は AnimeLink が指す対象を表し、PostgreSQL の anime_link_kind
// enum と対応する。official_site と wikipedia は works が source とし、other は編集者が
// anime に直接足すリンク向けに予約する。
type AnimeLinkKind string

const (
	AnimeLinkKindOfficialSite AnimeLinkKind = "official_site"
	AnimeLinkKindWikipedia    AnimeLinkKind = "wikipedia"
	AnimeLinkKindOther        AnimeLinkKind = "other"
)

// String returns the textual representation of the kind.
//
// [Ja] kind の文字列表現を返す。
func (k AnimeLinkKind) String() string { return string(k) }

// Language identifies the language of a localized row and mirrors the language
// PostgreSQL enum. It is defined here because anime_links is its first consumer; a
// later table that needs it can lift this type to a shared location.
//
// [Ja] Language はローカライズされた行の言語を表し、PostgreSQL の language enum と対応する。
// anime_links が最初の利用者のためここで定義する。後続のテーブルが必要とした時点で
// 共有の場所へ引き上げてよい。
type Language string

const (
	LanguageJa    Language = "ja"
	LanguageEn    Language = "en"
	LanguageOther Language = "other"
)

// String returns the textual representation of the language.
//
// [Ja] language の文字列表現を返す。
func (l Language) String() string { return string(l) }

// AnimeLink is the domain entity for anime_links: it holds an anime's external links
// (official site, Wikipedia, ...) keyed by (kind, language). It is keyed by the
// layer-1 anime identity, so the link stays attached across re-classification. Label
// and LabelEn are the optional display labels and are nil when absent (works do not
// source them, so synced rows leave them nil).
//
// [Ja] AnimeLink は anime_links のドメインエンティティ。anime の外部リンク (公式サイト・
// Wikipedia など) を (kind, language) をキーに持つ。第 1 層の anime 同一性をキーにするため、
// 再分類をまたいでもリンクが紐づき続ける。Label / LabelEn は任意の表示ラベルで、無い場合は
// nil (works は source しないため、同期した行では nil のまま)。
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
