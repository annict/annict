package usecase

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetDBWorkRevisionsUsecase_Execute(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	revisionRepo := repository.NewAnimeRevisionRepository(queries)
	uc := NewGetDBWorkRevisionsUsecase(repository.NewWorkRepository(queries), revisionRepo, repository.NewNumberFormatRepository(queries))
	ctx := context.Background()

	// 非公開の作品の履歴も見せる。
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴ユースケース作品").WithUnpublishedAt(time.Now()).Build()
	animeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	for range 3 {
		editID := testutil.NewDBEditBuilder(t, tx).Build()
		if err := revisionRepo.Create(ctx, editID, animeID, model.AnimeSnapshot{Status: model.AnimeStatusArchived, Kind: model.AnimeClassificationKindWork}); err != nil {
			t.Fatalf("版の作成に失敗: %v", err)
		}
	}

	output, err := uc.Execute(ctx, GetDBWorkRevisionsInput{WorkID: workID, Page: 1, PerPage: 2})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Work.ID != workID || output.TotalCount != 3 || len(output.Revisions) != 2 {
		t.Errorf("出力 = 作品%d・総数%d・%d件、期待値 = 作品%d・総数3・2件", output.Work.ID, output.TotalCount, len(output.Revisions), workID)
	}
}

func TestGetDBWorkRevisionsUsecase_Execute_Unmapped(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBWorkRevisionsUsecase(repository.NewWorkRepository(queries), repository.NewAnimeRevisionRepository(queries), repository.NewNumberFormatRepository(queries))
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("未マッピングの作品").Build()

	output, err := uc.Execute(context.Background(), GetDBWorkRevisionsInput{WorkID: workID, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.TotalCount != 0 || len(output.Revisions) != 0 {
		t.Errorf("出力 = 総数%d・%d件、空の履歴を期待", output.TotalCount, len(output.Revisions))
	}
}

func TestGetDBWorkRevisionsUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBWorkRevisionsUsecase(repository.NewWorkRepository(queries), repository.NewAnimeRevisionRepository(queries), repository.NewNumberFormatRepository(queries))

	for _, tt := range []struct {
		name   string
		workID model.WorkID
	}{
		{name: "存在しない作品", workID: model.WorkID(1 << 62)},
		{name: "削除した作品", workID: testutil.NewWorkBuilder(t, tx).WithDeletedAt(time.Now()).Build()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), GetDBWorkRevisionsInput{WorkID: tt.workID, Page: 1, PerPage: 10})
			if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("エラー = %v、AppErrCodeResourceNotFoundを期待", err)
			}
		})
	}
}

// 版の「話数フォーマット」を名前で示すため、話数フォーマットの一覧も返すことを確かめる。
func TestGetDBWorkRevisionsUsecase_Execute_NumberFormats(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBWorkRevisionsUsecase(repository.NewWorkRepository(queries), repository.NewAnimeRevisionRepository(queries), repository.NewNumberFormatRepository(queries))

	workID := testutil.NewWorkBuilder(t, tx).Build()
	animeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	var numberFormatID int64
	if err := tx.QueryRow(`INSERT INTO number_formats (name, sort_number, created_at, updated_at) VALUES ('履歴の話数フォーマット', 1, NOW(), NOW()) RETURNING id`).Scan(&numberFormatID); err != nil {
		t.Fatalf("話数フォーマットの作成に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), GetDBWorkRevisionsInput{WorkID: workID, Page: 1, PerPage: 30})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !slices.ContainsFunc(output.NumberFormats, func(f model.NumberFormat) bool {
		return f.ID == model.NumberFormatID(numberFormatID) && f.Name == "履歴の話数フォーマット"
	}) {
		t.Errorf("話数フォーマットの一覧 = %+v、作成した話数フォーマットを含むことを期待", output.NumberFormats)
	}
}
