package viewmodel

import "github.com/annict/annict/go/internal/model"

// StaffIDはPresentation層で使うスタッフIDのラッパー型
// TemplatesがModelに直接依存しないために定義する
type StaffID model.StaffID

// Stringは文字列表現を返す
func (id StaffID) String() string { return model.StaffID(id).String() }
