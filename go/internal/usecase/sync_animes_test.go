package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// fakeWorkIDPagerはbackingスライスからwork IDをkeysetページで返し、UseCaseが
// 渡すbatchSizeを尊重する。各呼び出しの (afterID, batchSize) を記録し、カーソルが正しく
// 前進することをテストで検証できるようにする。
type fakeWorkIDPager struct {
	ids       []model.WorkID
	calls     []fakeWorkPageCall
	err       error
	errOnCall int
	callCount int
}

type fakeWorkPageCall struct {
	afterID   model.WorkID
	batchSize int
}

func (p *fakeWorkIDPager) ListIDsAfter(_ context.Context, afterID model.WorkID, batchSize int) ([]model.WorkID, error) {
	p.callCount++
	p.calls = append(p.calls, fakeWorkPageCall{afterID: afterID, batchSize: batchSize})
	if p.err != nil && p.callCount == p.errOnCall {
		return nil, p.err
	}

	var page []model.WorkID
	for _, id := range p.ids {
		if id <= afterID {
			continue
		}
		page = append(page, id)
		if len(page) == batchSize {
			break
		}
	}
	return page, nil
}

// fakeEpisodeIDPagerはepisode ID用にfakeWorkIDPagerを写したもの。
type fakeEpisodeIDPager struct {
	ids   []model.EpisodeID
	calls []fakeEpisodePageCall
}

type fakeEpisodePageCall struct {
	afterID   model.EpisodeID
	batchSize int
}

func (p *fakeEpisodeIDPager) ListIDsAfter(_ context.Context, afterID model.EpisodeID, batchSize int) ([]model.EpisodeID, error) {
	p.calls = append(p.calls, fakeEpisodePageCall{afterID: afterID, batchSize: batchSize})

	var page []model.EpisodeID
	for _, id := range p.ids {
		if id <= afterID {
			continue
		}
		page = append(page, id)
		if len(page) == batchSize {
			break
		}
	}
	return page, nil
}

// fakeWorksSyncerは同期を依頼されたwork IDを記録し、固定の呼び出し結果を返して
// 集計をテストで検証できるようにする。
type fakeWorksSyncer struct {
	gotPages [][]model.WorkID
	result   SyncWorksToAnimesResult
}

func (s *fakeWorksSyncer) Execute(_ context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error) {
	s.gotPages = append(s.gotPages, input.WorkIDs)
	r := s.result
	r.Processed = len(input.WorkIDs)
	return &r, nil
}

// fakeEpisodesSyncerはepisodes用にfakeWorksSyncerを写したもの。works同期が
// 既に走ったかも記録し、works→episodesの順序をテストで検証できるようにする。
type fakeEpisodesSyncer struct {
	gotPages       [][]model.EpisodeID
	result         SyncEpisodesToAnimesResult
	worksDoneFirst func() bool
	worksWereDone  bool
}

func (s *fakeEpisodesSyncer) Execute(_ context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error) {
	if s.worksDoneFirst != nil {
		s.worksWereDone = s.worksDoneFirst()
	}
	s.gotPages = append(s.gotPages, input.EpisodeIDs)
	r := s.result
	r.Processed = len(input.EpisodeIDs)
	return &r, nil
}

// fakeWorkSatellitesSyncerは別表パス用にfakeWorksSyncerを写したもの。works同期が
// 既に走ったかも記録し、本パスが依存するworks→別表 の順序 (anime_idはworksパスが先に
// 書き戻す必要がある) をテストで検証できるようにする。
type fakeWorkSatellitesSyncer struct {
	gotPages       [][]model.WorkID
	result         SyncWorkSatellitesResult
	worksDoneFirst func() bool
	worksWereDone  bool
}

func (s *fakeWorkSatellitesSyncer) Execute(_ context.Context, input SyncWorkSatellitesInput) (*SyncWorkSatellitesResult, error) {
	if s.worksDoneFirst != nil {
		s.worksWereDone = s.worksDoneFirst()
	}
	s.gotPages = append(s.gotPages, input.WorkIDs)
	r := s.result
	r.Processed = len(input.WorkIDs)
	return &r, nil
}

