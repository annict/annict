package viewmodel

import "github.com/annict/annict/go/internal/model"

// WorkID は Presentation 層で使う作品 ID のラッパー型
// Templates が Model に直接依存しないために定義する
type WorkID model.WorkID

// String は文字列表現を返す
func (id WorkID) String() string { return model.WorkID(id).String() }
