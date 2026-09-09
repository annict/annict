package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// fakeWorkIDPager serves work IDs from a backing slice in keyset pages, honoring
// the batchSize the usecase passes. It records the (afterID, batchSize) of every
// call so the test can assert the cursor advances correctly.
//
// [Ja] fakeWorkIDPager は backing スライスから work ID を keyset ページで返し、UseCase が
// 渡す batchSize を尊重する。各呼び出しの (afterID, batchSize) を記録し、カーソルが正しく
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

// fakeEpisodeIDPager mirrors fakeWorkIDPager for episode IDs.
//
// [Ja] fakeEpisodeIDPager は episode ID 用に fakeWorkIDPager を写したもの。
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

// fakeWorksSyncer records the work IDs it was asked to sync and returns a fixed
// per-call result so the test can assert aggregation.
//
// [Ja] fakeWorksSyncer は同期を依頼された work ID を記録し、固定の呼び出し結果を返して
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

// fakeEpisodesSyncer mirrors fakeWorksSyncer for episodes. It also records whether
// the works sync had already run, so the test can assert the works-before-episodes
// ordering.
//
// [Ja] fakeEpisodesSyncer は episodes 用に fakeWorksSyncer を写したもの。works 同期が
// 既に走ったかも記録し、works→episodes の順序をテストで検証できるようにする。
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

