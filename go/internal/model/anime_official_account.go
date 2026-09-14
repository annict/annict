package model

import "time"

// AnimeAccountServiceはAnimeOfficialAccountが属するプラットフォームを表し、
// PostgreSQLのanime_account_service enumと対応する。worksがsourceとするのは
// xのアカウントのみ (twitter_username由来) で、他のサービスは編集者がanimeに直接
// 足すアカウント向けに予約する。
type AnimeAccountService string

const (
	AnimeAccountServiceBluesky   AnimeAccountService = "bluesky"
	AnimeAccountServiceInstagram AnimeAccountService = "instagram"
	AnimeAccountServiceLine      AnimeAccountService = "line"
	AnimeAccountServiceMastodon  AnimeAccountService = "mastodon"
	AnimeAccountServiceMixi2     AnimeAccountService = "mixi2"
	AnimeAccountServiceThreads   AnimeAccountService = "threads"
	AnimeAccountServiceTiktok    AnimeAccountService = "tiktok"
	AnimeAccountServiceX         AnimeAccountService = "x"
	AnimeAccountServiceYoutube   AnimeAccountService = "youtube"
)

// Stringはサービスの文字列表現を返す。
func (s AnimeAccountService) String() string { return string(s) }

// AnimeOfficialAccountはanime_official_accountsのドメインエンティティ。animeの
// 公式ソーシャルアカウント (Xアカウントなど) を (anime_id, service) をキーに持つ。第1層の
// anime同一性をキーにするため、再分類をまたいでもアカウントが紐づき続ける。Accountは素の
// ハンドル (先頭の '@' を含まない)。Label / LabelEnは任意の表示ラベルで、無い場合はnil
// (worksはsourceしないため、同期した行ではnilのまま)。
type AnimeOfficialAccount struct {
	ID         AnimeOfficialAccountID
	AnimeID    AnimeID
	Service    AnimeAccountService
	Account    string
	Label      *string
	LabelEn    *string
	SortNumber int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
