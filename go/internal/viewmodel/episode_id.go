package viewmodel

import "github.com/annict/annict/go/internal/model"

// EpisodeIDはPresentation層で使うエピソードIDのラッパー型。
// TemplatesがModelに直接依存しないために定義する。
type EpisodeID model.EpisodeID

// StringはIDの文字列表現を返す。
func (id EpisodeID) String() string { return model.EpisodeID(id).String() }
