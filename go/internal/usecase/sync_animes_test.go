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

// fakeWorksSyncerは同期を依頼されたwork IDを記録し、固定の呼び出し結果を返して
// 集計をテストで検証できるようにする。
type fakeWorksSyncer struct {
	gotPages [][]model.WorkID
	result   SyncWorksToAnimesResult
	err      error
}

func (s *fakeWorksSyncer) Execute(_ context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error) {
	s.gotPages = append(s.gotPages, input.WorkIDs)
	if s.err != nil {
		return nil, s.err
	}
	r := s.result
	r.Processed = len(input.WorkIDs)
	return &r, nil
}

// fakeEpisodesSyncerはepisodes用にfakeWorksSyncerを写したもの。各ページの前に
// beforeExecuteを呼び、works→episodesの順序をテストで検証できるようにする。
type fakeEpisodesSyncer struct {
	gotPages      [][]model.EpisodeID
	result        SyncEpisodesToAnimesResult
	beforeExecute func()
}

func (s *fakeEpisodesSyncer) Execute(_ context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error) {
	if s.beforeExecute != nil {
		s.beforeExecute()
	}
	s.gotPages = append(s.gotPages, input.EpisodeIDs)
	r := s.result
	r.Processed = len(input.EpisodeIDs)
	return &r, nil
}

// fakeWorkSatellitesSyncerは別表パス用にfakeWorksSyncerを写したもの。各ページの前に
// beforeExecuteを呼び、本パスが依存するworks→別表の順序 (anime_idはworksパスが先に
// 書き戻す必要がある) をテストで検証できるようにする。
type fakeWorkSatellitesSyncer struct {
	gotPages      [][]model.WorkID
	result        SyncWorkSatellitesResult
	beforeExecute func()
}

func (s *fakeWorkSatellitesSyncer) Execute(_ context.Context, input SyncWorkSatellitesInput) (*SyncWorkSatellitesResult, error) {
	if s.beforeExecute != nil {
		s.beforeExecute()
	}
	s.gotPages = append(s.gotPages, input.WorkIDs)
	r := s.result
	r.Processed = len(input.WorkIDs)
	return &r, nil
}

// newSyncAnimesUsecaseInTxは、テストのトランザクションでworks / episodesを読む
// Repositoryと、子のUseCaseのフェイクからSyncAnimesUsecaseを組み立てる。
func newSyncAnimesUsecaseInTx(
	db *sql.DB,
	tx *sql.Tx,
	worksSyncer worksToAnimesSyncer,
	episodesSyncer episodesToAnimesSyncer,
	satellitesSyncer workSatellitesSyncer,
	batchSize int,
) *SyncAnimesUsecase {
	queries := query.New(db).WithTx(tx)
	return NewSyncAnimesUsecase(
		repository.NewWorkRepository(queries),
		repository.NewEpisodeRepository(queries),
		worksSyncer,
		episodesSyncer,
		satellitesSyncer,
		batchSize,
	)
}

// assertIDPagesは、走査したページが空でなく、batchSize以下で、最後のページ以外は
// batchSizeちょうどで、ページをまたいでIDが厳密に増加し、wantIDsをすべて含むことを
// 確かめる。テスト用DBにはほかのテストがコミットした行もあるため、ページの中身そのものは
// 固定しない。
func assertIDPages[ID ~int64](t *testing.T, name string, pages [][]ID, batchSize int, wantIDs []ID) {
	t.Helper()

	seen := make(map[ID]bool)
	var prev ID
	for i, page := range pages {
		if len(page) == 0 {
			t.Errorf("%sのページ%dが空。空のページで子のUseCaseを呼ばないこと", name, i)
		}
		if len(page) > batchSize {
			t.Errorf("%sのページ%dの件数 = %d、期待値 = %d以下", name, i, len(page), batchSize)
		}
		if i < len(pages)-1 && len(page) != batchSize {
			t.Errorf("%sのページ%dの件数 = %d、期待値 = %d (最後のページ以外)", name, i, len(page), batchSize)
		}
		for _, id := range page {
			if id <= prev {
				t.Errorf("%sのページ%dのID %dが直前のID %d以下。IDは厳密に増加すること", name, i, id, prev)
			}
			prev = id
			seen[id] = true
		}
	}
	for _, id := range wantIDs {
		if !seen[id] {
			t.Errorf("%sのページにID %dが含まれない", name, id)
		}
	}
}

