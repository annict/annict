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

// newSyncAnimeExternalIDsUsecaseは共有テストDB上にリコンサイラとそのリポジトリを
// 組み立てる。本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) を
// アウターtxで包まずGetTestDB経由で直接コミットする。
func newSyncAnimeExternalIDsUsecase(db *sql.DB) (*SyncAnimeExternalIDsUsecase, *repository.AnimeExternalIDRepository) {
	repo := repository.NewAnimeExternalIDRepository(query.New(db))
	return NewSyncAnimeExternalIDsUsecase(db, repo), repo
}

// workForExternalIDSyncは外部IDリコンサイラが読むカラム (sc_tid / mal_anime_id)
// だけを持つanime解決済みのworkを組み立てる。
func workForExternalIDSync(animeID model.AnimeID, scTid, malAnimeID *int32) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, ScTid: scTid, MalAnimeID: malAnimeID}
}

// externalIDsByServiceはanimeの外部IDをサービスをキーに読み戻す。
func externalIDsByService(t *testing.T, repo *repository.AnimeExternalIDRepository, animeID model.AnimeID) map[model.AnimeExternalService]string {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
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
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 2 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ2", counts)
	}

	byService := externalIDsByService(t, repo, animeID)
	if byService[model.AnimeExternalServiceSyobocal] != "12345" {
		t.Errorf("syobocal = %q、期待値 = 12345", byService[model.AnimeExternalServiceSyobocal])
	}
	if byService[model.AnimeExternalServiceMal] != "678" {
		t.Errorf("mal = %q、期待値 = 678", byService[model.AnimeExternalServiceMal])
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), int32Ptr(200))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が
	// 依拠する不変条件 (同期済みのページは差分ゼロを報告する)。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 2 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ2", counts)
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_UpdatesChangedExternalID(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), nil)}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(111), nil)})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Updatedのみ1", counts)
	}

	if got := externalIDsByService(t, repo, animeID)[model.AnimeExternalServiceSyobocal]; got != "111" {
		t.Errorf("更新後のsyobocal = %q、期待値 = 111", got)
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), int32Ptr(200))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// mal_anime_idが消えた (NULL)。syobocalの行だけが残るべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForExternalIDSync(animeID, int32Ptr(100), nil)})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Unchanged != 1 || counts.Created != 0 || counts.Updated != 0 {
		t.Errorf("counts = %+v、期待値 = Deleted 1 / Unchanged 1", counts)
	}

	byService := externalIDsByService(t, repo, animeID)
	if _, ok := byService[model.AnimeExternalServiceMal]; ok {
		t.Error("malの行が削除されなかった")
	}
	if byService[model.AnimeExternalServiceSyobocal] != "100" {
		t.Errorf("syobocal = %q、期待値 = 100 (据え置き)", byService[model.AnimeExternalServiceSyobocal])
	}
}

func TestSyncAnimeExternalIDsUsecase_Reconcile_TreatsNullAndZeroAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeExternalIDsUsecase(db)

	animeID := insertBareAnime(t, db)
	// sc_tid = 0とmal_anime_id = NULLはどちらも「欠損」: 行は作られない。
	work := workForExternalIDSync(animeID, int32Ptr(0), nil)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if rows := externalIDsByService(t, repo, animeID); len(rows) != 0 {
		t.Errorf("同期元が無いときのrows = %v、期待値 = 0件", rows)
	}
}
