package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// editorImageDataは、履歴の版と操作の一覧の編集者に持たせるプロフィール画像のimage_data。
const editorImageData = `{"original":{"id":"avatar.jpg","storage":"store"}}`

func revisionPtr[T any](v T) *T { return &v }

// fullAnimeSnapshotはすべての項目と別表に値を持つスナップショットを返す。
func fullAnimeSnapshot(parentAnimeID model.AnimeID) model.AnimeSnapshot {
	numberFormatID := model.NumberFormatID(1)
	return model.AnimeSnapshot{
		Title:                 revisionPtr("作品"),
		TitleKana:             revisionPtr("さくひん"),
		TitleRo:               revisionPtr("sakuhin"),
		TitleEn:               revisionPtr("Work"),
		TitleAlter:            revisionPtr("別名"),
		TitleAlterRo:          revisionPtr("betsumei"),
		TitleAlterEn:          revisionPtr("Alias"),
		TitleAlterOther:       revisionPtr("Other"),
		Media:                 revisionPtr(model.AnimeMediaTV),
		ReleaseStatus:         revisionPtr(model.ReleaseStatusReleased),
		Synopsis:              revisionPtr("あらすじ"),
		SynopsisEn:            revisionPtr("Synopsis"),
		SynopsisSource:        revisionPtr("出典"),
		SynopsisSourceEn:      revisionPtr("Source"),
		Status:                model.AnimeStatusArchived,
		ArchiveMessage:        revisionPtr("非公開の案内"),
		Kind:                  model.AnimeClassificationKindEpisode,
		ParentAnimeID:         &parentAnimeID,
		Number:                revisionPtr("3.5"),
		NumberText:            revisionPtr("第3.5話"),
		SortNumber:            revisionPtr(int32(350)),
		Standalone:            true,
		NumberFormatID:        &numberFormatID,
		EpisodeStartNumber:    revisionPtr("1"),
		ExpectedEpisodesCount: revisionPtr(int32(12)),
		ExternalIDs:           []model.AnimeSnapshotExternalID{{Service: model.AnimeExternalServiceSyobocal, ExternalID: "123"}},
		Links:                 []model.AnimeSnapshotLink{{Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageJa, URL: "https://example.com/", Label: revisionPtr("公式"), SortNumber: 1}},
		OfficialAccounts:      []model.AnimeSnapshotOfficialAccount{{Service: model.AnimeAccountServiceX, Account: "example", LabelEn: revisionPtr("Official"), SortNumber: 1}},
		Hashtags:              []model.AnimeSnapshotHashtag{{Hashtag: "example", SortNumber: 1}},
		Seasons:               []model.AnimeSnapshotSeason{{Year: 2026, Name: revisionPtr(model.SeasonNameFall), IsPrimary: true}, {Year: 2027}},
		Events: []model.AnimeSnapshotEvent{{
			Kind:          model.AnimeEventKindBroadcast,
			StartedOn:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			EndedOn:       revisionPtr(time.Date(2026, 12, 24, 0, 0, 0, 0, time.UTC)),
			Title:         revisionPtr("放送"),
			TitleEn:       revisionPtr("Broadcast"),
			Description:   revisionPtr("説明"),
			DescriptionEn: revisionPtr("Description"),
			SortNumber:    1,
		}},
	}
}