func countIDs[ID any](pages [][]ID) int {
	n := 0
	for _, page := range pages {
		n += len(page)
	}
	return n
}

func TestSyncAnimesUsecase_Execute_PagesAndAggregates(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	// 並行テストのコミットで走査対象が変わらないよう、スナップショットを固定する。
	if _, err := tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"); err != nil {
		t.Fatalf("テスト用トランザクションの分離レベルの設定に失敗: %v", err)
	}

	var workIDs []model.WorkID
	for range 3 {
		workIDs = append(workIDs, testutil.NewWorkBuilder(t, tx).Build())
	}
	episodeIDs := []model.EpisodeID{
		testutil.NewEpisodeBuilder(t, tx, workIDs[0]).Build(),
		testutil.NewEpisodeBuilder(t, tx, workIDs[0]).Build(),
	}

	worksSyncer := &fakeWorksSyncer{result: SyncWorksToAnimesResult{Created: 1, Updated: 1}}
	episodesSyncer := &fakeEpisodesSyncer{result: SyncEpisodesToAnimesResult{Created: 1, SkippedNoParent: 1}}
	satellitesSyncer := &fakeWorkSatellitesSyncer{result: SyncWorkSatellitesResult{Created: 1, Deleted: 1}}
	// works同期のページ数を、episodesと別表の最初のページの時点で記録し、
	// works→episodes・works→別表の順序を検証する。
	worksPagesAtEpisodes, worksPagesAtSatellites := -1, -1
	episodesSyncer.beforeExecute = func() {
		if worksPagesAtEpisodes < 0 {
			worksPagesAtEpisodes = len(worksSyncer.gotPages)
		}
	}
	satellitesSyncer.beforeExecute = func() {
		if worksPagesAtSatellites < 0 {
			worksPagesAtSatellites = len(worksSyncer.gotPages)
		}
	}

	const batchSize = 2
	uc := newSyncAnimesUsecaseInTx(db, tx, worksSyncer, episodesSyncer, satellitesSyncer, batchSize)

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	assertIDPages(t, "works", worksSyncer.gotPages, batchSize, workIDs)
	assertIDPages(t, "episodes", episodesSyncer.gotPages, batchSize, episodeIDs)
	assertIDPages(t, "別表", satellitesSyncer.gotPages, batchSize, workIDs)

	// 子のUseCaseは各ページについて固定の件数を返すため、集計はページ数の倍数になる。
	// Processedは子のUseCaseに渡したIDの件数の合計になる。
	worksPages := len(worksSyncer.gotPages)
	if result.Works.Processed != countIDs(worksSyncer.gotPages) {
		t.Errorf("Works.Processed = %d、期待値 = %d", result.Works.Processed, countIDs(worksSyncer.gotPages))
	}
	if result.Works.Created != worksPages || result.Works.Updated != worksPages {
		t.Errorf("Works.Created / Updated = %d / %d、期待値 = %d / %d", result.Works.Created, result.Works.Updated, worksPages, worksPages)
	}
	episodesPages := len(episodesSyncer.gotPages)
	if result.Episodes.Processed != countIDs(episodesSyncer.gotPages) {
		t.Errorf("Episodes.Processed = %d、期待値 = %d", result.Episodes.Processed, countIDs(episodesSyncer.gotPages))
	}
	if result.Episodes.Created != episodesPages || result.Episodes.SkippedNoParent != episodesPages {
		t.Errorf("Episodes.Created / SkippedNoParent = %d / %d、期待値 = %d / %d", result.Episodes.Created, result.Episodes.SkippedNoParent, episodesPages, episodesPages)
	}
	satellitesPages := len(satellitesSyncer.gotPages)
	if result.Satellites.Processed != countIDs(satellitesSyncer.gotPages) {
		t.Errorf("Satellites.Processed = %d、期待値 = %d", result.Satellites.Processed, countIDs(satellitesSyncer.gotPages))
	}
	if result.Satellites.Created != satellitesPages || result.Satellites.Deleted != satellitesPages {
		t.Errorf("Satellites.Created / Deleted = %d / %d、期待値 = %d / %d", result.Satellites.Created, result.Satellites.Deleted, satellitesPages, satellitesPages)
	}

	// 順序: episodesと別表の最初のページは、worksの走査を終えてから実行される必要がある。
	if worksPagesAtEpisodes != worksPages {
		t.Errorf("episodesの開始時点のworksのページ数 = %d、期待値 = %d (worksの走査を終えてから)", worksPagesAtEpisodes, worksPages)
	}
	if worksPagesAtSatellites != worksPages {
		t.Errorf("別表の開始時点のworksのページ数 = %d、期待値 = %d (worksの走査を終えてから)", worksPagesAtSatellites, worksPages)
	}
}

