package work_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/work"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func TestHandler_ListWorks(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	classificationRepo := repository.NewAnimeClassificationRepository(queries)
	h := work.NewHandler(usecase.NewGetAPIWorksUsecase(animeRepo, classificationRepo, repository.NewAnimeLegacyIDRepository(queries), usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}))

	var ids []int64
	for _, title := range []string{"1つめ", "2つめ"} {
		a, err := animeRepo.Create(ctx, repository.CreateAnimeParams{Title: sql.NullString{String: title, Valid: true}})
		if err != nil {
			t.Fatalf("animeの作成に失敗しました: %v", err)
		}
		if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{AnimeID: a.ID, Kind: model.AnimeClassificationKindWork}); err != nil {
			t.Fatalf("分類の作成に失敗しました: %v", err)
		}
		ids = append(ids, int64(a.ID))
	}

	listWorks := func(t *testing.T, params apigen.ListWorksParams) apigen.ListWorks200JSONResponse {
		t.Helper()
		params.Ids = &ids
		res, err := h.ListWorks(ctx, apigen.ListWorksRequestObject{Params: params})
		if err != nil {
			t.Fatalf("ListWorks()のエラー = %v", err)
		}
		jsonRes, ok := res.(apigen.ListWorks200JSONResponse)
		if !ok {
			t.Fatalf("レスポンスの型 = %T、期待値 = ListWorks200JSONResponse", res)
		}
		return jsonRes
	}

	// サブテストは親のトランザクションを共有するため、並列にしない

	t.Run("next_cursorで次のページを取得できる", func(t *testing.T) {
		limit := int32(1)
		first := listWorks(t, apigen.ListWorksParams{Limit: &limit})
		if len(first.Body.Items) != 1 || first.Body.NextCursor == nil {
			t.Fatalf("1ページめ = %d件、next_cursor = %v、期待値 = 1件と続きのカーソル", len(first.Body.Items), first.Body.NextCursor)
		}
		second := listWorks(t, apigen.ListWorksParams{Limit: &limit, Cursor: first.Body.NextCursor})
		if len(second.Body.Items) != 1 || second.Body.NextCursor != nil {
			t.Fatalf("2ページめ = %d件、next_cursor = %v、期待値 = 1件で続き無し", len(second.Body.Items), second.Body.NextCursor)
		}
		w, err := second.Body.Items[0].AsWorkAnime()
		if err != nil || w.Id != ids[1] {
			t.Errorf("2ページめの作品 = %+v、期待値 = %d", w, ids[1])
		}
	})

	t.Run("空の一覧はitemsを空の配列にし、next_cursorをnullにする", func(t *testing.T) {
		title := "一致しないタイトル"
		res := listWorks(t, apigen.ListWorksParams{Title: &title})
		if res.Body.Items == nil || len(res.Body.Items) != 0 || res.Body.NextCursor != nil {
			t.Errorf("本文 = %+v、期待値 = 空の配列と続き無し", res.Body)
		}
	})

	t.Run("読めないカーソルはcursorのパラメーターの誤り", func(t *testing.T) {
		cursor := "invalid"
		_, err := h.ListWorks(ctx, apigen.ListWorksRequestObject{Params: apigen.ListWorksParams{Cursor: &cursor}})
		var pe *apierror.ParameterError
		if !errors.As(err, &pe) || pe.Parameter != "cursor" {
			t.Errorf("エラー = %v、期待値 = cursorのParameterError", err)
		}
	})
}