// fakeWorkSatellitesSyncer mirrors fakeWorksSyncer for the satellite pass. It also
// records whether the works sync had already run, so the test can assert the
// works-before-satellites ordering the pass depends on (anime_id must be written back
// by the works pass first).
//
// [Ja] fakeWorkSatellitesSyncer は別表パス用に fakeWorksSyncer を写したもの。works 同期が
// 既に走ったかも記録し、本パスが依存する works→別表 の順序 (anime_id は works パスが先に
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
		t.Fatalf("Execute() error = %v", err)
	}

	// Works: 5 IDs in pages of 2 -> [1,2] [3,4] [5]; the loop makes one more call
	// (cursor 5) that returns empty and stops. The satellite pass walks the same
	// works again, so it produces the same three pages.
	//
	// [Ja] works: 5 件を 2 件ずつ -> [1,2] [3,4] [5]。ループはもう 1 回 (カーソル 5)
	// 呼び出して空ページを受け取り停止する。別表パスは同じ works を再走査するため、
	// 同じ 3 ページになる。
	wantWorkPages := [][]model.WorkID{{1, 2}, {3, 4}, {5}}
	if len(worksSyncer.gotPages) != len(wantWorkPages) {
		t.Fatalf("works pages = %d, want %d (%v)", len(worksSyncer.gotPages), len(wantWorkPages), worksSyncer.gotPages)
	}
	for i, want := range wantWorkPages {
		if !equalWorkIDs(worksSyncer.gotPages[i], want) {
			t.Errorf("works page %d = %v, want %v", i, worksSyncer.gotPages[i], want)
		}
	}
	if len(satellitesSyncer.gotPages) != len(wantWorkPages) {
		t.Fatalf("satellite pages = %d, want %d (%v)", len(satellitesSyncer.gotPages), len(wantWorkPages), satellitesSyncer.gotPages)
	}
	for i, want := range wantWorkPages {
		if !equalWorkIDs(satellitesSyncer.gotPages[i], want) {
			t.Errorf("satellite page %d = %v, want %v", i, satellitesSyncer.gotPages[i], want)
		}
	}

	// The work pager is walked twice (works pass, then satellite pass); each walk's
	// cursor strictly advances 0, 2, 4, then 5 (empty page).
	//
	// [Ja] work pager は 2 度走査される (works パス、続いて別表パス)。各走査のカーソルは
	// 厳密に前進する: 0, 2, 4, 続いて 5 (空ページ)。
	wantWorkCursors := []model.WorkID{0, 2, 4, 5, 0, 2, 4, 5}
	if len(workPager.calls) != len(wantWorkCursors) {
		t.Fatalf("work pager calls = %d, want %d", len(workPager.calls), len(wantWorkCursors))
	}
	for i, want := range wantWorkCursors {
		if workPager.calls[i].afterID != want {
			t.Errorf("work pager call %d afterID = %d, want %d", i, workPager.calls[i].afterID, want)
		}
		if workPager.calls[i].batchSize != 2 {
			t.Errorf("work pager call %d batchSize = %d, want 2", i, workPager.calls[i].batchSize)
		}
	}

	// Aggregated counts: works synced 3 pages (created+1, updated+1 each), episodes
	// 2 pages (created+1, skipped+1 each), satellites 3 pages (created+1, deleted+1 each).
	//
	// [Ja] 集計件数: works は 3 ページ (各 created+1, updated+1)、episodes は 2 ページ
	// (各 created+1, skipped+1)、別表は 3 ページ (各 created+1, deleted+1)。
	if result.Works.Processed != 5 {
		t.Errorf("Works.Processed = %d, want 5", result.Works.Processed)
	}
	if result.Works.Created != 3 {
		t.Errorf("Works.Created = %d, want 3", result.Works.Created)
	}
	if result.Works.Updated != 3 {
		t.Errorf("Works.Updated = %d, want 3", result.Works.Updated)
	}
	if result.Episodes.Processed != 3 {
		t.Errorf("Episodes.Processed = %d, want 3", result.Episodes.Processed)
	}
	if result.Episodes.Created != 2 {
		t.Errorf("Episodes.Created = %d, want 2", result.Episodes.Created)
	}
	if result.Episodes.SkippedNoParent != 2 {
		t.Errorf("Episodes.SkippedNoParent = %d, want 2", result.Episodes.SkippedNoParent)
	}
	if result.Satellites.Processed != 5 {
		t.Errorf("Satellites.Processed = %d, want 5", result.Satellites.Processed)
	}
	if result.Satellites.Created != 3 {
		t.Errorf("Satellites.Created = %d, want 3", result.Satellites.Created)
	}
	if result.Satellites.Deleted != 3 {
		t.Errorf("Satellites.Deleted = %d, want 3", result.Satellites.Deleted)
	}

	// Ordering: every episodes page and every satellite page must have run after the
	// works pass completed.
	//
	// [Ja] 順序: episodes の各ページと別表の各ページは works の走査完了後に実行される必要がある。
	if !episodesSyncer.worksWereDone {
		t.Error("episodes were synced before works; works must run first")
	}
	if !satellitesSyncer.worksWereDone {
		t.Error("satellites were synced before works; works must run first")
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
		t.Fatalf("Execute() error = %v", err)
	}

	if len(worksSyncer.gotPages) != 0 {
		t.Errorf("works syncer called %d times, want 0", len(worksSyncer.gotPages))
	}
	if len(episodesSyncer.gotPages) != 0 {
		t.Errorf("episodes syncer called %d times, want 0", len(episodesSyncer.gotPages))
	}
	if len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("satellites syncer called %d times, want 0", len(satellitesSyncer.gotPages))
	}
	if result.Works.Processed != 0 || result.Episodes.Processed != 0 || result.Satellites.Processed != 0 {
		t.Errorf("Processed counts = (%d, %d, %d), want (0, 0, 0)", result.Works.Processed, result.Episodes.Processed, result.Satellites.Processed)
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
		t.Fatalf("Execute() error = %v, want wraps %v", err, wantErr)
	}

	// The error happened during the works pass, so neither episodes nor satellites
	// must be touched.
	//
	// [Ja] エラーは works の走査中に発生したため、episodes も別表も一切触れないこと。
	if len(episodesSyncer.gotPages) != 0 {
		t.Errorf("episodes syncer called %d times after works error, want 0", len(episodesSyncer.gotPages))
	}
	if len(satellitesSyncer.gotPages) != 0 {
		t.Errorf("satellites syncer called %d times after works error, want 0", len(satellitesSyncer.gotPages))
	}
}

func TestNewSyncAnimesUsecase_DefaultsBatchSize(t *testing.T) {
	t.Parallel()

	workPager := &fakeWorkIDPager{ids: []model.WorkID{1}}
	episodePager := &fakeEpisodeIDPager{}
	worksSyncer := &fakeWorksSyncer{}
	episodesSyncer := &fakeEpisodesSyncer{}
	satellitesSyncer := &fakeWorkSatellitesSyncer{}

	// A non-positive batch size must fall back to the default so the keyset loop
	// always makes progress (a LIMIT 0 page would return empty and stall).
	//
	// [Ja] 非正の batch size は既定値にフォールバックし、keyset ループが必ず前進する
	// こと (LIMIT 0 のページは空を返して停滞する)。
	uc := NewSyncAnimesUsecase(workPager, episodePager, worksSyncer, episodesSyncer, satellitesSyncer, 0)

	if _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if workPager.calls[0].batchSize != DefaultSyncAnimesBatchSize {
		t.Errorf("batchSize = %d, want default %d", workPager.calls[0].batchSize, DefaultSyncAnimesBatchSize)
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