func TestAnimeRevisionRepository_CreateAndListByAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	parentAnimeID := testutil.CreateTestAnime(t, tx)
	animeID := testutil.CreateTestAnime(t, tx)
	dbEditID := testutil.NewDBEditBuilder(t, tx).Build()

	initial := model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}
	after := fullAnimeSnapshot(parentAnimeID)
	if err := repo.CreateInitial(ctx, animeID, initial); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}
	if err := repo.Create(ctx, dbEditID, animeID, after); err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	revisions, err := repo.ListByAnimeID(ctx, animeID)
	if err != nil {
		t.Fatalf("ListByAnimeID()のエラー = %v", err)
	}
	if len(revisions) != 2 {
		t.Fatalf("版の件数 = %d、期待値 = 2", len(revisions))
	}

	if revisions[0].DBEditID != nil {
		t.Errorf("起点の版のDBEditID = %d、nilを期待", *revisions[0].DBEditID)
	}
	if revisions[1].DBEditID == nil || *revisions[1].DBEditID != dbEditID {
		t.Errorf("版のDBEditID = %v、期待値 = %d", revisions[1].DBEditID, dbEditID)
	}
	for _, r := range revisions {
		if r.AnimeID != animeID || r.CreatedAt.IsZero() {
			t.Errorf("版%dのAnimeID = %d、CreatedAt = %v、期待値 = %d、非ゼロ値", r.ID, r.AnimeID, r.CreatedAt, animeID)
		}
	}
	// 書き出して読み戻したスナップショットは、別表の行も含めて元と同じ内容になる。
	if diff := model.DiffAnimeSnapshots(&after, revisions[1].Snapshot); !diff.IsEmpty() {
		t.Errorf("読み戻した版の差分 = %+v、差分が無いことを期待", diff)
	}
	if !reflect.DeepEqual(revisions[1].Snapshot, after) {
		t.Errorf("読み戻した版 = %+v、期待値 = %+v", revisions[1].Snapshot, after)
	}
	if diff := model.DiffAnimeSnapshots(&initial, revisions[0].Snapshot); !diff.IsEmpty() {
		t.Errorf("読み戻した起点の版の差分 = %+v、差分が無いことを期待", diff)
	}

	// parent_anime_idの列にはスナップショットの親の作品を書く。
	if got := testutil.CountRows(t, tx, "anime_revisions", "anime_id = $1 AND parent_anime_id = $2 AND snapshot_version = 1", int64(animeID), int64(parentAnimeID)); got != 1 {
		t.Errorf("親の作品を持つ版の件数 = %d、期待値 = 1", got)
	}
}

func TestAnimeRevisionRepository_CreateInitial_Conflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	animeID := testutil.CreateTestAnime(t, tx)
	first := model.AnimeSnapshot{Title: revisionPtr("先"), Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}
	second := model.AnimeSnapshot{Title: revisionPtr("後"), Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}

	for _, s := range []model.AnimeSnapshot{first, second} {
		if err := repo.CreateInitial(ctx, animeID, s); err != nil {
			t.Fatalf("CreateInitial()のエラー = %v", err)
		}
	}

	revisions, err := repo.ListByAnimeID(ctx, animeID)
	if err != nil {
		t.Fatalf("ListByAnimeID()のエラー = %v", err)
	}
	// 起点の版は1つに保ち、先に作った版を残す。
	if len(revisions) != 1 {
		t.Fatalf("版の件数 = %d、期待値 = 1", len(revisions))
	}
	if got := revisions[0].Snapshot.Title; got == nil || *got != "先" {
		t.Errorf("起点の版のTitle = %v、期待値 = 先", got)
	}
}

