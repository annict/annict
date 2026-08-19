package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// insertEpisodeSyncWork inserts a minimal works row and returns its ID. animeID is
// the works.anime_id mapping column; pass an invalid NullInt64 for an unsynced
// parent.
//
// [Ja] insertEpisodeSyncWork は最小の works 行を挿入し ID を返す。animeID は
// works.anime_id マッピングカラムで、未同期の親には無効な NullInt64 を渡す。
func insertEpisodeSyncWork(t *testing.T, tx *sql.Tx, animeID sql.NullInt64) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(
		`INSERT INTO works (title, media, anime_id) VALUES ($1, $2, $3) RETURNING id`,
		"親作品", 1, animeID,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// insertEpisodeSyncParentAnime inserts a bare anime row to stand in for a synced
// parent work's anime and returns its ID.
//
// [Ja] insertEpisodeSyncParentAnime は同期済みの親作品の anime に見立てた素の anime 行を
// 挿入し ID を返す。
func insertEpisodeSyncParentAnime(t *testing.T, tx *sql.Tx) model.AnimeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("animes の挿入に失敗: %v", err)
	}
	return model.AnimeID(id)
}

// episodeSyncRow holds the episodes columns relevant to the episodes -> animes sync.
//
// [Ja] episodeSyncRow は episodes -> animes 同期に関係する episodes カラムを保持する。
type episodeSyncRow struct {
	workID         model.WorkID
	title          sql.NullString
	titleRo        string
	titleEn        string
	number         sql.NullString
	sortNumber     int32
	rawNumber      sql.NullFloat64
	status         string
	archiveMessage sql.NullString
	animeID        sql.NullInt64
}

