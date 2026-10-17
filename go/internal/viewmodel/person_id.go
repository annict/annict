package viewmodel

import "github.com/annict/annict/go/internal/model"

// PersonIDはPresentation層で使う人物IDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type PersonID model.PersonID

// Stringは文字列表現を返す
func (id PersonID) String() string { return model.PersonID(id).String() }
