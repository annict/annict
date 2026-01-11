package model

// WebhookEventStatus はWebhookイベントの処理状態を表す型です
type WebhookEventStatus string

const (
	// WebhookEventStatusPending は受信済み・処理待ちの状態
	WebhookEventStatusPending WebhookEventStatus = "pending"
	// WebhookEventStatusProcessed は処理完了の状態
	WebhookEventStatusProcessed WebhookEventStatus = "processed"
	// WebhookEventStatusFailed は処理失敗の状態
	WebhookEventStatusFailed WebhookEventStatus = "failed"
	// WebhookEventStatusSkipped は処理対象外のイベント（ログのみ）の状態
	WebhookEventStatusSkipped WebhookEventStatus = "skipped"
)

// String はWebhookEventStatusの文字列表現を返します
func (s WebhookEventStatus) String() string {
	return string(s)
}
