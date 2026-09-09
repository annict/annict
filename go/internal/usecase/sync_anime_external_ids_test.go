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

func int32Ptr(v int32) *int32 { return &v }

// newSyncAnimeExternalIDsUsecase builds the reconciler and its repository over the
// shared test DB. The usecase opens its own transaction, so the test commits its
// setup (animes) directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeExternalIDsUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを
// 組み立てる。本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) を
// アウター tx で包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeExternalIDsUsecase(db *sql.DB) (*SyncAnimeExternalIDsUsecase, *repository.AnimeExternalIDRepository) {
	repo := repository.NewAnimeExternalIDRepository(query.New(db))
	return NewSyncAnimeExternalIDsUsecase(db, repo), repo
}

// workForExternalIDSync builds an anime-resolved work carrying only the columns the
// external-ID reconciler reads (sc_tid / mal_anime_id).
//
// [Ja] workForExternalIDSync は外部 ID リコンサイラが読むカラム (sc_tid / mal_anime_id)
// だけを持つ anime 解決済みの work を組み立てる。
func workForExternalIDSync(animeID model.AnimeID, scTid, malAnimeID *int32) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, ScTid: scTid, MalAnimeID: malAnimeID}
}

// externalIDsByService reads back an anime's external IDs keyed by service.
//
// [Ja] externalIDsByService は anime の外部 ID をサービスをキーに読み戻す。
func externalIDsByService(t *testing.T, repo *repository.AnimeExternalIDRepository, animeID model.AnimeID) map[model.AnimeExternalService]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	byService := make(map[model.AnimeExternalService]string, len(rows))
	for _, r := range rows {
		byService[r.Service] = r.ExternalID
	}
	return byService
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_CreatesRowsFromWorkColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)
	work := workForExternalIDSync(animeID, int32Ptr(12345), int32Ptr(678))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 2 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 2 only", counts)
	}

	byService := externalIDsByService(t, repo, animeID)
	if byService[model.AnimeExternalServiceSyobocal] != "12345" {
		t.Errorf("syobocal = %q, want 12345", byService[model.AnimeExternalServiceSyobocal])
	}
	if byService[model.AnimeExternalServiceMal] != "678" {
		t.Errorf("mal = %q, want 678", byService[model.AnimeExternalServiceMal])
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), int32Ptr(200))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// Re-running with the same source must detect no diff and write nothing. This is
	// the invariant the cutover decision depends on (a synced page reports zero diff).
	//
	// [Ja] 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が
	// 依拠する不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 2 {
		t.Errorf("counts = %+v, want Unchanged 2 only", counts)
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_UpdatesChangedExternalID(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), nil)}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(111), nil)})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Updated 1 only", counts)
	}

	if got := externalIDsByService(t, repo, animeID)[model.AnimeExternalServiceSyobocal]; got != "111" {
		t.Errorf("syobocal = %q, want 111 after update", got)
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), int32Ptr(200))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// mal_anime_id is gone (NULL); only the syobocal row should survive.
	//
	// [Ja] mal_anime_id が消えた (NULL)。syobocal の行だけが残るべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), nil)})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Unchanged != 1 || counts.Created != 0 || counts.Updated != 0 {
		t.Errorf("counts = %+v, want Deleted 1 / Unchanged 1", counts)
	}

	byService := externalIDsByService(t, repo, animeID)
	if _, ok := byService[model.AnimeExternalServiceMal]; ok {
		t.Error("mal row should have been deleted")
	}
	if byService[model.AnimeExternalServiceSyobocal] != "100" {
		t.Errorf("syobocal = %q, want 100 (kept)", byService[model.AnimeExternalServiceSyobocal])
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_TreatsNullAndZeroAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)
	// sc_tid = 0 and mal_anime_id = NULL are both "absent": no rows are created.
	//
	// [Ja] sc_tid = 0 と mal_anime_id = NULL はどちらも「欠損」: 行は作られない。
	work := workForExternalIDSync(animeID, int32Ptr(0), nil)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if rows := externalIDsByService(t, repo, animeID); len(rows) != 0 {
		t.Errorf("rows = %v, want none for absent sources", rows)
	}
}
