package viewmodel

import "github.com/annict/annict/go/internal/model"

// EpisodeID is the Presentation-layer wrapper for an episode id, defined so templates do not
// depend on the model package directly.
//
// [Ja] EpisodeID は Presentation 層で使うエピソード ID のラッパー型。
// Templates が Model に直接依存しないために定義する。
type EpisodeID model.EpisodeID

// String returns the textual representation of the id.
//
// [Ja] ID の文字列表現を返す。
func (id EpisodeID) String() string { return model.EpisodeID(id).String() }
