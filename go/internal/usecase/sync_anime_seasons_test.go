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

// newSyncAnimeSeasonsUsecase builds the reconciler and its repository over the shared
// test DB. The usecase opens its own transaction, so the test commits its setup (animes)
// directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeSeasonsUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを
// 組み立てる。本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) を
// アウター tx で包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeSeasonsUsecase(db *sql.DB) (*SyncAnimeSeasonsUsecase, *repository.AnimeSeasonRepository) {
	repo := repository.NewAnimeSeasonRepository(query.New(db))
	return NewSyncAnimeSeasonsUsecase(db, repo), repo
}

// workForSeasonSync builds an anime-resolved work carrying only the columns the season
// reconciler reads (season_year and the legacy season_name integer).
//
// [Ja] workForSeasonSync は季節リコンサイラが読むカラム (season_year と旧 season_name の
// integer) だけを持つ anime 解決済みの work を組み立てる。
func workForSeasonSync(animeID model.AnimeID, year, seasonName *int32) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, SeasonYear: year, SeasonName: seasonName}
}

// seasonsOf reads back an anime's seasons ordered by (year, name).
//
// [Ja] seasonsOf は anime の季節を (year, name) 順で読み戻す。
func seasonsOf(t *testing.T, repo *repository.AnimeSeasonRepository, animeID model.AnimeID) []*model.AnimeSeason {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	return rows
}

func TestSyncAnimeSeasonsUsecase_Reconcile_CreatesRowFromWorkColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// The legacy season_name integer 2 maps to spring.
	//
	// [Ja] 旧 season_name の integer 2 は spring に写像される。
	work := workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 only", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 {
		t.Fatalf("len(seasons) = %d, want 1", len(got))
	}
	s := got[0]
	if s.Year != 2024 || s.Name == nil || *s.Name != model.SeasonNameSpring || !s.IsPrimary {
		t.Errorf("season = %+v, want 2024 spring is_primary", s)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_MapsAutumnToFall(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// The legacy works.season_name 4 = autumn, folded to the fall enum value.
	//
	// [Ja] 旧 works.season_name 4 = autumn は fall の enum 値に寄せる。
	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(4))}); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Name == nil || *got[0].Name != model.SeasonNameFall {
		t.Errorf("seasons = %+v, want a single fall row", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_CreatesYearOnlyRowWhenNameUndetermined(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	// season_year present but season_name NULL: a year-only row with an undetermined
	// (NULL) name is created.
	//
	// [Ja] season_year はあるが season_name は NULL: 名前未定 (NULL) の年のみの行を作る。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 {
		t.Errorf("counts = %+v, want Created 1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Year != 2024 || got[0].Name != nil {
		t.Errorf("seasons = %+v, want a single 2024 row with NULL name", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// Re-running with the same source must detect no diff and write nothing. This is the
	// invariant the cutover decision depends on (a synced page reports zero diff).
	//
	// [Ja] 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v, want Unchanged 1 only", counts)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_ReplacesChangedSeason(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// (year, name) is the natural key, so a changed season is a delete (old) plus a
	// create (new). Both rows are is_primary, so the delete must run before the create or
	// the partial UNIQUE index on (anime_id) WHERE is_primary would reject the second
	// primary row — this test guards that ordering.
	//
	// [Ja] (year, name) は自然キーのため、季節の変更は削除 (old) + 作成 (new) になる。両行と
	// も is_primary なので、削除を作成より先に走らせなければ (anime_id) WHERE is_primary の
	// 部分 UNIQUE インデックスが 2 つ目の主行を拒否する。本テストはその順序を担保する。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(3))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Deleted != 1 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 and Deleted 1", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Name == nil || *got[0].Name != model.SeasonNameSummer {
		t.Errorf("seasons = %+v, want a single summer row after replace", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, int32Ptr(2024), int32Ptr(2))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// season_year is gone (NULL); the works-managed row should be deleted.
	//
	// [Ja] season_year が消えた (NULL)。works 管理下の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, int32Ptr(2))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	if got := seasonsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("seasons = %+v, want none after source removed", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_TreatsNullYearAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	// A NULL season_year yields no row even when season_name is present: the row's
	// existence is gated on the year.
	//
	// [Ja] season_year が NULL なら season_name があっても行は作られない: 行の有無は year で
	// 決まる。
	animeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, int32Ptr(2))})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if got := seasonsOf(t, repo, animeID); len(got) != 0 {
		t.Errorf("seasons = %+v, want none for NULL year", got)
	}
}

func TestSyncAnimeSeasonsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeSeasonsUsecase(db)

	animeID := insertBareAnime(t, db)

	// Seed a works-managed primary season (is_primary true, via the repo) and an
	// editor-added secondary season (is_primary false, inserted directly since works
	// never source a non-primary slot). anime_seasons has no kind / service
	// discriminator, so the is_primary slot is what marks a row as works-managed and thus
	// deletable.
	//
	// [Ja] works 管理下の主季節 (is_primary true、リポジトリ経由で作成) と編集者追加の副次
	// シーズン (is_primary false、works は非主スロットを source しないため直接挿入) を用意する。
	// anime_seasons には kind / service の判別列が無いため、is_primary スロットが「works 管理下
	// (= 削除対象)」であることを示す。
	spring := model.SeasonNameSpring
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2024, Name: &spring, IsPrimary: true,
	}); err != nil {
		t.Fatalf("seed managed Create() error = %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO anime_seasons (anime_id, year, name, is_primary, created_at, updated_at) VALUES ($1, $2, NULL, $3, NOW(), NOW())`,
		int64(animeID), 2099, false,
	); err != nil {
		t.Fatalf("seed editor INSERT error = %v", err)
	}

	// The work sources nothing, so the managed is_primary row is deleted while the
	// editor-added is_primary=false row outside the works-managed slot is preserved.
	//
	// [Ja] work は何も source しないため、管理下の is_primary 行は削除されるが、works 管理下の
	// スロットの外にある編集者追加の is_primary=false の行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForSeasonSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	got := seasonsOf(t, repo, animeID)
	if len(got) != 1 || got[0].Year != 2099 || got[0].IsPrimary {
		t.Errorf("seasons = %+v, want the editor-added 2099 row preserved", got)
	}
}
