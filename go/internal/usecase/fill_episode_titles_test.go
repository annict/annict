package usecase

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"os"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// fakeSubTitleLookupは、しょぼいカレンダーの代わりに決まったサブタイトルを返す。
type fakeSubTitleLookup struct {
	subTitles    map[int32]map[int32]string
	err          error
	tids         []int32
	beforeReturn func()
}

func (f *fakeSubTitleLookup) LookupSubTitles(_ context.Context, tids []int32) (map[int32]map[int32]string, error) {
	f.tids = tids
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
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

// TestFillEpisodeTitlesUsecase_Execute_RejectsChangedTargetは、HTTP待機中に対象の識別値や
// 公開状態が変わっても、古いサブタイトルや編集者のタイトルを上書きしないことを検証する。
func TestFillEpisodeTitlesUsecase_Execute_RejectsChangedTarget(t *testing.T) {
	t.Parallel()

	for _, change := range []struct {
		name   string
		update string
	}{
		{"話数変更", `UPDATE episodes SET raw_number = 2 WHERE id = $1`},
		{"話数なし", `UPDATE episodes SET raw_number = NULL WHERE id = $1`},
		{"非公開化", `UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`},
		{"削除", `UPDATE episodes SET deleted_at = NOW() WHERE id = $1`},
		{"古いエピソード", `UPDATE episodes SET created_at = NOW() - INTERVAL '8 days' WHERE id = $1`},
		{"編集者のタイトル", `UPDATE episodes SET title = '編集者のタイトル' WHERE id = $1`},
		{"TID変更", ""},
		{"作品を非公開化", ""},
		{"作品を削除", ""},
		{"作品の変更", ""},
		{"放送枠の紐づけを解除", ""},
		{"番組の紐づけを解除", ""},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			db := testutil.GetTestDB()
			workID, tid := insertTitleFillWork(t, db)
			var otherWorkID model.WorkID
			if change.name == "作品の変更" {
				otherWorkID = insertCreateTargetWork(t, db, sql.NullInt64{})
			}
			program := insertGenerationProgram(t, db, workID, 1)
			episodeID := insertTitleFillEpisode(t, db, program, 1)
			animeID := insertBareAnime(t, db)
			if _, err := db.Exec(`UPDATE episodes SET anime_id = $2 WHERE id = $1`, int64(episodeID), int64(animeID)); err != nil {
				t.Fatalf("animeの対応付けに失敗: %v", err)
			}
			lookup := &fakeSubTitleLookup{
				subTitles: map[int32]map[int32]string{tid: {1: "問い合わせ前の第1話"}},
				beforeReturn: func() {
					var err error
					switch change.name {
					case "TID変更":
						_, err = db.Exec(`UPDATE works SET sc_tid = $2 WHERE id = $1`, int64(workID), tid+100000)
					case "作品を非公開化":
						_, err = db.Exec(`UPDATE works SET unpublished_at = NOW() WHERE id = $1`, int64(workID))
					case "作品を削除":
						_, err = db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID))
					case "作品の変更":
						_, err = db.Exec(`UPDATE episodes SET work_id = $2 WHERE id = $1`, int64(episodeID), int64(otherWorkID))
					case "放送枠の紐づけを解除":
						_, err = db.Exec(`UPDATE slots SET episode_id = NULL WHERE episode_id = $1`, int64(episodeID))
					case "番組の紐づけを解除":
						_, err = db.Exec(`UPDATE slots SET program_id = NULL WHERE episode_id = $1`, int64(episodeID))
					default:
						_, err = db.Exec(change.update, int64(episodeID))
					}
					if err != nil {
						t.Fatalf("問い合わせ中の変更に失敗: %v", err)
					}
				},
			}
			output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(context.Background(), FillEpisodeTitlesInput{Now: time.Now()})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.FilledCount != 0 || len(output.FailedEpisodes) != 0 {
				t.Errorf("変更後の結果 = %+v、期待値 = 更新・エラーなし", output)
			}
			title, _ := readEpisodeTitle(t, db, episodeID)
			if change.name == "編集者のタイトル" {
				if title.String != "編集者のタイトル" {
					t.Errorf("タイトル = %q、期待値 = 編集者のタイトル", title.String)
				}
			} else if title.Valid {
				t.Errorf("対象が変わったエピソードのタイトル = %q、期待値 = NULL", title.String)
			}
			anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(context.Background(), animeID)
			if err != nil || anime == nil || anime.Title.Valid {
				t.Errorf("anime = %+v、エラー = %v、期待値 = タイトルがNULLのまま", anime, err)
			}
		})
	}
}

