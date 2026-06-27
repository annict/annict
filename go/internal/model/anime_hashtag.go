package model

import "time"

// AnimeHashtag is the domain entity for anime_hashtags: it holds an anime's hashtags
// (the tags shown / searched for the work on X), keyed by (anime_id, hashtag). It is
// keyed by the layer-1 anime identity, so a hashtag stays attached across
// re-classification. Hashtag is the bare tag (no leading '#'). SortNumber orders the
// hashtags within an anime; works fix it at 0 for the single tag they source, while a
// hashtag an editor adds directly (a later phase) gets a non-zero value, which is how
// the sync tells its own rows apart from editor-added ones.
//
// [Ja] AnimeHashtag は anime_hashtags のドメインエンティティ。anime のハッシュタグ (X で
// 作品に表示・検索されるタグ) を (anime_id, hashtag) をキーに持つ。第 1 層の anime 同一性を
// キーにするため、再分類をまたいでもハッシュタグが紐づき続ける。Hashtag は素のタグ (先頭の
// '#' を含まない)。SortNumber は anime 内でのハッシュタグの並び順で、works は自身が source
// する単一タグについて 0 で固定する。編集者が直接足すハッシュタグ (後続フェーズ) は非ゼロ値を
// 持ち、同期が自身の行と編集者追加の行を見分ける手がかりになる。
type AnimeHashtag struct {
	ID         AnimeHashtagID
	AnimeID    AnimeID
	Hashtag    string
	SortNumber int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
