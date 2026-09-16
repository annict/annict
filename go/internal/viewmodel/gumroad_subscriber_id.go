package viewmodel

import "github.com/annict/annict/go/internal/model"

// GumroadSubscriberIDはPresentation層で使うGumroadサブスクライバーIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type GumroadSubscriberID model.GumroadSubscriberID

// Stringは文字列表現を返す
func (id GumroadSubscriberID) String() string { return model.GumroadSubscriberID(id).String() }