// TestFillEpisodeTitlesUsecase_Execute_UsesCurrentAnimeは、HTTP待機中に対応付けや付け替えが
// あっても、現在のanimeだけにタイトルを両書きすることを検証する。
func TestFillEpisodeTitlesUsecase_Execute_UsesCurrentAnime(t *testing.T) {
	t.Parallel()

	for _, remapped := range []bool{false, true} {
		name := "未対応から対応付け"
		if remapped {
			name = "別のanimeへ付け替え"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := testutil.GetTestDB()
			workID, tid := insertTitleFillWork(t, db)
			program := insertGenerationProgram(t, db, workID, 1)
			episodeID := insertTitleFillEpisode(t, db, program, 1)
			currentAnimeID := insertBareAnime(t, db)
			var oldAnimeID model.AnimeID
			if remapped {
				oldAnimeID = insertBareAnime(t, db)
				if _, err := db.Exec(`UPDATE episodes SET anime_id = $2 WHERE id = $1`, int64(episodeID), int64(oldAnimeID)); err != nil {
					t.Fatalf("以前のanimeの対応付けに失敗: %v", err)
				}
			}
			lookup := &fakeSubTitleLookup{
				subTitles: map[int32]map[int32]string{tid: {1: "第1話のサブタイトル"}},
				beforeReturn: func() {
					if _, err := db.Exec(`UPDATE episodes SET anime_id = $2 WHERE id = $1`, int64(episodeID), int64(currentAnimeID)); err != nil {
						t.Fatalf("問い合わせ中のanimeの対応付けに失敗: %v", err)
					}
				},
			}
			output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(context.Background(), FillEpisodeTitlesInput{Now: time.Now()})
			if err != nil || output == nil || output.FilledCount != 1 || len(output.FailedEpisodes) != 0 {
				t.Fatalf("結果 = %+v、エラー = %v、期待値 = 更新1件", output, err)
			}
			title, _ := readEpisodeTitle(t, db, episodeID)
			repo := repository.NewAnimeRepository(query.New(db))
			current, err := repo.GetByID(context.Background(), currentAnimeID)
			if err != nil || current == nil || current.Title.String != "第1話のサブタイトル" || current.Title != title {
				t.Errorf("現在のanime = %+v、エラー = %v、episodeのタイトル = %+v", current, err, title)
			}
			if remapped {
				old, err := repo.GetByID(context.Background(), oldAnimeID)
				if err != nil || old == nil || old.Title.Valid {
					t.Errorf("以前のanime = %+v、エラー = %v、期待値 = タイトルがNULLのまま", old, err)
				}
			}
		})
	}
}

