// Package db_episode provides HTTP handlers for episodes in the Annict DB admin UI. The
// episode endpoints live under two URL prefixes: the ones addressing a work's episodes as a
// collection are nested under the work (/db/works/{work_id}/episodes), while the ones
// addressing a single episode are keyed by its own id (/db/episodes/{id}). Both belong to the
// episode resource, so one package holds them, as the Rails Db::EpisodesController does.
//
// [Ja] Package db_episode は Annict DB 管理画面のエピソード関連 HTTP ハンドラーを定義する。
// エピソードのエンドポイントは 2 つの URL 接頭辞に分かれ、ある作品のエピソードをコレクション
// として扱うものは作品の下 (/db/works/{work_id}/episodes) に、単一のエピソードを扱うものは
// エピソード自身の id (/db/episodes/{id}) に紐づく。いずれもエピソードというリソースに属する
// ため、Rails の Db::EpisodesController と同じく 1 つのパッケージにまとめる。
package db_episode

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies shared by the episode HTTP handlers in the Annict DB
// admin UI.
//
// [Ja] Handler は Annict DB 管理画面のエピソード関連 HTTP ハンドラーが共有する依存を
// まとめる。
type Handler struct {
	cfg             *config.Config
	getDBEpisodesUC *usecase.GetDBEpisodesUsecase
}

func NewHandler(
	cfg *config.Config,
	getDBEpisodesUC *usecase.GetDBEpisodesUsecase,
) *Handler {
	return &Handler{
		cfg:             cfg,
		getDBEpisodesUC: getDBEpisodesUC,
	}
}