func TestSyncAnimesUsecase_Execute_PagesAndAggregates(t *testing.T) {
	t.Parallel()

	workPager := &fakeWorkIDPager{ids: []model.WorkID{1, 2, 3, 4, 5}}
	episodePager := &fakeEpisodeIDPager{ids: []model.EpisodeID{10, 20, 30}}
	worksSyncer := &fakeWorksSyncer{result: SyncWorksToAnimesResult{Created: 1, Updated: 1}}
	episodesSyncer := &fakeEpisodesSyncer{result: SyncEpisodesToAnimesResult{Created: 1, SkippedNoParent: 1}}
	episodesSyncer.worksDoneFirst = func() bool { return len(worksSyncer.gotPages) > 0 }
	satellitesSyncer := &fakeWorkSatellitesSyncer{result: SyncWorkSatellitesResult{Created: 1, Deleted: 1}}
	satellitesSyncer.worksDoneFirst = func() bool { return len(worksSyncer.gotPages) > 0 }

	uc := NewSyncAnimesUsecase(workPager, episodePager, worksSyncer, episodesSyncer, satellitesSyncer, 2)

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works: 5件を2件ずつ -> [1,2] [3,4] [5]。ループはもう1回 (カーソル5)
	// 呼び出して空ページを受け取り停止する。別表パスは同じworksを再走査するため、
	// 同じ3ページになる。
	wantWorkPages := [][]model.WorkID{{1, 2}, {3, 4}, {5}}
	if len(worksSyncer.gotPages) != len(wantWorkPages) {
		t.Fatalf("作品のページ数 = %d、期待値 = %d (%v)", len(worksSyncer.gotPages), len(wantWorkPages), worksSyncer.gotPages)
	}
	for i, want := range wantWorkPages {
		if !equalWorkIDs(worksSyncer.gotPages[i], want) {
			t.Errorf("作品のページ%d = %v、期待値 = %v", i, worksSyncer.gotPages[i], want)
		}
	}
	if len(satellitesSyncer.gotPages) != len(wantWorkPages) {
		t.Fatalf("サテライトのページ数 = %d、期待値 = %d (%v)", len(satellitesSyncer.gotPages), len(wantWorkPages), satellitesSyncer.gotPages)
	}
	for i, want := range wantWorkPages {
		if !equalWorkIDs(satellitesSyncer.gotPages[i], want) {
			t.Errorf("サテライトのページ%d = %v、期待値 = %v", i, satellitesSyncer.gotPages[i], want)
		}
	}

	// work pagerは2度走査される (worksパス、続いて別表パス)。各走査のカーソルは
	// 厳密に前進する: 0, 2, 4, 続いて5 (空ページ)。
	wantWorkCursors := []model.WorkID{0, 2, 4, 5, 0, 2, 4, 5}
	if len(workPager.calls) != len(wantWorkCursors) {
		t.Fatalf("作品のページャーの呼び出し回数 = %d、期待値 = %d", len(workPager.calls), len(wantWorkCursors))
	}
	for i, want := range wantWorkCursors {
		if workPager.calls[i].afterID != want {
			t.Errorf("作品のページャーの呼び出し%dのafterID = %d、期待値 = %d", i, workPager.calls[i].afterID, want)
		}
		if workPager.calls[i].batchSize != 2 {
			t.Errorf("作品のページャーの呼び出し%dのbatchSize = %d、期待値 = 2", i, workPager.calls[i].batchSize)
		}
	}

	// 集計件数: worksは3ページ (各created+1, updated+1)、episodesは2ページ
	// (各created+1, skipped+1)、別表は3ページ (各created+1, deleted+1)。
	if result.Works.Processed != 5 {
		t.Errorf("Works.Processed = %d、期待値 = 5", result.Works.Processed)
	}
	if result.Works.Created != 3 {
		t.Errorf("Works.Created = %d、期待値 = 3", result.Works.Created)
	}
	if result.Works.Updated != 3 {
		t.Errorf("Works.Updated = %d、期待値 = 3", result.Works.Updated)
	}
	if result.Episodes.Processed != 3 {
		t.Errorf("Episodes.Processed = %d、期待値 = 3", result.Episodes.Processed)
	}
	if result.Episodes.Created != 2 {
		t.Errorf("Episodes.Created = %d、期待値 = 2", result.Episodes.Created)
	}
	if result.Episodes.SkippedNoParent != 2 {
		t.Errorf("Episodes.SkippedNoParent = %d、期待値 = 2", result.Episodes.SkippedNoParent)
	}
	if result.Satellites.Processed != 5 {
		t.Errorf("Satellites.Processed = %d、期待値 = 5", result.Satellites.Processed)
	}
	if result.Satellites.Created != 3 {
		t.Errorf("Satellites.Created = %d、期待値 = 3", result.Satellites.Created)
	}
	if result.Satellites.Deleted != 3 {
		t.Errorf("Satellites.Deleted = %d、期待値 = 3", result.Satellites.Deleted)
	}

	// 順序: episodesの各ページと別表の各ページはworksの走査完了後に実行される必要がある。
	if !episodesSyncer.worksWereDone {
		t.Error("worksより先にepisodesが同期された。worksが先に動くこと")
	}
	if !satellitesSyncer.worksWereDone {
		t.Error("worksより先にsatellitesが同期された。worksが先に動くこと")
	}
}