func TestAnimeRevisionRepository_ListAnimeIDsWithRevisions(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	withRevisionID := testutil.CreateTestAnime(t, tx)
	withoutRevisionID := testutil.CreateTestAnime(t, tx)
	if err := repo.CreateInitial(ctx, withRevisionID, model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}

	got, err := repo.ListAnimeIDsWithRevisions(ctx, []model.AnimeID{withRevisionID, withoutRevisionID})
	if err != nil {
		t.Fatalf("ListAnimeIDsWithRevisions()のエラー = %v", err)
	}
	want := map[model.AnimeID]bool{withRevisionID: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("版を持つ作品・エピソード = %v、期待値 = %v", got, want)
	}

	empty, err := repo.ListAnimeIDsWithRevisions(ctx, nil)
	if err != nil {
		t.Fatalf("空の入力のListAnimeIDsWithRevisions()のエラー = %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("空の入力の結果の件数 = %d、期待値 = 0", len(empty))
	}
}

func TestAnimeRevisionRepository_ListHistoryByAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	animeID := testutil.CreateTestAnime(t, tx)
	userID := testutil.NewUserBuilder(t, tx).WithProfileImageData(editorImageData).Build()
	title := func(v string) model.AnimeSnapshot {
		return model.AnimeSnapshot{Title: &v, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}
	}
	if err := repo.CreateInitial(ctx, animeID, title("起点")); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}
	userEditID := testutil.NewDBEditBuilder(t, tx).WithUser(userID).Build()
	if err := repo.Create(ctx, userEditID, animeID, title("1")); err != nil {
		t.Fatalf("1つ目のCreate()のエラー = %v", err)
	}
	systemEditID := testutil.NewDBEditBuilder(t, tx).Build()
	if err := repo.Create(ctx, systemEditID, animeID, title("2")); err != nil {
		t.Fatalf("2つ目のCreate()のエラー = %v", err)
	}

	count, err := repo.CountHistoryByAnimeID(ctx, animeID)
	if err != nil {
		t.Fatalf("CountHistoryByAnimeID()のエラー = %v", err)
	}
	// 起点の版は数えない。
	if count != 2 {
		t.Errorf("版の数 = %d、期待値 = 2", count)
	}

	first, err := repo.ListHistoryByAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: animeID, Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("1ページ目のListHistoryByAnimeID()のエラー = %v", err)
	}
	second, err := repo.ListHistoryByAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: animeID, Page: 2, PerPage: 1})
	if err != nil {
		t.Fatalf("2ページ目のListHistoryByAnimeID()のエラー = %v", err)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("各ページの件数 = %d・%d、期待値 = 1・1", len(first), len(second))
	}

	// 新しい順に並び、1つ前の版はページの外にあっても付く。
	newest, oldest := first[0], second[0]
	if *newest.Snapshot.Title != "2" || newest.Previous == nil || *newest.Previous.Title != "1" {
		t.Errorf("1ページ目の版 = %q (前の版 %v)、期待値 = 2 (前の版 1)", *newest.Snapshot.Title, newest.Previous)
	}
	if newest.DBEdit == nil || newest.DBEdit.ID != systemEditID || newest.DBEdit.Actor != model.DBEditActorSystem || newest.DBEdit.User != nil {
		t.Errorf("1ページ目の操作 = %+v、期待値 = 編集者のいないシステムの操作%d", newest.DBEdit, systemEditID)
	}
	// 最初の操作の版は、起点の版と比べる。
	if *oldest.Snapshot.Title != "1" || oldest.Previous == nil || *oldest.Previous.Title != "起点" {
		t.Errorf("2ページ目の版 = %q (前の版 %v)、期待値 = 1 (前の版 起点)", *oldest.Snapshot.Title, oldest.Previous)
	}
	if oldest.DBEdit == nil || oldest.DBEdit.User == nil || oldest.DBEdit.User.ID != userID || oldest.DBEdit.User.Username == "" {
		t.Fatalf("2ページ目の操作の編集者 = %+v、期待値 = アットネームを持つユーザー%d", oldest.DBEdit, userID)
	}
	if got := oldest.DBEdit.User.ProfileImageData; !got.Valid || got.String != editorImageData {
		t.Errorf("2ページ目の操作の編集者のProfileImageData = %+v、期待値 = %q", got, editorImageData)
	}
}

func TestAnimeRevisionRepository_ListHistoryByAnimeID_NoPrevious(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	animeID := testutil.CreateTestAnime(t, tx)
	// 退会したユーザーの作成の操作 (主体がuserで編集者が空)。
	editID := testutil.NewDBEditBuilder(t, tx).WithActor(model.DBEditActorUser).WithAction(model.DBEditActionCreate).Build()
	if err := repo.Create(ctx, editID, animeID, model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	revisions, err := repo.ListHistoryByAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: animeID, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListHistoryByAnimeID()のエラー = %v", err)
	}
	if len(revisions) != 1 {
		t.Fatalf("版の件数 = %d、期待値 = 1", len(revisions))
	}
	if revisions[0].Previous != nil {
		t.Errorf("Previous = %+v、nilを期待", revisions[0].Previous)
	}
	if e := revisions[0].DBEdit; e.Actor != model.DBEditActorUser || e.UserID != nil || e.User != nil || e.Action != model.DBEditActionCreate {
		t.Errorf("操作 = %+v、期待値 = 編集者が空のuserの作成", e)
	}
}

