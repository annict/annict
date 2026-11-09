package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// fakeSatelliteReconcilerは渡されたworksを記録し固定の件数を返す。オーケストレーション
// のフィルタと集計が正しいかをテストで検証できるようにする。
type fakeSatelliteReconciler struct {
	gotWorks [][]*model.Work
	counts   satelliteReconcileCounts
	err      error
}

func (r *fakeSatelliteReconciler) Reconcile(_ context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	r.gotWorks = append(r.gotWorks, works)
	if r.err != nil {
		return satelliteReconcileCounts{}, r.err
	}
	return r.counts, nil
}

// newSyncWorkSatellitesUsecaseInTxは、テストのトランザクションでworksを読む
// Repositoryと、リコンサイラのフェイクからSyncWorkSatellitesUsecaseを組み立てる。
func newSyncWorkSatellitesUsecaseInTx(db *sql.DB, tx *sql.Tx, reconcilers ...satelliteReconciler) *SyncWorkSatellitesUsecase {
	return NewSyncWorkSatellitesUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)), reconcilers...)
}

// buildWorkWithAnimeInTxは、animeにマッピング済みの作品をテストのトランザクションに作る。
func buildWorkWithAnimeInTx(t *testing.T, tx *sql.Tx) model.WorkID {
	t.Helper()

	workID := testutil.NewWorkBuilder(t, tx).Build()
	var animeID int64
	if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&animeID); err != nil {
		t.Fatalf("animesの挿入に失敗: %v", err)
	}
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, animeID, int64(workID)); err != nil {
		t.Fatalf("works.anime_idの更新に失敗: %v", err)
	}
	return workID
}

func TestSyncWorkSatellitesUsecase_Execute_FiltersUnresolvedAndAggregates(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	// anime解決済みの2件と、anime_id未解決の1件。
	resolved1 := buildWorkWithAnimeInTx(t, tx)
	unresolved := testutil.NewWorkBuilder(t, tx).Build()
	resolved2 := buildWorkWithAnimeInTx(t, tx)

	r1 := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Created: 1, Updated: 2, Unchanged: 1}}
	r2 := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Deleted: 3}}

	uc := newSyncWorkSatellitesUsecaseInTx(db, tx, r1, r2)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{resolved1, unresolved, resolved2}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if result.Processed != 3 {
		t.Errorf("Processed = %d、期待値 = 3", result.Processed)
	}
	if result.SkippedNoAnime != 1 {
		t.Errorf("SkippedNoAnime = %d、期待値 = 1", result.SkippedNoAnime)
	}

	// 各リコンサイラはanime解決済みの2件だけを受け取る必要がある。
	for name, r := range map[string]*fakeSatelliteReconciler{"r1": r1, "r2": r2} {
		if len(r.gotWorks) != 1 {
			t.Fatalf("%sの呼び出し回数 = %d、期待値 = 1", name, len(r.gotWorks))
		}
		got := workIDsOf(r.gotWorks[0])
		if len(got) != 2 || got[0] != resolved1 || got[1] != resolved2 {
			t.Errorf("%sのworks = %v、期待値 = idが [%d %d]", name, got, resolved1, resolved2)
		}
	}

	// 件数は両リコンサイラで合算される。
	if result.Created != 1 || result.Updated != 2 || result.Unchanged != 1 || result.Deleted != 3 {
		t.Errorf("counts = created %d / updated %d / unchanged %d / deleted %d、期待値 = 1 / 2 / 1 / 3",
			result.Created, result.Updated, result.Unchanged, result.Deleted)
	}
}

func TestSyncWorkSatellitesUsecase_Execute_NoResolvedWorksSkipsReconcilers(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workIDs := []model.WorkID{
		testutil.NewWorkBuilder(t, tx).Build(),
		testutil.NewWorkBuilder(t, tx).Build(),
	}
	reconciler := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Created: 99}}

	uc := newSyncWorkSatellitesUsecaseInTx(db, tx, reconciler)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: workIDs})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(reconciler.gotWorks) != 0 {
		t.Errorf("reconcilerの呼び出し回数 = %d、期待値 = 0", len(reconciler.gotWorks))
	}
	if result.Processed != 2 || result.SkippedNoAnime != 2 {
		t.Errorf("Processed/SkippedNoAnime = %d/%d、期待値 = 2/2", result.Processed, result.SkippedNoAnime)
	}
	if result.Created != 0 {
		t.Errorf("Created = %d、期待値 = 0 (reconcilerが動かないこと)", result.Created)
	}
}

func TestSyncWorkSatellitesUsecase_Execute_NoReconcilersRegistered(t *testing.T) {
	t.Parallel()

	// リコンサイラを1つも登録しないとき: worksはanimeに解決するが何も書かれず、
	// メトリクスだけが動く。
	db, tx := testutil.SetupTx(t)
	workID := buildWorkWithAnimeInTx(t, tx)

	uc := newSyncWorkSatellitesUsecaseInTx(db, tx)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if result.Processed != 1 || result.SkippedNoAnime != 0 {
		t.Errorf("Processed/SkippedNoAnime = %d/%d、期待値 = 1/0", result.Processed, result.SkippedNoAnime)
	}
	if result.Created != 0 || result.Updated != 0 || result.Deleted != 0 || result.Unchanged != 0 {
		t.Errorf("書き込みのcounts = %d/%d/%d/%d、期待値 = すべて0", result.Created, result.Updated, result.Deleted, result.Unchanged)
	}
}

func TestSyncWorkSatellitesUsecase_Execute_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := buildWorkWithAnimeInTx(t, tx)
	reconciler := &fakeSatelliteReconciler{}

	uc := newSyncWorkSatellitesUsecaseInTx(db, tx, reconciler)

	// キャンセル済みのコンテキストで、worksの取得を失敗させる。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Execute(ctx, SyncWorkSatellitesInput{WorkIDs: []model.WorkID{workID}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, context.Canceled)
	}
	if len(reconciler.gotWorks) != 0 {
		t.Errorf("worksの取得のエラー後にreconcilerが呼ばれた。期待値 = 0回")
	}
}

func TestSyncWorkSatellitesUsecase_Execute_PropagatesReconcilerError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := buildWorkWithAnimeInTx(t, tx)
	wantErr := errors.New("reconciler boom")
	reconciler := &fakeSatelliteReconciler{err: wantErr}

	uc := newSyncWorkSatellitesUsecaseInTx(db, tx, reconciler)

	_, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{workID}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, wantErr)
	}
}

func workIDsOf(works []*model.Work) []model.WorkID {
	ids := make([]model.WorkID, len(works))
	for i, w := range works {
		ids[i] = w.ID
	}
	return ids
}
