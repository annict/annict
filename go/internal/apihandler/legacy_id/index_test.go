package legacy_id_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/legacy_id"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

func TestHandler_ListLegacyIds(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	classificationRepo := repository.NewAnimeClassificationRepository(queries)
	h := legacy_id.NewHandler(usecase.NewGetAPILegacyIDsUsecase(
		validator.NewAPILegacyIDIndexValidator(),
		animeRepo,
		classificationRepo,
		repository.NewAnimeRedirectRepository(queries),
		repository.NewAnimeLegacyIDRepository(queries),
	))

	work, err := animeRepo.Create(ctx, repository.CreateAnimeParams{Title: sql.NullString{String: "作品", Valid: true}})
	if err != nil {
		t.Fatalf("作品の作成に失敗しました: %v", err)
	}
	if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{AnimeID: work.ID, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作品の分類の作成に失敗しました: %v", err)
	}
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("旧作品").Build()
	if err := repository.NewWorkRepository(queries).UpdateAnimeID(ctx, workID, work.ID); err != nil {
		t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
	}

	// サブテストは親のトランザクションを共有するため、並列にしない

	t.Run("旧IDごとのanime_idを返し、変換できない旧IDはnullにする", func(t *testing.T) {
		workIDs := []int64{int64(workID), 0}
		episodeIDs := []int64{0}
		res, err := h.ListLegacyIds(ctx, apigen.ListLegacyIdsRequestObject{Params: apigen.ListLegacyIdsParams{WorkIds: &workIDs, EpisodeIds: &episodeIDs}})
		if err != nil {
			t.Fatalf("ListLegacyIds()のエラー = %v", err)
		}
		body := res.(apigen.ListLegacyIds200JSONResponse)
		if len(body.WorkIds) != 2 || body.WorkIds[0].LegacyWorkId != int64(workID) || body.WorkIds[0].AnimeId == nil || *body.WorkIds[0].AnimeId != int64(work.ID) {
			t.Errorf("作品の結果 = %+v、期待値 = 旧作品 %d がanime %d に変換される", body.WorkIds, workID, work.ID)
		}
		if len(body.WorkIds) == 2 && body.WorkIds[1].AnimeId != nil {
			t.Errorf("存在しない旧作品の結果 = %+v、期待値 = anime_idがnil", body.WorkIds[1])
		}
		if len(body.EpisodeIds) != 1 || body.EpisodeIds[0].LegacyEpisodeId != 0 || body.EpisodeIds[0].AnimeId != nil {
			t.Errorf("エピソードの結果 = %+v、期待値 = anime_idがnilの1件", body.EpisodeIds)
		}
	})

	t.Run("UseCaseのエラーはそのまま返し、共通のヘルパーに変換を任せる", func(t *testing.T) {
		_, err := h.ListLegacyIds(ctx, apigen.ListLegacyIdsRequestObject{})
		if model.AsValidationError(err) == nil {
			t.Errorf("エラー = %v、期待値 = ValidationError", err)
		}
	})
}
