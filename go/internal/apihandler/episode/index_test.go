package episode_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/episode"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func TestHandler_ListEpisodes(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	classificationRepo := repository.NewAnimeClassificationRepository(queries)
	h := episode.NewHandler(usecase.NewGetAPIEpisodesUsecase(animeRepo, classificationRepo, repository.NewAnimeLegacyIDRepository(queries), usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}))

	work, err := animeRepo.Create(ctx, repository.CreateAnimeParams{Title: sql.NullString{String: "作品", Valid: true}})
	if err != nil {
		t.Fatalf("作品の作成に失敗しました: %v", err)
	}
	if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{AnimeID: work.ID, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作品の分類の作成に失敗しました: %v", err)
	}
	var episodeIDs []int64
	for i := range 2 {
		e, err := animeRepo.Create(ctx, repository.CreateAnimeParams{})
		if err != nil {
			t.Fatalf("エピソードの作成に失敗しました: %v", err)
		}
		if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{
			AnimeID: e.ID, Kind: model.AnimeClassificationKindEpisode, ParentAnimeID: &work.ID,
			SortNumber: sql.NullInt32{Int32: int32(i + 1), Valid: true},
		}); err != nil {
			t.Fatalf("エピソードの分類の作成に失敗しました: %v", err)
		}
		episodeIDs = append(episodeIDs, int64(e.ID))
	}

	// サブテストは親のトランザクションを共有するため、並列にしない

	t.Run("next_cursorで次のページを取得できる", func(t *testing.T) {
		limit := int32(1)
		res, err := h.ListEpisodes(ctx, apigen.ListEpisodesRequestObject{AnimeId: int64(work.ID), Params: apigen.ListEpisodesParams{Limit: &limit}})
		if err != nil {
			t.Fatalf("ListEpisodes()のエラー = %v", err)
		}
		first := res.(apigen.ListEpisodes200JSONResponse)
		if len(first.Body.Items) != 1 || first.Body.NextCursor == nil {
			t.Fatalf("1ページめ = %d件、next_cursor = %v、期待値 = 1件と続きのカーソル", len(first.Body.Items), first.Body.NextCursor)
		}

		res, err = h.ListEpisodes(ctx, apigen.ListEpisodesRequestObject{AnimeId: int64(work.ID), Params: apigen.ListEpisodesParams{Limit: &limit, Cursor: first.Body.NextCursor}})
		if err != nil {
			t.Fatalf("2ページめのListEpisodes()のエラー = %v", err)
		}
		second := res.(apigen.ListEpisodes200JSONResponse)
		e, err := second.Body.Items[0].AsEpisodeAnime()
		if err != nil || e.Id != episodeIDs[1] || second.Body.NextCursor != nil {
			t.Errorf("2ページめ = %+v、next_cursor = %v、期待値 = %d で続き無し", e, second.Body.NextCursor, episodeIDs[1])
		}
	})

	t.Run("読めないカーソルはcursorのパラメーターの誤り", func(t *testing.T) {
		cursor := "invalid"
		_, err := h.ListEpisodes(ctx, apigen.ListEpisodesRequestObject{AnimeId: int64(work.ID), Params: apigen.ListEpisodesParams{Cursor: &cursor}})
		var pe *apierror.ParameterError
		if !errors.As(err, &pe) || pe.Parameter != "cursor" {
			t.Errorf("エラー = %v、期待値 = cursorのParameterError", err)
		}
	})

	t.Run("UseCaseのエラーはそのまま返し、共通のヘルパーに変換を任せる", func(t *testing.T) {
		_, err := h.ListEpisodes(ctx, apigen.ListEpisodesRequestObject{AnimeId: episodeIDs[0]})
		if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Errorf("エラー = %v、期待値 = 未存在のAppError", err)
		}
	})
}
