package viewmodel

import "github.com/annict/annict/go/internal/model"

// StripeSubscriberID は Presentation 層で使う Stripe サブスクライバー ID のラッパー型
// Templates が Model に直接依存しないために定義する
type StripeSubscriberID model.StripeSubscriberID

// String は文字列表現を返す
func (id StripeSubscriberID) String() string { return model.StripeSubscriberID(id).String() }