func TestAnimeRevisionRepository_ListHistoryByWorkAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	workAnimeID := testutil.CreateTestAnime(t, tx)
	workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).Build()
	newEpisodeAnime := func(deleted bool) (model.EpisodeID, model.AnimeID) {
		animeID := testutil.CreateTestAnime(t, tx)
		episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
		if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1, deleted_at = CASE WHEN $3 THEN NOW() END WHERE id = $2`, int64(animeID), int64(episodeID), deleted); err != nil {
			t.Fatalf("エピソードの写像の設定に失敗: %v", err)
		}
		return episodeID, animeID
	}
	episodeID, episodeAnimeID := newEpisodeAnime(false)
	_, deletedEpisodeAnimeID := newEpisodeAnime(true)
	otherAnimeID := testutil.CreateTestAnime(t, tx)
	userID := testutil.NewUserBuilder(t, tx).WithProfileImageData(editorImageData).Build()

	snapshot := func(title string, kind model.AnimeClassificationKind) model.AnimeSnapshot {
		s := model.AnimeSnapshot{Title: &title, Status: model.AnimeStatusPublished, Kind: kind}
		if kind == model.AnimeClassificationKindEpisode {
			s.ParentAnimeID = &workAnimeID
		}
		return s
	}
	create := func(animeID model.AnimeID, s model.AnimeSnapshot) {
		if err := repo.Create(ctx, testutil.NewDBEditBuilder(t, tx).WithUser(userID).Build(), animeID, s); err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}
	}
	create(workAnimeID, snapshot("作品", model.AnimeClassificationKindWork))
	if err := repo.CreateInitial(ctx, episodeAnimeID, snapshot("起点", model.AnimeClassificationKindEpisode)); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}
	create(episodeAnimeID, snapshot("1", model.AnimeClassificationKindEpisode))
	create(episodeAnimeID, snapshot("2", model.AnimeClassificationKindEpisode))
	create(deletedEpisodeAnimeID, snapshot("削除したエピソード", model.AnimeClassificationKindEpisode))
	create(otherAnimeID, snapshot("ほかの作品", model.AnimeClassificationKindWork))

	count, err := repo.CountHistoryByWorkAnimeID(ctx, workAnimeID)
	if err != nil {
		t.Fatalf("CountHistoryByWorkAnimeID()のエラー = %v", err)
	}
	// 作品の版1つとエピソードの操作の版2つ。起点の版・削除したエピソード・ほかの作品の版は数えない。
	if count != 3 {
		t.Errorf("版の数 = %d、期待値 = 3", count)
	}

	revisions, err := repo.ListHistoryByWorkAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: workAnimeID, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListHistoryByWorkAnimeID()のエラー = %v", err)
	}
	if len(revisions) != 3 {
		t.Fatalf("版の件数 = %d、期待値 = 3", len(revisions))
	}
	// 新しい順に並び、エピソードの版の1つ前の版は、作品の版ではなく同じエピソードの版になる。
	for i, want := range []struct {
		title    string
		previous string
		episode  bool
	}{
		{"2", "1", true},
		{"1", "起点", true},
		{"作品", "", false},
	} {
		r := revisions[i]
		if *r.Snapshot.Title != want.title {
			t.Errorf("%d件目の版 = %q、期待値 = %q", i+1, *r.Snapshot.Title, want.title)
		}
		if want.previous == "" {
			if r.Previous != nil {
				t.Errorf("%d件目の1つ前の版 = %v、nilを期待", i+1, r.Previous)
			}
		} else if r.Previous == nil || *r.Previous.Title != want.previous {
			t.Errorf("%d件目の1つ前の版 = %v、期待値 = %q", i+1, r.Previous, want.previous)
		}
		if want.episode != (r.Episode != nil) || (r.Episode != nil && r.Episode.ID != episodeID) {
			t.Errorf("%d件目のEpisode = %+v、エピソード%dの版か = %v", i+1, r.Episode, episodeID, want.episode)
		}
		if r.DBEdit == nil || r.DBEdit.User == nil || r.DBEdit.User.ProfileImageData.String != editorImageData {
			t.Errorf("%d件目の操作の編集者 = %+v、期待値 = プロフィール画像%qを持つユーザー%d", i+1, r.DBEdit, editorImageData, userID)
		}
	}
}

// 移動前の作品にはその時点の版を残すが、現在の親作品が削除されたエピソードの版は、
// 直接の履歴ページと同じく見せない。除外した版がページの枠や総数にも残らないことを確かめる。
func TestAnimeRevisionRepository_ListHistoryByWorkAnimeID_HidesDeletedCurrentParent(t *testing.T) {
	t.Parallel()

	_, tx := testutil.SetupTx(t)
	ctx := context.Background()
	repo := repository.NewAnimeRevisionRepository(query.New(tx))
	workAnimeID := testutil.CreateTestAnime(t, tx)
	workID := testutil.NewWorkBuilder(t, tx).Build()
	movedWorkID := testutil.NewWorkBuilder(t, tx).Build()
	episodeAnimeID := testutil.CreateTestAnime(t, tx)
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, int64(episodeAnimeID), int64(episodeID)); err != nil {
		t.Fatalf("エピソードの写像の設定に失敗: %v", err)
	}
	for _, revision := range []struct {
		animeID  model.AnimeID
		snapshot model.AnimeSnapshot
	}{
		{workAnimeID, model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}},
		{episodeAnimeID, model.AnimeSnapshot{Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindEpisode, ParentAnimeID: &workAnimeID}},
	} {
		if err := repo.Create(ctx, testutil.NewDBEditBuilder(t, tx).Build(), revision.animeID, revision.snapshot); err != nil {
			t.Fatalf("版の作成に失敗: %v", err)
		}
	}
	if _, err := tx.Exec(`UPDATE episodes SET work_id = $1 WHERE id = $2`, int64(movedWorkID), int64(episodeID)); err != nil {
		t.Fatalf("エピソードの移動に失敗: %v", err)
	}

	assertFirstPage := func(wantCount int64, wantAnimeID model.AnimeID) {
		t.Helper()
		count, err := repo.CountHistoryByWorkAnimeID(ctx, workAnimeID)
		if err != nil {
			t.Fatalf("版の数の取得に失敗: %v", err)
		}
		if count != wantCount {
			t.Errorf("版の数 = %d、期待値 = %d", count, wantCount)
		}
		revisions, err := repo.ListHistoryByWorkAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: workAnimeID, Page: 1, PerPage: 1})
		if err != nil {
			t.Fatalf("版の取得に失敗: %v", err)
		}
		if len(revisions) != 1 || revisions[0].AnimeID != wantAnimeID {
			t.Fatalf("1ページ目の版 = %+v、期待値 = anime%dの1件", revisions, wantAnimeID)
		}
	}
	// 移動先の作品が見える間は、移動前の作品の履歴にもエピソードの版が残る。
	assertFirstPage(2, episodeAnimeID)
	if _, err := tx.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(movedWorkID)); err != nil {
		t.Fatalf("移動先の作品の削除に失敗: %v", err)
	}
	// 新しいエピソードの版が除外され、古い作品の版が空き枠を埋める。
	assertFirstPage(1, workAnimeID)
	revisions, err := repo.ListHistoryByWorkAnimeID(ctx, repository.AnimeRevisionHistoryParams{AnimeID: workAnimeID, Page: 2, PerPage: 1})
	if err != nil {
		t.Fatalf("2ページ目の版の取得に失敗: %v", err)
	}
	if len(revisions) != 0 {
		t.Errorf("2ページ目の版の数 = %d、期待値 = 0", len(revisions))
	}
}
