package viewmodel

import (
	"fmt"
	"strconv"
)

// ExternalServiceLink is one external-service cell value in the work list: the
// display label (the external id) and its target URL. Both are "" when the work
// has no id registered for that service, in which case the template renders "-".
//
// [Ja] ExternalServiceLink は作品一覧の外部サービス列 1 セル分の値。表示ラベル (外部 ID) と
// リンク先 URL を持つ。作品にその外部 ID が無い場合は両方 "" になり、テンプレートは "-" を描画する。
type ExternalServiceLink struct {
	Label string
	URL   string
}

// SyobocalURL builds the Syoboi Calendar program URL for a works.sc_tid value,
// mirroring the Rails Work#syobocal_url helper so the list and form links stay in sync.
//
// [Ja] SyobocalURL は works.sc_tid の値に対応するしょぼいカレンダーの番組 URL を生成する。
// 一覧とフォームのリンクを揃えるため、Rails の Work#syobocal_url ヘルパーと対応させている。
func SyobocalURL(scTid int32) string {
	return fmt.Sprintf("http://cal.syoboi.jp/tid/%d", scTid)
}

// MalAnimeURL builds the MyAnimeList anime URL for a works.mal_anime_id value,
// mirroring the Rails Work#mal_anime_url helper so the list and form links stay in sync.
//
// [Ja] MalAnimeURL は works.mal_anime_id の値に対応する MyAnimeList のアニメ URL を生成する。
// 一覧とフォームのリンクを揃えるため、Rails の Work#mal_anime_url ヘルパーと対応させている。
func MalAnimeURL(malAnimeID int32) string {
	return fmt.Sprintf("https://myanimelist.net/anime/%d", malAnimeID)
}

// TwitterUsernameURL builds the X (formerly Twitter) profile URL for a works.twitter_username
// value. The Go side targets x.com even though the column and helper keep the historical
// "twitter" name. It returns "" when the username is empty so the work form can decide whether
// to show the external link.
//
// [Ja] TwitterUsernameURL は works.twitter_username の値に対応する X (旧 Twitter) のプロフィール
// URL を生成する。列名・ヘルパー名は歴史的な "twitter" のままだが、Go 側では x.com を参照する。
// ユーザー名が空のときは "" を返し、作品フォーム側で外部リンクを出すかどうかを決められるようにする。
func TwitterUsernameURL(username string) string {
	if username == "" {
		return ""
	}
	return "https://x.com/" + username
}

// TwitterHashtagURL builds the X (formerly Twitter) hashtag-search URL for a works.twitter_hashtag
// value. The Go side targets x.com even though the column and helper keep the historical
// "twitter" name. It returns "" when the hashtag is empty.
//
// [Ja] TwitterHashtagURL は works.twitter_hashtag の値に対応する X (旧 Twitter) のハッシュタグ
// 検索 URL を生成する。列名・ヘルパー名は歴史的な "twitter" のままだが、Go 側では x.com を参照する。
// ハッシュタグが空のときは "" を返す。
func TwitterHashtagURL(hashtag string) string {
	if hashtag == "" {
		return ""
	}
	return "https://x.com/search?q=%23" + hashtag
}

// externalIDURL builds a service URL from a submitted external-id form value (a string).
// It returns "" when the value is empty or not a valid 32-bit integer, so the work form
// only links values that map to a real id. urlFor maps the parsed id to its service URL
// (e.g. SyobocalURL / MalAnimeURL).
//
// [Ja] externalIDURL は送信された外部 ID のフォーム値 (文字列) からサービス URL を生成する。
// 値が空、または 32bit 整数として不正なときは "" を返し、実在の ID に対応する値だけをフォームで
// リンクする。urlFor はパース済みの ID をサービス URL に写像する関数 (例: SyobocalURL /
// MalAnimeURL)。
func externalIDURL(value string, urlFor func(int32) string) string {
	if value == "" {
		return ""
	}
	id, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return ""
	}
	return urlFor(int32(id))
}

// newExternalServiceLink builds an ExternalServiceLink from a nullable external id,
// returning the zero value (empty label / URL) when the id is unset. urlFor maps a
// present id to its service URL (e.g. SyobocalURL / MalAnimeURL).
//
// [Ja] newExternalServiceLink は NULL 許容の外部 ID から ExternalServiceLink を生成する。
// ID が未設定のときはゼロ値 (空のラベル / URL) を返す。urlFor は設定済みの ID をサービス URL に
// 写像する関数 (例: SyobocalURL / MalAnimeURL)。
func newExternalServiceLink(id *int32, urlFor func(int32) string) ExternalServiceLink {
	if id == nil {
		return ExternalServiceLink{}
	}
	return ExternalServiceLink{
		Label: strconv.FormatInt(int64(*id), 10),
		URL:   urlFor(*id),
	}
}