// TestFillEpisodeTitlesUsecase_Execute_ContinuesAfterRollbackは、行ロックによってDB更新を
// 失敗させ、失敗したエピソードは戻し、次のエピソードは更新することを検証する。
func TestFillEpisodeTitlesUsecase_Execute_ContinuesAfterRollback(t *testing.T) {
	t.Parallel()

	for _, lockedTable := range []string{"works", "animes"} {
		name := "作品ロックの取得失敗"
		if lockedTable == "animes" {
			name = "anime側の更新失敗"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := testutil.GetTestDB()
			workID, tid := insertTitleFillWork(t, db)
			program := insertGenerationProgram(t, db, workID, 1)
			failedID := insertTitleFillEpisode(t, db, program, 1)
			goodWorkID, goodTid := insertTitleFillWork(t, db)
			goodProgram := insertGenerationProgram(t, db, goodWorkID, 1)
			goodID := insertTitleFillEpisode(t, db, goodProgram, 2)
			animeID := insertBareAnime(t, db)
			if _, err := db.Exec(`UPDATE episodes SET anime_id = $2 WHERE id = $1`, int64(failedID), int64(animeID)); err != nil {
				t.Fatalf("animeの対応付けに失敗: %v", err)
			}

			// タイムアウトはこのUseCase専用の接続に設定し、共有プールやほかのテストには適用しない。
			timeoutDB, err := sql.Open("postgres", cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres@postgresql:5432/annict_test?sslmode=disable"))
			if err != nil {
				t.Fatalf("タイムアウト付きの接続の作成に失敗: %v", err)
			}
			t.Cleanup(func() { _ = timeoutDB.Close() })
			timeoutDB.SetMaxOpenConns(1)
			if _, err := timeoutDB.Exec(`SET lock_timeout = '100ms'`); err != nil {
				t.Fatalf("lock_timeoutの設定に失敗: %v", err)
			}
			blocker, err := db.Begin()
			if err != nil {
				t.Fatalf("ロック用のトランザクションの開始に失敗: %v", err)
			}
			defer func() { _ = blocker.Rollback() }()
			if lockedTable == "works" {
				_, err = blocker.Exec(`UPDATE works SET updated_at = updated_at WHERE id = $1`, int64(workID))
			} else {
				_, err = blocker.Exec(`UPDATE animes SET title = title WHERE id = $1`, int64(animeID))
			}
			if err != nil {
				t.Fatalf("%sの行のロックに失敗: %v", lockedTable, err)
			}
			lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{tid: {1: "失敗するタイトル"}, goodTid: {2: "後続のタイトル"}}}
			output, err := newFillEpisodeTitlesUsecase(timeoutDB, lookup).Execute(context.Background(), FillEpisodeTitlesInput{Now: time.Now()})
			if err != nil || output == nil || output.FilledCount != 1 || len(output.FailedEpisodes) != 1 {
				t.Fatalf("結果 = %+v、エラー = %v、期待値 = 更新1件・失敗1件", output, err)
			}
			var pgError *pq.Error
			failure := output.FailedEpisodes[0]
			if failure.EpisodeID != failedID || !errors.As(failure.Err, &pgError) || pgError.Code != "55P03" {
				t.Errorf("失敗 = %+v、期待値 = エピソード%dのロックタイムアウト", failure, failedID)
			}
			if title, _ := readEpisodeTitle(t, db, failedID); title.Valid {
				t.Errorf("失敗したエピソードのタイトル = %q、期待値 = ロールバックされてNULL", title.String)
			}
			if title, _ := readEpisodeTitle(t, db, goodID); title.String != "後続のタイトル" {
				t.Errorf("後続のエピソードのタイトル = %q、期待値 = 後続のタイトル", title.String)
			}
			anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(context.Background(), animeID)
			if err != nil || anime == nil || anime.Title.Valid {
				t.Errorf("失敗したanime = %+v、エラー = %v、期待値 = タイトルがNULLのまま", anime, err)
			}
		})
	}
}

// TestFillEpisodeTitlesUsecase_Execute_SkipsLockedEpisodeは、編集中のエピソードを待たずに
// 飛ばし、作品とエピソードのロック順の逆転によるデッドロックを避けることを検証する。
func TestFillEpisodeTitlesUsecase_Execute_SkipsLockedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	workID, tid := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	episodeID := insertTitleFillEpisode(t, db, program, 1)
	blocker, err := db.Begin()
	if err != nil {
		t.Fatalf("ロック用のトランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = blocker.Rollback() }()
	if _, err := blocker.Exec(`UPDATE episodes SET title = title WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("エピソードの行のロックに失敗: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{tid: {1: "サブタイトル"}}}
	output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(ctx, FillEpisodeTitlesInput{Now: time.Now()})
	if err != nil || output == nil || output.FilledCount != 0 || len(output.FailedEpisodes) != 0 {
		t.Errorf("編集中の結果 = %+v、エラー = %v、期待値 = 更新・エラーなし", output, err)
	}
}