func insertEpisodeSyncEpisode(t *testing.T, tx *sql.Tx, in episodeSyncRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, title, title_ro, title_en, number, sort_number,
			raw_number, status, archive_message, anime_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		int64(in.workID), in.title, in.titleRo, in.titleEn, in.number, in.sortNumber,
		in.rawNumber, in.status, in.archiveMessage, in.animeID,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// dbListEpisodeRow holds the episodes columns the Annict DB screens read.
//
// The status field mirrors the dormant episodes.status column. Test cases
// deliberately make it disagree with unpublished_at / deleted_at to verify that
// the filters and the derived status use those timestamps instead.
//
// [Ja] dbListEpisodeRow は Annict DB の画面が読む episodes カラムを保持する。
//
// status フィールドは休眠中の episodes.status カラムを写す。テストケースでは
// unpublished_at / deleted_at と意図的に食い違わせ、各絞り込みと状態導出が
// これらのタイムスタンプを使うことを検証する。
type dbListEpisodeRow struct {
	workID              model.WorkID
	number              sql.NullString
	rawNumber           sql.NullFloat64
	sortNumber          int32
	title               sql.NullString
	titleRo             string
	titleEn             string
	episodeRecordsCount int32
	status              string
	unpublishedAt       sql.NullTime
	deletedAt           sql.NullTime
}

func insertDBListEpisode(t *testing.T, tx *sql.Tx, in dbListEpisodeRow) model.EpisodeID {
	t.Helper()
	var id int64
	// created_at / updated_at are written the way Rails writes them, so a fixture row
	// carries the version the edit form reads.
	//
	// [Ja] created_at / updated_at は Rails が書くのと同じように入れる。フィクスチャの行が
	// 編集フォームの読む版を持つようにするため。
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, number, raw_number, sort_number, title, title_ro, title_en,
			episode_records_count, status, unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()) RETURNING id`,
		int64(in.workID), in.number, in.rawNumber, in.sortNumber, in.title, in.titleRo, in.titleEn,
		in.episodeRecordsCount, in.status, in.unpublishedAt, in.deletedAt,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertDBListWork inserts a minimal works row to own the listed episodes.
//
// [Ja] insertDBListWork は一覧対象のエピソードを持たせる最小の works 行を挿入する。
func insertDBListWork(t *testing.T, tx *sql.Tx) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(
		`INSERT INTO works (title, media) VALUES ($1, $2) RETURNING id`,
		"一覧対象の作品", 1,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

func TestEpisodeRepository_ListForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 作品のエピソードを sort_number 降順で取得する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:              workID,
			number:              sql.NullString{String: "第1話", Valid: true},
			rawNumber:           sql.NullFloat64{Float64: 1, Valid: true},
			sortNumber:          1,
			title:               sql.NullString{String: "はじまり", Valid: true},
			titleRo:             "Hajimari",
			titleEn:             "The Beginning",
			episodeRecordsCount: 42,
			status:              "published",
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			sortNumber: 2,
			status:     "published",
		})

		// Another work's episode must not leak into the list.
		//
		// [Ja] 別作品のエピソードが一覧に混ざらないこと。
		otherWorkID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     otherWorkID,
			number:     sql.NullString{String: "別作品の第1話", Valid: true},
			sortNumber: 1,
			status:     "published",
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, want 2", len(got))
		}
		if got[0].Number == nil || *got[0].Number != "第2話" {
			t.Errorf("got[0].Number = %v, want 第2話 (sort_number 降順)", got[0].Number)
		}
		// The second episode leaves title and raw_number NULL, so they must map to nil
		// rather than to a pointer holding the zero value.
		//
		// [Ja] 第 2 話は title と raw_number を NULL のままにしているため、ゼロ値を指す
		// ポインタではなく nil に写像されること。
		if got[0].Title != nil {
			t.Errorf("got[0].Title = %v, want nil", got[0].Title)
		}
		if got[0].RawNumber != nil {
			t.Errorf("got[0].RawNumber = %v, want nil", got[0].RawNumber)
		}

		second := got[1]
		if second.Number == nil || *second.Number != "第1話" {
			t.Errorf("got[1].Number = %v, want 第1話", second.Number)
		}
		if second.WorkID != workID {
			t.Errorf("got[1].WorkID = %d, want %d", second.WorkID, workID)
		}
		if second.Title == nil || *second.Title != "はじまり" {
			t.Errorf("got[1].Title = %v, want はじまり", second.Title)
		}
		if second.TitleRo != "Hajimari" {
			t.Errorf("got[1].TitleRo = %q, want Hajimari", second.TitleRo)
		}
		if second.TitleEn != "The Beginning" {
			t.Errorf("got[1].TitleEn = %q, want The Beginning", second.TitleEn)
		}
		if second.RawNumber == nil || *second.RawNumber != 1 {
			t.Errorf("got[1].RawNumber = %v, want 1", second.RawNumber)
		}
		if second.SortNumber != 1 {
			t.Errorf("got[1].SortNumber = %d, want 1", second.SortNumber)
		}
		if second.EpisodeRecordsCount != 42 {
			t.Errorf("got[1].EpisodeRecordsCount = %d, want 42", second.EpisodeRecordsCount)
		}
		if second.DerivedStatus() != model.EpisodeStatusPublished {
			t.Errorf("got[1].DerivedStatus() = %q, want published", second.DerivedStatus())
		}
	})

	t.Run("正常系: ページ単位で取得する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		for i := int32(1); i <= 3; i++ {
			insertDBListEpisode(t, tx, dbListEpisodeRow{
				workID:     workID,
				number:     sql.NullString{String: "第" + strconv.Itoa(int(i)) + "話", Valid: true},
				sortNumber: i,
				status:     "published",
			})
		}

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d, want 2", len(firstPage))
		}
		if *firstPage[0].Number != "第3話" || *firstPage[1].Number != "第2話" {
			t.Errorf("firstPage = [%q %q], want [第3話 第2話]", *firstPage[0].Number, *firstPage[1].Number)
		}

		secondPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    2,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(secondPage) != 1 {
			t.Fatalf("len(secondPage) = %d, want 1", len(secondPage))
		}
		if *secondPage[0].Number != "第1話" {
			t.Errorf("secondPage[0].Number = %q, want 第1話", *secondPage[0].Number)
		}
	})

	t.Run("正常系: sort_number が同じ行は id 降順でページ境界をまたいで取得する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		oldestID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
			status:     "published",
		})
		middleID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
			status:     "published",
		})
		newestID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
			status:     "published",
		})

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB() first page error = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d, want 2", len(firstPage))
		}
		wantFirstPage := []model.EpisodeID{newestID, middleID}
		for i, wantID := range wantFirstPage {
			if firstPage[i].ID != wantID {
				t.Errorf("firstPage[%d].ID = %d, want %d", i, firstPage[i].ID, wantID)
			}
		}

		secondPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    2,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB() second page error = %v", err)
		}
		if len(secondPage) != 1 {
			t.Fatalf("len(secondPage) = %d, want 1", len(secondPage))
		}
		if secondPage[0].ID != oldestID {
			t.Errorf("secondPage[0].ID = %d, want %d", secondPage[0].ID, oldestID)
		}
	})

	t.Run("正常系: 除外と状態は deleted_at / unpublished_at で決まり休眠 status は読まない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "公開中の話", Valid: true},
			sortNumber: 1,
			status:     "published",
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:        workID,
			number:        sql.NullString{String: "非公開の話", Valid: true},
			sortNumber:    2,
			status:        "published",
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "休眠 status だけが deleted の話", Valid: true},
			sortNumber: 3,
			status:     "deleted",
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "deleted_at で削除された話", Valid: true},
			sortNumber: 4,
			status:     "published",
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("len(got) = %d, want 3 (deleted_at の行だけを除外)", len(got))
		}

		wantNumbers := []string{"休眠 status だけが deleted の話", "非公開の話", "公開中の話"}
		for i, want := range wantNumbers {
			if *got[i].Number != want {
				t.Errorf("got[%d].Number = %q, want %q", i, *got[i].Number, want)
			}
		}

		wantStatuses := []model.EpisodeStatus{
			model.EpisodeStatusPublished,
			model.EpisodeStatusArchived,
			model.EpisodeStatusPublished,
		}
		for i, want := range wantStatuses {
			if got[i].DerivedStatus() != want {
				t.Errorf("got[%d].DerivedStatus() = %q, want %q", i, got[i].DerivedStatus(), want)
			}
		}
	})

	t.Run("正常系: 直前のエピソードを sort_number 順の隣接行から導出する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第1話", Valid: true},
			rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
			sortNumber: 100,
			status:     "published",
		})
		// The deleted episode sits between the other two in sort_number order. The list
		// drops it, so the derivation must skip it too instead of naming a row the page
		// does not show.
		//
		// [Ja] 削除済みのエピソードは sort_number 順で他 2 話の間に位置する。一覧はこれを
		// 落とすため、導出も飛ばし、ページに出ない行を名指ししないこと。
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "削除済みの話", Valid: true},
			sortNumber: 150,
			status:     "published",
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			sortNumber: 200,
			status:     "published",
		})
		// Another work's episode must not become anyone's preceding episode.
		//
		// [Ja] 別作品のエピソードが直前のエピソードになってはならない。
		otherWorkID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     otherWorkID,
			number:     sql.NullString{String: "別作品の話", Valid: true},
			sortNumber: 50,
			status:     "published",
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, want 2", len(got))
		}
		if got[0].PrevNumber == nil || *got[0].PrevNumber != "第1話" {
			t.Errorf("got[0].PrevNumber = %v, want 第1話", got[0].PrevNumber)
		}
		if got[0].PrevRawNumber == nil || *got[0].PrevRawNumber != 1 {
			t.Errorf("got[0].PrevRawNumber = %v, want 1", got[0].PrevRawNumber)
		}
		if got[1].PrevNumber != nil {
			t.Errorf("got[1].PrevNumber = %v, want nil (作品の最初のエピソード)", got[1].PrevNumber)
		}
		if got[1].PrevRawNumber != nil {
			t.Errorf("got[1].PrevRawNumber = %v, want nil (作品の最初のエピソード)", got[1].PrevRawNumber)
		}
	})

	// The derivation runs over the work's whole list before the page is cut out, so the last
	// row of a page still names the episode that lands on the next page.
	//
	// [Ja] 導出はページを切り出す前に作品の一覧全体に対して行われるため、ページ末尾の行も
	// 次ページに載るエピソードを名指しできる。
	t.Run("正常系: ページ境界でも直前のエピソードを導出する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		for i := int32(1); i <= 3; i++ {
			insertDBListEpisode(t, tx, dbListEpisodeRow{
				workID:     workID,
				number:     sql.NullString{String: "第" + strconv.Itoa(int(i)) + "話", Valid: true},
				sortNumber: i * 100,
				status:     "published",
			})
		}

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d, want 2", len(firstPage))
		}
		// The second episode closes the first page while its predecessor, the first
		// episode, opens the second page.
		//
		// [Ja] 第2話は 1 ページ目の末尾で、その直前の第1話は 2 ページ目の先頭に載る。
		if firstPage[1].PrevNumber == nil || *firstPage[1].PrevNumber != "第1話" {
			t.Errorf("firstPage[1].PrevNumber = %v, want 第1話", firstPage[1].PrevNumber)
		}
	})

	t.Run("正常系: エピソードが無い作品では空スライスを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d, want 0", len(got))
		}
	})

	t.Run("境界値: 最大ページ番号でも OFFSET がオーバーフローしない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    math.MaxInt32,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d, want 0", len(got))
		}
	})
}

func TestEpisodeRepository_CountForDB(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

	workID := insertDBListWork(t, tx)
	insertDBListEpisode(t, tx, dbListEpisodeRow{workID: workID, sortNumber: 1, status: "published"})
	insertDBListEpisode(t, tx, dbListEpisodeRow{
		workID:        workID,
		sortNumber:    2,
		status:        "published",
		unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	// The count must use the same filter as the list: the deleted_at row and another
	// work's episode are left out, while the row whose dormant status alone says
	// deleted is still counted.
	//
	// [Ja] 件数は一覧と同じ絞り込みを使うため、deleted_at の行と別作品のエピソードは
	// 数えず、休眠 status だけが deleted の行は数える。
	insertDBListEpisode(t, tx, dbListEpisodeRow{workID: workID, sortNumber: 3, status: "deleted"})
	insertDBListEpisode(t, tx, dbListEpisodeRow{
		workID:     workID,
		sortNumber: 4,
		status:     "published",
		deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
	otherWorkID := insertDBListWork(t, tx)
	insertDBListEpisode(t, tx, dbListEpisodeRow{workID: otherWorkID, sortNumber: 1, status: "published"})

	got, err := repo.CountForDB(context.Background(), workID)
	if err != nil {
		t.Fatalf("CountForDB() error = %v", err)
	}
	if got != 3 {
		t.Errorf("CountForDB() = %d, want 3", got)
	}
}

func TestEpisodeRepository_GetForEditByID(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 編集対象のカラムと親作品を射影する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("編集対象の作品").WithNoEpisodes(true).Build()
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			rawNumber:  sql.NullFloat64{Float64: 2.5, Valid: true},
			sortNumber: 200,
			title:      sql.NullString{String: "もう、お婿にいけません", Valid: true},
			titleEn:    "No Longer Marriageable",
			status:     "published",
		})

		got, err := repo.GetForEditByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForEditByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForEditByID() = nil, want エピソード")
		}
		if got.Episode.ID != episodeID {
			t.Errorf("Episode.ID = %d, want %d", int64(got.Episode.ID), int64(episodeID))
		}
		if got.Episode.Number == nil || *got.Episode.Number != "第2話" {
			t.Errorf("Episode.Number = %v, want %q", got.Episode.Number, "第2話")
		}
		if got.Episode.RawNumber == nil || *got.Episode.RawNumber != 2.5 {
			t.Errorf("Episode.RawNumber = %v, want 2.5", got.Episode.RawNumber)
		}
		if got.Episode.SortNumber != 200 {
			t.Errorf("Episode.SortNumber = %d, want 200", got.Episode.SortNumber)
		}
		if got.Episode.Title == nil || *got.Episode.Title != "もう、お婿にいけません" {
			t.Errorf("Episode.Title = %v, want %q", got.Episode.Title, "もう、お婿にいけません")
		}
		if got.Episode.TitleEn != "No Longer Marriageable" {
			t.Errorf("Episode.TitleEn = %q, want %q", got.Episode.TitleEn, "No Longer Marriageable")
		}
		// The form carries updated_at as the version its submit is made against, so the
		// loader has to populate it.
		//
		// [Ja] フォームは updated_at を送信が前提とする版として運ぶため、ローダーが値を
		// 入れる必要がある。
		if got.Episode.UpdatedAt == nil {
			t.Error("Episode.UpdatedAt = nil, want 更新時刻")
		}
		// The page reads the work's title for its heading and no_episodes for the shared
		// subnav, so both ride along with the episode.
		//
		// [Ja] ページは見出しに作品の title を、共有サブナビに no_episodes を読むため、
		// どちらもエピソードと一緒に返る。
		if got.Work.ID != workID {
			t.Errorf("Work.ID = %d, want %d", int64(got.Work.ID), int64(workID))
		}
		if got.Work.Title != "編集対象の作品" {
			t.Errorf("Work.Title = %q, want %q", got.Work.Title, "編集対象の作品")
		}
		if !got.Work.NoEpisodes {
			t.Error("Work.NoEpisodes = false, want true")
		}
	})

	t.Run("正常系: 未設定の任意カラムは nil になる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("未設定カラムの作品").Build()
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{workID: workID, sortNumber: 100, status: "published"})

		got, err := repo.GetForEditByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForEditByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForEditByID() = nil, want エピソード")
		}
		if got.Episode.Number != nil {
			t.Errorf("Episode.Number = %v, want nil", got.Episode.Number)
		}
		if got.Episode.RawNumber != nil {
			t.Errorf("Episode.RawNumber = %v, want nil", got.Episode.RawNumber)
		}
		if got.Episode.Title != nil {
			t.Errorf("Episode.Title = %v, want nil", got.Episode.Title)
		}
	})

	t.Run("異常系: 編集できないエピソードは (nil, nil) を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("除外テストの作品").Build()
		deletedEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 100,
			status:     "published",
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})
		// The dormant status column is deliberately made to disagree with the timestamps:
		// a row it alone calls deleted is still editable.
		//
		// [Ja] 休眠 status カラムは意図的にタイムスタンプと食い違わせる。status だけが
		// deleted の行は編集できる。
		dormantDeletedEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 200,
			status:     "deleted",
		})

		deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済みの作品").WithDeletedAt(time.Now()).Build()
		episodeOfDeletedWorkID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     deletedWorkID,
			sortNumber: 100,
			status:     "published",
		})

		for _, tt := range []struct {
			name      string
			episodeID model.EpisodeID
			wantNil   bool
		}{
			{name: "存在しないエピソード", episodeID: model.EpisodeID(999999999), wantNil: true},
			{name: "削除済みのエピソード", episodeID: deletedEpisodeID, wantNil: true},
			{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID, wantNil: true},
			{name: "休眠 status だけが deleted のエピソード", episodeID: dormantDeletedEpisodeID, wantNil: false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.GetForEditByID(context.Background(), tt.episodeID)
				if err != nil {
					t.Fatalf("GetForEditByID() error = %v", err)
				}
				if tt.wantNil && got != nil {
					t.Errorf("GetForEditByID() = %+v, want nil", got)
				}
				if !tt.wantNil && got == nil {
					t.Error("GetForEditByID() = nil, want エピソード")
				}
			})
		}
	})
}

func TestEpisodeRepository_ListForAnimeSyncByIDs(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 全カラムを射影し親 anime_id を JOIN で解決する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})

		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:     workID,
			title:      sql.NullString{String: "第3話タイトル", Valid: true},
			titleRo:    "Episode 3",
			titleEn:    "Episode Three",
			number:     sql.NullString{String: "第3話", Valid: true},
			sortNumber: 3,
			rawNumber:  sql.NullFloat64{Float64: 3.5, Valid: true},
			status:     "published",
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d, want 1", len(episodes))
		}
		e := episodes[0]

		if e.ID != episodeID {
			t.Errorf("ID = %d, want %d", e.ID, episodeID)
		}
		if e.WorkID != workID {
			t.Errorf("WorkID = %d, want %d", e.WorkID, workID)
		}
		if e.Title == nil || *e.Title != "第3話タイトル" {
			t.Errorf("Title = %v, want 第3話タイトル", e.Title)
		}
		if e.TitleRo != "Episode 3" {
			t.Errorf("TitleRo = %q, want Episode 3", e.TitleRo)
		}
		if e.TitleEn != "Episode Three" {
			t.Errorf("TitleEn = %q, want Episode Three", e.TitleEn)
		}
		if e.Number == nil || *e.Number != "第3話" {
			t.Errorf("Number = %v, want 第3話", e.Number)
		}
		if e.SortNumber != 3 {
			t.Errorf("SortNumber = %d, want 3", e.SortNumber)
		}
		if e.RawNumber == nil || *e.RawNumber != 3.5 {
			t.Errorf("RawNumber = %v, want 3.5", e.RawNumber)
		}
		if e.Status != model.EpisodeStatusPublished {
			t.Errorf("Status = %q, want published", e.Status)
		}
		if e.ParentAnimeID == nil || *e.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v, want %d", e.ParentAnimeID, parentAnimeID)
		}
		// The episode itself is not yet mapped to an anime.
		//
		// [Ja] episode 自体はまだ anime にマッピングされていない。
		if e.AnimeID != nil {
			t.Errorf("AnimeID = %v, want nil", e.AnimeID)
		}
	})

	t.Run("正常系: NULL 許容カラムは nil、未同期の親は ParentAnimeID nil", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		// Parent work has no anime_id (unsynced), and the episode leaves the
		// nullable title / number / raw_number columns NULL.
		//
		// [Ja] 親作品は anime_id を持たず (未同期)、episode は NULL 許容の
		// title / number / raw_number カラムを NULL のままにする。
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{})
		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:     workID,
			sortNumber: 1,
			status:     "published",
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d, want 1", len(episodes))
		}
		e := episodes[0]

		if e.Title != nil {
			t.Errorf("Title = %v, want nil", e.Title)
		}
		if e.Number != nil {
			t.Errorf("Number = %v, want nil", e.Number)
		}
		if e.RawNumber != nil {
			t.Errorf("RawNumber = %v, want nil", e.RawNumber)
		}
		if e.ParentAnimeID != nil {
			t.Errorf("ParentAnimeID = %v, want nil (unsynced parent)", e.ParentAnimeID)
		}
	})

	t.Run("正常系: 空入力はクエリせず空スライスを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
		}
		if len(episodes) != 0 {
			t.Errorf("len(episodes) = %d, want 0", len(episodes))
		}
	})
}

// TestEpisodeRepository_Create covers the insert the bulk create writes each row with: the
// optional columns of a partly filled row become NULL rather than empty values, and the two
// columns the create fills in itself (the anime mapping and the preceding episode) are stored
// with the row.
//
// [Ja] TestEpisodeRepository_Create は一括作成が行ごとに書く INSERT を検証する。一部だけ入力
// された行の任意カラムは空の値ではなく NULL になり、作成が自ら埋める 2 つのカラム (anime の
// マッピングと直前のエピソード) が行と一緒に保存される。
func TestEpisodeRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
	animeID := insertEpisodeSyncParentAnime(t, tx)
	workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(animeID), Valid: true})

	number := "#1"
	rawNumber := 1.5
	title := "はじまり"
	firstID, err := repo.Create(ctx, repository.CreateEpisodeParams{
		UserID:     userID,
		WorkID:     workID,
		Number:     &number,
		RawNumber:  &rawNumber,
		Title:      &title,
		SortNumber: 100,
		AnimeID:    &animeID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	secondID, err := repo.Create(ctx, repository.CreateEpisodeParams{
		UserID:        userID,
		WorkID:        workID,
		Title:         &title,
		SortNumber:    200,
		PrevEpisodeID: &firstID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var (
		gotNumber        sql.NullString
		gotRawNumber     sql.NullFloat64
		gotTitle         sql.NullString
		gotSortNumber    int32
		gotPrevEpisodeID sql.NullInt64
		gotAnimeID       sql.NullInt64
	)
	if err := tx.QueryRow(`
		SELECT number, raw_number, title, sort_number, prev_episode_id, anime_id
		FROM episodes
		WHERE id = $1
	`, int64(firstID)).Scan(&gotNumber, &gotRawNumber, &gotTitle, &gotSortNumber, &gotPrevEpisodeID, &gotAnimeID); err != nil {
		t.Fatalf("作成されたエピソードの読み込みに失敗: %v", err)
	}

	if gotNumber.String != number || gotTitle.String != title {
		t.Errorf("(number, title) = (%q, %q), want (%q, %q)", gotNumber.String, gotTitle.String, number, title)
	}
	if gotRawNumber.Float64 != rawNumber {
		t.Errorf("raw_number = %v, want %v", gotRawNumber, rawNumber)
	}
	if gotSortNumber != 100 {
		t.Errorf("sort_number = %d, want 100", gotSortNumber)
	}
	if gotAnimeID.Int64 != int64(animeID) {
		t.Errorf("anime_id = %+v, want %d", gotAnimeID, int64(animeID))
	}
	// The first row of a work has no preceding episode.
	//
	// [Ja] 作品の最初の行には直前のエピソードが無い。
	if gotPrevEpisodeID.Valid {
		t.Errorf("prev_episode_id = %+v, want NULL", gotPrevEpisodeID)
	}

	if err := tx.QueryRow(`
		SELECT number, raw_number, prev_episode_id, anime_id
		FROM episodes
		WHERE id = $1
	`, int64(secondID)).Scan(&gotNumber, &gotRawNumber, &gotPrevEpisodeID, &gotAnimeID); err != nil {
		t.Fatalf("作成されたエピソードの読み込みに失敗: %v", err)
	}

	// The columns the row left empty are stored as NULL: no existing episode carries an
	// empty string in them.
	//
	// [Ja] 行が空のままにしたカラムは NULL として保存される。既存のどのエピソードもそれらに
	// 空文字列を持たないため。
	if gotNumber.Valid {
		t.Errorf("number = %+v, want NULL", gotNumber)
	}
	if gotRawNumber.Valid {
		t.Errorf("raw_number = %+v, want NULL", gotRawNumber)
	}
	// An episode under a work that is not mapped yet carries no anime.
	//
	// [Ja] 未マッピングの作品配下のエピソードは anime を持たない。
	if gotAnimeID.Valid {
		t.Errorf("anime_id = %+v, want NULL", gotAnimeID)
	}
	if gotPrevEpisodeID.Int64 != int64(firstID) {
		t.Errorf("prev_episode_id = %+v, want %d", gotPrevEpisodeID, int64(firstID))
	}

	var (
		activityUserID           int64
		activityTrackableID      int64
		activityTrackableType    string
		activityAction           string
		activityRootResourceID   int64
		activityRootResourceType string
		activityNewID            string
	)
	if err := tx.QueryRow(`
		SELECT
			user_id,
			trackable_id,
			trackable_type,
			action,
			root_resource_id,
			root_resource_type,
			parameters->'new'->>'id'
		FROM db_activities
		WHERE trackable_id = $1
			AND trackable_type = 'Episode'
	`, int64(secondID)).Scan(
		&activityUserID, &activityTrackableID, &activityTrackableType, &activityAction,
		&activityRootResourceID, &activityRootResourceType, &activityNewID,
	); err != nil {
		t.Fatalf("DB 活動履歴の読み込みに失敗: %v", err)
	}
	if activityUserID != int64(userID) || activityTrackableID != int64(secondID) ||
		activityTrackableType != "Episode" || activityAction != "episodes.create" ||
		activityRootResourceID != int64(workID) || activityRootResourceType != "Work" ||
		activityNewID != secondID.String() {
		t.Errorf("DB 活動履歴が作成内容と一致しません")
	}
}

func TestEpisodeRepository_UpdateAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

	parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
	workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})
	episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
		workID:     workID,
		sortNumber: 1,
		status:     "published",
	})

	// The anime the episode gets mapped to (its own identity row).
	//
	// [Ja] episode がマッピングされる anime (episode 自身の同一性の行)。
	episodeAnimeID := insertEpisodeSyncParentAnime(t, tx)

	if err := repo.UpdateAnimeID(context.Background(), episodeID, episodeAnimeID); err != nil {
		t.Fatalf("UpdateAnimeID() error = %v", err)
	}

	episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
	if err != nil {
		t.Fatalf("ListForAnimeSyncByIDs() error = %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("len(episodes) = %d, want 1", len(episodes))
	}
	if episodes[0].AnimeID == nil || *episodes[0].AnimeID != episodeAnimeID {
		t.Errorf("AnimeID = %v, want %d", episodes[0].AnimeID, episodeAnimeID)
	}
}

// TestEpisodeRepository_ListIDsAfter verifies keyset pagination. As with the works
// variant, other tests commit episodes to the shared test DB, so it asserts the
// keyset invariants that hold regardless of foreign rows (first id strictly greater
// than the cursor, ascending order, the limit, strict cursor advancement) rather
// than exact page content.
//
// [Ja] TestEpisodeRepository_ListIDsAfter は keyset ページネーションを検証する。works 版と
// 同様、他テストが共有テスト DB に episodes をコミットするため、ページ内容の厳密一致では
// なく、他行の有無に依らず成立する keyset の不変条件 (カーソルより厳密に大きい最初の id・
// 昇順・LIMIT・カーソルの厳密前進) を検証する。
func TestEpisodeRepository_ListIDsAfter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{})
	// Three episodes in ascending id order; the middle one only needs to exist so
	// the first page (limit 2) is full and id3 stays ahead of the second-page cursor.
	//
	// [Ja] id 昇順の 3 件。中間の 1 件は、最初のページ (limit 2) が満杯になり id3 が
	// 2 ページ目のカーソルより先に残るために存在させるだけ。
	id1 := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 1, status: "published"})
	insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 2, status: "published"})
	id3 := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 3, status: "published"})

	t.Run("カーソル直後の id を LIMIT どおり 1 件返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 1)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d, want %d", got[0], id1)
		}
	})

	t.Run("昇順かつカーソルより大きい id だけを LIMIT 件数まで返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d, want %d", got[0], id1)
		}
		if got[0] >= got[1] {
			t.Errorf("not ascending: %v", got)
		}
	})

	t.Run("カーソルを進めると重複なく前進する", func(t *testing.T) {
		page1, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		cursor := page1[len(page1)-1]

		page2, err := repo.ListIDsAfter(ctx, cursor, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(page2) == 0 {
			t.Fatal("page2 is empty, want at least id3")
		}
		if page2[0] <= cursor {
			t.Errorf("page2[0] = %d, want > cursor %d", page2[0], cursor)
		}
	})

	t.Run("全 id より大きいカーソルでは空を返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id3+1_000_000_000, 10)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
	})
}

// dbUpdateEpisodeRow holds the episodes columns an update test starts from: what the form
// edits, the two mapping columns, and the state timestamps the update has to leave alone.
//
// [Ja] dbUpdateEpisodeRow は更新テストの出発点となる episodes カラムを保持する。フォームが
// 編集するもの、2 つのマッピングカラム、および更新が触れてはならない状態のタイムスタンプ。
type dbUpdateEpisodeRow struct {
	workID        model.WorkID
	number        sql.NullString
	rawNumber     sql.NullFloat64
	sortNumber    int32
	title         sql.NullString
	titleRo       string
	titleEn       string
	animeID       sql.NullInt64
	prevEpisodeID sql.NullInt64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
	// nullUpdatedAt leaves updated_at NULL, which is the version an episode written before
	// the column was populated carries.
	//
	// [Ja] nullUpdatedAt は updated_at を NULL のままにする。カラムが埋まる前に書かれた
	// エピソードが持つ版がこれにあたる。
	nullUpdatedAt bool
}

// insertDBUpdateEpisode inserts the episode an update test edits. Its timestamps are taken from
// the database an hour back rather than from the Go clock: NOW() is the transaction's start
// time, so a fixture stamped by the test process could sit after the value the update writes
// and the version would appear to move backwards.
//
// [Ja] insertDBUpdateEpisode は更新テストが編集するエピソードを挿入する。タイムスタンプは Go の
// 時計ではなく DB の 1 時間前を使う。NOW() はトランザクション開始時刻のため、テストプロセスが
// 打刻したフィクスチャは更新が書く値より後になりえ、版が巻き戻ったように見えてしまう。
func insertDBUpdateEpisode(t *testing.T, tx *sql.Tx, in dbUpdateEpisodeRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, number, raw_number, sort_number, title, title_ro, title_en,
			anime_id, prev_episode_id, unpublished_at, deleted_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			NOW() - INTERVAL '1 hour',
			CASE WHEN $12::boolean THEN NULL ELSE NOW() - INTERVAL '1 hour' END
		) RETURNING id`,
		int64(in.workID), in.number, in.rawNumber, in.sortNumber, in.title, in.titleRo, in.titleEn,
		in.animeID, in.prevEpisodeID, in.unpublishedAt, in.deletedAt, in.nullUpdatedAt,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertDBUpdateWork inserts the parent work of an update test with its updated_at set back a
// day, so a touch performed by the update is observable.
//
// [Ja] insertDBUpdateWork は更新テストの親作品を、updated_at を 1 日前にして挿入する。更新が
// 行う touch を観測できるようにするため。
func insertDBUpdateWork(t *testing.T, tx *sql.Tx, animeID sql.NullInt64) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO works (title, media, anime_id, created_at, updated_at)
		VALUES ($1, 1, $2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day') RETURNING id`,
		"更新対象の作品", animeID,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// storedDBUpdateEpisode is what the assertions read back after an update: the columns the form
// writes, the recomputed navigation column, and the version the write advanced.
//
// [Ja] storedDBUpdateEpisode は更新後にアサーションが読み戻す内容。フォームが書くカラム、
// 再計算された導線のカラム、書き込みが進めた版。
type storedDBUpdateEpisode struct {
	number        sql.NullString
	rawNumber     sql.NullFloat64
	sortNumber    int32
	title         sql.NullString
	titleEn       string
	prevEpisodeID sql.NullInt64
	updatedAt     sql.NullTime
}

func readDBUpdateEpisode(t *testing.T, tx *sql.Tx, id model.EpisodeID) storedDBUpdateEpisode {
	t.Helper()
	var row storedDBUpdateEpisode
	if err := tx.QueryRow(`
		SELECT number, raw_number, sort_number, title, title_en, prev_episode_id, updated_at
		FROM episodes WHERE id = $1`, int64(id),
	).Scan(&row.number, &row.rawNumber, &row.sortNumber, &row.title, &row.titleEn, &row.prevEpisodeID, &row.updatedAt); err != nil {
		t.Fatalf("更新後のエピソードの読み込みに失敗: %v", err)
	}
	return row
}

// readDBUpdateEpisodeVersion returns the stored version of an episode, which a submit has to
// state to be accepted.
//
// [Ja] readDBUpdateEpisodeVersion はエピソードの保存済みの版を返す。送信が受理されるには、
// この版を名乗る必要がある。
func readDBUpdateEpisodeVersion(t *testing.T, tx *sql.Tx, id model.EpisodeID) *time.Time {
	t.Helper()
	stored := readDBUpdateEpisode(t, tx, id)
	if !stored.updatedAt.Valid {
		return nil
	}
	return &stored.updatedAt.Time
}

// countDBEpisodeUpdateActivities counts the change-history rows the Rails admin reads for an
// episode.
//
// [Ja] countDBEpisodeUpdateActivities は、あるエピソードについて Rails 側の管理画面が変更履歴と
// して読む行を数える。
func countDBEpisodeUpdateActivities(t *testing.T, tx *sql.Tx, id model.EpisodeID) int {
	t.Helper()
	var count int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM db_activities
		WHERE trackable_type = 'Episode' AND trackable_id = $1 AND action = 'episodes.update'`,
		int64(id),
	).Scan(&count); err != nil {
		t.Fatalf("DB 活動履歴の読み込みに失敗: %v", err)
	}
	return count
}

func TestEpisodeRepository_GetForUpdateByID(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 送信された値が運ばないカラムだけを射影する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
		episodeAnimeID := insertEpisodeSyncParentAnime(t, tx)
		workID := insertDBUpdateWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})
		unpublishedAt := time.Now()
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:        workID,
			sortNumber:    100,
			titleRo:       "Episode 1",
			animeID:       sql.NullInt64{Int64: int64(episodeAnimeID), Valid: true},
			unpublishedAt: sql.NullTime{Time: unpublishedAt, Valid: true},
		})

		got, err := repo.GetForUpdateByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForUpdateByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForUpdateByID() = nil, want エピソード")
		}
		if got.ID != episodeID || got.WorkID != workID {
			t.Errorf("(ID, WorkID) = (%d, %d), want (%d, %d)", int64(got.ID), int64(got.WorkID), int64(episodeID), int64(workID))
		}
		if got.TitleRo != "Episode 1" {
			t.Errorf("TitleRo = %q, want %q", got.TitleRo, "Episode 1")
		}
		// The state timestamps ride along because the anime dual-write derives anime.status
		// from them; without them a content edit would republish an archived anime.
		//
		// [Ja] 状態のタイムスタンプが一緒に返るのは、anime への両書きがこれらから anime.status を
		// 導出するため。これが無いと内容編集がアーカイブ済みの anime を再公開してしまう。
		if got.UnpublishedAt == nil {
			t.Error("UnpublishedAt = nil, want 非公開時刻")
		}
		if got.DerivedStatus() != model.EpisodeStatusArchived {
			t.Errorf("DerivedStatus() = %q, want %q", got.DerivedStatus(), model.EpisodeStatusArchived)
		}
		if got.AnimeID == nil || *got.AnimeID != episodeAnimeID {
			t.Errorf("AnimeID = %v, want %d", got.AnimeID, int64(episodeAnimeID))
		}
		if got.ParentAnimeID == nil || *got.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v, want %d", got.ParentAnimeID, int64(parentAnimeID))
		}
	})

	t.Run("正常系: 未マッピングのエピソードは両方の anime_id が nil になる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})

		got, err := repo.GetForUpdateByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForUpdateByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForUpdateByID() = nil, want エピソード")
		}
		if got.AnimeID != nil || got.ParentAnimeID != nil {
			t.Errorf("(AnimeID, ParentAnimeID) = (%v, %v), want (nil, nil)", got.AnimeID, got.ParentAnimeID)
		}
	})

	t.Run("異常系: 更新できないエピソードは (nil, nil) を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		deletedEpisodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			sortNumber: 100,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})
		// The dormant status column is deliberately made to disagree with the timestamps: a
		// row it alone calls deleted is still updatable.
		//
		// [Ja] 休眠 status カラムは意図的にタイムスタンプと食い違わせる。status だけが deleted の
		// 行は更新できる。
		dormantDeletedEpisodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 200})
		if _, err := tx.Exec(`UPDATE episodes SET status = 'deleted' WHERE id = $1`, int64(dormantDeletedEpisodeID)); err != nil {
			t.Fatalf("休眠 status の更新に失敗: %v", err)
		}

		deletedWorkID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		if _, err := tx.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(deletedWorkID)); err != nil {
			t.Fatalf("作品の削除に失敗: %v", err)
		}
		episodeOfDeletedWorkID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: deletedWorkID, sortNumber: 100})

		for _, tt := range []struct {
			name      string
			episodeID model.EpisodeID
			wantNil   bool
		}{
			{name: "存在しないエピソード", episodeID: model.EpisodeID(999999999), wantNil: true},
			{name: "削除済みのエピソード", episodeID: deletedEpisodeID, wantNil: true},
			{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID, wantNil: true},
			{name: "休眠 status だけが deleted のエピソード", episodeID: dormantDeletedEpisodeID, wantNil: false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.GetForUpdateByID(context.Background(), tt.episodeID)
				if err != nil {
					t.Fatalf("GetForUpdateByID() error = %v", err)
				}
				if tt.wantNil && got != nil {
					t.Errorf("GetForUpdateByID() = %+v, want nil", got)
				}
				if !tt.wantNil && got == nil {
					t.Error("GetForUpdateByID() = nil, want エピソード")
				}
			})
		}
	})
}

func TestEpisodeRepository_Update(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 送信された値を書き、版を進め、Rails の保存副作用を再現する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "#1", Valid: true},
			rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
			sortNumber: 100,
			title:      sql.NullString{String: "はじまり", Valid: true},
		})
		version := readDBUpdateEpisodeVersion(t, tx, episodeID)

		number := "第2話"
		rawNumber := 2.5
		title := "もう、お婿にいけません"
		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Number:     &number,
			RawNumber:  &rawNumber,
			Title:      &title,
			TitleEn:    "No Longer Marriageable",
			SortNumber: 200,
			Version:    version,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false, want true")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.number.String != number || stored.title.String != title {
			t.Errorf("(number, title) = (%q, %q), want (%q, %q)", stored.number.String, stored.title.String, number, title)
		}
		if stored.rawNumber.Float64 != rawNumber {
			t.Errorf("raw_number = %v, want %v", stored.rawNumber, rawNumber)
		}
		if stored.titleEn != "No Longer Marriageable" {
			t.Errorf("title_en = %q, want %q", stored.titleEn, "No Longer Marriageable")
		}
		if stored.sortNumber != 200 {
			t.Errorf("sort_number = %d, want 200", stored.sortNumber)
		}
		// The version has to move on, or a second submit from the same form would be accepted
		// as if nothing had been written.
		//
		// [Ja] 版は必ず進む必要がある。進まなければ、同じフォームからの 2 件目の送信が、何も
		// 書かれていないかのように受理されてしまう。
		if !stored.updatedAt.Valid || version == nil || !stored.updatedAt.Time.After(*version) {
			t.Errorf("updated_at = %+v, want > %v", stored.updatedAt, version)
		}

		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 1 {
			t.Errorf("DB 活動履歴 = %d 件, want 1", got)
		}
		var parameters string
		if err := tx.QueryRow(`
			SELECT parameters::text FROM db_activities
			WHERE trackable_type = 'Episode' AND trackable_id = $1`, int64(episodeID),
		).Scan(&parameters); err != nil {
			t.Fatalf("DB 活動履歴の読み込みに失敗: %v", err)
		}
		// Rails records the row before and after the save, so the admin change history can show
		// what the edit replaced.
		//
		// [Ja] Rails は保存前後の行を記録する。管理画面の変更履歴が、その編集で何が置き換わったか
		// を示せるようにするため。
		for _, want := range []string{`"old"`, `"new"`, "はじまり", title} {
			if !strings.Contains(parameters, want) {
				t.Errorf("parameters に %q が含まれていません: %s", want, parameters)
			}
		}

		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if !workTouched {
			t.Error("works.updated_at が更新されていません")
		}
	})

	t.Run("正常系: prev_episode_id を送信された sort_number の位置から再計算する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		firstID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})
		secondID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 200})
		// The edited episode starts at the front of the work, so it has no preceding episode.
		//
		// [Ja] 編集対象のエピソードは作品の先頭から始まるため、直前のエピソードを持たない。
		targetID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 50})
		// A deleted episode is not part of the order the list shows, so it never becomes the
		// preceding one.
		//
		// [Ja] 削除済みのエピソードは一覧が示す並びに含まれないため、直前のエピソードにはならない。
		insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			sortNumber: 150,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         targetID,
			WorkID:     workID,
			SortNumber: 250,
			Version:    readDBUpdateEpisodeVersion(t, tx, targetID),
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false, want true")
		}

		stored := readDBUpdateEpisode(t, tx, targetID)
		if stored.prevEpisodeID.Int64 != int64(secondID) {
			t.Errorf("prev_episode_id = %+v, want %d (sort_number 順の直前行)", stored.prevEpisodeID, int64(secondID))
		}
		if stored.prevEpisodeID.Int64 == int64(firstID) {
			t.Error("prev_episode_id が sort_number の直前ではない行を指しています")
		}
	})

	t.Run("正常系: 内容が変わらない送信では活動履歴も作品の touch も行わない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		number := "#1"
		title := "はじまり"
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: number, Valid: true},
			sortNumber: 100,
			title:      sql.NullString{String: title, Valid: true},
		})

		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Number:     &number,
			Title:      &title,
			SortNumber: 100,
			Version:    readDBUpdateEpisodeVersion(t, tx, episodeID),
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false, want true")
		}

		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB 活動履歴 = %d 件, want 0 (内容が変わっていないため)", got)
		}
		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if workTouched {
			t.Error("内容が変わっていないのに works.updated_at が更新されています")
		}
	})

	t.Run("正常系: prev_episode_id だけが動いた送信では活動履歴も作品の touch も行わない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		precedingID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})
		// The edited episode carries no prev_episode_id although an earlier episode exists, which
		// is the state a work is left in whenever the ordering moved after the column was last
		// written. The submit repeats the stored values, so the recomputation is the only
		// difference it produces.
		//
		// [Ja] 編集対象のエピソードは、先行するエピソードがあるにも関わらず prev_episode_id を
		// 持たない。カラムが最後に書かれた後に並び順が動いた作品が置かれる状態である。送信は保存済み
		// の値をそのまま繰り返すため、再計算だけが唯一の差分になる。
		title := "はじまり"
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			sortNumber: 200,
			title:      sql.NullString{String: title, Valid: true},
		})

		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &title,
			SortNumber: 200,
			Version:    readDBUpdateEpisodeVersion(t, tx, episodeID),
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false, want true")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.prevEpisodeID.Int64 != int64(precedingID) {
			t.Fatalf("prev_episode_id = %+v, want %d (再計算されていない)", stored.prevEpisodeID, int64(precedingID))
		}

		// prev_episode_id is derived from the ordering rather than typed, so its recomputation is
		// not an edit. Recording it would put a change the editor never made into the change
		// history the Rails admin reads.
		//
		// [Ja] prev_episode_id は入力ではなく並び順から導出されるため、その再計算は編集ではない。
		// 記録すると、編集者が行っていない変更が、Rails 側の管理画面が読む変更履歴に載ってしまう。
		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB 活動履歴 = %d 件, want 0 (導出列の再計算は編集ではない)", got)
		}
		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if workTouched {
			t.Error("導出列の再計算だけで works.updated_at が更新されています")
		}
	})

	t.Run("異常系: 古い版を名乗る送信は何も書かずに false を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		title := "はじまり"
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:     workID,
			sortNumber: 100,
			title:      sql.NullString{String: title, Valid: true},
		})

		staleVersion := time.Now().Add(-time.Hour)
		submitted := "上書きされたタイトル"
		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &submitted,
			SortNumber: 100,
			Version:    &staleVersion,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated {
			t.Fatal("Update() = true, want false")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.title.String != title {
			t.Errorf("title = %q, want %q (却下された送信は行を書かない)", stored.title.String, title)
		}
		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB 活動履歴 = %d 件, want 0", got)
		}
	})

	t.Run("正常系: updated_at が NULL の行は 1 回目だけ受理される", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{
			workID:        workID,
			sortNumber:    100,
			nullUpdatedAt: true,
		})

		first := "1 回目"
		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &first,
			SortNumber: 100,
			Version:    nil,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("1 回目の Update() = false, want true")
		}

		// The first write advances updated_at to a timestamp, so a second submit still naming
		// the NULL version no longer matches.
		//
		// [Ja] 最初の書き込みが updated_at を timestamp へ進めるため、NULL の版を名乗り続ける
		// 2 件目はもう一致しない。
		second := "2 回目"
		updated, err = repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &second,
			SortNumber: 100,
			Version:    nil,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated {
			t.Fatal("2 回目の Update() = true, want false")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.title.String != first {
			t.Errorf("title = %q, want %q", stored.title.String, first)
		}
	})
}

// dbUpdateNeighbourEpisode is one row of the work an ordering test starts from.
//
// unlinked leaves prev_episode_id NULL although a preceding episode exists, which is the state a
// row stays in for as long as nothing relinks it. A case that expects no write can then tell an
// untouched row apart from one a write happened to land the same value on.
//
// [Ja] dbUpdateNeighbourEpisode は並び順のテストが出発点とする作品の 1 行を表す。
//
// unlinked は、先行するエピソードがあるにも関わらず prev_episode_id を NULL のままにする。何も
// 張り替えなければ行が留まる状態である。これにより、書き込みが起きないことを期待するケースが、
// 触られていない行と、書き込みがたまたま同じ値を入れた行とを区別できる。
type dbUpdateNeighbourEpisode struct {
	sortNumber int32
	deleted    bool
	unlinked   bool
}

// insertDBUpdateNeighbourEpisodes inserts the given rows under one work and returns their IDs in
// the same order. The rows are chained in insertion order, each naming the one inserted before
// it, so the work starts out with prev_episode_id agreeing with the ordering and any value that
// disagrees afterwards is one this update wrote.
//
// [Ja] insertDBUpdateNeighbourEpisodes は与えられた行を 1 つの作品の下に挿入し、同じ順序で ID を
// 返す。行は挿入順に連結し、それぞれが直前に挿入された行を名乗るため、作品は prev_episode_id が
// 並び順と一致した状態から始まる。以降に食い違う値があれば、それはこの更新が書いたものである。
func insertDBUpdateNeighbourEpisodes(t *testing.T, tx *sql.Tx, workID model.WorkID, rows []dbUpdateNeighbourEpisode) []model.EpisodeID {
	t.Helper()
	ids := make([]model.EpisodeID, 0, len(rows))
	var prevEpisodeID sql.NullInt64
	for _, row := range rows {
		in := dbUpdateEpisodeRow{workID: workID, sortNumber: row.sortNumber, prevEpisodeID: prevEpisodeID}
		if row.deleted {
			in.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
		if row.unlinked {
			in.prevEpisodeID = sql.NullInt64{}
		}
		id := insertDBUpdateEpisode(t, tx, in)
		ids = append(ids, id)
		prevEpisodeID = sql.NullInt64{Int64: int64(id), Valid: true}
	}
	return ids
}

// assertDBUpdatePrevEpisodeIDs checks the whole work at once: wantPrev holds, for each episode,
// the index of the episode its prev_episode_id has to name, or -1 for NULL. Reading every row
// rather than only the relinked ones also fixes that no other row was touched.
//
// [Ja] assertDBUpdatePrevEpisodeIDs は作品全体を一度に検査する。wantPrev はエピソードごとに、その
// prev_episode_id が名乗るべきエピソードの添字 (NULL なら -1) を保持する。張り替え対象だけでなく
// 全行を読むことで、他の行が触られていないことも固定する。
func assertDBUpdatePrevEpisodeIDs(t *testing.T, tx *sql.Tx, ids []model.EpisodeID, wantPrev []int) {
	t.Helper()
	for i, id := range ids {
		want := sql.NullInt64{}
		if wantPrev[i] >= 0 {
			want = sql.NullInt64{Int64: int64(ids[wantPrev[i]]), Valid: true}
		}
		if got := readDBUpdateEpisode(t, tx, id).prevEpisodeID; got != want {
			t.Errorf("episodes[%d].prev_episode_id = %+v, want %+v", i, got, want)
		}
	}
}

// TestEpisodeRepository_UpdateRelinksNeighbours covers the two rows around a moved episode.
// Recomputing the edited row alone (TestEpisodeRepository_Update) leaves the row it used to
// precede naming an episode that is no longer in front of it, and the row it comes to precede
// naming the one before that, so the disagreement between the ordering and the column would
// merely move from the edited episode to its neighbours.
//
// [Ja] TestEpisodeRepository_UpdateRelinksNeighbours は移動したエピソードの前後 2 行を検証する。
// 編集対象の行だけを再計算すると (TestEpisodeRepository_Update)、移動前に直後だった行はもう前に
// いないエピソードを名乗り、移動後に直後になる行はその 1 つ前を名乗ったままになる。並び順と
// カラムの食い違いが、編集対象から隣接行へ移るだけで解消しない。
func TestEpisodeRepository_UpdateRelinksNeighbours(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 並び順に応じて隣接する 2 行を張り替える", func(t *testing.T) {
		t.Parallel()

		for _, tt := range []struct {
			name string
			// episodes describes the work in insertion order, which is also ascending
			// sort_number order.
			//
			// [Ja] episodes は挿入順 (sort_number 昇順でもある) に作品を記述する。
			episodes []dbUpdateNeighbourEpisode
			// targetIndex is the episode the submit edits and newSortNumber where it moves to.
			//
			// [Ja] targetIndex は送信が編集するエピソード、newSortNumber はその移動先。
			targetIndex   int
			newSortNumber int32
			wantPrev      []int
		}{
			{
				name: "後ろへ移動すると、跨がれた行と着地点の次の行が張り替わる",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   1,
				newSortNumber: 350,
				wantPrev:      []int{-1, 2, 0, 1},
			},
			{
				name: "前へ移動すると、着地点の次の行が編集対象を名乗る",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   3,
				newSortNumber: 150,
				wantPrev:      []int{-1, 3, 1, 0},
			},
			{
				name: "どの行も跨がない移動では隣接行を書かない",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300, unlinked: true},
				},
				targetIndex:   1,
				newSortNumber: 250,
				wantPrev:      []int{-1, 0, -1},
			},
			{
				name: "sort_number が変わらない送信では隣接行を書かない",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300, unlinked: true},
				},
				targetIndex:   1,
				newSortNumber: 200,
				wantPrev:      []int{-1, 0, -1},
			},
			{
				name: "先頭のエピソードを後ろへ移動すると、移動前の次の行が先頭になる",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   0,
				newSortNumber: 350,
				wantPrev:      []int{2, -1, 1, 0},
			},
			{
				name: "移動前の同値 sort_number は id をタイブレーカに直前行を選ぶ",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 200}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   1,
				newSortNumber: 350,
				wantPrev:      []int{-1, 2, 0, 1},
			},
			{
				name: "移動前後の同値 sort_number は id をタイブレーカに次の行を選ぶ",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   1,
				newSortNumber: 300,
				wantPrev:      []int{-1, 2, 0, 1, 3},
			},
			{
				// A deleted episode is not part of the order the list shows, so it is neither
				// relinked nor picked as the neighbour another row comes to name.
				//
				// [Ja] 削除済みのエピソードは一覧が示す並びに含まれないため、張り替えの対象にも、
				// 他の行が名乗る隣接行にもならない。
				name: "削除済みのエピソードは張り替えの対象にも隣接行にもならない",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 250, deleted: true}, {sortNumber: 300},
				},
				targetIndex:   1,
				newSortNumber: 400,
				wantPrev:      []int{-1, 3, 1, 0},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				db, tx := testutil.SetupTx(t)
				repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
				userID := testutil.NewUserBuilder(t, tx).Build()

				workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
				ids := insertDBUpdateNeighbourEpisodes(t, tx, workID, tt.episodes)
				targetID := ids[tt.targetIndex]

				updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
					ID:         targetID,
					WorkID:     workID,
					SortNumber: tt.newSortNumber,
					Version:    readDBUpdateEpisodeVersion(t, tx, targetID),
					UserID:     userID,
				})
				if err != nil {
					t.Fatalf("Update() error = %v", err)
				}
				if !updated {
					t.Fatal("Update() = false, want true")
				}

				assertDBUpdatePrevEpisodeIDs(t, tx, ids, tt.wantPrev)
			})
		}
	})

	t.Run("正常系: 張り替えは隣接行の版も変更履歴も動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		ids := insertDBUpdateNeighbourEpisodes(t, tx, workID, []dbUpdateNeighbourEpisode{
			{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
		})
		relinked := []model.EpisodeID{ids[2], ids[3]}
		versions := make([]*time.Time, len(relinked))
		for i, id := range relinked {
			versions[i] = readDBUpdateEpisodeVersion(t, tx, id)
		}

		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         ids[1],
			WorkID:     workID,
			SortNumber: 350,
			Version:    readDBUpdateEpisodeVersion(t, tx, ids[1]),
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false, want true")
		}
		// Read the links first: without the relink having happened, the assertions below would
		// hold for the wrong reason.
		//
		// [Ja] 先に張り替えを確認する。張り替え自体が起きていなければ、以降のアサーションは
		// 別の理由で成立してしまう。
		assertDBUpdatePrevEpisodeIDs(t, tx, ids, []int{-1, 2, 0, 1})

		for i, id := range relinked {
			stored := readDBUpdateEpisode(t, tx, id)
			// Advancing a neighbour's version would reject the next submit from an editor whose
			// form was open, over a column no form submits.
			//
			// [Ja] 隣接行の版を進めると、フォームを開いていた編集者の次の送信が、どのフォームも
			// 送信しないカラムを理由に却下されてしまう。
			if versions[i] == nil || !stored.updatedAt.Valid || !stored.updatedAt.Time.Equal(*versions[i]) {
				t.Errorf("張り替えられた行の updated_at = %+v, want %v (進めない)", stored.updatedAt, versions[i])
			}
			// The relink maintains a value the ordering derives, so recording it would put an
			// edit nobody made into the change history the Rails admin reads.
			//
			// [Ja] 張り替えは並び順から導出される値の維持のため、記録すると、誰も行っていない
			// 編集が Rails 側の管理画面が読む変更履歴に載ってしまう。
			if got := countDBEpisodeUpdateActivities(t, tx, id); got != 0 {
				t.Errorf("張り替えられた行の DB 活動履歴 = %d 件, want 0", got)
			}
		}
	})

	t.Run("異常系: 却下された送信は隣接行を張り替えない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		ids := insertDBUpdateNeighbourEpisodes(t, tx, workID, []dbUpdateNeighbourEpisode{
			{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
		})

		staleVersion := time.Now().Add(-time.Hour)
		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         ids[1],
			WorkID:     workID,
			SortNumber: 350,
			Version:    &staleVersion,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated {
			t.Fatal("Update() = true, want false")
		}

		// The edited row was not moved, so relinking its neighbours would leave the work naming
		// an ordering it does not have.
		//
		// [Ja] 編集対象の行は移動していないため、隣接行だけを張り替えると、作品は実際には持たない
		// 並び順を名乗ることになる。
		assertDBUpdatePrevEpisodeIDs(t, tx, ids, []int{-1, 0, 1, 2})
	})
}

// TestEpisodeRepository_UpdateRejectsConcurrentParentDeletion reproduces the boundary between
// the edit pre-read and the write. A parent soft-delete holds its row lock while Update starts;
// after that delete commits, Update must re-evaluate the kept-parent condition and return false
// without changing the episode.
//
// [Ja] TestEpisodeRepository_UpdateRejectsConcurrentParentDeletion は編集用の事前読み取りと
// 書き込みの境界を再現する。親作品の論理削除が行ロックを保持している間に Update を開始し、削除が
// commit した後は、削除されていない親という条件を再評価して false を返し、episode を変更しない
// ことを検証する。
func TestEpisodeRepository_UpdateRejectsConcurrentParentDeletion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	episodeID := insertDBUpdateEpisode(t, setupTx, dbUpdateEpisodeRow{
		workID:     workID,
		sortNumber: 100,
		title:      sql.NullString{String: "削除前のタイトル", Valid: true},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, episodeID)
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	// These fixtures are committed because both competing transactions must see them. Remove
	// the dependent rows explicitly after both transactions finish.
	//
	// [Ja] 競合する双方のトランザクションから見える必要があるため、fixture は commit する。
	// 両トランザクションの終了後に依存行から明示的に削除する。
	t.Cleanup(func() {
		statements := []string{
			`DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`,
			`DELETE FROM episodes WHERE work_id = $1`,
			`DELETE FROM works WHERE id = $1`,
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("並行更新fixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deleteTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("delete transaction BeginTx() error = %v", err)
	}
	defer func() { _ = deleteTx.Rollback() }()
	if _, err := deleteTx.ExecContext(ctx, `UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("親作品の論理削除に失敗: %v", err)
	}

	updateTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("update transaction BeginTx() error = %v", err)
	}
	defer func() { _ = updateTx.Rollback() }()
	var updateBackendPID int
	if err := updateTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&updateBackendPID); err != nil {
		t.Fatalf("更新側 backend PID の取得に失敗: %v", err)
	}

	title := "削除後に届いたタイトル"
	type updateResult struct {
		updated bool
		err     error
	}
	resultCh := make(chan updateResult, 1)
	repo := repository.NewEpisodeRepository(query.New(db)).WithTx(updateTx)
	go func() {
		updated, err := repo.Update(ctx, repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &title,
			SortNumber: 100,
			Version:    version,
			// No activity is inserted when the parent guard rejects the update, so this ID need
			// not name a persisted user.
			//
			// [Ja] 親のガードが更新を却下すれば活動履歴は挿入されないため、この ID が保存済み
			// ユーザーを指す必要はない。
			UserID: model.UserID(1),
		})
		resultCh <- updateResult{updated: updated, err: err}
	}()

	// Observe the update waiting on the row lock before allowing the delete to commit. Polling
	// pg_stat_activity makes the interleaving deterministic instead of relying on a sleep.
	//
	// [Ja] 削除の commit を許可する前に、更新が行ロック待ちになったことを観測する。
	// sleep に依存せず決定的な実行順にするため pg_stat_activity をポーリングする。
	lockDeadline := time.NewTimer(5 * time.Second)
	defer lockDeadline.Stop()
	lockTicker := time.NewTicker(10 * time.Millisecond)
	defer lockTicker.Stop()
	waitingForLock := false
	for !waitingForLock {
		select {
		case result := <-resultCh:
			t.Fatalf("Update() completed before parent deletion committed: %+v", result)
		case <-lockTicker.C:
			var waitEventType sql.NullString
			if err := deleteTx.QueryRowContext(ctx, `
				SELECT wait_event_type
				FROM pg_stat_activity
				WHERE pid = $1`, updateBackendPID).Scan(&waitEventType); err != nil {
				t.Fatalf("更新側の待機状態の取得に失敗: %v", err)
			}
			waitingForLock = waitEventType.Valid && waitEventType.String == "Lock"
		case <-lockDeadline.C:
			t.Fatal("Update() did not wait for the parent row lock")
		}
	}

	if err := deleteTx.Commit(); err != nil {
		t.Fatalf("delete transaction Commit() error = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("Update() did not finish after parent deletion committed: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("Update() error = %v", result.err)
	}
	if result.updated {
		t.Fatal("Update() = true, want false after parent deletion")
	}
	if err := updateTx.Commit(); err != nil {
		t.Fatalf("update transaction Commit() error = %v", err)
	}

	var storedTitle sql.NullString
	if err := db.QueryRow(`SELECT title FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&storedTitle); err != nil {
		t.Fatalf("エピソードの再読み込みに失敗: %v", err)
	}
	if storedTitle.String != "削除前のタイトル" {
		t.Errorf("title = %q, want %q", storedTitle.String, "削除前のタイトル")
	}
}

// TestEpisodeRepository_UpdateSerializesConcurrentSiblingUpdates pins the strength of the lock
// Update takes on the parent work. Two submits for episodes of the same work both touch that
// single row, so taking it shared first and exclusively later within the same statement would
// let each hold the shared lock while waiting for the other's, which PostgreSQL breaks by
// aborting one with a deadlock error the handler can only turn into a 500. Taking it up front
// at the strength the touch needs makes the second submit wait for the first to commit.
//
// The interleaving is built from a submit that changes nothing, which is the only way to hold
// the lock without also writing the work: it leaves episode_change empty, so touched_work
// writes no row and the lock taken by current_episode is the only one held.
//
// [Ja] TestEpisodeRepository_UpdateSerializesConcurrentSiblingUpdates は Update が親作品に対して
// 取るロックの強さを固定する。同じ作品のエピソードへの 2 つの送信はいずれもその 1 行に触れるため、
// 同一文の中で先に共有・後から排他と取ると、双方が共有ロックを保持したまま相手の共有ロックを
// 待つ状態になる。PostgreSQL はこれを片方のデッドロック中断で解消するが、ハンドラーはそれを 500 に
// するしかない。touch が必要とする強さで最初から取れば、2 つ目の送信は 1 つ目の commit を待つ。
//
// 交差の組み立てには内容が変わらない送信を使う。作品を書かずにロックだけを保持する唯一の方法で
// あるため。この送信は episode_change を空にするので touched_work は 1 行も書かず、
// current_episode が取ったロックだけが保持される。
func TestEpisodeRepository_UpdateSerializesConcurrentSiblingUpdates(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	firstNumber := "#1"
	firstTitle := "第 1 話"
	firstID := insertDBUpdateEpisode(t, setupTx, dbUpdateEpisodeRow{
		workID:     workID,
		number:     sql.NullString{String: firstNumber, Valid: true},
		rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
		sortNumber: 100,
		title:      sql.NullString{String: firstTitle, Valid: true},
		titleEn:    "Episode 1",
	})
	secondID := insertDBUpdateEpisode(t, setupTx, dbUpdateEpisodeRow{
		workID:     workID,
		sortNumber: 200,
		title:      sql.NullString{String: "第 2 話", Valid: true},
	})
	firstVersion := readDBUpdateEpisodeVersion(t, setupTx, firstID)
	secondVersion := readDBUpdateEpisodeVersion(t, setupTx, secondID)
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	// Both competing transactions have to see these rows, so the fixtures are committed and
	// removed explicitly afterwards. The activities the submits record reference the user, so
	// they go before the user's own rows.
	//
	// [Ja] 競合する双方のトランザクションから見える必要があるため fixture は commit し、後から
	// 明示的に削除する。送信が記録した活動履歴がユーザーを参照するため、ユーザー自身の行より先に
	// 消す。
	t.Cleanup(func() {
		statements := []string{
			`DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`,
			`DELETE FROM episodes WHERE work_id = $1`,
			`DELETE FROM works WHERE id = $1`,
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("並行更新fixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	firstTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("first transaction BeginTx() error = %v", err)
	}
	defer func() { _ = firstTx.Rollback() }()
	firstRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(firstTx)

	rawNumber := 1.0
	updated, err := firstRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID:         firstID,
		WorkID:     workID,
		Number:     &firstNumber,
		RawNumber:  &rawNumber,
		Title:      &firstTitle,
		TitleEn:    "Episode 1",
		SortNumber: 100,
		Version:    firstVersion,
		UserID:     userID,
	})
	if err != nil {
		t.Fatalf("内容が変わらない Update() error = %v", err)
	}
	if !updated {
		t.Fatal("内容が変わらない Update() = false, want true")
	}
	// The whole interleaving rests on this submit not writing the work: if it recorded a change,
	// the transaction would already hold the work exclusively and the second submit would be
	// blocked by that instead of by the lock under test.
	//
	// [Ja] 以降の交差はこの送信が作品を書かないことに依存する。変更として記録されていれば、
	// このトランザクションは既に作品を排他で保持しており、2 つ目の送信は検証対象のロックではなく
	// そちらに阻まれてしまう。
	if got := countDBEpisodeUpdateActivities(t, firstTx, firstID); got != 0 {
		t.Fatalf("内容が変わらない送信の DB 活動履歴 = %d 件, want 0", got)
	}
	revisedVersion := readDBUpdateEpisodeVersion(t, firstTx, firstID)

	secondTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("second transaction BeginTx() error = %v", err)
	}
	defer func() { _ = secondTx.Rollback() }()
	var secondBackendPID int
	if err := secondTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&secondBackendPID); err != nil {
		t.Fatalf("2 つ目の送信の backend PID の取得に失敗: %v", err)
	}

	secondTitle := "第 2 話 (改題)"
	type updateResult struct {
		updated bool
		err     error
	}
	resultCh := make(chan updateResult, 1)
	secondRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(secondTx)
	go func() {
		updated, err := secondRepo.Update(ctx, repository.UpdateEpisodeParams{
			ID:         secondID,
			WorkID:     workID,
			Title:      &secondTitle,
			SortNumber: 200,
			Version:    secondVersion,
			UserID:     userID,
		})
		resultCh <- updateResult{updated: updated, err: err}
	}()

	// Wait until the second submit is parked on the work row lock, so the first transaction's
	// next statement runs while the second is holding whatever lock it took. Polling
	// pg_stat_activity makes the interleaving deterministic instead of relying on a sleep.
	//
	// [Ja] 2 つ目の送信が作品の行ロックで停止するまで待ち、1 つ目のトランザクションの次の文が、
	// 2 つ目が取ったロックを保持したままの状態で走るようにする。sleep に依存せず決定的な実行順に
	// するため pg_stat_activity をポーリングする。
	lockDeadline := time.NewTimer(10 * time.Second)
	defer lockDeadline.Stop()
	lockTicker := time.NewTicker(10 * time.Millisecond)
	defer lockTicker.Stop()
	waitingForLock := false
	for !waitingForLock {
		select {
		case result := <-resultCh:
			t.Fatalf("2 つ目の Update() が 1 つ目のコミット前に完了した: %+v", result)
		case <-lockTicker.C:
			var waitEventType sql.NullString
			if err := firstTx.QueryRowContext(ctx, `
				SELECT wait_event_type
				FROM pg_stat_activity
				WHERE pid = $1`, secondBackendPID).Scan(&waitEventType); err != nil {
				t.Fatalf("2 つ目の送信の待機状態の取得に失敗: %v", err)
			}
			waitingForLock = waitEventType.Valid && waitEventType.String == "Lock"
		case <-lockDeadline.C:
			t.Fatal("2 つ目の Update() が作品の行ロックを待たなかった")
		}
	}

	// The first transaction now submits a change, which needs the work exclusively. With the
	// lock already held at that strength this proceeds; a shared lock would have to wait for
	// the second transaction's shared lock, which is itself waiting on the first, and the
	// deadlock would surface here as an error.
	//
	// [Ja] ここで 1 つ目のトランザクションが内容の変わる送信を行い、作品を排他で必要とする。
	// 既に同じ強さで保持していればそのまま進む。共有ロックであれば 2 つ目のトランザクションの
	// 共有ロックを待つことになり、その 2 つ目は 1 つ目を待っているため、デッドロックがここで
	// エラーとして現れる。
	revisedTitle := "第 1 話 (改題)"
	updated, err = firstRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID:         firstID,
		WorkID:     workID,
		Number:     &firstNumber,
		RawNumber:  &rawNumber,
		Title:      &revisedTitle,
		TitleEn:    "Episode 1",
		SortNumber: 100,
		Version:    revisedVersion,
		UserID:     userID,
	})
	if err != nil {
		t.Fatalf("同一作品の並行更新で 1 つ目の Update() error = %v", err)
	}
	if !updated {
		t.Fatal("1 つ目の Update() = false, want true")
	}
	if err := firstTx.Commit(); err != nil {
		t.Fatalf("first transaction Commit() error = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("1 つ目のコミット後も 2 つ目の Update() が完了しない: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("同一作品の並行更新で 2 つ目の Update() error = %v", result.err)
	}
	if !result.updated {
		t.Fatal("2 つ目の Update() = false, want true")
	}
	if err := secondTx.Commit(); err != nil {
		t.Fatalf("second transaction Commit() error = %v", err)
	}

	// Serialising the two must not cost either write.
	//
	// [Ja] 直列化しても、どちらの書き込みも失われてはならない。
	for _, want := range []struct {
		id    model.EpisodeID
		title string
	}{{firstID, revisedTitle}, {secondID, secondTitle}} {
		var storedTitle sql.NullString
		if err := db.QueryRow(`SELECT title FROM episodes WHERE id = $1`, int64(want.id)).Scan(&storedTitle); err != nil {
			t.Fatalf("エピソードの再読み込みに失敗: %v", err)
		}
		if storedTitle.String != want.title {
			t.Errorf("title = %q, want %q", storedTitle.String, want.title)
		}
	}
}

// TestEpisodeRepository_UpdateSerializesConcurrentMoves verifies that a submit which waited on
// the work derives both its former and destination neighbours from the winner's committed
// ordering. Using the snapshot from before the wait would leave a cycle in prev_episode_id.
//
// [Ja] TestEpisodeRepository_UpdateSerializesConcurrentMoves は、作品ロックを待った送信が、先行更新の
// commit 後の並びから移動前・移動先双方の隣接行を導出することを検証する。待機前のスナップショット
// を使うと prev_episode_id に循環が残る。
func TestEpisodeRepository_UpdateSerializesConcurrentMoves(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400}, {sortNumber: 500},
	})
	firstVersion := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	secondVersion := readDBUpdateEpisodeVersion(t, setupTx, ids[2])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("並行移動fixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	firstTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("first transaction BeginTx() error = %v", err)
	}
	defer func() { _ = firstTx.Rollback() }()
	firstRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(firstTx)

	updated, err := firstRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 450, Version: firstVersion, UserID: userID,
	})
	if err != nil {
		t.Fatalf("1 つ目の Update() error = %v", err)
	}
	if !updated {
		t.Fatal("1 つ目の Update() = false, want true")
	}

	secondTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("second transaction BeginTx() error = %v", err)
	}
	defer func() { _ = secondTx.Rollback() }()
	var secondBackendPID int
	if err := secondTx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&secondBackendPID); err != nil {
		t.Fatalf("2 つ目の送信の backend PID の取得に失敗: %v", err)
	}

	type updateResult struct {
		updated bool
		err     error
	}
	resultCh := make(chan updateResult, 1)
	secondRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(secondTx)
	go func() {
		updated, err := secondRepo.Update(ctx, repository.UpdateEpisodeParams{
			ID: ids[2], WorkID: workID, SortNumber: 50, Version: secondVersion, UserID: userID,
		})
		resultCh <- updateResult{updated: updated, err: err}
	}()

	lockDeadline := time.NewTimer(10 * time.Second)
	defer lockDeadline.Stop()
	lockTicker := time.NewTicker(10 * time.Millisecond)
	defer lockTicker.Stop()
	for {
		select {
		case result := <-resultCh:
			t.Fatalf("2 つ目の Update() が 1 つ目の commit 前に完了した: %+v", result)
		case <-lockTicker.C:
			var waitEventType sql.NullString
			if err := firstTx.QueryRowContext(
				ctx,
				"SELECT wait_event_type FROM pg_stat_activity WHERE pid = $1",
				secondBackendPID,
			).Scan(&waitEventType); err != nil {
				t.Fatalf("2 つ目の送信の待機状態の取得に失敗: %v", err)
			}
			if waitEventType.Valid && waitEventType.String == "Lock" {
				goto secondIsWaiting
			}
		case <-lockDeadline.C:
			t.Fatal("2 つ目の Update() が作品の行ロックを待たなかった")
		}
	}

secondIsWaiting:
	if err := firstTx.Commit(); err != nil {
		t.Fatalf("first transaction Commit() error = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("1 つ目の commit 後も 2 つ目の Update() が完了しない: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("2 つ目の Update() error = %v", result.err)
	}
	if !result.updated {
		t.Fatal("2 つ目の Update() = false, want true")
	}
	if err := secondTx.Commit(); err != nil {
		t.Fatalf("second transaction Commit() error = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("assert transaction BeginTx() error = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	assertDBUpdatePrevEpisodeIDs(t, assertTx, ids, []int{2, 3, -1, 0, 1})
}

// TestEpisodeRepository_UpdateBreaksRailsLockOrderCycle fixes the shared-DB lock protocol.
// Rails saves in episode -> work order, while Go protects neighbour derivation in work ->
// episodes order. If Rails already owns a sibling, Go must fail NOWAIT and roll its whole
// transaction back, allowing Rails to touch the work before the Go retry.
//
// [Ja] TestEpisodeRepository_UpdateBreaksRailsLockOrderCycle は共有 DB のロックプロトコルを固定する。
// Rails は episode -> work、Go は隣接導出を守るため work -> episodes の順でロックする。Rails が
// sibling を既に保持している場合、Go は NOWAIT で失敗してトランザクション全体を rollback し、
// Go の再試行前に Rails が作品を touch できるようにする必要がある。
func TestEpisodeRepository_UpdateBreaksRailsLockOrderCycle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("Rails順ロックfixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	railsTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Rails transaction BeginTx() error = %v", err)
	}
	defer func() { _ = railsTx.Rollback() }()
	if _, err := railsTx.ExecContext(ctx, "UPDATE episodes SET title = title WHERE id = $1", int64(ids[2])); err != nil {
		t.Fatalf("Rails順の episode lock に失敗: %v", err)
	}

	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go transaction BeginTx() error = %v", err)
	}
	// The explicit Rollback below is what releases the work lock at the point the test needs it
	// released. This defer only covers an assertion failing before that: without it the lock
	// would outlive the test and t.Cleanup's DELETE would wait on it forever, turning a failed
	// assertion into a hung suite.
	//
	// [Ja] 作品ロックをテストが必要とする時点で解放するのは下の明示的な Rollback である。この
	// defer はその前にアサーションが落ちた場合だけを受け持つ。これが無いとロックがテストより
	// 長く残り、t.Cleanup の DELETE が永久に待つため、失敗したアサーションがスイートのハングに
	// 変わってしまう。
	defer func() { _ = goTx.Rollback() }()
	goRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	updated, err := goRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if updated {
		t.Fatal("Rails が sibling を保持中の Update() = true, want false")
	}
	if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
		t.Fatalf("Rails が sibling を保持中の Update() error = %v, want ErrEpisodeLockUnavailable", err)
	}
	if err := goTx.Rollback(); err != nil {
		t.Fatalf("Go transaction Rollback() error = %v", err)
	}

	// This is the belongs_to :work, touch: true half of Rails' save order. It can proceed only
	// after the failed Go attempt has released its work lock.
	//
	// [Ja] これは Rails の保存順序における belongs_to :work, touch: true 側。失敗した Go の試行が
	// 作品ロックを解放した後にだけ進める。
	if _, err := railsTx.ExecContext(ctx, "UPDATE works SET updated_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("Rails順の work touch に失敗: %v", err)
	}
	if err := railsTx.Commit(); err != nil {
		t.Fatalf("Rails transaction Commit() error = %v", err)
	}

	retryTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("retry transaction BeginTx() error = %v", err)
	}
	defer func() { _ = retryTx.Rollback() }()
	retryRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(retryTx)
	updated, err = retryRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("Rails commit 後の Update() error = %v", err)
	}
	if !updated {
		t.Fatal("Rails commit 後の Update() = false, want true")
	}
	if err := retryTx.Commit(); err != nil {
		t.Fatalf("retry transaction Commit() error = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("assert transaction BeginTx() error = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	assertDBUpdatePrevEpisodeIDs(t, assertTx, ids, []int{-1, 2, 0, 1})
}

// TestEpisodeRepository_UpdateBreaksRailsDestinationPredecessorDeleteCycle covers the row the
// moved episode will reference through prev_episode_id. Rails destroy locks that episode before
// touching its work, while Go locks the work before deriving the destination predecessor. Go
// must include that predecessor in its NOWAIT lock set so the foreign-key check cannot complete
// the cycle. Once Go rolls back, Rails can touch the work and commit; the retry then derives its
// predecessor without the deleted row.
//
// [Ja] TestEpisodeRepository_UpdateBreaksRailsDestinationPredecessorDeleteCycle は、移動後の
// prev_episode_id から参照する行を検証する。Rails の destroy はそのエピソードをロックしてから作品を
// touch し、Go は作品をロックしてから移動先の直前行を導出する。外部キーの検査が循環を完成させない
// よう、Go はその直前行も NOWAIT の対象に含める必要がある。Go の rollback 後は Rails が作品を
// touch して commit でき、再試行は削除済み行を除いて直前行を導出する。
func TestEpisodeRepository_UpdateBreaksRailsDestinationPredecessorDeleteCycle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("移動先直前行の削除fixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	railsTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Rails transaction BeginTx() error = %v", err)
	}
	defer func() { _ = railsTx.Rollback() }()
	if _, err := railsTx.ExecContext(ctx, "DELETE FROM episodes WHERE id = $1", int64(ids[3])); err != nil {
		t.Fatalf("Rails順の destination predecessor delete に失敗: %v", err)
	}

	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go transaction BeginTx() error = %v", err)
	}
	// The explicit Rollback below releases the work lock before Rails touches it. This defer
	// covers an assertion failing before that point so cleanup cannot hang behind the lock.
	//
	// [Ja] 下の明示的な Rollback が Rails の touch 前に作品ロックを解放する。この defer はその前に
	// アサーションが落ちた場合を受け持ち、後始末がロック待ちで止まるのを防ぐ。
	defer func() { _ = goTx.Rollback() }()
	goRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	attemptCtx, cancelAttempt := context.WithTimeout(ctx, 2*time.Second)
	updated, err := goRepo.Update(attemptCtx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 500, Version: version, UserID: userID,
	})
	cancelAttempt()
	if updated {
		t.Fatal("Rails が destination predecessor を削除中の Update() = true, want false")
	}
	if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
		t.Fatalf(
			"Rails が destination predecessor を削除中の Update() error = %v, want ErrEpisodeLockUnavailable",
			err,
		)
	}
	if err := goTx.Rollback(); err != nil {
		t.Fatalf("Go transaction Rollback() error = %v", err)
	}

	// This is the belongs_to :work, touch: true half of Rails' destroy order. The NOWAIT failure
	// has released the Go work lock, so this write and the delete can commit without a cycle.
	//
	// [Ja] これは Rails の destroy 順序における belongs_to :work, touch: true 側。NOWAIT の失敗で
	// Go の作品ロックは解放済みなので、この書き込みと削除は循環せず commit できる。
	if _, err := railsTx.ExecContext(ctx, "UPDATE works SET updated_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("Rails順の work touch に失敗: %v", err)
	}
	if err := railsTx.Commit(); err != nil {
		t.Fatalf("Rails transaction Commit() error = %v", err)
	}

	retryTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("retry transaction BeginTx() error = %v", err)
	}
	defer func() { _ = retryTx.Rollback() }()
	retryRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(retryTx)
	updated, err = retryRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 500, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("Rails commit 後の Update() error = %v", err)
	}
	if !updated {
		t.Fatal("Rails commit 後の Update() = false, want true")
	}
	if err := retryTx.Commit(); err != nil {
		t.Fatalf("retry transaction Commit() error = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("assert transaction BeginTx() error = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	wantTargetPrev := sql.NullInt64{Int64: int64(ids[2]), Valid: true}
	if got := readDBUpdateEpisode(t, assertTx, ids[1]).prevEpisodeID; got != wantTargetPrev {
		t.Errorf("移動した episode の prev_episode_id = %+v, want %+v", got, wantTargetPrev)
	}
	wantFormerFollowingPrev := sql.NullInt64{Int64: int64(ids[0]), Valid: true}
	if got := readDBUpdateEpisode(t, assertTx, ids[2]).prevEpisodeID; got != wantFormerFollowingPrev {
		t.Errorf("移動前の直後 episode の prev_episode_id = %+v, want %+v", got, wantFormerFollowingPrev)
	}
	var deletedReferenceCount int
	if err := assertTx.QueryRow(
		"SELECT COUNT(*) FROM episodes WHERE work_id = $1 AND prev_episode_id = $2",
		int64(workID), int64(ids[3]),
	).Scan(&deletedReferenceCount); err != nil {
		t.Fatalf("削除済み episode への参照数の取得に失敗: %v", err)
	}
	if deletedReferenceCount != 0 {
		t.Errorf("削除済み episode を指す prev_episode_id = %d 件, want 0", deletedReferenceCount)
	}
}

// TestEpisodeRepository_UpdateIgnoresLocksOnNonNeighbours bounds what one edit is allowed to
// wait on. Rails locks an episode row for writes that have nothing to do with the ordering:
// counter_culture :episode on EpisodeRecord means every record a user makes holds its episode
// for the rest of that transaction. If the update pre-empted the whole work, such a record
// anywhere in a currently airing work would abort an unrelated edit and, once the retries ran
// out, be reported to the editor as a failure.
//
// [Ja] TestEpisodeRepository_UpdateIgnoresLocksOnNonNeighbours は、1 回の編集が何を待ち得るかを
// 限定する。Rails は並び順と無関係な書き込みでもエピソード行をロックする。EpisodeRecord の
// counter_culture :episode により、ユーザーが記録を作るたびにそのエピソードをトランザクション
// の間ずっと保持する。更新が作品全体を先取りしていると、放送中の作品のどこかで作られた記録が
// 無関係な編集を中断させ、再試行を使い切ると編集者へ失敗として報告されてしまう。
func TestEpisodeRepository_UpdateIgnoresLocksOnNonNeighbours(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("fixture transaction Begin() error = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300},
		{sortNumber: 400}, {sortNumber: 500}, {sortNumber: 600},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("fixture transaction Commit() error = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("非隣接ロックfixtureの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// This is the shape of counter_culture's write: it changes no ordering column, but it holds
	// the row for the rest of its transaction all the same.
	//
	// [Ja] これが counter_culture の書き込みの形である。並び順のカラムは変えないが、それでも
	// トランザクションの間ずっとその行を保持する。
	recordTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("記録トランザクションの BeginTx() error = %v", err)
	}
	defer func() { _ = recordTx.Rollback() }()
	if _, err := recordTx.ExecContext(
		ctx,
		"UPDATE episodes SET episode_records_count = episode_records_count + 1 WHERE id = $1",
		int64(ids[5]),
	); err != nil {
		t.Fatalf("隣接行ではないエピソードのロック取得に失敗: %v", err)
	}

	// Moving ids[1] to 350 makes ids[0] / ids[2] / ids[3] its neighbours, leaving ids[4] and
	// ids[5] outside the set the update reads or writes.
	//
	// [Ja] ids[1] を 350 へ動かすと ids[0] / ids[2] / ids[3] が隣接行になり、ids[4] と ids[5] は
	// 更新が読み書きする範囲の外に残る。
	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go transaction BeginTx() error = %v", err)
	}
	defer func() { _ = goTx.Rollback() }()
	repo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	updated, err := repo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("隣接行ではない行がロックされている状態の Update() error = %v", err)
	}
	if !updated {
		t.Fatal("隣接行ではない行がロックされている状態の Update() = false, want true")
	}

	assertDBUpdatePrevEpisodeIDs(t, goTx, ids, []int{-1, 2, 0, 1, 3, 4})
}

// TestEpisodeRepository_UpdateRejectsMovedParent covers the guard on the parent observed by the
// edit pre-read. Rails' Annict::DataCare::MoveEpisode writes episodes.work_id with
// update_column, which takes no work lock, so the target can end up under a different work
// between the form being opened and the submit. Its ordering says nothing about the new
// parent's list, so the submit is refused rather than applied there.
//
// [Ja] TestEpisodeRepository_UpdateRejectsMovedParent は、編集用の事前読み取りが観測した親作品に
// 対するガードを検証する。Rails の Annict::DataCare::MoveEpisode は episodes.work_id を
// update_column で書き、作品ロックを取らないため、フォームを開いてから送信までの間に対象が別の
// 作品の配下へ移りうる。その並び順は新しい親の一覧について何も述べていないため、送信はそこへ
// 適用せず却下する。
func TestEpisodeRepository_UpdateRejectsMovedParent(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).Build()

	originalWorkID := insertDBUpdateWork(t, tx, sql.NullInt64{})
	otherWorkID := insertDBUpdateWork(t, tx, sql.NullInt64{})
	episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: originalWorkID, sortNumber: 100})
	version := readDBUpdateEpisodeVersion(t, tx, episodeID)

	if _, err := tx.Exec("UPDATE episodes SET work_id = $1 WHERE id = $2", int64(otherWorkID), int64(episodeID)); err != nil {
		t.Fatalf("別作品への付け替えに失敗: %v", err)
	}

	title := "移動後に届いた送信"
	updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
		ID:         episodeID,
		WorkID:     originalWorkID,
		Title:      &title,
		SortNumber: 250,
		Version:    version,
		UserID:     userID,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated {
		t.Fatal("別作品へ移された行への Update() = true, want false")
	}

	stored := readDBUpdateEpisode(t, tx, episodeID)
	if stored.title.Valid && stored.title.String == title {
		t.Errorf("episodes.title = %q, want 送信前の値のまま", stored.title.String)
	}
	if stored.sortNumber != 100 {
		t.Errorf("episodes.sort_number = %d, want 100", stored.sortNumber)
	}
}

// dbArchiveEpisodeRow holds the mapping and state timestamps an archive or re-publish test
// starts an episode from. AnimeID lets the success cases verify Archive / Unarchive return the
// mapping from the updated row.
//
// [Ja] dbArchiveEpisodeRow は非公開・再公開テストがエピソードの初期値として与える写像と状態
// タイムスタンプを保持する。AnimeID により、正常系は Archive / Unarchive が更新した行の写像を
// 返すことを検証できる。
type dbArchiveEpisodeRow struct {
	workID        model.WorkID
	animeID       sql.NullInt64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
}

// insertDBArchiveWork inserts the parent work of an archive, re-publish or delete test with the
// counter cache those writes move and an updated_at set back a day, so the touch they perform is
// observable.
//
// [Ja] insertDBArchiveWork は非公開・再公開・削除テストの親作品を、3 者が動かすカウンター
// キャッシュ付きで、updated_at を 1 日前にして挿入する。3 者が行う touch を観測できるように
// するため。
func insertDBArchiveWork(t *testing.T, tx *sql.Tx, episodesCount int32) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO works (title, media, episodes_count, created_at, updated_at)
		VALUES ($1, 1, $2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day') RETURNING id`,
		"非公開対象の作品", episodesCount,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// insertDBArchiveEpisode inserts the episode an archive test unpublishes or a re-publish test
// publishes again. Its timestamps come from the database an hour back rather than from the Go
// clock, for the reason insertDBUpdateEpisode states: NOW() is the transaction's start time, so
// a fixture stamped by the test process could sit after the value the write produces.
//
// [Ja] insertDBArchiveEpisode は非公開テストが非公開にする、または再公開テストが公開に戻す
// エピソードを挿入する。タイムスタンプは insertDBUpdateEpisode が述べる理由により、Go の時計では
// なく DB の 1 時間前を使う。NOW() はトランザクション開始時刻のため、テストプロセスが打刻した
// フィクスチャは書き込みが打つ値より後になりうる。
func insertDBArchiveEpisode(t *testing.T, tx *sql.Tx, in dbArchiveEpisodeRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, number, sort_number, title, anime_id, unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES (
			$1, '第1話', 100, '教えてティーチャー', $2, $3, $4,
			NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour'
		) RETURNING id`,
		int64(in.workID), in.animeID, in.unpublishedAt, in.deletedAt,
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// storedDBArchiveEpisode is the episode state an archive or re-publish test reads back: the
// state column both directions write and the version they advance alongside it.
//
// [Ja] storedDBArchiveEpisode は非公開・再公開テストが読み戻すエピソードの状態。両方向が書く
// 状態カラムと、それと併せて進める版。
type storedDBArchiveEpisode struct {
	unpublishedAt sql.NullTime
	updatedAt     sql.NullTime
}

func readDBArchiveEpisode(t *testing.T, tx *sql.Tx, id model.EpisodeID) storedDBArchiveEpisode {
	t.Helper()
	var row storedDBArchiveEpisode
	if err := tx.QueryRow(
		`SELECT unpublished_at, updated_at FROM episodes WHERE id = $1`, int64(id),
	).Scan(&row.unpublishedAt, &row.updatedAt); err != nil {
		t.Fatalf("非公開後のエピソードの読み込みに失敗: %v", err)
	}
	return row
}

// storedDBArchiveWork is the parent-work state an archive, re-publish or delete test reads back:
// the counter cache the Rails API serves and the timestamp Rails advances through
// belongs_to :work, touch: true.
//
// [Ja] storedDBArchiveWork は非公開・再公開・削除テストが読み戻す親作品の状態。Rails API が
// 配信するカウンターキャッシュと、Rails が belongs_to :work, touch: true で進めるタイム
// スタンプ。
type storedDBArchiveWork struct {
	episodesCount int32
	updatedAt     time.Time
}

func readDBArchiveWork(t *testing.T, tx *sql.Tx, id model.WorkID) storedDBArchiveWork {
	t.Helper()
	var row storedDBArchiveWork
	if err := tx.QueryRow(
		`SELECT episodes_count, updated_at FROM works WHERE id = $1`, int64(id),
	).Scan(&row.episodesCount, &row.updatedAt); err != nil {
		t.Fatalf("非公開後の作品の読み込みに失敗: %v", err)
	}
	return row
}

// countDBEpisodeActivities counts every change-history row the Rails admin reads for an
// episode, whatever the action. The Rails unpublish and publish are both a plain
// update(unpublished_at:) and record none, so neither direction may change the count.
//
// [Ja] countDBEpisodeActivities は、あるエピソードについて Rails 側の管理画面が変更履歴として
// 読む行を、action を問わず数える。Rails の非公開・再公開はいずれも素の
// update(unpublished_at:) で 1 件も記録しないため、どちらの方向も件数を変えてはならない。
func countDBEpisodeActivities(t *testing.T, tx *sql.Tx, id model.EpisodeID) int {
	t.Helper()
	var count int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM db_activities WHERE trackable_type = 'Episode' AND trackable_id = $1`,
		int64(id),
	).Scan(&count); err != nil {
		t.Fatalf("DB 活動履歴の読み込みに失敗: %v", err)
	}
	return count
}

func TestEpisodeRepository_Archive(t *testing.T) {
	t.Parallel()

	// The published episode is unpublished, its version advances with it, and the parent work
	// sees the two side effects the Rails unpublish has: the counter cache loses the row it no
	// longer counts and the work is touched. No change history is recorded, matching
	// update(unpublished_at:), which does not go through save_and_create_activity!.
	//
	// [Ja] 公開中のエピソードが非公開になり、版もそれと併せて進む。親作品には Rails の非公開が
	// 持つ 2 つの副作用が現れる。カウンターキャッシュはもう数えない行を失い、作品は touch される。
	// 変更履歴は記録されない。save_and_create_activity! を通らない update(unpublished_at:) と
	// 揃えるため。
	t.Run("正常系: 公開中のエピソードを非公開にし、作品のカウンターと更新時刻を動かす", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		var animeID int64
		if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&animeID); err != nil {
			t.Fatalf("anime の挿入に失敗: %v", err)
		}
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:  workID,
			animeID: sql.NullInt64{Int64: animeID, Valid: true},
		})
		before := readDBArchiveEpisode(t, tx, episodeID)
		workBefore := readDBArchiveWork(t, tx, workID)

		result, err := repo.Archive(context.Background(), repository.ArchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if result == nil {
			t.Fatal("Archive() = nil, want result")
		}
		if result.AnimeID == nil || *result.AnimeID != model.AnimeID(animeID) {
			t.Errorf("Archive().AnimeID = %v, want %d", result.AnimeID, animeID)
		}

		stored := readDBArchiveEpisode(t, tx, episodeID)
		if !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL, want 非公開の時刻")
		}
		if !stored.updatedAt.Valid || !stored.updatedAt.Time.After(before.updatedAt.Time) {
			t.Errorf("episodes.updated_at = %v, want %v より後", stored.updatedAt, before.updatedAt.Time)
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 2 {
			t.Errorf("works.episodes_count = %d, want 2", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v, want %v より後", work.updatedAt, workBefore.updatedAt)
		}
		if count := countDBEpisodeActivities(t, tx, episodeID); count != 0 {
			t.Errorf("DB 活動履歴 = %d 件, want 0 件", count)
		}
	})

	// A submit from a confirmation page opened before someone else archived the episode finds
	// no publishable row. Reporting that instead of writing keeps the counter from being
	// decremented twice for one transition.
	//
	// [Ja] 他者が先に非公開にした後の確認ページからの送信は、公開中の行を見つけない。書き込まず
	// それを報告することで、1 回の遷移に対してカウンターが 2 度減算されるのを防ぐ。
	t.Run("非公開済みのエピソードは nil を返し、カウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:        workID,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		result, err := repo.Archive(context.Background(), repository.ArchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if result != nil {
			t.Fatal("非公開済みのエピソードへの Archive() returned a result, want nil")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d, want 3", work.episodesCount)
		}
	})

	// A deleted episode is out of reach of the confirmation page, so its submit is refused too
	// rather than stamping unpublished_at on a row nothing displays.
	//
	// [Ja] 削除済みのエピソードは確認ページの対象外のため、その送信も拒否する。何も表示しない行に
	// unpublished_at を打たないようにするため。
	t.Run("削除済みのエピソードは nil を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:    workID,
			deletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		result, err := repo.Archive(context.Background(), repository.ArchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if result != nil {
			t.Fatal("削除済みのエピソードへの Archive() returned a result, want nil")
		}
	})

	// A soft-deleted parent is outside the admin list scope even when the published episode row
	// still exists. Refusing it before the episode update keeps both the state and counter
	// unchanged, as the re-publish direction does.
	//
	// [Ja] 論理削除済みの親作品は、公開中の episode 行が残っていても管理画面の一覧対象外。
	// 再公開方向と同じく、episode の更新前に拒否し、状態とカウンターの両方を変えないことを検証する。
	t.Run("削除済み作品のエピソードは nil を返し、非公開にしない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{workID: workID})
		if _, err := tx.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
			t.Fatalf("親作品の削除に失敗: %v", err)
		}

		result, err := repo.Archive(context.Background(), repository.ArchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if result != nil {
			t.Fatal("削除済み作品のエピソードへの Archive() returned a result, want nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at に値が入った, want NULL のまま")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d, want 3", work.episodesCount)
		}
	})

	// An episode moved to another work between the confirmation page and its submit no longer
	// belongs to the work the page counted, so neither the episode nor either work is written.
	//
	// [Ja] 確認ページと送信の間に別作品へ移されたエピソードは、そのページが数えていた作品にもう
	// 属していない。したがってエピソードもどちらの作品も書かない。
	t.Run("別作品へ移されたエピソードは nil を返し、元の作品のカウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		originalWorkID := insertDBArchiveWork(t, tx, 3)
		movedWorkID := insertDBArchiveWork(t, tx, 1)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{workID: movedWorkID})

		result, err := repo.Archive(context.Background(), repository.ArchiveEpisodeParams{
			ID:     episodeID,
			WorkID: originalWorkID,
		})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if result != nil {
			t.Fatal("別作品へ移された行への Archive() returned a result, want nil")
		}
		if work := readDBArchiveWork(t, tx, originalWorkID); work.episodesCount != 3 {
			t.Errorf("移動元の works.episodes_count = %d, want 3", work.episodesCount)
		}
		if work := readDBArchiveWork(t, tx, movedWorkID); work.episodesCount != 1 {
			t.Errorf("移動先の works.episodes_count = %d, want 1", work.episodesCount)
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at に値が入った, want NULL のまま")
		}
	})
}