func TestSyncAnimesUsecase_Execute_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	testutil.NewWorkBuilder(t, tx).Build()

	worksSyncer := &fakeWorksSyncer{}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}
	uc := newSyncAnimesUsecaseInTx(db, tx, worksSyncer, episodesSyncer, satellitesSyncer, 1)

	// キャンセル済みのコンテキストで、worksのIDのページの取得を失敗させる。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Execute(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, context.Canceled)
	}

	if len(worksSyncer.gotPages) != 0 || len(episodesSyncer.gotPages) != 0 || len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("子のUseCaseの呼び出し回数 = (%d, %d, %d)、期待値 = (0, 0, 0)",
			len(worksSyncer.gotPages), len(episodesSyncer.gotPages), len(satellitesSyncer.gotPages))
	}
}

func TestSyncAnimesUsecase_Execute_PropagatesSyncerError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	testutil.NewWorkBuilder(t, tx).Build()

	wantErr := errors.New("works同期の失敗")
	worksSyncer := &fakeWorksSyncer{err: wantErr}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}
	uc := newSyncAnimesUsecaseInTx(db, tx, worksSyncer, episodesSyncer, satellitesSyncer, 1)

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, wantErr)
	}

	// エラーはworksの走査中に発生したため、episodesも別表も一切触れないこと。
	if len(worksSyncer.gotPages) != 1 {
		t.Errorf("worksのsyncerの呼び出し回数 = %d、期待値 = 1 (最初のページで止まる)", len(worksSyncer.gotPages))
	}
	if len(episodesSyncer.gotPages) != 0 {
		t.Errorf("worksのエラー後のepisodesのsyncerの呼び出し回数 = %d、期待値 = 0", len(episodesSyncer.gotPages))
	}
	if len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("worksのエラー後のsatellitesのsyncerの呼び出し回数 = %d、期待値 = 0", len(satellitesSyncer.gotPages))
	}
}

func TestNewSyncAnimesUsecase_DefaultsBatchSize(t *testing.T) {
	t.Parallel()

	// 非正のbatch sizeは既定値にフォールバックし、keysetループが必ず前進する
	// こと (LIMIT 0のページは空を返して停滞する)。
	for _, batchSize := range []int{0, -1} {
		uc := NewSyncAnimesUsecase(nil, nil, &fakeWorksSyncer{}, &fakeEpisodesSyncer{}, &fakeWorkSatellitesSyncer{}, batchSize)
		if uc.batchSize != DefaultSyncAnimesBatchSize {
			t.Errorf("batchSize %dを渡したときのbatchSize = %d、期待値 = 既定値の%d", batchSize, uc.batchSize, DefaultSyncAnimesBatchSize)
		}
	}
}
