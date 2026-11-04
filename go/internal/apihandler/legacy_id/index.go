package legacy_id

import (
	"context"

	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// ListLegacyIdsは旧IDごとの `anime_id` を返す (GET /api/v2/legacy_ids)
func (h *Handler) ListLegacyIds(ctx context.Context, request apigen.ListLegacyIdsRequestObject) (apigen.ListLegacyIdsResponseObject, error) {
	var input usecase.GetAPILegacyIDsInput
	if request.Params.WorkIds != nil {
		for _, id := range *request.Params.WorkIds {
			input.WorkIDs = append(input.WorkIDs, model.WorkID(id))
		}
	}
	if request.Params.EpisodeIds != nil {
		for _, id := range *request.Params.EpisodeIds {
			input.EpisodeIDs = append(input.EpisodeIDs, model.EpisodeID(id))
		}
	}

	output, err := h.getAPILegacyIDsUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	res := apigen.ListLegacyIds200JSONResponse{
		WorkIds:    make([]apigen.LegacyWorkId, len(output.Works)),
		EpisodeIds: make([]apigen.LegacyEpisodeId, len(output.Episodes)),
	}
	for i, w := range output.Works {
		res.WorkIds[i] = apigen.LegacyWorkId{LegacyWorkId: int64(w.WorkID), AnimeId: toAnimeID(w.AnimeID)}
	}
	for i, e := range output.Episodes {
		res.EpisodeIds[i] = apigen.LegacyEpisodeId{LegacyEpisodeId: int64(e.EpisodeID), AnimeId: toAnimeID(e.AnimeID)}
	}
	return res, nil
}

func toAnimeID(id *model.AnimeID) *int64 {
	if id == nil {
		return nil
	}
	v := int64(*id)
	return &v
}
