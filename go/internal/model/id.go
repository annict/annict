package model

import "strconv"

// UserID はユーザーのID型
type UserID int64

// String は文字列表現を返す
func (id UserID) String() string { return strconv.FormatInt(int64(id), 10) }

// WorkID は作品のID型
type WorkID int64

// String は文字列表現を返す
func (id WorkID) String() string { return strconv.FormatInt(int64(id), 10) }

// EpisodeID はエピソードのID型
type EpisodeID int64

// String は文字列表現を返す
func (id EpisodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// CastID はキャストのID型
type CastID int64

// String は文字列表現を返す
func (id CastID) String() string { return strconv.FormatInt(int64(id), 10) }

// StaffID はスタッフのID型
type StaffID int64

// String は文字列表現を返す
func (id StaffID) String() string { return strconv.FormatInt(int64(id), 10) }

// StripeSubscriberID はStripeサブスクライバーのID型
type StripeSubscriberID int64

// String は文字列表現を返す
func (id StripeSubscriberID) String() string { return strconv.FormatInt(int64(id), 10) }

// GumroadSubscriberID はGumroadサブスクライバーのID型
type GumroadSubscriberID int64

// String は文字列表現を返す
func (id GumroadSubscriberID) String() string { return strconv.FormatInt(int64(id), 10) }

// NumberFormatID はエピソード番号フォーマットのID型
type NumberFormatID int64

// String は文字列表現を返す
func (id NumberFormatID) String() string { return strconv.FormatInt(int64(id), 10) }

// StripeWebhookEventID はStripe WebhookイベントのID型
type StripeWebhookEventID int64

// String は文字列表現を返す
func (id StripeWebhookEventID) String() string { return strconv.FormatInt(int64(id), 10) }

// SlotID は放送枠のID型
type SlotID int64

// String は文字列表現を返す
func (id SlotID) String() string { return strconv.FormatInt(int64(id), 10) }

// ProfileID はプロフィールのID型
type ProfileID int64

// String は文字列表現を返す
func (id ProfileID) String() string { return strconv.FormatInt(int64(id), 10) }

// SettingID は設定のID型
type SettingID int64

// String は文字列表現を返す
func (id SettingID) String() string { return strconv.FormatInt(int64(id), 10) }

// EmailNotificationID はメール通知設定のID型
type EmailNotificationID int64

// String は文字列表現を返す
func (id EmailNotificationID) String() string { return strconv.FormatInt(int64(id), 10) }

// PasswordResetTokenID はパスワードリセットトークンのID型
type PasswordResetTokenID int64

// String は文字列表現を返す
func (id PasswordResetTokenID) String() string { return strconv.FormatInt(int64(id), 10) }

// SignInCodeID はサインインコードのID型
type SignInCodeID int64

// String は文字列表現を返す
func (id SignInCodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// SignUpCodeID はサインアップコードのID型
type SignUpCodeID int64

// String は文字列表現を返す
func (id SignUpCodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeID is the ID type for the content-identity table (animes).
//
// [Ja] AnimeID はコンテンツ同一性テーブル (animes) の ID 型。
type AnimeID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeClassificationID is the ID type for the catalog-classification table
// (anime_classifications).
//
// [Ja] AnimeClassificationID はカタログ分類テーブル (anime_classifications) の ID 型。
type AnimeClassificationID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeClassificationID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeExternalIDID is the primary-key ID type for the anime_external_ids table.
// The doubled "ID" is the mechanical result of the {Entity}ID convention applied
// to the AnimeExternalID entity; it names the row's own id, distinct from the
// external service's id stored in the external_id column.
//
// [Ja] AnimeExternalIDID は anime_external_ids テーブルの主キー ID 型。"ID" が
// 重なるのは {Entity}ID 規約を AnimeExternalID エンティティに機械的に適用した結果で、
// external_id カラムに格納する外部サービスの id とは別の、行自身の id を表す。
type AnimeExternalIDID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeExternalIDID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeLinkID is the primary-key ID type for the anime_links table.
//
// [Ja] AnimeLinkID は anime_links テーブルの主キー ID 型。
type AnimeLinkID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeLinkID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeOfficialAccountID is the primary-key ID type for the anime_official_accounts table.
//
// [Ja] AnimeOfficialAccountID は anime_official_accounts テーブルの主キー ID 型。
type AnimeOfficialAccountID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeOfficialAccountID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeHashtagID is the primary-key ID type for the anime_hashtags table.
//
// [Ja] AnimeHashtagID は anime_hashtags テーブルの主キー ID 型。
type AnimeHashtagID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeHashtagID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeSeasonID is the primary-key ID type for the anime_seasons table.
//
// [Ja] AnimeSeasonID は anime_seasons テーブルの主キー ID 型。
type AnimeSeasonID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeSeasonID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeEventID is the primary-key ID type for the anime_events table.
//
// [Ja] AnimeEventID は anime_events テーブルの主キー ID 型。
type AnimeEventID int64

// String returns the textual representation of the ID.
//
// [Ja] String は ID の文字列表現を返す。
func (id AnimeEventID) String() string { return strconv.FormatInt(int64(id), 10) }

// FeatureFlagID はフィーチャーフラグのID型
type FeatureFlagID int64

// String は文字列表現を返す
func (id FeatureFlagID) String() string { return strconv.FormatInt(int64(id), 10) }

// FeatureFlagName はフィーチャーフラグ名の型
type FeatureFlagName string

// String は文字列表現を返す
func (n FeatureFlagName) String() string { return string(n) }
