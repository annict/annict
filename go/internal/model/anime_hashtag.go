package model

import "time"

// AnimeHashtagはanime_hashtagsのドメインエンティティ。animeのハッシュタグ (Xで
// 作品に表示・検索されるタグ) を (anime_id, hashtag) をキーに持つ。第1層のanime同一性を
// キーにするため、再分類をまたいでもハッシュタグが紐づき続ける。Hashtagは素のタグ (先頭の
// '#' を含まない)。SortNumberはanime内でのハッシュタグの並び順で、worksは自身がsource
// する単一タグについて0で固定する。編集者が直接足すハッシュタグ (後続フェーズ) は非ゼロ値を
// 持ち、同期が自身の行と編集者追加の行を見分ける手がかりになる。
type AnimeHashtag struct {
	ID         AnimeHashtagID
	AnimeID    AnimeID
	Hashtag    string
	SortNumber int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
