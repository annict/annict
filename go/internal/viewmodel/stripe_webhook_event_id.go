package viewmodel

import "github.com/annict/annict/go/internal/model"

// StripeWebhookEventIDはPresentation層で使うStripe WebhookイベントIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type StripeWebhookEventID model.StripeWebhookEventID

// Stringは文字列表現を返す
func (id StripeWebhookEventID) String() string { return model.StripeWebhookEventID(id).String() }
