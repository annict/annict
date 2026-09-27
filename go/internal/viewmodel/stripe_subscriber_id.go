package viewmodel

import "github.com/annict/annict/go/internal/model"

// StripeSubscriberIDはPresentation層で使うStripeサブスクライバーIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type StripeSubscriberID model.StripeSubscriberID

// Stringは文字列表現を返す
func (id StripeSubscriberID) String() string { return model.StripeSubscriberID(id).String() }
