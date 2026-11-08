package anime

import (
	"context"

	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// GetAnimeは作品かエピソードを返す (GET /api/v2/anime/{anime_id})
func (h *Handler) GetAnime(ctx context.Context, request apigen.GetAnimeRequestObject) (apigen.GetAnimeResponseObject, error) {
	output, err := h.getAPIAnimeUC.Execute(ctx, usecase.GetAPIAnimeInput{AnimeID: model.AnimeID(request.AnimeId)})
	if err != nil {
		return nil, err
	}

	anime, err := ToAPIAnime(output.Anime)
	if err != nil {
		return nil, err
	}
	return apigen.GetAnime200JSONResponse(anime), nil
}
