package usecase

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// fakeSubTitleLookupは、しょぼいカレンダーの代わりに決まったサブタイトルを返す。
type fakeSubTitleLookup struct {
	subTitles map[int32]map[int32]string
	err       error
	tids      []int32
}

func (f *fakeSubTitleLookup) LookupSubTitles(_ context.Context, tids []int32) (map[int32]map[int32]string, error) {
	f.tids = tids
	if f.err != nil {
		return nil, f.err
	}
	return f.subTitles, nil
}

func newFillEpisodeTitlesUsecase(db *sql.DB, lookup SyobocalSubTitleLookup) *FillEpisodeTitlesUsecase {
	queries := query.New(db)
	return NewFillEpisodeTitlesUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
		lookup,
	)
}

var titleFillTidCounter atomic.Int32

// insertTitleFillWorkは、しょぼいカレンダーのTIDを持つ作品を共有DBにコミットする。
// 対象のエピソードはDB全体から読まれるため、テストごとに別のTIDを使い、偽のしょぼいカレンダーが
// ほかのテストのエピソードにサブタイトルを返さないようにする。
func insertTitleFillWork(t *testing.T, db *sql.DB) (model.WorkID, int32) {
	t.Helper()

	tid := 900000000 + titleFillTidCounter.Add(1)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	if _, err := db.Exec(`UPDATE works SET sc_tid = $2 WHERE id = $1`, int64(workID), tid); err != nil {
		t.Fatalf("TIDの設定に失敗: %v", err)
	}
	return workID, tid
}

// insertTitleFillEpisodeは、タイトルが空のエピソードを番組の放送枠に紐づけて共有DBに
// コミットする。
func insertTitleFillEpisode(t *testing.T, db *sql.DB, program *generationProgram, rawNumber float64) model.EpisodeID {
	t.Helper()

	episodeID := insertSyncEpisode(t, db, generationEpisodeInput(program.workID, rawNumber, int32(rawNumber*100)))
	program.insertSlot(int32(rawNumber*10), time.Now(), func(b *testutil.SlotBuilder) { b.WithEpisodeID(episodeID) })
	return episodeID
}

func readEpisodeTitle(t *testing.T, db *sql.DB, episodeID model.EpisodeID) (sql.NullString, time.Time) {
	t.Helper()

	var (
		title     sql.NullString
		updatedAt time.Time
	)
	if err := db.QueryRow(`SELECT title, updated_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&title, &updatedAt); err != nil {
		t.Fatalf("エピソードの読み込みに失敗: %v", err)
	}
	return title, updatedAt
}

// TestFillEpisodeTitlesUsecase_Execute_FillsEpisodeAndAnimeは、サブタイトルをエピソードと
// そのanimeに両書きし、DB活動の記録・作品のtouch・updated_atの更新を行わないことを検証する。
func TestFillEpisodeTitlesUsecase_Execute_FillsEpisodeAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID, tid := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	mappedID := insertTitleFillEpisode(t, db, program, 1)
	animeID := insertBareAnime(t, db)
	if _, err := db.Exec(`UPDATE episodes SET anime_id = $2, updated_at = '2001-01-01' WHERE id = $1`, int64(mappedID), int64(animeID)); err != nil {
		t.Fatalf("エピソードの更新に失敗: %v", err)
	}
	unmappedID := insertTitleFillEpisode(t, db, program, 2)

	var workUpdatedAt time.Time
	if err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&workUpdatedAt); err != nil {
		t.Fatalf("作品の読み込みに失敗: %v", err)
	}

	lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{
		tid: {1: "オペレーション〈梟(ストリクス)〉", 2: "妻役を確保せよ"},
	}}
	output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(ctx, FillEpisodeTitlesInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if output.FilledCount != 2 || len(output.FailedEpisodes) != 0 {
		t.Errorf("output = %+v、期待値 = 2件を入れる", output)
	}
	if !slices.Contains(lookup.tids, tid) {
		t.Errorf("問い合わせたTID = %v、%dを含むことを期待", lookup.tids, tid)
	}

	title, updatedAt := readEpisodeTitle(t, db, mappedID)
	if title.String != "オペレーション〈梟(ストリクス)〉" || updatedAt.Year() != 2001 {
		t.Errorf("第1話 = (%q, %v)、期待値 = (オペレーション〈梟(ストリクス)〉, 2001年のまま)", title.String, updatedAt)
	}
	if title, _ := readEpisodeTitle(t, db, unmappedID); title.String != "妻役を確保せよ" {
		t.Errorf("第2話のtitle = %q、期待値 = 妻役を確保せよ", title.String)
	}

	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "オペレーション〈梟(ストリクス)〉" {
		t.Errorf("anime.Title = %q、期待値 = オペレーション〈梟(ストリクス)〉", anime.Title.String)
	}

	var (
		currentWorkUpdatedAt time.Time
		activityCount        int
	)
	if err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&currentWorkUpdatedAt); err != nil {
		t.Fatalf("作品の読み込みに失敗: %v", err)
	}
	if !currentWorkUpdatedAt.Equal(workUpdatedAt) {
		t.Errorf("works.updated_at = %v、期待値 = %v のまま", currentWorkUpdatedAt, workUpdatedAt)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`, int64(workID)).Scan(&activityCount); err != nil {
		t.Fatalf("DB活動の読み込みに失敗: %v", err)
	}
	if activityCount != 0 {
		t.Errorf("DB活動 = %d件、期待値 = 0", activityCount)
	}
}

// TestFillEpisodeTitlesUsecase_Execute_LeavesEpisodesWithoutSubTitleは、しょぼいカレンダーに
// その話数のサブタイトルが無いエピソードと、小数の話数のエピソードに触れないことを検証する。
func TestFillEpisodeTitlesUsecase_Execute_LeavesEpisodesWithoutSubTitle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID, tid := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	missingID := insertTitleFillEpisode(t, db, program, 3)
	decimalID := insertTitleFillEpisode(t, db, program, 2.5)

	lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{
		tid: {1: "オペレーション〈梟(ストリクス)〉", 2: "妻役を確保せよ"},
	}}
	output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(ctx, FillEpisodeTitlesInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.FailedEpisodes) != 0 {
		t.Errorf("FailedEpisodes = %+v、期待値 = 空", output.FailedEpisodes)
	}
	for name, id := range map[string]model.EpisodeID{"サブタイトルの無い話数": missingID, "小数の話数": decimalID} {
		if title, _ := readEpisodeTitle(t, db, id); title.Valid {
			t.Errorf("%s: title = %q、期待値 = NULL", name, title.String)
		}
	}
}

// TestFillEpisodeTitlesUsecase_Execute_ReturnsLookupErrorは、しょぼいカレンダーへの問い合わせが
// 失敗したときにエラーを返し、タイトルを書かないことを検証する。
func TestFillEpisodeTitlesUsecase_Execute_ReturnsLookupError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID, _ := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	episodeID := insertTitleFillEpisode(t, db, program, 1)

	lookupErr := errors.New("タイムアウト")
	_, err := newFillEpisodeTitlesUsecase(db, &fakeSubTitleLookup{err: lookupErr}).Execute(ctx, FillEpisodeTitlesInput{Now: time.Now()})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %v をラップしたエラー", err, lookupErr)
	}
	if title, _ := readEpisodeTitle(t, db, episodeID); title.Valid {
		t.Errorf("title = %q、期待値 = NULL", title.String)
	}
}
