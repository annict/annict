package viewmodel

import "github.com/annict/annict/go/internal/model"

// CastID は Presentation 層で使うキャスト ID のラッパー型
// Templates が Model に直接依存しないために定義する
type CastID model.CastID

// String は文字列表現を返す
func (id CastID) String() string { return model.CastID(id).String() }
