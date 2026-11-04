package episode

import (
	"context"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/anime"
	"github.com/annict/annict/go/internal/apipagination"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// episodeCursorはエピソードの一覧のカーソルの中身。並び順のキー (sort_number, anime_id) を持つ
type episodeCursor struct {
	SortNumber *int32 `json:"s"`
	AnimeID    *int64 `json:"i"`
}

// ListEpisodesは作品のエピソードの一覧を返す (GET /api/v2/anime/{anime_id}/episodes)
func (h *Handler) ListEpisodes(ctx context.Context, request apigen.ListEpisodesRequestObject) (apigen.ListEpisodesResponseObject, error) {
	after, err := decodeEpisodeCursor(ctx, request.Params.Cursor)
	if err != nil {
		return nil, err
	}

	output, err := h.getAPIEpisodesUC.Execute(ctx, usecase.GetAPIEpisodesInput{
		WorkID: model.AnimeID(request.AnimeId),
		After:  after,
		Limit:  apipagination.Limit(request.Params.Limit),
	})
	if err != nil {
		return nil, err
	}

	items := make([]apigen.Anime, len(output.Episodes))
	for i, e := range output.Episodes {
		items[i], err = anime.ToAPIAnime(e)
		if err != nil {
			return nil, err
		}
	}

	res := apigen.ListEpisodes200JSONResponse{Body: apigen.AnimeList{Items: items}}
	if output.Next != nil {
		id := int64(output.Next.AnimeID)
		nextCursor, err := apipagination.EncodeCursor(episodeCursor{SortNumber: &output.Next.SortNumber, AnimeID: &id})
		if err != nil {
			return nil, err
		}
		res.Body.NextCursor = &nextCursor
		if link := apipagination.NextLink(ctx, nextCursor); link != "" {
			res.Headers.Link = &link
		}
	}
	return res, nil
}

// decodeEpisodeCursorはリクエストのカーソルから、前のページの最後のエピソードを取り出す。
// カーソルが無ければnilを返す
func decodeEpisodeCursor(ctx context.Context, cursor *string) (*usecase.APIEpisodeCursor, error) {
	if cursor == nil {
		return nil, nil
	}

	var c episodeCursor
	if err := apipagination.DecodeCursor(*cursor, &c); err != nil || c.SortNumber == nil || c.AnimeID == nil {
		return nil, &apierror.ParameterError{Parameter: "cursor", Detail: i18n.T(ctx, "api_error_invalid_cursor")}
	}
	return &usecase.APIEpisodeCursor{SortNumber: *c.SortNumber, AnimeID: model.AnimeID(*c.AnimeID)}, nil
}