func TestSyncAnimesUsecase_Execute_EmptyTablesDoNotCallSyncers(t *testing.T) {
	t.Parallel()

	workPager := &fakeWorkIDPager{}
	episodePager := &fakeEpisodeIDPager{}
	worksSyncer := &fakeWorksSyncer{}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}

	uc := NewSyncAnimesUsecase(workPager, episodePager, worksSyncer, episodesSyncer, satellitesSyncer, 100)

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(worksSyncer.gotPages) != 0 {
		t.Errorf("worksのsyncerの呼び出し回数 = %d、期待値 = 0", len(worksSyncer.gotPages))
	}
	if len(episodesSyncer.gotPages) != 0 {
		t.Errorf("episodesのsyncerの呼び出し回数 = %d、期待値 = 0", len(episodesSyncer.gotPages))
	}
	if len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("satellitesのsyncerの呼び出し回数 = %d、期待値 = 0", len(satellitesSyncer.gotPages))
	}
	if result.Works.Processed != 0 || result.Episodes.Processed != 0 || result.Satellites.Processed != 0 {
		t.Errorf("Processedの件数 = (%d, %d, %d)、期待値 = (0, 0, 0)", result.Works.Processed, result.Episodes.Processed, result.Satellites.Processed)
	}
}

func TestSyncAnimesUsecase_Execute_PropagatesPagerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("pager boom")
	workPager := &fakeWorkIDPager{ids: []model.WorkID{1, 2, 3}, err: wantErr, errOnCall: 2}
	episodePager := &fakeEpisodeIDPager{}
	worksSyncer := &fakeWorksSyncer{}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}

	uc := NewSyncAnimesUsecase(workPager, episodePager, worksSyncer, episodesSyncer, satellitesSyncer, 1)

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, wantErr)
	}

	// エラーはworksの走査中に発生したため、episodesも別表も一切触れないこと。
	if len(episodesSyncer.gotPages) != 0 {
		t.Errorf("worksのエラー後のepisodesのsyncerの呼び出し回数 = %d、期待値 = 0", len(episodesSyncer.gotPages))
	}
	if len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("worksのエラー後のsatellitesのsyncerの呼び出し回数 = %d、期待値 = 0", len(satellitesSyncer.gotPages))
	}
}

func TestNewSyncAnimesUsecase_DefaultsBatchSize(t *testing.T) {
	t.Parallel()

	workPager := &fakeWorkIDPager{ids: []model.WorkID{1}}
	episodePager := &fakeEpisodeIDPager{}
	worksSyncer := &fakeWorksSyncer{}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}

	// 非正のbatch sizeは既定値にフォールバックし、keysetループが必ず前進する
	// こと (LIMIT 0のページは空を返して停滞する)。
	uc := NewSyncAnimesUsecase(workPager, episodePager, worksSyncer, episodesSyncer, satellitesSyncer, 0)

	if _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if workPager.calls[0].batchSize != DefaultSyncAnimesBatchSize {
		t.Errorf("batchSize = %d、期待値 = 既定値の%d", workPager.calls[0].batchSize, DefaultSyncAnimesBatchSize)
	}
}

func equalWorkIDs(a, b []model.WorkID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
