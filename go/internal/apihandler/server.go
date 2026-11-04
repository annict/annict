// Package apihandlerはREST API V2のハンドラーをまとめる。
// oapi-codegenが生成する `apigen.StrictServerInterface` は全operationを1つのインターフェースにまとめるため、
// リソースごとのディレクトリに置いたHandlerを埋め込んだServerでこのインターフェースを満たす。
package apihandler

import (
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/anime"
	"github.com/annict/annict/go/internal/apihandler/episode"
	"github.com/annict/annict/go/internal/apihandler/legacy_id"
	"github.com/annict/annict/go/internal/apihandler/openapi_description"
	"github.com/annict/annict/go/internal/apihandler/work"
)

// リソースごとのHandlerはどれも `Handler` という名前のため、そのまま埋め込むとフィールド名が重複する。
// 別名を付けて埋め込み、別名をフィールド名にする
type (
	animeHandler              = anime.Handler
	episodeHandler            = episode.Handler
	legacyIDHandler           = legacy_id.Handler
	openAPIDescriptionHandler = openapi_description.Handler
	workHandler               = work.Handler
)

// Serverはリソースごとのハンドラーを埋め込み、`apigen.StrictServerInterface` を満たす
type Server struct {
	*animeHandler
	*episodeHandler
	*legacyIDHandler
	*openAPIDescriptionHandler
	*workHandler
}

var _ apigen.StrictServerInterface = (*Server)(nil)

// NewServerは新しいServerを作成する
func NewServer(
	anime *anime.Handler,
	episode *episode.Handler,
	legacyID *legacy_id.Handler,
	openAPIDescription *openapi_description.Handler,
	work *work.Handler,
) *Server {
	return &Server{
		animeHandler:              anime,
		episodeHandler:            episode,
		legacyIDHandler:           legacyID,
		openAPIDescriptionHandler: openAPIDescription,
		workHandler:               work,
	}
}
