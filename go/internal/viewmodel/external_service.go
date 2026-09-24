package viewmodel

import (
	"fmt"
	"strconv"
)

// ExternalServiceLinkは作品一覧の外部サービス列1セル分の値。表示ラベル (外部ID) と
// リンク先URLを持つ。作品にその外部IDが無い場合は両方 "" になり、テンプレートは "-" を描画する。
type ExternalServiceLink struct {
	Label string
	URL   string
}

// SyobocalURLはworks.sc_tidの値に対応するしょぼいカレンダーの番組URLを生成する。
// 一覧とフォームのリンクを揃えるため、RailsのWork#syobocal_urlヘルパーと対応させている。
func SyobocalURL(scTid int32) string {
	return fmt.Sprintf("http://cal.syoboi.jp/tid/%d", scTid)
}

// MalAnimeURLはworks.mal_anime_idの値に対応するMyAnimeListのアニメURLを生成する。
// 一覧とフォームのリンクを揃えるため、RailsのWork#mal_anime_urlヘルパーと対応させている。
func MalAnimeURL(malAnimeID int32) string {
	return fmt.Sprintf("https://myanimelist.net/anime/%d", malAnimeID)
}

// TwitterUsernameURLはworks.twitter_usernameの値に対応するX (旧Twitter) のプロフィール
// URLを生成する。列名・ヘルパー名は歴史的な "twitter" のままだが、Go側ではx.comを参照する。
// ユーザー名が空のときは "" を返し、作品フォーム側で外部リンクを出すかどうかを決められるようにする。
func TwitterUsernameURL(username string) string {
	if username == "" {
		return ""
	}
	return "https://x.com/" + username
}

// TwitterHashtagURLはworks.twitter_hashtagの値に対応するX (旧Twitter) のハッシュタグ
// 検索URLを生成する。列名・ヘルパー名は歴史的な "twitter" のままだが、Go側ではx.comを参照する。
// ハッシュタグが空のときは "" を返す。
func TwitterHashtagURL(hashtag string) string {
	if hashtag == "" {
		return ""
	}
	return "https://x.com/search?q=%23" + hashtag
}

// externalIDURLは送信された外部IDのフォーム値 (文字列) からサービスURLを生成する。
// 値が空、または32bit整数として不正なときは "" を返し、実在のIDに対応する値だけをフォームで
// リンクする。urlForはパース済みのIDをサービスURLに写像する関数 (例: SyobocalURL /
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

// newExternalServiceLinkはNULL許容の外部IDからExternalServiceLinkを生成する。
// IDが未設定のときはゼロ値 (空のラベル / URL) を返す。urlForは設定済みのIDをサービスURLに
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
