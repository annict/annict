package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newDBEditsSnapshotは、ほかのテストがコミットする操作で一覧と件数の前提が変わらない読み取りを用意する。
// 固定したスナップショットの中でも、このテスト自身が作った操作と版は見える。
func newDBEditsSnapshot(t *testing.T) (*GetDBEditsUsecase, *sql.Tx) {
	t.Helper()
	tx, err := testutil.GetTestDB().BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		t.Fatalf("テスト用トランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	queries := query.New(tx)
	return NewGetDBEditsUsecase(
		repository.NewDBEditRepository(queries),
		repository.NewAnimeRevisionRepository(queries),
		repository.NewPersonRevisionRepository(queries),
		repository.NewNumberFormatRepository(queries),
		repository.NewPrefectureRepository(queries),
	), tx
}

// 見える版と隠す版が1つの操作に混在しても、閲覧者に見せる版とそのリンク先だけを返す。
// エピソード自身と現在の親作品の削除も、操作の選択と版の読み戻しの両方に反映する。
func TestGetDBEditsUsecase_Execute_VisibleRevisions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uc, tx := newDBEditsSnapshot(t)
	queries := query.New(tx)
	animeRepo := repository.NewAnimeRevisionRepository(queries)
	personRepo := repository.NewPersonRevisionRepository(queries)
	baseline := make(map[bool]int64)
	for _, admin := range []bool{false, true} {
		count, err := repository.NewDBEditRepository(queries).Count(ctx, admin)
		if err != nil {
			t.Fatalf("準備前の件数の取得に失敗: %v", err)
		}
		baseline[admin] = count
	}

	newWork := func(deleted bool) (model.WorkID, model.AnimeID) {
		t.Helper()
		builder := testutil.NewWorkBuilder(t, tx)
		if deleted {
			builder.WithDeletedAt(time.Now())
		}
		workID := builder.Build()
		animeID := testutil.CreateTestAnime(t, tx)
		if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
			t.Fatalf("作品の写像の設定に失敗: %v", err)
		}
		return workID, animeID
	}
	newEpisode := func(workID model.WorkID, deleted bool) (model.EpisodeID, model.AnimeID) {
		t.Helper()
		builder := testutil.NewEpisodeBuilder(t, tx, workID)
		if deleted {
			builder.WithDeletedAt(time.Now())
		}
		episodeID := builder.Build()
		animeID := testutil.CreateTestAnime(t, tx)
		if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(episodeID)); err != nil {
			t.Fatalf("エピソードの写像の設定に失敗: %v", err)
		}
		return episodeID, animeID
	}
	workID, workAnimeID := newWork(false)
	deletedWorkID, deletedWorkAnimeID := newWork(true)
	episodeID, episodeAnimeID := newEpisode(workID, false)
	_, deletedEpisodeAnimeID := newEpisode(workID, true)
	_, deletedParentEpisodeAnimeID := newEpisode(deletedWorkID, false)

	mixedEdit := testutil.NewDBEditBuilder(t, tx).Build()
	animeSnapshots := []struct {
		id     model.AnimeID
		parent *model.AnimeID
	}{
		{workAnimeID, nil},
		{episodeAnimeID, &workAnimeID},
		{deletedWorkAnimeID, nil},
		{deletedEpisodeAnimeID, &workAnimeID},
		{deletedParentEpisodeAnimeID, &deletedWorkAnimeID},
	}
	recordAnime := func(editID model.DBEditID, id model.AnimeID, parent *model.AnimeID) {
		t.Helper()
		kind := model.AnimeClassificationKindWork
		if parent != nil {
			kind = model.AnimeClassificationKindEpisode
		}
		if err := animeRepo.Create(ctx, editID, id, model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: kind, ParentAnimeID: parent}); err != nil {
			t.Fatalf("作品・エピソードの版の作成に失敗: %v", err)
		}
	}
	for _, a := range animeSnapshots {
		recordAnime(mixedEdit, a.id, a.parent)
	}
	people := []struct {
		name              string
		current, snapshot model.PersonStatus
		id                model.PersonID
	}{
		{name: "見える人物", current: model.PersonStatusPublished, snapshot: model.PersonStatusPublished},
		{name: "掲載停止中に記録した版", current: model.PersonStatusPublished, snapshot: model.PersonStatusDelisted},
		{name: "現在掲載停止中の人物", current: model.PersonStatusDelisted, snapshot: model.PersonStatusPublished},
		{name: "削除した人物", current: model.PersonStatusDeleted, snapshot: model.PersonStatusPublished},
	}
	for i := range people {
		p := &people[i]
		p.id = testutil.NewPersonBuilder(t, tx).WithStatus(p.current).Build()
		if err := personRepo.Create(ctx, mixedEdit, p.id, model.PersonSnapshot{Name: p.name, Status: p.snapshot}); err != nil {
			t.Fatalf("人物の版の作成に失敗: %v", err)
		}
	}
	// より新しい操作でも、削除したリソースの版だけなら最初のページの枠を使わない。
	hiddenEdit := testutil.NewDBEditBuilder(t, tx).Build()
	for _, a := range animeSnapshots[2:] {
		recordAnime(hiddenEdit, a.id, a.parent)
	}

	for _, tt := range []struct {
		name      string
		admin     bool
		user      *model.User
		personIDs []model.PersonID
	}{
		{"管理者以外", false, nil, []model.PersonID{people[0].id}},
		{"管理者", true, &model.User{ID: 1, Role: model.RoleAdmin}, []model.PersonID{people[0].id, people[1].id, people[2].id}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, GetDBEditsInput{User: tt.user, Page: 1, PerPage: 1})
			if err != nil {
				t.Fatalf("履歴の取得に失敗: %v", err)
			}
			if len(output.Edits) != 1 || output.Edits[0].ID != mixedEdit {
				t.Fatalf("最初のページ = %+v、期待値 = 混在する操作%dの1件", output.Edits, mixedEdit)
			}
			if want := baseline[tt.admin] + 1; output.TotalCount != want {
				t.Errorf("総数 = %d、期待値 = %d (見える版を持つ操作1件だけ増える)", output.TotalCount, want)
			}
			edit := output.Edits[0]
			animeIDs := make(map[model.AnimeID]bool)
			for _, r := range edit.AnimeRevisions {
				animeIDs[r.AnimeID] = true
				if r.DBEdit != edit {
					t.Error("作品・エピソードの版が元の操作に結び付いていない")
				}
				if r.AnimeID == workAnimeID && (r.Work == nil || r.Work.ID != workID) {
					t.Errorf("作品のリンク先 = %+v、期待値 = 作品%d", r.Work, workID)
				}
				if r.AnimeID == episodeAnimeID && (r.Episode == nil || r.Episode.ID != episodeID) {
					t.Errorf("エピソードのリンク先 = %+v、期待値 = エピソード%d", r.Episode, episodeID)
				}
				// エピソードの版には、名前に添える今の親作品を付ける。
				if r.AnimeID == episodeAnimeID && (r.Work == nil || r.Work.ID != workID || r.Work.Title != "テストアニメ") {
					t.Errorf("エピソードの親作品 = %+v、期待値 = 作品%d「テストアニメ」", r.Work, workID)
				}
			}
			if len(edit.AnimeRevisions) != 2 || !animeIDs[workAnimeID] || !animeIDs[episodeAnimeID] {
				t.Errorf("作品・エピソードの版 = %v、期待値 = anime%d・%dの2件", animeIDs, workAnimeID, episodeAnimeID)
			}
			personIDs := make(map[model.PersonID]bool)
			for _, r := range edit.PersonRevisions {
				personIDs[r.PersonID] = true
				if r.DBEdit != edit {
					t.Error("人物の版が元の操作に結び付いていない")
				}
			}
			if len(edit.PersonRevisions) != len(tt.personIDs) {
				t.Errorf("人物の版の数 = %d、期待値 = %d", len(edit.PersonRevisions), len(tt.personIDs))
			}
			for _, id := range tt.personIDs {
				if !personIDs[id] {
					t.Errorf("人物%dの版が無い", id)
				}
			}
		})
	}
}

