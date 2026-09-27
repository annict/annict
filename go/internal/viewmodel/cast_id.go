package viewmodel

import "github.com/annict/annict/go/internal/model"

// CastIDはPresentation層で使うキャストIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type CastID model.CastID

// Stringは文字列表現を返す
func (id CastID) String() string { return model.CastID(id).String() }
