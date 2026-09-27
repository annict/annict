package viewmodel

import "github.com/annict/annict/go/internal/model"

// WorkIDはPresentation層で使う作品IDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type WorkID model.WorkID

// Stringは文字列表現を返す
func (id WorkID) String() string { return model.WorkID(id).String() }