// 隠す操作がページの枠や総数を消費せず、別ページの版も差分の比較に使うことを確かめる。
func TestGetDBEditsUsecase_Execute_Pagination(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uc, tx := newDBEditsSnapshot(t)
	queries := query.New(tx)
	editRepo := repository.NewDBEditRepository(queries)
	personRepo := repository.NewPersonRevisionRepository(queries)
	baseline := make(map[bool]int64)
	for _, admin := range []bool{false, true} {
		count, err := editRepo.Count(ctx, admin)
		if err != nil {
			t.Fatalf("準備前の件数の取得に失敗: %v", err)
		}
		baseline[admin] = count
	}
	personID := testutil.NewPersonBuilder(t, tx).Build()
	record := func(action model.DBEditAction, status model.PersonStatus, name string) model.DBEditID {
		t.Helper()
		id := testutil.NewDBEditBuilder(t, tx).WithAction(action).Build()
		if err := personRepo.Create(ctx, id, personID, model.PersonSnapshot{Name: name, Status: status}); err != nil {
			t.Fatalf("人物の版の作成に失敗: %v", err)
		}
		return id
	}
	first := record(model.DBEditActionCreate, model.PersonStatusPublished, "掲載停止前")
	delist := record(model.DBEditActionDelist, model.PersonStatusDelisted, "掲載停止前")
	whileDelisted := record(model.DBEditActionUpdate, model.PersonStatusDelisted, "掲載停止中の編集")
	undelist := record(model.DBEditActionUndelist, model.PersonStatusPublished, "掲載停止中の編集")
	latest := record(model.DBEditActionUpdate, model.PersonStatusPublished, "掲載再開後の編集")
	// 版の無い操作と、削除した人物の版しか持たない操作は、管理者にも数えない。
	deletedPersonID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()
	if err := personRepo.Create(ctx, testutil.NewDBEditBuilder(t, tx).Build(), deletedPersonID, model.PersonSnapshot{Name: "削除した人物", Status: model.PersonStatusPublished}); err != nil {
		t.Fatalf("削除した人物の版の作成に失敗: %v", err)
	}
	testutil.NewDBEditBuilder(t, tx).Build()

	for _, tt := range []struct {
		name  string
		admin bool
		user  *model.User
		ids   []model.DBEditID
	}{
		{"管理者以外", false, nil, []model.DBEditID{latest, first}},
		{"管理者", true, &model.User{ID: 1, Role: model.RoleAdmin}, []model.DBEditID{latest, undelist, whileDelisted, delist, first}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			page := int32(1)
			for _, wantID := range tt.ids {
				output, err := uc.Execute(ctx, GetDBEditsInput{User: tt.user, Page: page, PerPage: 1})
				if err != nil {
					t.Fatalf("%dページ目の取得に失敗: %v", page, err)
				}
				if want := baseline[tt.admin] + int64(len(tt.ids)); output.TotalCount != want {
					t.Errorf("総数 = %d、期待値 = %d", output.TotalCount, want)
				}
				if len(output.Edits) != 1 || output.Edits[0].ID != wantID {
					t.Fatalf("%dページ目 = %+v、期待値 = 操作%dの1件", page, output.Edits, wantID)
				}
				if len(output.Edits[0].PersonRevisions) != 1 {
					t.Fatalf("人物の版の数 = %d、期待値 = 1", len(output.Edits[0].PersonRevisions))
				}
				if page == 1 {
					r := output.Edits[0].PersonRevisions[0]
					if r.Previous == nil || r.Previous.Name != "掲載停止中の編集" || r.Previous.Status != model.PersonStatusPublished {
						t.Fatalf("直前の版 = %+v、期待値 = 別ページの掲載再開の版", r.Previous)
					}
					diff := model.DiffPersonSnapshots(r.Previous, r.Snapshot)
					if len(diff.Fields) != 1 || diff.Fields[0].Field != model.RevisionField("name") {
						t.Errorf("最新の版の差分 = %+v、期待値 = 名前だけ", diff)
					}
				}
				page++
			}
		})
	}
}
