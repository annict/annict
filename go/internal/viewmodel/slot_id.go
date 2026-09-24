package viewmodel

import "github.com/annict/annict/go/internal/model"

// SlotIDはPresentation層で使う放送枠IDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type SlotID model.SlotID

// Stringは文字列表現を返す
func (id SlotID) String() string { return model.SlotID(id).String() }
