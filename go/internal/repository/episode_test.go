package repository_test

import (
	"context"
	"database/sql"
	"math"
	"strconv"
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
