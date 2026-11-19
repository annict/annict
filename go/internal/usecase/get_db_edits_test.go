package usecase

import (
	"context"
	"slices"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBEditsUsecase_Execute_HidesDelistingは、掲載を再開した人物の掲載停止・掲載再開の操作が、管理者以外には
// 並ばず、管理者には並ぶことを検証する。作品の版を持つ操作は誰にでも並び、各操作は自分の版を持つ。
func TestGetDBEditsUsecase_Execute_HidesDelisting(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	personRevisionRepo := repository.NewPersonRevisionRepository(queries)
	animeRevisionRepo := repository.NewAnimeRevisionRepository(queries)
	uc := NewGetDBEditsUsecase(repository.NewDBEditRepository(queries), animeRevisionRepo, personRevisionRepo, repository.NewNumberFormatRepository(queries), repository.NewPrefectureRepository(queries))

	// 作品の作成と、人物の掲載停止・掲載再開。人物は掲載を再開して公開中に戻っている。
	animeID := testutil.CreateTestAnime(t, tx)
	workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).Build()
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	title := "全体の履歴の作品"
	workEdit := testutil.NewDBEditBuilder(t, tx).WithAction(model.DBEditActionCreate).Build()
	if err := animeRevisionRepo.Create(ctx, workEdit, animeID, model.AnimeSnapshot{Title: &title, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作品の版の作成に失敗: %v", err)
	}
	personID := testutil.NewPersonBuilder(t, tx).Build()
	delist := testutil.NewDBEditBuilder(t, tx).WithAction(model.DBEditActionDelist).Build()
	undelist := testutil.NewDBEditBuilder(t, tx).WithAction(model.DBEditActionUndelist).Build()
	// 直前の版は版のID順で決まるため、掲載停止・掲載再開の順に挿入する。
	for _, v := range []struct {
		editID model.DBEditID
		status model.PersonStatus
	}{
		{delist, model.PersonStatusDelisted},
		{undelist, model.PersonStatusPublished},
	} {
		if err := personRevisionRepo.Create(ctx, v.editID, personID, model.PersonSnapshot{Name: "掲載を再開した人物", Status: v.status}); err != nil {
			t.Fatalf("人物の版の作成に失敗: %v", err)
		}
	}

	for _, tt := range []struct {
		name       string
		user       *model.User
		wantListed map[model.DBEditID]bool
	}{
		{name: "ログインしていない閲覧者", wantListed: map[model.DBEditID]bool{workEdit: true, delist: false, undelist: false}},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}, wantListed: map[model.DBEditID]bool{workEdit: true, delist: false, undelist: false}},
		{name: "管理者", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantListed: map[model.DBEditID]bool{workEdit: true, delist: true, undelist: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, GetDBEditsInput{User: tt.user, Page: 1, PerPage: 1000})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			listed := map[model.DBEditID]*model.DBEdit{}
			for _, e := range output.Edits {
				listed[e.ID] = e
			}
			for id, want := range tt.wantListed {
				if (listed[id] != nil) != want {
					t.Errorf("操作%dが並ぶか = %v、期待値 = %v", id, listed[id] != nil, want)
				}
			}
			if e := listed[workEdit]; e == nil || len(e.AnimeRevisions) != 1 || len(e.PersonRevisions) != 0 || e.AnimeRevisions[0].Work == nil || e.AnimeRevisions[0].Work.ID != workID {
				t.Errorf("作品の操作の版 = %+v、期待値 = 作品%dの版1つ", e, workID)
			}
			if e := listed[undelist]; e != nil {
				// 掲載再開の版の1つ前の版は、掲載停止の版。
				if len(e.PersonRevisions) != 1 || e.PersonRevisions[0].Previous == nil || e.PersonRevisions[0].Previous.Status != model.PersonStatusDelisted {
					t.Errorf("掲載再開の操作の版 = %+v、期待値 = 1つ前が掲載停止の版", e.PersonRevisions)
				}
			}
		})
	}
}

// 版の「話数フォーマット」と「出身地」を名前で示すため、話数フォーマットと都道府県の一覧も返すことを確かめる。
func TestGetDBEditsUsecase_Execute_IDNames(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBEditsUsecase(repository.NewDBEditRepository(queries), repository.NewAnimeRevisionRepository(queries), repository.NewPersonRevisionRepository(queries), repository.NewNumberFormatRepository(queries), repository.NewPrefectureRepository(queries))
	var numberFormatID int64
	if err := tx.QueryRow(`INSERT INTO number_formats (name, sort_number, created_at, updated_at) VALUES ('全体の履歴の話数フォーマット', 1, NOW(), NOW()) RETURNING id`).Scan(&numberFormatID); err != nil {
		t.Fatalf("話数フォーマットの作成に失敗: %v", err)
	}
	prefectureID := testutil.CreateTestPrefecture(t, tx, "全体の履歴の出身地")

	output, err := uc.Execute(context.Background(), GetDBEditsInput{Page: 1, PerPage: 30})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !slices.ContainsFunc(output.NumberFormats, func(f model.NumberFormat) bool {
		return f.ID == model.NumberFormatID(numberFormatID) && f.Name == "全体の履歴の話数フォーマット"
	}) {
		t.Errorf("話数フォーマットの一覧 = %+v、作成した話数フォーマットを含むことを期待", output.NumberFormats)
	}
	if !slices.ContainsFunc(output.Prefectures, func(p model.Prefecture) bool {
		return p.ID == prefectureID && p.Name == "全体の履歴の出身地"
	}) {
		t.Errorf("都道府県の一覧 = %+v、作成した出身地を含むことを期待", output.Prefectures)
	}
}
