package model

// WebhookEventStatusはWebhookイベントの処理状態を表す型です
type WebhookEventStatus string

const (
	// WebhookEventStatusPendingは受信済み・処理待ちの状態
	WebhookEventStatusPending WebhookEventStatus = "pending"
	// WebhookEventStatusProcessedは処理完了の状態
	WebhookEventStatusProcessed WebhookEventStatus = "processed"
	// WebhookEventStatusFailedは処理失敗の状態
	WebhookEventStatusFailed WebhookEventStatus = "failed"
	// WebhookEventStatusSkippedは処理対象外のイベント (ログのみ) の状態
	WebhookEventStatusSkipped WebhookEventStatus = "skipped"
)

// StringはWebhookEventStatusの文字列表現を返します
func (s WebhookEventStatus) String() string {
	return string(s)
}
