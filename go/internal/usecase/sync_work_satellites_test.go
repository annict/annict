package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// fakeSatelliteWorkLoaderは要求IDに関係なく固定のworksを返し、DBなしで
// オーケストレーションを動かせるようにする (ローダーSQL自体はリポジトリテストでカバー)。
type fakeSatelliteWorkLoader struct {
	works []*model.Work
	err   error
}

func (l *fakeSatelliteWorkLoader) ListForSatelliteSyncByIDs(_ context.Context, _ []model.WorkID) ([]*model.Work, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.works, nil
}

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

func workWithAnimeID(id model.WorkID, animeID int64) *model.Work {
	aid := model.AnimeID(animeID)
	return &model.Work{ID: id, AnimeID: &aid}
}

func TestSyncWorkSatellitesUsecase_Execute_FiltersUnresolvedAndAggregates(t *testing.T) {
	t.Parallel()

	// anime解決済みの2件と、anime_id未解決の1件。
	works := []*model.Work{
		workWithAnimeID(1, 1001),
		{ID: 2, AnimeID: nil},
		workWithAnimeID(3, 1003),
	}
	loader := &fakeSatelliteWorkLoader{works: works}
	r1 := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Created: 1, Updated: 2, Unchanged: 1}}
	r2 := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Deleted: 3}}

	uc := NewSyncWorkSatellitesUsecase(loader, r1, r2)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{1, 2, 3}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if result.Processed != 3 {
		t.Errorf("Processed = %d、期待値 = 3", result.Processed)
	}
	if result.SkippedNoAnime != 1 {
		t.Errorf("SkippedNoAnime = %d、期待値 = 1", result.SkippedNoAnime)
	}

	// 各リコンサイラはanime解決済みの2件 (id 1と3) だけを受け取る必要がある。
	for name, r := range map[string]*fakeSatelliteReconciler{"r1": r1, "r2": r2} {
		if len(r.gotWorks) != 1 {
			t.Fatalf("%sの呼び出し回数 = %d、期待値 = 1", name, len(r.gotWorks))
		}
		got := r.gotWorks[0]
		if len(got) != 2 || got[0].ID != 1 || got[1].ID != 3 {
			t.Errorf("%sのworks = %v、期待値 = idが [1 3]", name, workIDsOf(got))
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

	works := []*model.Work{{ID: 1, AnimeID: nil}, {ID: 2, AnimeID: nil}}
	loader := &fakeSatelliteWorkLoader{works: works}
	reconciler := &fakeSatelliteReconciler{counts: satelliteReconcileCounts{Created: 99}}

	uc := NewSyncWorkSatellitesUsecase(loader, reconciler)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{1, 2}})
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

	// タスク2-7の本番配線を写したもの: worksはanimeに解決するが、リコンサイラは
	// まだ未登録のため何も書かれず、メトリクスだけが動く。
	loader := &fakeSatelliteWorkLoader{works: []*model.Work{workWithAnimeID(1, 1001)}}

	uc := NewSyncWorkSatellitesUsecase(loader)

	result, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{1}})
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

func TestSyncWorkSatellitesUsecase_Execute_PropagatesLoaderError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("loader boom")
	loader := &fakeSatelliteWorkLoader{err: wantErr}
	reconciler := &fakeSatelliteReconciler{}

	uc := NewSyncWorkSatellitesUsecase(loader, reconciler)

	_, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{1}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, wantErr)
	}
	if len(reconciler.gotWorks) != 0 {
		t.Errorf("loaderのエラー後にreconcilerが呼ばれた。期待値 = 0回")
	}
}

func TestSyncWorkSatellitesUsecase_Execute_PropagatesReconcilerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("reconciler boom")
	loader := &fakeSatelliteWorkLoader{works: []*model.Work{workWithAnimeID(1, 1001)}}
	reconciler := &fakeSatelliteReconciler{err: wantErr}

	uc := NewSyncWorkSatellitesUsecase(loader, reconciler)

	_, err := uc.Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{1}})
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
