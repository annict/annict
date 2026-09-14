package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newSyncAnimeSeasonsUsecaseは共有テストDB上にリコンサイラとそのリポジトリを
// 組み立てる。本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) を
// アウターtxで包まずGetTestDB経由で直接コミットする。
func newSyncAnimeSeasonsUsecase(db *sql.DB) (*SyncAnimeSeasonsUsecase, *repository.AnimeSeasonRepository) {
	repo := repository.NewAnimeSeasonRepository(query.New(db))
	return NewSyncAnimeSeasonsUsecase(db, repo), repo
}

// workForSeasonSyncは季節リコンサイラが読むカラム (season_yearと旧season_nameの
// integer) だけを持つanime解決済みのworkを組み立てる。
func workForSeasonSync(animeID model.AnimeID, year, seasonName *int32) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, SeasonYear: year, SeasonName: seasonName}
}

// seasonsOfはanimeの季節を (year, name) 順で読み戻す。
func seasonsOf(t *testing.T, repo *repository.AnimeSeasonRepository, animeID model.AnimeID) []*model.AnimeSeason {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	return rows
}

func TestSyncAnimeSeasonsUsecase_Reconcile_CreatesRowFromWorkColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// 旧season_nameのinteger 2はspringに写像される。
	work := workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 {
		t.Fatalf("len(seasons) = %d、期待値 = 1", len(got))
	}
	s := got[0]
	if s.Year != 2024 || s.Name == nil || *s.Name != model.SeasonNameSpring || !s.IsPrimary {
		t.Errorf("season = %+v、期待値 = 2024 spring is_primary", s)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_MapsAutumnToFall(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// 旧works.season_name 4 = autumnはfallのenum値に寄せる。
	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(4))}); err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Name == nil || *got[0].Name != model.SeasonNameFall {
		t.Errorf("seasons = %+v、期待値 = fallの行が1件", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_CreatesYearOnlyRowWhenNameUndetermined(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// season_yearはあるがseason_nameはNULL: 名前未定 (NULL) の年のみの行を作る。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 {
		t.Errorf("counts = %+v、期待値 = Created 1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Year != 2024 || got[0].Name != nil {
		t.Errorf("seasons = %+v、期待値 = nameがNULLの2024の行が1件", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ1", counts)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_ReplacesChangedSeason(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// (year, name) は自然キーのため、季節の変更は削除 (old) + 作成 (new) になる。両行と
	// もis_primaryなので、削除を作成より先に走らせなければ (anime_id) WHERE is_primaryの
	// 部分UNIQUEインデックスが2つ目の主行を拒否する。本テストはその順序を担保する。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(3))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Deleted != 1 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Created 1とDeleted 1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Name == nil || *got[0].Name != model.SeasonNameSummer {
		t.Errorf("置き換え後のseasons = %+v、期待値 = summerの行が1件", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// season_yearが消えた (NULL)。works管理下の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, int32Ptr(2))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	if got := seasonsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("同期元の削除後のseasons = %+v、期待値 = 0件", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_TreatsNullYearAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	// season_yearがNULLならseason_nameがあっても行は作られない: 行の有無はyearで
	// 決まる。
	animeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, int32Ptr(2))})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if got := seasonsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("yearがNULLのときのseasons = %+v、期待値 = 0件", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	// works管理下の主季節 (is_primary true、リポジトリ経由で作成) と編集者追加の副次
	// シーズン (is_primary false、worksは非主スロットをsourceしないため直接挿入) を用意する。
	// anime_seasonsにはkind / serviceの判別列が無いため、is_primaryスロットが「works管理下
	// (= 削除対象)」であることを示す。
	spring := model.SeasonNameSpring
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2024, Name: &spring, IsPrimary: true,
	}); err != nil {
		t.Fatalf("シードの管理対象のCreate()のエラー = %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO anime_seasons (anime_id, year, name, is_primary, created_at, updated_at) VALUES ($1, $2, NULL, $3, NOW(), NOW())`,
		int64(animeID), 2099, false,
	); err != nil {
		t.Fatalf("シードの編集者追加行のINSERTのエラー = %v", err)
	}

	// workは何もsourceしないため、管理下のis_primary行は削除されるが、works管理下の
	// スロットの外にある編集者追加のis_primary=falseの行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Year != 2099 || got[0].IsPrimary {
		t.Errorf("seasons = %+v、期待値 = 編集者が追加した2099の行が保持されること", got)
	}
}
