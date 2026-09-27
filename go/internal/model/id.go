package model

import "strconv"

// UserIDはユーザーのID型
type UserID int64

// Stringは文字列表現を返す
func (id UserID) String() string { return strconv.FormatInt(int64(id), 10) }

// WorkIDは作品のID型
type WorkID int64

// Stringは文字列表現を返す
func (id WorkID) String() string { return strconv.FormatInt(int64(id), 10) }

// EpisodeIDはエピソードのID型
type EpisodeID int64

// Stringは文字列表現を返す
func (id EpisodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// CastIDはキャストのID型
type CastID int64

// Stringは文字列表現を返す
func (id CastID) String() string { return strconv.FormatInt(int64(id), 10) }

// StaffIDはスタッフのID型
type StaffID int64

// Stringは文字列表現を返す
func (id StaffID) String() string { return strconv.FormatInt(int64(id), 10) }

// StripeSubscriberIDはStripeサブスクライバーのID型
type StripeSubscriberID int64

// Stringは文字列表現を返す
func (id StripeSubscriberID) String() string { return strconv.FormatInt(int64(id), 10) }

// GumroadSubscriberIDはGumroadサブスクライバーのID型
type GumroadSubscriberID int64

// Stringは文字列表現を返す
func (id GumroadSubscriberID) String() string { return strconv.FormatInt(int64(id), 10) }

// NumberFormatIDはエピソード番号フォーマットのID型
type NumberFormatID int64

// Stringは文字列表現を返す
func (id NumberFormatID) String() string { return strconv.FormatInt(int64(id), 10) }

// StripeWebhookEventIDはStripe WebhookイベントのID型
type StripeWebhookEventID int64

// Stringは文字列表現を返す
func (id StripeWebhookEventID) String() string { return strconv.FormatInt(int64(id), 10) }

// SlotIDは放送枠のID型
type SlotID int64

// Stringは文字列表現を返す
func (id SlotID) String() string { return strconv.FormatInt(int64(id), 10) }

// ProfileIDはプロフィールのID型
type ProfileID int64

// Stringは文字列表現を返す
func (id ProfileID) String() string { return strconv.FormatInt(int64(id), 10) }

// SettingIDは設定のID型
type SettingID int64

// Stringは文字列表現を返す
func (id SettingID) String() string { return strconv.FormatInt(int64(id), 10) }

// EmailNotificationIDはメール通知設定のID型
type EmailNotificationID int64

// Stringは文字列表現を返す
func (id EmailNotificationID) String() string { return strconv.FormatInt(int64(id), 10) }

// PasswordResetTokenIDはパスワードリセットトークンのID型
type PasswordResetTokenID int64

// Stringは文字列表現を返す
func (id PasswordResetTokenID) String() string { return strconv.FormatInt(int64(id), 10) }

// SignInCodeIDはサインインコードのID型
type SignInCodeID int64

// Stringは文字列表現を返す
func (id SignInCodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// SignUpCodeIDはサインアップコードのID型
type SignUpCodeID int64

// Stringは文字列表現を返す
func (id SignUpCodeID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeIDはコンテンツ同一性テーブル (animes) のID型。
type AnimeID int64

// StringはIDの文字列表現を返す。
func (id AnimeID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeClassificationIDはカタログ分類テーブル (anime_classifications) のID型。
type AnimeClassificationID int64

// StringはIDの文字列表現を返す。
func (id AnimeClassificationID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeExternalIDIDはanime_external_idsテーブルの主キーID型。"ID" が
// 重なるのは {Entity}ID規約をAnimeExternalIDエンティティに機械的に適用した結果で、
// external_idカラムに格納する外部サービスのidとは別の、行自身のidを表す。
type AnimeExternalIDID int64

// StringはIDの文字列表現を返す。
func (id AnimeExternalIDID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeLinkIDはanime_linksテーブルの主キーID型。
type AnimeLinkID int64

// StringはIDの文字列表現を返す。
func (id AnimeLinkID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeOfficialAccountIDはanime_official_accountsテーブルの主キーID型。
type AnimeOfficialAccountID int64

// StringはIDの文字列表現を返す。
func (id AnimeOfficialAccountID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeHashtagIDはanime_hashtagsテーブルの主キーID型。
type AnimeHashtagID int64

// StringはIDの文字列表現を返す。
func (id AnimeHashtagID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeSeasonIDはanime_seasonsテーブルの主キーID型。
type AnimeSeasonID int64

// StringはIDの文字列表現を返す。
func (id AnimeSeasonID) String() string { return strconv.FormatInt(int64(id), 10) }

// AnimeEventIDはanime_eventsテーブルの主キーID型。
type AnimeEventID int64

// StringはIDの文字列表現を返す。
func (id AnimeEventID) String() string { return strconv.FormatInt(int64(id), 10) }

// FeatureFlagIDはフィーチャーフラグのID型
type FeatureFlagID int64

// Stringは文字列表現を返す
func (id FeatureFlagID) String() string { return strconv.FormatInt(int64(id), 10) }

// FeatureFlagNameはフィーチャーフラグ名の型
type FeatureFlagName string

// Stringは文字列表現を返す
func (n FeatureFlagName) String() string { return string(n) }
