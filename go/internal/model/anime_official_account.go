package model

import "time"

// AnimeAccountService identifies the platform an AnimeOfficialAccount lives on and
// mirrors the anime_account_service PostgreSQL enum. Works are the source of the x
// account only (from twitter_username); the other services are reserved for accounts
// an editor adds to an anime directly.
//
// [Ja] AnimeAccountService は AnimeOfficialAccount が属するプラットフォームを表し、
// PostgreSQL の anime_account_service enum と対応する。works が source とするのは
// x のアカウントのみ (twitter_username 由来) で、他のサービスは編集者が anime に直接
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

// String returns the textual representation of the service.
//
// [Ja] サービスの文字列表現を返す。
func (s AnimeAccountService) String() string { return string(s) }

// AnimeOfficialAccount is the domain entity for anime_official_accounts: it holds an
// anime's official social accounts (the X account, ...) keyed by (anime_id, service).
// It is keyed by the layer-1 anime identity, so the account stays attached across
// re-classification. Account is the bare handle (no leading '@'). Label and LabelEn
// are the optional display labels and are nil when absent (works do not source them,
// so synced rows leave them nil).
//
// [Ja] AnimeOfficialAccount は anime_official_accounts のドメインエンティティ。anime の
// 公式ソーシャルアカウント (X アカウントなど) を (anime_id, service) をキーに持つ。第 1 層の
// anime 同一性をキーにするため、再分類をまたいでもアカウントが紐づき続ける。Account は素の
// ハンドル (先頭の '@' を含まない)。Label / LabelEn は任意の表示ラベルで、無い場合は nil
// (works は source しないため、同期した行では nil のまま)。
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
