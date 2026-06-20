package viewmodel

import "github.com/annict/annict/go/internal/model"

// StripeWebhookEventID は Presentation 層で使う Stripe Webhook イベント ID のラッパー型
// Templates が Model に直接依存しないために定義する
type StripeWebhookEventID model.StripeWebhookEventID

// String は文字列表現を返す
func (id StripeWebhookEventID) String() string { return model.StripeWebhookEventID(id).String() }
