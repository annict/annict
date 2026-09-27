package viewmodel

import "github.com/annict/annict/go/internal/model"

// NumberFormatIDはPresentation層で使うエピソード番号フォーマットIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type NumberFormatID model.NumberFormatID

// Stringは文字列表現を返す
func (id NumberFormatID) String() string { return model.NumberFormatID(id).String() }
