package viewmodel

import "github.com/annict/annict/go/internal/model"

// SlotID は Presentation 層で使う放送枠 ID のラッパー型
// Templates が Model に直接依存しないために定義する
type SlotID model.SlotID

// String は文字列表現を返す
func (id SlotID) String() string { return model.SlotID(id).String() }