func TestEpisodeRepository_Unarchive(t *testing.T) {
	t.Parallel()

	// The archived episode is published again, its version advances with it, and the parent work
	// sees the two side effects the Rails publish has: the counter cache regains the row it
	// counts once more and the work is touched. No change history is recorded, matching
	// update(unpublished_at: nil), which does not go through save_and_create_activity!.
	//
	// [Ja] 非公開のエピソードが再び公開になり、版もそれと併せて進む。親作品には Rails の再公開が
	// 持つ 2 つの副作用が現れる。カウンターキャッシュは再び数える行を取り戻し、作品は touch
	// される。変更履歴は記録されない。save_and_create_activity! を通らない
	// update(unpublished_at: nil) と揃えるため。
	t.Run("正常系: 非公開のエピソードを公開に戻し、作品のカウンターと更新時刻を動かす", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 2)
		var animeID int64
		if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('archived') RETURNING id`).Scan(&animeID); err != nil {
			t.Fatalf("anime の挿入に失敗: %v", err)
		}
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:        workID,
			animeID:       sql.NullInt64{Int64: animeID, Valid: true},
			unpublishedAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		})
		before := readDBArchiveEpisode(t, tx, episodeID)
		workBefore := readDBArchiveWork(t, tx, workID)

		result, err := repo.Unarchive(context.Background(), repository.UnarchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Unarchive() error = %v", err)
		}
		if result == nil {
			t.Fatal("Unarchive() = nil, want result")
		}
		if result.AnimeID == nil || *result.AnimeID != model.AnimeID(animeID) {
			t.Errorf("Unarchive().AnimeID = %v, want %d", result.AnimeID, animeID)
		}

		stored := readDBArchiveEpisode(t, tx, episodeID)
		if stored.unpublishedAt.Valid {
			t.Errorf("episodes.unpublished_at = %v, want NULL", stored.unpublishedAt.Time)
		}
		if !stored.updatedAt.Valid || !stored.updatedAt.Time.After(before.updatedAt.Time) {
			t.Errorf("episodes.updated_at = %v, want %v より後", stored.updatedAt, before.updatedAt.Time)
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d, want 3", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v, want %v より後", work.updatedAt, workBefore.updatedAt)
		}
		if count := countDBEpisodeActivities(t, tx, episodeID); count != 0 {
			t.Errorf("DB 活動履歴 = %d 件, want 0 件", count)
		}
	})

	// A submit from a list opened before someone else re-published the episode finds no archived
	// row. Reporting that instead of writing keeps the counter from being incremented twice for
	// one transition.
	//
	// [Ja] 他者が先に再公開した後の一覧からの送信は、非公開の行を見つけない。書き込まずそれを
	// 報告することで、1 回の遷移に対してカウンターが 2 度加算されるのを防ぐ。
	t.Run("公開中のエピソードは nil を返し、カウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{workID: workID})

		result, err := repo.Unarchive(context.Background(), repository.UnarchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Unarchive() error = %v", err)
		}
		if result != nil {
			t.Fatal("公開中のエピソードへの Unarchive() returned a result, want nil")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d, want 3", work.episodesCount)
		}
	})

	// A deleted episode is out of reach of the list the re-publish is submitted from, so its
	// submit is refused too rather than publishing a row nothing displays.
	//
	// [Ja] 削除済みのエピソードは再公開の送信元である一覧の対象外のため、その送信も拒否する。
	// 何も表示しない行を公開しないようにするため。
	t.Run("削除済みのエピソードは nil を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:        workID,
			unpublishedAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
			deletedAt:     sql.NullTime{Time: time.Now(), Valid: true},
		})

		result, err := repo.Unarchive(context.Background(), repository.UnarchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Unarchive() error = %v", err)
		}
		if result != nil {
			t.Fatal("削除済みのエピソードへの Unarchive() returned a result, want nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL, want 非公開の時刻のまま")
		}
	})

	// A soft-deleted parent is outside the admin list scope even when the archived episode row
	// still exists. Refusing it before the episode update keeps both the state and counter
	// unchanged.
	//
	// [Ja] 論理削除済みの親作品は、非公開の episode 行が残っていても管理画面の一覧対象外。
	// episode の更新前に拒否し、状態とカウンターの両方を変えないことを検証する。
	t.Run("削除済み作品のエピソードは nil を返し、再公開しない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:        workID,
			unpublishedAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		})
		if _, err := tx.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
			t.Fatalf("親作品の削除に失敗: %v", err)
		}

		result, err := repo.Unarchive(context.Background(), repository.UnarchiveEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Unarchive() error = %v", err)
		}
		if result != nil {
			t.Fatal("削除済み作品のエピソードへの Unarchive() returned a result, want nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL, want 非公開の時刻のまま")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d, want 3", work.episodesCount)
		}
	})

	// An episode moved to another work between the list and its submit no longer belongs to the
	// work that list counted, so neither the episode nor either work is written.
	//
	// [Ja] 一覧と送信の間に別作品へ移されたエピソードは、その一覧が数えていた作品にもう属して
	// いない。したがってエピソードもどちらの作品も書かない。
	t.Run("別作品へ移されたエピソードは nil を返し、元の作品のカウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		originalWorkID := insertDBArchiveWork(t, tx, 3)
		movedWorkID := insertDBArchiveWork(t, tx, 1)
		episodeID := insertDBArchiveEpisode(t, tx, dbArchiveEpisodeRow{
			workID:        movedWorkID,
			unpublishedAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		})

		result, err := repo.Unarchive(context.Background(), repository.UnarchiveEpisodeParams{
			ID:     episodeID,
			WorkID: originalWorkID,
		})
		if err != nil {
			t.Fatalf("Unarchive() error = %v", err)
		}
		if result != nil {
			t.Fatal("別作品へ移された行への Unarchive() returned a result, want nil")
		}
		if work := readDBArchiveWork(t, tx, originalWorkID); work.episodesCount != 3 {
			t.Errorf("移動元の works.episodes_count = %d, want 3", work.episodesCount)
		}
		if work := readDBArchiveWork(t, tx, movedWorkID); work.episodesCount != 1 {
			t.Errorf("移動先の works.episodes_count = %d, want 1", work.episodesCount)
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at が NULL になった, want 非公開の時刻のまま")
		}
	})
}

func TestEpisodeRepository_GetForArchiveByID(t *testing.T) {
	t.Parallel()

	// The loader projects what the confirmation page shows (the episode's number and title,
	// the parent work's title and no_episodes) together with the state timestamps the submit
	// checks. The anime mapping comes from ArchiveDBEpisode's updated row instead.
	//
	// [Ja] ローダーは確認ページが表示するもの (エピソードの話数とタイトル、親作品のタイトルと
	// no_episodes) と、送信が検査する状態タイムスタンプを射影する。anime の写像は代わりに
	// ArchiveDBEpisode が更新した行から得る。
	t.Run("正常系: 確認ページと送信が必要とするカラムを射影する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開確認の作品").WithNoEpisodes(true).Build()
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			sortNumber: 200,
			title:      sql.NullString{String: "もう、お婿にいけません", Valid: true},
			titleRo:    "Mou, Oyome ni Ikemasen",
			titleEn:    "No Longer Marriageable",
			status:     "published",
		})

		got, err := repo.GetForArchiveByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForArchiveByID() = nil, want エピソード")
		}
		if got.Episode.Number == nil || *got.Episode.Number != "第2話" {
			t.Errorf("Episode.Number = %v, want %q", got.Episode.Number, "第2話")
		}
		if got.Episode.Title == nil || *got.Episode.Title != "もう、お婿にいけません" {
			t.Errorf("Episode.Title = %v, want %q", got.Episode.Title, "もう、お婿にいけません")
		}

		// The state columns travel so the caller derives the status itself; a published
		// episode carries neither.
		//
		// [Ja] 状態カラムは呼び出し側が status を導出できるよう運ばれる。公開中のエピソードは
		// どちらも持たない。
		if got.Episode.DerivedStatus() != model.EpisodeStatusPublished {
			t.Errorf("Episode.DerivedStatus() = %q, want %q", got.Episode.DerivedStatus(), model.EpisodeStatusPublished)
		}
		if got.Work.ID != workID {
			t.Errorf("Work.ID = %d, want %d", int64(got.Work.ID), int64(workID))
		}
		if got.Work.Title != "非公開確認の作品" {
			t.Errorf("Work.Title = %q, want %q", got.Work.Title, "非公開確認の作品")
		}
		if !got.Work.NoEpisodes {
			t.Error("Work.NoEpisodes = false, want true")
		}
	})

	// An archived episode is still returned: the state is reported through the timestamps so
	// the caller decides what to do with it, which is what lets the same loader serve the
	// un-archive path later.
	//
	// [Ja] 非公開のエピソードも返す。状態はタイムスタンプで報告し、その扱いは呼び出し側が決める。
	// これにより同じローダーを後の再公開の経路でも使える。
	t.Run("非公開のエピソードは archived として返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:        workID,
			sortNumber:    100,
			status:        "published",
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		got, err := repo.GetForArchiveByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetForArchiveByID() = nil, want エピソード")
		}
		if got.Episode.DerivedStatus() != model.EpisodeStatusArchived {
			t.Errorf("Episode.DerivedStatus() = %q, want %q", got.Episode.DerivedStatus(), model.EpisodeStatusArchived)
		}
	})

	// The exclusions match the edit form's loader, so the archive confirmation page is
	// reachable for exactly the episodes the edit form is.
	//
	// [Ja] 除外条件は編集フォームのローダーと揃える。非公開の確認ページに到達できるエピソードを、
	// 編集フォームに到達できるエピソードと一致させるため。
	t.Run("削除済みのエピソードと削除済み作品のエピソードは nil を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		keptWorkID := insertDBListWork(t, tx)
		deletedEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     keptWorkID,
			sortNumber: 100,
			status:     "published",
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		deletedWorkID := testutil.NewWorkBuilder(t, tx).WithDeletedAt(time.Now()).Build()
		orphanEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     deletedWorkID,
			sortNumber: 100,
			status:     "published",
		})

		for name, id := range map[string]model.EpisodeID{
			"削除済みのエピソード":   deletedEpisodeID,
			"削除済み作品のエピソード": orphanEpisodeID,
			"存在しないエピソード":   model.EpisodeID(-1),
		} {
			got, err := repo.GetForArchiveByID(context.Background(), id)
			if err != nil {
				t.Fatalf("%s: GetForArchiveByID() error = %v", name, err)
			}
			if got != nil {
				t.Errorf("%s: GetForArchiveByID() = %+v, want nil", name, got)
			}
		}
	})
}
