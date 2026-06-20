package viewmodel

import "github.com/annict/annict/go/internal/model"

// GumroadSubscriberID は Presentation 層で使う Gumroad サブスクライバー ID のラッパー型
// Templates が Model に直接依存しないために定義する
type GumroadSubscriberID model.GumroadSubscriberID

// String は文字列表現を返す
func (id GumroadSubscriberID) String() string { return model.GumroadSubscriberID(id).String() }
