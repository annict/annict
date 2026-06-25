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

// FeatureFlagID はフィーチャーフラグのID型
type FeatureFlagID int64

// String は文字列表現を返す
func (id FeatureFlagID) String() string { return strconv.FormatInt(int64(id), 10) }

// FeatureFlagName はフィーチャーフラグ名の型
type FeatureFlagName string

// String は文字列表現を返す
func (n FeatureFlagName) String() string { return string(n) }
