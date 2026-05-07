package viewmodel

import "github.com/annict/annict/go/internal/model"

// NumberFormatID は Presentation 層で使うエピソード番号フォーマット ID のラッパー型
// Templates が Model に直接依存しないために定義する
type NumberFormatID model.NumberFormatID

// String は文字列表現を返す
func (id NumberFormatID) String() string { return model.NumberFormatID(id).String() }
