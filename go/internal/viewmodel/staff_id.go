package viewmodel

import "github.com/annict/annict/go/internal/model"

// StaffID は Presentation 層で使うスタッフ ID のラッパー型
// Templates が Model に直接依存しないために定義する
type StaffID model.StaffID

// String は文字列表現を返す
func (id StaffID) String() string { return model.StaffID(id).String() }
