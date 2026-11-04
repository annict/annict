package work

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

// workCursorは作品の一覧のカーソルの中身。
// 別の並び順で作ったカーソルを受け付けないよう、並び順も入れる
type workCursor struct {
	Order   string `json:"o"`
	AnimeID *int64 `json:"i"`
	// SeasonKeyは季節の順のときの代表の季節のキー。代表の季節の無い作品ではnull
	SeasonKey *int32 `json:"k"`
}

// ListWorksは作品の一覧を返す (GET /api/v2/anime/works)
func (h *Handler) ListWorks(ctx context.Context, request apigen.ListWorksRequestObject) (apigen.ListWorksResponseObject, error) {
	params := request.Params

	order := usecase.APIWorkOrderID
	if params.Sort != nil {
		order = usecase.APIWorkOrder(*params.Sort)
	}
	after, err := decodeWorkCursor(ctx, params.Cursor, order)
	if err != nil {
		return nil, err
	}

	input := usecase.GetAPIWorksInput{
		Year:  params.Year,
		Title: params.Title,
		Order: order,
		After: after,
		Limit: apipagination.Limit(params.Limit),
	}
	if params.Ids != nil {
		input.IDs = make([]model.AnimeID, len(*params.Ids))
		for i, id := range *params.Ids {
			input.IDs[i] = model.AnimeID(id)
		}
	}
	if params.Season != nil {
		input.Seasons = make([]model.SeasonBucket, len(*params.Season))
		for i, s := range *params.Season {
			bucket, err := model.ParseSeasonBucket(s)
			if err != nil {
				return nil, &apierror.ParameterError{Parameter: "season", Detail: err.Error()}
			}
			input.Seasons[i] = bucket
		}
	}

	output, err := h.getAPIWorksUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	items := make([]apigen.Anime, len(output.Works))
	for i, w := range output.Works {
		items[i], err = anime.ToAPIAnime(w)
		if err != nil {
			return nil, err
		}
	}

	res := apigen.ListWorks200JSONResponse{Body: apigen.AnimeList{Items: items}}
	if output.Next != nil {
		id := int64(output.Next.AnimeID)
		nextCursor, err := apipagination.EncodeCursor(workCursor{Order: string(order), AnimeID: &id, SeasonKey: output.Next.SeasonKey})
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

// decodeWorkCursorはリクエストのカーソルから、前のページの最後の作品を取り出す。
// カーソルが無ければnilを返す。別の並び順で作ったカーソルは受け付けない
func decodeWorkCursor(ctx context.Context, cursor *string, order usecase.APIWorkOrder) (*usecase.APIWorkCursor, error) {
	if cursor == nil {
		return nil, nil
	}

	var c workCursor
	if err := apipagination.DecodeCursor(*cursor, &c); err != nil || c.AnimeID == nil || c.Order != string(order) {
		return nil, &apierror.ParameterError{Parameter: "cursor", Detail: i18n.T(ctx, "api_error_invalid_cursor")}
	}
	return &usecase.APIWorkCursor{AnimeID: model.AnimeID(*c.AnimeID), SeasonKey: c.SeasonKey}, nil
}
