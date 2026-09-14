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

// insertEpisodeSyncWorkは最小のworks行を挿入しIDを返す。animeIDは
// works.anime_idマッピングカラムで、未同期の親には無効なNullInt64を渡す。
func insertEpisodeSyncWork(t *testing.T, tx *sql.Tx, animeID sql.NullInt64) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(
		`INSERT INTO works (title, media, anime_id) VALUES ($1, $2, $3) RETURNING id`,
		"親作品", 1, animeID,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// insertEpisodeSyncParentAnimeは同期済みの親作品のanimeに見立てた素のanime行を
// 挿入しIDを返す。
func insertEpisodeSyncParentAnime(t *testing.T, tx *sql.Tx) model.AnimeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("animesの挿入に失敗: %v", err)
	}
	return model.AnimeID(id)
}

// episodeSyncRowはepisodes -> animes同期に関係するepisodesカラムを保持する。
// unpublishedAt / deletedAtは同期がDerivedStatus経由でanime.statusに写像する状態
// カラムで、フィクスチャは3つの状態のいずれにも行を置ける。
type episodeSyncRow struct {
	workID        model.WorkID
	title         sql.NullString
	titleRo       string
	titleEn       string
	number        sql.NullString
	sortNumber    int32
	rawNumber     sql.NullFloat64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
	animeID       sql.NullInt64
}

func insertEpisodeSyncEpisode(t *testing.T, tx *sql.Tx, in episodeSyncRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, title, title_ro, title_en, number, sort_number,
			raw_number, unpublished_at, deleted_at, anime_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		int64(in.workID), in.title, in.titleRo, in.titleEn, in.number, in.sortNumber,
		in.rawNumber, in.unpublishedAt, in.deletedAt, in.animeID,
	).Scan(&id); err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// dbListEpisodeRowはAnnict DBの画面が読むepisodesカラムを保持する。
type dbListEpisodeRow struct {
	workID              model.WorkID
	number              sql.NullString
	rawNumber           sql.NullFloat64
	sortNumber          int32
	title               sql.NullString
	titleRo             string
	titleEn             string
	episodeRecordsCount int32
	unpublishedAt       sql.NullTime
	deletedAt           sql.NullTime
}

func insertDBListEpisode(t *testing.T, tx *sql.Tx, in dbListEpisodeRow) model.EpisodeID {
	t.Helper()
	var id int64
	// created_at / updated_atはRailsが書くのと同じように入れる。フィクスチャの行が
	// 編集フォームの読む版を持つようにするため。
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, number, raw_number, sort_number, title, title_ro, title_en,
			episode_records_count, unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) RETURNING id`,
		int64(in.workID), in.number, in.rawNumber, in.sortNumber, in.title, in.titleRo, in.titleEn,
		in.episodeRecordsCount, in.unpublishedAt, in.deletedAt,
	).Scan(&id); err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertDBListWorkは一覧対象のエピソードを持たせる最小のworks行を挿入する。
func insertDBListWork(t *testing.T, tx *sql.Tx) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(
		`INSERT INTO works (title, media) VALUES ($1, $2) RETURNING id`,
		"一覧対象の作品", 1,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

func TestEpisodeRepository_ListForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 作品のエピソードをsort_number降順で取得する", func(t *testing.T) {
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
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			sortNumber: 2,
		})

		// 別作品のエピソードが一覧に混ざらないこと。
		otherWorkID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     otherWorkID,
			number:     sql.NullString{String: "別作品の第1話", Valid: true},
			sortNumber: 1,
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d、期待値 = 2", len(got))
		}
		if got[0].Number == nil || *got[0].Number != "第2話" {
			t.Errorf("got[0].Number = %v、期待値 = 第2話 (sort_number降順)", got[0].Number)
		}
		// 第2話はtitleとraw_numberをNULLのままにしているため、ゼロ値を指す
		// ポインタではなくnilに写像されること。
		if got[0].Title != nil {
			t.Errorf("got[0].Title = %v、期待値 = nil", got[0].Title)
		}
		if got[0].RawNumber != nil {
			t.Errorf("got[0].RawNumber = %v、期待値 = nil", got[0].RawNumber)
		}

		second := got[1]
		if second.Number == nil || *second.Number != "第1話" {
			t.Errorf("got[1].Number = %v、期待値 = 第1話", second.Number)
		}
		if second.WorkID != workID {
			t.Errorf("got[1].WorkID = %d、期待値 = %d", second.WorkID, workID)
		}
		if second.Title == nil || *second.Title != "はじまり" {
			t.Errorf("got[1].Title = %v、期待値 = はじまり", second.Title)
		}
		if second.TitleRo != "Hajimari" {
			t.Errorf("got[1].TitleRo = %q、期待値 = Hajimari", second.TitleRo)
		}
		if second.TitleEn != "The Beginning" {
			t.Errorf("got[1].TitleEn = %q、期待値 = The Beginning", second.TitleEn)
		}
		if second.RawNumber == nil || *second.RawNumber != 1 {
			t.Errorf("got[1].RawNumber = %v、期待値 = 1", second.RawNumber)
		}
		if second.SortNumber != 1 {
			t.Errorf("got[1].SortNumber = %d、期待値 = 1", second.SortNumber)
		}
		if second.EpisodeRecordsCount != 42 {
			t.Errorf("got[1].EpisodeRecordsCount = %d、期待値 = 42", second.EpisodeRecordsCount)
		}
		if second.DerivedStatus() != model.EpisodeStatusPublished {
			t.Errorf("got[1].DerivedStatus() = %q、期待値 = published", second.DerivedStatus())
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
			})
		}

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d、期待値 = 2", len(firstPage))
		}
		if *firstPage[0].Number != "第3話" || *firstPage[1].Number != "第2話" {
			t.Errorf("firstPage = [%q %q]、期待値 = [第3話 第2話]", *firstPage[0].Number, *firstPage[1].Number)
		}

		secondPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    2,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(secondPage) != 1 {
			t.Fatalf("len(secondPage) = %d、期待値 = 1", len(secondPage))
		}
		if *secondPage[0].Number != "第1話" {
			t.Errorf("secondPage[0].Number = %q、期待値 = 第1話", *secondPage[0].Number)
		}
	})

	t.Run("正常系: sort_numberが同じ行はid降順でページ境界をまたいで取得する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		oldestID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
		})
		middleID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
		})
		newestID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 1,
		})

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB()の1ページ目のエラー = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d、期待値 = 2", len(firstPage))
		}
		wantFirstPage := []model.EpisodeID{newestID, middleID}
		for i, wantID := range wantFirstPage {
			if firstPage[i].ID != wantID {
				t.Errorf("firstPage[%d].ID = %d、期待値 = %d", i, firstPage[i].ID, wantID)
			}
		}

		secondPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    2,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB()の2ページ目のエラー = %v", err)
		}
		if len(secondPage) != 1 {
			t.Fatalf("len(secondPage) = %d、期待値 = 1", len(secondPage))
		}
		if secondPage[0].ID != oldestID {
			t.Errorf("secondPage[0].ID = %d、期待値 = %d", secondPage[0].ID, oldestID)
		}
	})

	t.Run("正常系: 除外と状態はdeleted_at / unpublished_atで決まる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "公開中の話", Valid: true},
			sortNumber: 1,
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:        workID,
			number:        sql.NullString{String: "非公開の話", Valid: true},
			sortNumber:    2,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "deleted_atで削除された話", Valid: true},
			sortNumber: 3,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d、期待値 = 2 (deleted_atの行だけを除外)", len(got))
		}

		wantNumbers := []string{"非公開の話", "公開中の話"}
		for i, want := range wantNumbers {
			if *got[i].Number != want {
				t.Errorf("got[%d].Number = %q、期待値 = %q", i, *got[i].Number, want)
			}
		}

		wantStatuses := []model.EpisodeStatus{
			model.EpisodeStatusArchived,
			model.EpisodeStatusPublished,
		}
		for i, want := range wantStatuses {
			if got[i].DerivedStatus() != want {
				t.Errorf("got[%d].DerivedStatus() = %q、期待値 = %q", i, got[i].DerivedStatus(), want)
			}
		}
	})

	t.Run("正常系: 直前のエピソードをsort_number順の隣接行から導出する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第1話", Valid: true},
			rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
			sortNumber: 100,
		})
		// 削除済みのエピソードはsort_number順で他2話の間に位置する。一覧はこれを
		// 落とすため、導出も飛ばし、ページに出ない行を名指ししないこと。
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "削除済みの話", Valid: true},
			sortNumber: 150,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			number:     sql.NullString{String: "第2話", Valid: true},
			sortNumber: 200,
		})
		// 別作品のエピソードが直前のエピソードになってはならない。
		otherWorkID := insertDBListWork(t, tx)
		insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     otherWorkID,
			number:     sql.NullString{String: "別作品の話", Valid: true},
			sortNumber: 50,
		})

		got, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d、期待値 = 2", len(got))
		}
		if got[0].PrevNumber == nil || *got[0].PrevNumber != "第1話" {
			t.Errorf("got[0].PrevNumber = %v、期待値 = 第1話", got[0].PrevNumber)
		}
		if got[0].PrevRawNumber == nil || *got[0].PrevRawNumber != 1 {
			t.Errorf("got[0].PrevRawNumber = %v、期待値 = 1", got[0].PrevRawNumber)
		}
		if got[1].PrevNumber != nil {
			t.Errorf("got[1].PrevNumber = %v、期待値 = nil (作品の最初のエピソード)", got[1].PrevNumber)
		}
		if got[1].PrevRawNumber != nil {
			t.Errorf("got[1].PrevRawNumber = %v、期待値 = nil (作品の最初のエピソード)", got[1].PrevRawNumber)
		}
	})

	// 導出はページを切り出す前に作品の一覧全体に対して行われるため、ページ末尾の行も
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
			})
		}

		firstPage, err := repo.ListForDB(context.Background(), repository.DBEpisodeListParams{
			WorkID:  workID,
			Page:    1,
			PerPage: 2,
		})
		if err != nil {
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(firstPage) != 2 {
			t.Fatalf("len(firstPage) = %d、期待値 = 2", len(firstPage))
		}
		// 第2話は1ページ目の末尾で、その直前の第1話は2ページ目の先頭に載る。
		if firstPage[1].PrevNumber == nil || *firstPage[1].PrevNumber != "第1話" {
			t.Errorf("firstPage[1].PrevNumber = %v、期待値 = 第1話", firstPage[1].PrevNumber)
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
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d、期待値 = 0", len(got))
		}
	})

	t.Run("境界値: 最大ページ番号でもOFFSETがオーバーフローしない", func(t *testing.T) {
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
			t.Fatalf("ListForDB()のエラー = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d、期待値 = 0", len(got))
		}
	})
}

func TestEpisodeRepository_CountForDB(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

	workID := insertDBListWork(t, tx)
	insertDBListEpisode(t, tx, dbListEpisodeRow{workID: workID, sortNumber: 1})
	insertDBListEpisode(t, tx, dbListEpisodeRow{
		workID:        workID,
		sortNumber:    2,
		unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	// 件数は一覧と同じ絞り込みを使うため、deleted_atの行と別作品のエピソードは
	// 数えない。
	insertDBListEpisode(t, tx, dbListEpisodeRow{
		workID:     workID,
		sortNumber: 3,
		deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
	otherWorkID := insertDBListWork(t, tx)
	insertDBListEpisode(t, tx, dbListEpisodeRow{workID: otherWorkID, sortNumber: 1})

	got, err := repo.CountForDB(context.Background(), workID)
	if err != nil {
		t.Fatalf("CountForDB()のエラー = %v", err)
	}
	if got != 2 {
		t.Errorf("CountForDB() = %d、期待値 = 2", got)
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
		})

		got, err := repo.GetForEditByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForEditByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForEditByID() = nil、期待値 = エピソード")
		}
		if got.Episode.ID != episodeID {
			t.Errorf("Episode.ID = %d、期待値 = %d", int64(got.Episode.ID), int64(episodeID))
		}
		if got.Episode.Number == nil || *got.Episode.Number != "第2話" {
			t.Errorf("Episode.Number = %v、期待値 = %q", got.Episode.Number, "第2話")
		}
		if got.Episode.RawNumber == nil || *got.Episode.RawNumber != 2.5 {
			t.Errorf("Episode.RawNumber = %v、期待値 = 2.5", got.Episode.RawNumber)
		}
		if got.Episode.SortNumber != 200 {
			t.Errorf("Episode.SortNumber = %d、期待値 = 200", got.Episode.SortNumber)
		}
		if got.Episode.Title == nil || *got.Episode.Title != "もう、お婿にいけません" {
			t.Errorf("Episode.Title = %v、期待値 = %q", got.Episode.Title, "もう、お婿にいけません")
		}
		if got.Episode.TitleEn != "No Longer Marriageable" {
			t.Errorf("Episode.TitleEn = %q、期待値 = %q", got.Episode.TitleEn, "No Longer Marriageable")
		}
		// フォームはupdated_atを送信が前提とする版として運ぶため、ローダーが値を
		// 入れる必要がある。
		if got.Episode.UpdatedAt == nil {
			t.Error("Episode.UpdatedAt = nil、期待値 = 更新時刻")
		}
		// ページは見出しに作品のtitleを、共有サブナビにno_episodesを読むため、
		// どちらもエピソードと一緒に返る。
		if got.Work.ID != workID {
			t.Errorf("Work.ID = %d、期待値 = %d", int64(got.Work.ID), int64(workID))
		}
		if got.Work.Title != "編集対象の作品" {
			t.Errorf("Work.Title = %q、期待値 = %q", got.Work.Title, "編集対象の作品")
		}
		if !got.Work.NoEpisodes {
			t.Error("Work.NoEpisodes = false、期待値 = true")
		}
	})

	t.Run("正常系: 未設定の任意カラムはnilになる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("未設定カラムの作品").Build()
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{workID: workID, sortNumber: 100})

		got, err := repo.GetForEditByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForEditByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForEditByID() = nil、期待値 = エピソード")
		}
		if got.Episode.Number != nil {
			t.Errorf("Episode.Number = %v、期待値 = nil", got.Episode.Number)
		}
		if got.Episode.RawNumber != nil {
			t.Errorf("Episode.RawNumber = %v、期待値 = nil", got.Episode.RawNumber)
		}
		if got.Episode.Title != nil {
			t.Errorf("Episode.Title = %v、期待値 = nil", got.Episode.Title)
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
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})
		editableEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 200,
		})

		deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済みの作品").WithDeletedAt(time.Now()).Build()
		episodeOfDeletedWorkID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     deletedWorkID,
			sortNumber: 100,
		})

		for _, tt := range []struct {
			name      string
			episodeID model.EpisodeID
			wantNil   bool
		}{
			{name: "存在しないエピソード", episodeID: model.EpisodeID(999999999), wantNil: true},
			{name: "削除済みのエピソード", episodeID: deletedEpisodeID, wantNil: true},
			{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID, wantNil: true},
			{name: "生きているエピソード", episodeID: editableEpisodeID, wantNil: false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.GetForEditByID(context.Background(), tt.episodeID)
				if err != nil {
					t.Fatalf("GetForEditByID()のエラー = %v", err)
				}
				if tt.wantNil && got != nil {
					t.Errorf("GetForEditByID() = %+v、期待値 = nil", got)
				}
				if !tt.wantNil && got == nil {
					t.Error("GetForEditByID() = nil、期待値 = エピソード")
				}
			})
		}
	})
}

func TestEpisodeRepository_ListForAnimeSyncByIDs(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 全カラムを射影し親anime_idをJOINで解決する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		parentAnimeID := insertEpisodeSyncParentAnime(t, tx)
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(parentAnimeID), Valid: true})

		// エピソードは非公開にされた後に削除された状態にする。削除はdeleted_atを打ち、
		// 非公開が打ったタイムスタンプはそのまま残すため、これが削除後の行の形になる。両方を
		// 立てることで、状態カラム2つの射影を1行で通す。
		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:        workID,
			title:         sql.NullString{String: "第3話タイトル", Valid: true},
			titleRo:       "Episode 3",
			titleEn:       "Episode Three",
			number:        sql.NullString{String: "第3話", Valid: true},
			sortNumber:    3,
			rawNumber:     sql.NullFloat64{Float64: 3.5, Valid: true},
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
			deletedAt:     sql.NullTime{Time: time.Now(), Valid: true},
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs()のエラー = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d、期待値 = 1", len(episodes))
		}
		e := episodes[0]

		if e.ID != episodeID {
			t.Errorf("ID = %d、期待値 = %d", e.ID, episodeID)
		}
		if e.WorkID != workID {
			t.Errorf("WorkID = %d、期待値 = %d", e.WorkID, workID)
		}
		if e.Title == nil || *e.Title != "第3話タイトル" {
			t.Errorf("Title = %v、期待値 = 第3話タイトル", e.Title)
		}
		if e.TitleRo != "Episode 3" {
			t.Errorf("TitleRo = %q、期待値 = Episode 3", e.TitleRo)
		}
		if e.TitleEn != "Episode Three" {
			t.Errorf("TitleEn = %q、期待値 = Episode Three", e.TitleEn)
		}
		if e.Number == nil || *e.Number != "第3話" {
			t.Errorf("Number = %v、期待値 = 第3話", e.Number)
		}
		if e.SortNumber != 3 {
			t.Errorf("SortNumber = %d、期待値 = 3", e.SortNumber)
		}
		if e.RawNumber == nil || *e.RawNumber != 3.5 {
			t.Errorf("RawNumber = %v、期待値 = 3.5", e.RawNumber)
		}
		// 状態のタイムスタンプが一緒に返るのは、同期がこれらからanime.statusを導出する
		// ため。射影から落ちると、すべてのエピソードがpublishedにリコンサイルされてしまう。
		if e.UnpublishedAt == nil {
			t.Error("UnpublishedAt = nil、期待値 = 非公開時刻")
		}
		if e.DeletedAt == nil {
			t.Error("DeletedAt = nil、期待値 = 削除時刻")
		}
		if e.DerivedStatus() != model.EpisodeStatusDeleted {
			t.Errorf("DerivedStatus() = %q、期待値 = %q", e.DerivedStatus(), model.EpisodeStatusDeleted)
		}
		if e.ParentAnimeID == nil || *e.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v、期待値 = %d", e.ParentAnimeID, parentAnimeID)
		}
		// episode自体はまだanimeにマッピングされていない。
		if e.AnimeID != nil {
			t.Errorf("AnimeID = %v、期待値 = nil", e.AnimeID)
		}
	})

	t.Run("正常系: NULL許容カラムはnil、未同期の親はParentAnimeID nil", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		// 親作品はanime_idを持たず (未同期)、episodeはNULL許容の
		// title / number / raw_numberカラムと状態のタイムスタンプ2つをNULLのままにする。
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{})
		episodeID := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{
			workID:     workID,
			sortNumber: 1,
		})

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs()のエラー = %v", err)
		}
		if len(episodes) != 1 {
			t.Fatalf("len(episodes) = %d、期待値 = 1", len(episodes))
		}
		e := episodes[0]

		if e.Title != nil {
			t.Errorf("Title = %v、期待値 = nil", e.Title)
		}
		if e.Number != nil {
			t.Errorf("Number = %v、期待値 = nil", e.Number)
		}
		if e.RawNumber != nil {
			t.Errorf("RawNumber = %v、期待値 = nil", e.RawNumber)
		}
		if e.UnpublishedAt != nil {
			t.Errorf("UnpublishedAt = %v、期待値 = nil", e.UnpublishedAt)
		}
		if e.DeletedAt != nil {
			t.Errorf("DeletedAt = %v、期待値 = nil", e.DeletedAt)
		}
		if e.DerivedStatus() != model.EpisodeStatusPublished {
			t.Errorf("DerivedStatus() = %q、期待値 = %q", e.DerivedStatus(), model.EpisodeStatusPublished)
		}
		if e.ParentAnimeID != nil {
			t.Errorf("ParentAnimeID = %v、期待値 = nil (親が未同期のため)", e.ParentAnimeID)
		}
	})

	t.Run("正常系: 空入力はクエリせず空スライスを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListForAnimeSyncByIDs()のエラー = %v", err)
		}
		if len(episodes) != 0 {
			t.Errorf("len(episodes) = %d、期待値 = 0", len(episodes))
		}
	})
}

// TestEpisodeRepository_Createは一括作成が行ごとに書くINSERTを検証する。一部だけ入力
// された行の任意カラムは空の値ではなくNULLになり、作成が自ら埋める2つのカラム (animeの
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
		t.Fatalf("Create()のエラー = %v", err)
	}

	secondID, err := repo.Create(ctx, repository.CreateEpisodeParams{
		UserID:        userID,
		WorkID:        workID,
		Title:         &title,
		SortNumber:    200,
		PrevEpisodeID: &firstID,
	})
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
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
		t.Errorf("(number, title) = (%q, %q)、期待値 = (%q, %q)", gotNumber.String, gotTitle.String, number, title)
	}
	if gotRawNumber.Float64 != rawNumber {
		t.Errorf("raw_number = %v、期待値 = %v", gotRawNumber, rawNumber)
	}
	if gotSortNumber != 100 {
		t.Errorf("sort_number = %d、期待値 = 100", gotSortNumber)
	}
	if gotAnimeID.Int64 != int64(animeID) {
		t.Errorf("anime_id = %+v、期待値 = %d", gotAnimeID, int64(animeID))
	}
	// 作品の最初の行には直前のエピソードが無い。
	if gotPrevEpisodeID.Valid {
		t.Errorf("prev_episode_id = %+v、期待値 = NULL", gotPrevEpisodeID)
	}

	if err := tx.QueryRow(`
		SELECT number, raw_number, prev_episode_id, anime_id
		FROM episodes
		WHERE id = $1
	`, int64(secondID)).Scan(&gotNumber, &gotRawNumber, &gotPrevEpisodeID, &gotAnimeID); err != nil {
		t.Fatalf("作成されたエピソードの読み込みに失敗: %v", err)
	}

	// 行が空のままにしたカラムはNULLとして保存される。既存のどのエピソードもそれらに
	// 空文字列を持たないため。
	if gotNumber.Valid {
		t.Errorf("number = %+v、期待値 = NULL", gotNumber)
	}
	if gotRawNumber.Valid {
		t.Errorf("raw_number = %+v、期待値 = NULL", gotRawNumber)
	}
	// 未マッピングの作品配下のエピソードはanimeを持たない。
	if gotAnimeID.Valid {
		t.Errorf("anime_id = %+v、期待値 = NULL", gotAnimeID)
	}
	if gotPrevEpisodeID.Int64 != int64(firstID) {
		t.Errorf("prev_episode_id = %+v、期待値 = %d", gotPrevEpisodeID, int64(firstID))
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
		t.Fatalf("DB活動履歴の読み込みに失敗: %v", err)
	}
	if activityUserID != int64(userID) || activityTrackableID != int64(secondID) ||
		activityTrackableType != "Episode" || activityAction != "episodes.create" ||
		activityRootResourceID != int64(workID) || activityRootResourceType != "Work" ||
		activityNewID != secondID.String() {
		t.Errorf("DB活動履歴が作成内容と一致しません")
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
	})

	// episodeがマッピングされるanime (episode自身の同一性の行)。
	episodeAnimeID := insertEpisodeSyncParentAnime(t, tx)

	if err := repo.UpdateAnimeID(context.Background(), episodeID, episodeAnimeID); err != nil {
		t.Fatalf("UpdateAnimeID()のエラー = %v", err)
	}

	episodes, err := repo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
	if err != nil {
		t.Fatalf("ListForAnimeSyncByIDs()のエラー = %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("len(episodes) = %d、期待値 = 1", len(episodes))
	}
	if episodes[0].AnimeID == nil || *episodes[0].AnimeID != episodeAnimeID {
		t.Errorf("AnimeID = %v、期待値 = %d", episodes[0].AnimeID, episodeAnimeID)
	}
}

// TestEpisodeRepository_ListIDsAfterはkeysetページネーションを検証する。works版と
// 同様、他テストが共有テストDBにepisodesをコミットするため、ページ内容の厳密一致では
// なく、他行の有無に依らず成立するkeysetの不変条件 (カーソルより厳密に大きい最初のid・
// 昇順・LIMIT・カーソルの厳密前進) を検証する。
func TestEpisodeRepository_ListIDsAfter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{})
	// id昇順の3件。中間の1件は、最初のページ (limit 2) が満杯になりid3が
	// 2ページ目のカーソルより先に残るために存在させるだけ。
	id1 := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 1})
	insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 2})
	id3 := insertEpisodeSyncEpisode(t, tx, episodeSyncRow{workID: workID, sortNumber: 3})

	t.Run("カーソル直後のidをLIMITどおり1件返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 1)
		if err != nil {
			t.Fatalf("ListIDsAfter()のエラー = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len = %d、期待値 = 1", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d、期待値 = %d", got[0], id1)
		}
	})

	t.Run("昇順かつカーソルより大きいidだけをLIMIT件数まで返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter()のエラー = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len = %d、期待値 = 2", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d、期待値 = %d", got[0], id1)
		}
		if got[0] >= got[1] {
			t.Errorf("取得順 = %v、期待値 = ID昇順", got)
		}
	})

	t.Run("カーソルを進めると重複なく前進する", func(t *testing.T) {
		page1, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter()のエラー = %v", err)
		}
		cursor := page1[len(page1)-1]

		page2, err := repo.ListIDsAfter(ctx, cursor, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter()のエラー = %v", err)
		}
		if len(page2) == 0 {
			t.Fatal("page2が空だった。少なくともid3を期待")
		}
		if page2[0] <= cursor {
			t.Errorf("page2[0] = %d、期待値 = cursor (%d) より大きいID", page2[0], cursor)
		}
	})

	t.Run("全idより大きいカーソルでは空を返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id3+1_000_000_000, 10)
		if err != nil {
			t.Fatalf("ListIDsAfter()のエラー = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d、期待値 = 0", len(got))
		}
	})
}

// dbUpdateEpisodeRowは更新テストの出発点となるepisodesカラムを保持する。フォームが
// 編集するもの、2つのマッピングカラム、および更新が触れてはならない状態のタイムスタンプ。
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
	// nullUpdatedAtはupdated_atをNULLのままにする。カラムが埋まる前に書かれた
	// エピソードが持つ版がこれにあたる。
	nullUpdatedAt bool
}

// insertDBUpdateEpisodeは更新テストが編集するエピソードを挿入する。タイムスタンプはGoの
// 時計ではなくDBの1時間前を使う。NOW() はトランザクション開始時刻のため、テストプロセスが
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
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertDBUpdateWorkは更新テストの親作品を、updated_atを1日前にして挿入する。更新が
// 行うupdated_atの更新を観測できるようにするため。
func insertDBUpdateWork(t *testing.T, tx *sql.Tx, animeID sql.NullInt64) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO works (title, media, anime_id, created_at, updated_at)
		VALUES ($1, 1, $2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day') RETURNING id`,
		"更新対象の作品", animeID,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// storedDBUpdateEpisodeは更新後にアサーションが読み戻す内容。フォームが書くカラム、
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

// readDBUpdateEpisodeVersionはエピソードの保存済みの版を返す。送信が受理されるには、
// この版を名乗る必要がある。
func readDBUpdateEpisodeVersion(t *testing.T, tx *sql.Tx, id model.EpisodeID) *time.Time {
	t.Helper()
	stored := readDBUpdateEpisode(t, tx, id)
	if !stored.updatedAt.Valid {
		return nil
	}
	return &stored.updatedAt.Time
}

// countDBEpisodeUpdateActivitiesは、あるエピソードについてRails側の管理画面が変更履歴と
// して読む行を数える。
func countDBEpisodeUpdateActivities(t *testing.T, tx *sql.Tx, id model.EpisodeID) int {
	t.Helper()
	var count int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM db_activities
		WHERE trackable_type = 'Episode' AND trackable_id = $1 AND action = 'episodes.update'`,
		int64(id),
	).Scan(&count); err != nil {
		t.Fatalf("DB活動履歴の読み込みに失敗: %v", err)
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
			t.Fatalf("GetForUpdateByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForUpdateByID() = nil、期待値 = エピソード")
		}
		if got.ID != episodeID || got.WorkID != workID {
			t.Errorf("(ID, WorkID) = (%d, %d)、期待値 = (%d, %d)", int64(got.ID), int64(got.WorkID), int64(episodeID), int64(workID))
		}
		if got.TitleRo != "Episode 1" {
			t.Errorf("TitleRo = %q、期待値 = %q", got.TitleRo, "Episode 1")
		}
		// 状態のタイムスタンプが一緒に返るのは、animeへの両書きがこれらからanime.statusを
		// 導出するため。これが無いと内容編集がアーカイブ済みのanimeを再公開してしまう。
		if got.UnpublishedAt == nil {
			t.Error("UnpublishedAt = nil、期待値 = 非公開時刻")
		}
		if got.DerivedStatus() != model.EpisodeStatusArchived {
			t.Errorf("DerivedStatus() = %q、期待値 = %q", got.DerivedStatus(), model.EpisodeStatusArchived)
		}
		if got.AnimeID == nil || *got.AnimeID != episodeAnimeID {
			t.Errorf("AnimeID = %v、期待値 = %d", got.AnimeID, int64(episodeAnimeID))
		}
		if got.ParentAnimeID == nil || *got.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v、期待値 = %d", got.ParentAnimeID, int64(parentAnimeID))
		}
	})

	t.Run("正常系: 未マッピングのエピソードは両方のanime_idがnilになる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		episodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})

		got, err := repo.GetForUpdateByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForUpdateByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForUpdateByID() = nil、期待値 = エピソード")
		}
		if got.AnimeID != nil || got.ParentAnimeID != nil {
			t.Errorf("(AnimeID, ParentAnimeID) = (%v, %v)、期待値 = (nil, nil)", got.AnimeID, got.ParentAnimeID)
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
		updatableEpisodeID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 200})

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
			{name: "生きているエピソード", episodeID: updatableEpisodeID, wantNil: false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.GetForUpdateByID(context.Background(), tt.episodeID)
				if err != nil {
					t.Fatalf("GetForUpdateByID()のエラー = %v", err)
				}
				if tt.wantNil && got != nil {
					t.Errorf("GetForUpdateByID() = %+v、期待値 = nil", got)
				}
				if !tt.wantNil && got == nil {
					t.Error("GetForUpdateByID() = nil、期待値 = エピソード")
				}
			})
		}
	})
}

func TestEpisodeRepository_Update(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 送信された値を書き、版を進め、Railsの保存副作用を再現する", func(t *testing.T) {
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false、期待値 = true")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.number.String != number || stored.title.String != title {
			t.Errorf("(number, title) = (%q, %q)、期待値 = (%q, %q)", stored.number.String, stored.title.String, number, title)
		}
		if stored.rawNumber.Float64 != rawNumber {
			t.Errorf("raw_number = %v、期待値 = %v", stored.rawNumber, rawNumber)
		}
		if stored.titleEn != "No Longer Marriageable" {
			t.Errorf("title_en = %q、期待値 = %q", stored.titleEn, "No Longer Marriageable")
		}
		if stored.sortNumber != 200 {
			t.Errorf("sort_number = %d、期待値 = 200", stored.sortNumber)
		}
		// 版は必ず進む必要がある。進まなければ、同じフォームからの2件目の送信が、何も
		// 書かれていないかのように受理されてしまう。
		if !stored.updatedAt.Valid || version == nil || !stored.updatedAt.Time.After(*version) {
			t.Errorf("updated_at = %+v、期待値 = %vより後", stored.updatedAt, version)
		}

		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 1 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 1", got)
		}
		var parameters string
		if err := tx.QueryRow(`
			SELECT parameters::text FROM db_activities
			WHERE trackable_type = 'Episode' AND trackable_id = $1`, int64(episodeID),
		).Scan(&parameters); err != nil {
			t.Fatalf("DB活動履歴の読み込みに失敗: %v", err)
		}
		// Railsは保存前後の行を記録する。管理画面の変更履歴が、その編集で何が置き換わったか
		// を示せるようにするため。
		for _, want := range []string{`"old"`, `"new"`, "はじまり", title} {
			if !strings.Contains(parameters, want) {
				t.Errorf("parametersに%qが含まれていません: %s", want, parameters)
			}
		}

		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if !workTouched {
			t.Error("works.updated_atが更新されていません")
		}
	})

	t.Run("正常系: prev_episode_idを送信されたsort_numberの位置から再計算する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		firstID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})
		secondID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 200})
		// 編集対象のエピソードは作品の先頭から始まるため、直前のエピソードを持たない。
		targetID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 50})
		// 削除済みのエピソードは一覧が示す並びに含まれないため、直前のエピソードにはならない。
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false、期待値 = true")
		}

		stored := readDBUpdateEpisode(t, tx, targetID)
		if stored.prevEpisodeID.Int64 != int64(secondID) {
			t.Errorf("prev_episode_id = %+v、期待値 = %d (sort_number順の直前行)", stored.prevEpisodeID, int64(secondID))
		}
		if stored.prevEpisodeID.Int64 == int64(firstID) {
			t.Error("prev_episode_idがsort_numberの直前ではない行を指しています")
		}
	})

	t.Run("正常系: 内容が変わらない送信では活動履歴も作品のupdated_at更新も行わない", func(t *testing.T) {
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false、期待値 = true")
		}

		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0 (内容が変わっていないため)", got)
		}
		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if workTouched {
			t.Error("内容が変わっていないのにworks.updated_atが更新されています")
		}
	})

	t.Run("正常系: prev_episode_idだけが動いた送信では活動履歴も作品のupdated_at更新も行わない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))
		userID := testutil.NewUserBuilder(t, tx).Build()

		workID := insertDBUpdateWork(t, tx, sql.NullInt64{})
		precedingID := insertDBUpdateEpisode(t, tx, dbUpdateEpisodeRow{workID: workID, sortNumber: 100})
		// 編集対象のエピソードは、先行するエピソードがあるにも関わらずprev_episode_idを
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false、期待値 = true")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.prevEpisodeID.Int64 != int64(precedingID) {
			t.Fatalf("prev_episode_id = %+v、期待値 = %d (再計算されていない)", stored.prevEpisodeID, int64(precedingID))
		}

		// prev_episode_idは入力ではなく並び順から導出されるため、その再計算は編集ではない。
		// 記録すると、編集者が行っていない変更が、Rails側の管理画面が読む変更履歴に載ってしまう。
		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0 (導出列の再計算は編集ではない)", got)
		}
		var workTouched bool
		if err := tx.QueryRow(`SELECT updated_at > created_at FROM works WHERE id = $1`, int64(workID)).Scan(&workTouched); err != nil {
			t.Fatalf("作品の保存副作用の読み込みに失敗: %v", err)
		}
		if workTouched {
			t.Error("導出列の再計算だけでworks.updated_atが更新されています")
		}
	})

	t.Run("異常系: 古い版を名乗る送信は何も書かずにfalseを返す", func(t *testing.T) {
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if updated {
			t.Fatal("Update() = true、期待値 = false")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.title.String != title {
			t.Errorf("title = %q、期待値 = %q (却下された送信は行を書かない)", stored.title.String, title)
		}
		if got := countDBEpisodeUpdateActivities(t, tx, episodeID); got != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0", got)
		}
	})

	t.Run("正常系: updated_atがNULLの行は1回目だけ受理される", func(t *testing.T) {
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

		first := "1回目"
		updated, err := repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &first,
			SortNumber: 100,
			Version:    nil,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("1回目のUpdate() = false、期待値 = true")
		}

		// 最初の書き込みがupdated_atをtimestampへ進めるため、NULLの版を名乗り続ける
		// 2件目はもう一致しない。
		second := "2回目"
		updated, err = repo.Update(context.Background(), repository.UpdateEpisodeParams{
			ID:         episodeID,
			WorkID:     workID,
			Title:      &second,
			SortNumber: 100,
			Version:    nil,
			UserID:     userID,
		})
		if err != nil {
			t.Fatalf("Update()のエラー = %v", err)
		}
		if updated {
			t.Fatal("2回目のUpdate() = true、期待値 = false")
		}

		stored := readDBUpdateEpisode(t, tx, episodeID)
		if stored.title.String != first {
			t.Errorf("title = %q、期待値 = %q", stored.title.String, first)
		}
	})
}

// dbUpdateNeighbourEpisodeは並び順のテストが出発点とする作品の1行を表す。
//
// unlinkedは、先行するエピソードがあるにも関わらずprev_episode_idをNULLのままにする。何も
// 張り替えなければ行が留まる状態である。これにより、書き込みが起きないことを期待するケースが、
// 触られていない行と、書き込みがたまたま同じ値を入れた行とを区別できる。
type dbUpdateNeighbourEpisode struct {
	sortNumber int32
	deleted    bool
	unlinked   bool
}

// insertDBUpdateNeighbourEpisodesは与えられた行を1つの作品の下に挿入し、同じ順序でIDを
// 返す。行は挿入順に連結し、それぞれが直前に挿入された行を名乗るため、作品はprev_episode_idが
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

// assertDBUpdatePrevEpisodeIDsは作品全体を一度に検査する。wantPrevはエピソードごとに、その
// prev_episode_idが名乗るべきエピソードの添字 (NULLなら -1) を保持する。張り替え対象だけでなく
// 全行を読むことで、他の行が触られていないことも固定する。
func assertDBUpdatePrevEpisodeIDs(t *testing.T, tx *sql.Tx, ids []model.EpisodeID, wantPrev []int) {
	t.Helper()
	for i, id := range ids {
		want := sql.NullInt64{}
		if wantPrev[i] >= 0 {
			want = sql.NullInt64{Int64: int64(ids[wantPrev[i]]), Valid: true}
		}
		if got := readDBUpdateEpisode(t, tx, id).prevEpisodeID; got != want {
			t.Errorf("episodes[%d].prev_episode_id = %+v、期待値 = %+v", i, got, want)
		}
	}
}

// TestEpisodeRepository_UpdateRelinksNeighboursは移動したエピソードの前後2行を検証する。
// 編集対象の行だけを再計算すると (TestEpisodeRepository_Update)、移動前に直後だった行はもう前に
// いないエピソードを名乗り、移動後に直後になる行はその1つ前を名乗ったままになる。並び順と
// カラムの食い違いが、編集対象から隣接行へ移るだけで解消しない。
func TestEpisodeRepository_UpdateRelinksNeighbours(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 並び順に応じて隣接する2行を張り替える", func(t *testing.T) {
		t.Parallel()

		for _, tt := range []struct {
			name string
			// episodesは挿入順 (sort_number昇順でもある) に作品を記述する。
			episodes []dbUpdateNeighbourEpisode
			// targetIndexは送信が編集するエピソード、newSortNumberはその移動先。
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
				name: "sort_numberが変わらない送信では隣接行を書かない",
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
				name: "移動前の同値sort_numberはidをタイブレーカに直前行を選ぶ",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 200}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   1,
				newSortNumber: 350,
				wantPrev:      []int{-1, 2, 0, 1},
			},
			{
				name: "移動前後の同値sort_numberはidをタイブレーカに次の行を選ぶ",
				episodes: []dbUpdateNeighbourEpisode{
					{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
				},
				targetIndex:   1,
				newSortNumber: 300,
				wantPrev:      []int{-1, 2, 0, 1, 3},
			},
			{
				// 削除済みのエピソードは一覧が示す並びに含まれないため、張り替えの対象にも、
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
					t.Fatalf("Update()のエラー = %v", err)
				}
				if !updated {
					t.Fatal("Update() = false、期待値 = true")
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if !updated {
			t.Fatal("Update() = false、期待値 = true")
		}
		// 先に張り替えを確認する。張り替え自体が起きていなければ、以降のアサーションは
		// 別の理由で成立してしまう。
		assertDBUpdatePrevEpisodeIDs(t, tx, ids, []int{-1, 2, 0, 1})

		for i, id := range relinked {
			stored := readDBUpdateEpisode(t, tx, id)
			// 隣接行の版を進めると、フォームを開いていた編集者の次の送信が、どのフォームも
			// 送信しないカラムを理由に却下されてしまう。
			if versions[i] == nil || !stored.updatedAt.Valid || !stored.updatedAt.Time.Equal(*versions[i]) {
				t.Errorf("張り替えられた行のupdated_at = %+v、期待値 = %v (進めない)", stored.updatedAt, versions[i])
			}
			// 張り替えは並び順から導出される値の維持のため、記録すると、誰も行っていない
			// 編集がRails側の管理画面が読む変更履歴に載ってしまう。
			if got := countDBEpisodeUpdateActivities(t, tx, id); got != 0 {
				t.Errorf("張り替えられた行のDB活動履歴 = %d件、期待値 = 0", got)
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
			t.Fatalf("Update()のエラー = %v", err)
		}
		if updated {
			t.Fatal("Update() = true、期待値 = false")
		}

		// 編集対象の行は移動していないため、隣接行だけを張り替えると、作品は実際には持たない
		// 並び順を名乗ることになる。
		assertDBUpdatePrevEpisodeIDs(t, tx, ids, []int{-1, 0, 1, 2})
	})
}

// TestEpisodeRepository_UpdateRejectsConcurrentParentDeletionは編集用の事前読み取りと
// 書き込みの境界を再現する。親作品の論理削除が行ロックを保持している間にUpdateを開始し、削除が
// コミットした後は、削除されていない親という条件を再評価してfalseを返し、エピソードを変更
// しない
// ことを検証する。
func TestEpisodeRepository_UpdateRejectsConcurrentParentDeletion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
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
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	// 競合する双方のトランザクションから見える必要があるため、テストデータはコミットする。
	// 両トランザクションの終了後に依存行から明示的に削除する。
	t.Cleanup(func() {
		statements := []string{
			`DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`,
			`DELETE FROM episodes WHERE work_id = $1`,
			`DELETE FROM works WHERE id = $1`,
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("並行更新のテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deleteTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("削除用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = deleteTx.Rollback() }()
	if _, err := deleteTx.ExecContext(ctx, `UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("親作品の論理削除に失敗: %v", err)
	}

	updateTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("更新用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = updateTx.Rollback() }()
	var updateBackendPID int
	if err := updateTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&updateBackendPID); err != nil {
		t.Fatalf("更新側backend PIDの取得に失敗: %v", err)
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
			// 親のガードが更新を却下すれば活動履歴は挿入されないため、このIDが保存済み
			// ユーザーを指す必要はない。
			UserID: model.UserID(1),
		})
		resultCh <- updateResult{updated: updated, err: err}
	}()

	// 削除のコミットを許可する前に、更新が行ロック待ちになったことを観測する。
	// sleepに依存せず決定的な実行順にするためpg_stat_activityをポーリングする。
	lockDeadline := time.NewTimer(5 * time.Second)
	defer lockDeadline.Stop()
	lockTicker := time.NewTicker(10 * time.Millisecond)
	defer lockTicker.Stop()
	waitingForLock := false
	for !waitingForLock {
		select {
		case result := <-resultCh:
			t.Fatalf("親の削除のコミット前に返ったUpdate()の結果 = %+v、期待値 = 親の行ロックを待つこと", result)
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
			t.Fatal("Update()が親の行ロックを待たなかった")
		}
	}

	if err := deleteTx.Commit(); err != nil {
		t.Fatalf("削除用トランザクションのCommit()のエラー = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("親の削除のコミット後にUpdate()を待つcontextのエラー = %v、期待値 = nil (Update()が完了すること)", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("Update()のエラー = %v", result.err)
	}
	if result.updated {
		t.Fatal("親の削除後のUpdate() = true、期待値 = false")
	}
	if err := updateTx.Commit(); err != nil {
		t.Fatalf("更新用トランザクションのCommit()のエラー = %v", err)
	}

	var storedTitle sql.NullString
	if err := db.QueryRow(`SELECT title FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&storedTitle); err != nil {
		t.Fatalf("エピソードの再読み込みに失敗: %v", err)
	}
	if storedTitle.String != "削除前のタイトル" {
		t.Errorf("title = %q、期待値 = %q", storedTitle.String, "削除前のタイトル")
	}
}

// TestEpisodeRepository_UpdateSerializesConcurrentSiblingUpdatesはUpdateが親作品に対して
// 取るロックの強さを固定する。同じ作品のエピソードへの2つの送信はいずれもその1行に触れるため、
// 同一文の中で先に共有・後から排他と取ると、双方が共有ロックを保持したまま相手の共有ロックを
// 待つ状態になる。PostgreSQLはこれを片方のデッドロック中断で解消するが、ハンドラーはそれを500に
// するしかない。updated_atの更新が必要とする強さで最初から取れば、2つ目の送信は1つ目の
// コミットを待つ。
//
// 交差の組み立てには内容が変わらない送信を使う。作品を書かずにロックだけを保持する唯一の方法で
// あるため。この送信はepisode_changeを空にするのでtouched_workは1行も書かず、
// current_episodeが取ったロックだけが保持される。
func TestEpisodeRepository_UpdateSerializesConcurrentSiblingUpdates(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	firstNumber := "#1"
	firstTitle := "第1話"
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
		title:      sql.NullString{String: "第2話", Valid: true},
	})
	firstVersion := readDBUpdateEpisodeVersion(t, setupTx, firstID)
	secondVersion := readDBUpdateEpisodeVersion(t, setupTx, secondID)
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	// 競合する双方のトランザクションから見える必要があるためテストデータはコミットし、後から
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
				t.Logf("並行更新のテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	firstTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("1つ目のトランザクションのBeginTx()のエラー = %v", err)
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
		t.Fatalf("内容が変わらないUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("内容が変わらないUpdate() = false、期待値 = true")
	}
	// 以降の交差はこの送信が作品を書かないことに依存する。変更として記録されていれば、
	// このトランザクションは既に作品を排他で保持しており、2つ目の送信は検証対象のロックではなく
	// そちらに阻まれてしまう。
	if got := countDBEpisodeUpdateActivities(t, firstTx, firstID); got != 0 {
		t.Fatalf("内容が変わらない送信のDB活動履歴 = %d件、期待値 = 0", got)
	}
	revisedVersion := readDBUpdateEpisodeVersion(t, firstTx, firstID)

	secondTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("2つ目のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = secondTx.Rollback() }()
	var secondBackendPID int
	if err := secondTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&secondBackendPID); err != nil {
		t.Fatalf("2つ目の送信のbackend PIDの取得に失敗: %v", err)
	}

	secondTitle := "第2話 (改題)"
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

	// 2つ目の送信が作品の行ロックで停止するまで待ち、1つ目のトランザクションの次の文が、
	// 2つ目が取ったロックを保持したままの状態で走るようにする。sleepに依存せず決定的な実行順に
	// するためpg_stat_activityをポーリングする。
	lockDeadline := time.NewTimer(10 * time.Second)
	defer lockDeadline.Stop()
	lockTicker := time.NewTicker(10 * time.Millisecond)
	defer lockTicker.Stop()
	waitingForLock := false
	for !waitingForLock {
		select {
		case result := <-resultCh:
			t.Fatalf("2つ目のUpdate()が1つ目のコミット前に完了した: %+v", result)
		case <-lockTicker.C:
			var waitEventType sql.NullString
			if err := firstTx.QueryRowContext(ctx, `
				SELECT wait_event_type
				FROM pg_stat_activity
				WHERE pid = $1`, secondBackendPID).Scan(&waitEventType); err != nil {
				t.Fatalf("2つ目の送信の待機状態の取得に失敗: %v", err)
			}
			waitingForLock = waitEventType.Valid && waitEventType.String == "Lock"
		case <-lockDeadline.C:
			t.Fatal("2つ目のUpdate()が作品の行ロックを待たなかった")
		}
	}

	// ここで1つ目のトランザクションが内容の変わる送信を行い、作品を排他で必要とする。
	// 既に同じ強さで保持していればそのまま進む。共有ロックであれば2つ目のトランザクションの
	// 共有ロックを待つことになり、その2つ目は1つ目を待っているため、デッドロックがここで
	// エラーとして現れる。
	revisedTitle := "第1話 (改題)"
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
		t.Fatalf("同一作品の並行更新で1つ目のUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("1つ目のUpdate() = false、期待値 = true")
	}
	if err := firstTx.Commit(); err != nil {
		t.Fatalf("1つ目のトランザクションのCommit()のエラー = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("1つ目のコミット後も2つ目のUpdate()が完了しない: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("同一作品の並行更新で2つ目のUpdate()のエラー = %v", result.err)
	}
	if !result.updated {
		t.Fatal("2つ目のUpdate() = false、期待値 = true")
	}
	if err := secondTx.Commit(); err != nil {
		t.Fatalf("2つ目のトランザクションのCommit()のエラー = %v", err)
	}

	// 直列化しても、どちらの書き込みも失われてはならない。
	for _, want := range []struct {
		id    model.EpisodeID
		title string
	}{{firstID, revisedTitle}, {secondID, secondTitle}} {
		var storedTitle sql.NullString
		if err := db.QueryRow(`SELECT title FROM episodes WHERE id = $1`, int64(want.id)).Scan(&storedTitle); err != nil {
			t.Fatalf("エピソードの再読み込みに失敗: %v", err)
		}
		if storedTitle.String != want.title {
			t.Errorf("title = %q、期待値 = %q", storedTitle.String, want.title)
		}
	}
}

// TestEpisodeRepository_UpdateSerializesConcurrentMovesは、作品ロックを待った送信が、先行更新の
// コミット後の並びから移動前・移動先双方の隣接行を導出することを検証する。待機前の
// スナップショット
// を使うとprev_episode_idに循環が残る。
func TestEpisodeRepository_UpdateSerializesConcurrentMoves(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
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
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("並行移動のテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	firstTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("1つ目のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = firstTx.Rollback() }()
	firstRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(firstTx)

	updated, err := firstRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 450, Version: firstVersion, UserID: userID,
	})
	if err != nil {
		t.Fatalf("1つ目のUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("1つ目のUpdate() = false、期待値 = true")
	}

	secondTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("2つ目のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = secondTx.Rollback() }()
	var secondBackendPID int
	if err := secondTx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&secondBackendPID); err != nil {
		t.Fatalf("2つ目の送信のbackend PIDの取得に失敗: %v", err)
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
			t.Fatalf("2つ目のUpdate()が1つ目のコミット前に完了した: %+v", result)
		case <-lockTicker.C:
			var waitEventType sql.NullString
			if err := firstTx.QueryRowContext(
				ctx,
				"SELECT wait_event_type FROM pg_stat_activity WHERE pid = $1",
				secondBackendPID,
			).Scan(&waitEventType); err != nil {
				t.Fatalf("2つ目の送信の待機状態の取得に失敗: %v", err)
			}
			if waitEventType.Valid && waitEventType.String == "Lock" {
				goto secondIsWaiting
			}
		case <-lockDeadline.C:
			t.Fatal("2つ目のUpdate()が作品の行ロックを待たなかった")
		}
	}

secondIsWaiting:
	if err := firstTx.Commit(); err != nil {
		t.Fatalf("1つ目のトランザクションのCommit()のエラー = %v", err)
	}

	var result updateResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("1つ目のコミット後も2つ目のUpdate()が完了しない: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("2つ目のUpdate()のエラー = %v", result.err)
	}
	if !result.updated {
		t.Fatal("2つ目のUpdate() = false、期待値 = true")
	}
	if err := secondTx.Commit(); err != nil {
		t.Fatalf("2つ目のトランザクションのCommit()のエラー = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("検証用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	assertDBUpdatePrevEpisodeIDs(t, assertTx, ids, []int{2, 3, -1, 0, 1})
}

// TestEpisodeRepository_UpdateBreaksRailsLockOrderCycleは共有DBのロックプロトコルを固定する。
// Railsはepisodes -> works、Goは隣接導出を守るためworks -> episodesの順でロックする。Railsが
// 同じ作品の別エピソードの行ロックを既に保持している場合、GoはNOWAITで失敗してトランザクション
// 全体をロールバックし、Goの再試行前にRailsが作品のupdated_atを更新できるようにする必要がある。
func TestEpisodeRepository_UpdateBreaksRailsLockOrderCycle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("Rails順ロックのテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	railsTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Rails側のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = railsTx.Rollback() }()
	if _, err := railsTx.ExecContext(ctx, "UPDATE episodes SET title = title WHERE id = $1", int64(ids[2])); err != nil {
		t.Fatalf("Railsの保存順でのエピソードの行ロックに失敗: %v", err)
	}

	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go側のトランザクションのBeginTx()のエラー = %v", err)
	}
	// 作品ロックをテストが必要とする時点で解放するのは下の明示的なRollbackである。この
	// deferはその前にアサーションが落ちた場合だけを受け持つ。これが無いとロックがテストより
	// 長く残り、t.CleanupのDELETEが永久に待つため、失敗したアサーションがスイートのハングに
	// 変わってしまう。
	defer func() { _ = goTx.Rollback() }()
	goRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	updated, err := goRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if updated {
		t.Fatal("Railsが同じ作品の別エピソードの行ロックを保持中のUpdate() = true、期待値 = false")
	}
	if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
		t.Fatalf("Railsが同じ作品の別エピソードの行ロックを保持中のUpdate()のエラー = %v、期待値 = ErrEpisodeLockUnavailable", err)
	}
	if err := goTx.Rollback(); err != nil {
		t.Fatalf("Go側のトランザクションのRollback()のエラー = %v", err)
	}

	// これはRailsの保存順序におけるbelongs_to :work, touch: true側。失敗したGoの試行が
	// 作品ロックを解放した後にだけ進める。
	if _, err := railsTx.ExecContext(ctx, "UPDATE works SET updated_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("Railsの保存順での作品のupdated_at更新に失敗: %v", err)
	}
	if err := railsTx.Commit(); err != nil {
		t.Fatalf("Rails側のトランザクションのCommit()のエラー = %v", err)
	}

	retryTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("リトライ用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = retryTx.Rollback() }()
	retryRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(retryTx)
	updated, err = retryRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("Rails側のコミット後のUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("Rails側のコミット後のUpdate() = false、期待値 = true")
	}
	if err := retryTx.Commit(); err != nil {
		t.Fatalf("リトライ用トランザクションのCommit()のエラー = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("検証用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	assertDBUpdatePrevEpisodeIDs(t, assertTx, ids, []int{-1, 2, 0, 1})
}

// TestEpisodeRepository_UpdateBreaksRailsDestinationPredecessorDeleteCycleは、移動後の
// prev_episode_idから参照する行を検証する。Railsのdestroyはそのエピソードをロックしてから
// 作品のupdated_atを更新し、Goは作品をロックしてから移動先の直前行を導出する。外部キーの検査が
// 循環を完成させないよう、Goはその直前行もNOWAITの対象に含める必要がある。Goのロールバック後は
// Railsが作品のupdated_atを更新してコミットでき、再試行は削除済み行を除いて直前行を導出する。
func TestEpisodeRepository_UpdateBreaksRailsDestinationPredecessorDeleteCycle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := insertDBUpdateWork(t, setupTx, sql.NullInt64{})
	ids := insertDBUpdateNeighbourEpisodes(t, setupTx, workID, []dbUpdateNeighbourEpisode{
		{sortNumber: 100}, {sortNumber: 200}, {sortNumber: 300}, {sortNumber: 400},
	})
	version := readDBUpdateEpisodeVersion(t, setupTx, ids[1])
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("移動先直前行の削除のテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	railsTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Rails側のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = railsTx.Rollback() }()
	if _, err := railsTx.ExecContext(ctx, "DELETE FROM episodes WHERE id = $1", int64(ids[3])); err != nil {
		t.Fatalf("Rails順の移動先の直前エピソードの削除に失敗: %v", err)
	}

	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go側のトランザクションのBeginTx()のエラー = %v", err)
	}
	// 下の明示的なRollbackがRailsのupdated_at更新前に作品ロックを解放する。このdeferはその前に
	// アサーションが落ちた場合を受け持ち、後始末がロック待ちで止まるのを防ぐ。
	defer func() { _ = goTx.Rollback() }()
	goRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	attemptCtx, cancelAttempt := context.WithTimeout(ctx, 2*time.Second)
	updated, err := goRepo.Update(attemptCtx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 500, Version: version, UserID: userID,
	})
	cancelAttempt()
	if updated {
		t.Fatal("Railsが移動先の直前エピソードを削除中のUpdate() = true、期待値 = false")
	}
	if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
		t.Fatalf(
			"Railsが移動先の直前エピソードを削除中のUpdate()のエラー = %v、期待値 = ErrEpisodeLockUnavailable",
			err,
		)
	}
	if err := goTx.Rollback(); err != nil {
		t.Fatalf("Go側のトランザクションのRollback()のエラー = %v", err)
	}

	// これはRailsのdestroy順序におけるbelongs_to :work, touch: true側。NOWAITの失敗で
	// Goの作品ロックは解放済みなので、この書き込みと削除は循環せずコミットできる。
	if _, err := railsTx.ExecContext(ctx, "UPDATE works SET updated_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("Railsの保存順での作品のupdated_at更新に失敗: %v", err)
	}
	if err := railsTx.Commit(); err != nil {
		t.Fatalf("Rails側のトランザクションのCommit()のエラー = %v", err)
	}

	retryTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("リトライ用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = retryTx.Rollback() }()
	retryRepo := repository.NewEpisodeRepository(query.New(db)).WithTx(retryTx)
	updated, err = retryRepo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 500, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("Rails側のコミット後のUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("Rails側のコミット後のUpdate() = false、期待値 = true")
	}
	if err := retryTx.Commit(); err != nil {
		t.Fatalf("リトライ用トランザクションのCommit()のエラー = %v", err)
	}

	assertTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("検証用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = assertTx.Rollback() }()
	wantTargetPrev := sql.NullInt64{Int64: int64(ids[2]), Valid: true}
	if got := readDBUpdateEpisode(t, assertTx, ids[1]).prevEpisodeID; got != wantTargetPrev {
		t.Errorf("移動したepisodeのprev_episode_id = %+v、期待値 = %+v", got, wantTargetPrev)
	}
	wantFormerFollowingPrev := sql.NullInt64{Int64: int64(ids[0]), Valid: true}
	if got := readDBUpdateEpisode(t, assertTx, ids[2]).prevEpisodeID; got != wantFormerFollowingPrev {
		t.Errorf("移動前の直後episodeのprev_episode_id = %+v、期待値 = %+v", got, wantFormerFollowingPrev)
	}
	var deletedReferenceCount int
	if err := assertTx.QueryRow(
		"SELECT COUNT(*) FROM episodes WHERE work_id = $1 AND prev_episode_id = $2",
		int64(workID), int64(ids[3]),
	).Scan(&deletedReferenceCount); err != nil {
		t.Fatalf("削除済みepisodeへの参照数の取得に失敗: %v", err)
	}
	if deletedReferenceCount != 0 {
		t.Errorf("削除済みepisodeを指すprev_episode_id = %d件、期待値 = 0", deletedReferenceCount)
	}
}

// TestEpisodeRepository_UpdateIgnoresLocksOnNonNeighboursは、1回の編集が何を待ち得るかを
// 限定する。Railsは並び順と無関係な書き込みでもエピソード行をロックする。EpisodeRecordの
// counter_culture :episodeにより、ユーザーが記録を作るたびにそのエピソードをトランザクション
// の間ずっと保持する。更新が作品全体を先取りしていると、放送中の作品のどこかで作られた記録が
// 無関係な編集を中断させ、再試行を使い切ると編集者へ失敗として報告されてしまう。
func TestEpisodeRepository_UpdateIgnoresLocksOnNonNeighbours(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("テストデータ用トランザクションのBegin()のエラー = %v", err)
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
		t.Fatalf("テストデータ用トランザクションのCommit()のエラー = %v", err)
	}

	t.Cleanup(func() {
		statements := []string{
			"DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1",
			"DELETE FROM episodes WHERE work_id = $1",
			"DELETE FROM works WHERE id = $1",
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement, int64(workID)); err != nil {
				t.Logf("非隣接ロックのテストデータの後始末に失敗 (%s): %v", statement, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// これがcounter_cultureの書き込みの形である。並び順のカラムは変えないが、それでも
	// トランザクションの間ずっとその行を保持する。
	recordTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("記録トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = recordTx.Rollback() }()
	if _, err := recordTx.ExecContext(
		ctx,
		"UPDATE episodes SET episode_records_count = episode_records_count + 1 WHERE id = $1",
		int64(ids[5]),
	); err != nil {
		t.Fatalf("隣接行ではないエピソードのロック取得に失敗: %v", err)
	}

	// ids[1] を350へ動かすとids[0] / ids[2] / ids[3] が隣接行になり、ids[4] とids[5] は
	// 更新が読み書きする範囲の外に残る。
	goTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Go側のトランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = goTx.Rollback() }()
	repo := repository.NewEpisodeRepository(query.New(db)).WithTx(goTx)
	updated, err := repo.Update(ctx, repository.UpdateEpisodeParams{
		ID: ids[1], WorkID: workID, SortNumber: 350, Version: version, UserID: userID,
	})
	if err != nil {
		t.Fatalf("隣接行ではない行がロックされている状態のUpdate()のエラー = %v", err)
	}
	if !updated {
		t.Fatal("隣接行ではない行がロックされている状態のUpdate() = false、期待値 = true")
	}

	assertDBUpdatePrevEpisodeIDs(t, goTx, ids, []int{-1, 2, 0, 1, 3, 4})
}

// TestEpisodeRepository_UpdateRejectsMovedParentは、編集用の事前読み取りが観測した親作品に
// 対するガードを検証する。RailsのAnnict::DataCare::MoveEpisodeはepisodes.work_idを
// update_columnで書き、作品ロックを取らないため、フォームを開いてから送信までの間に対象が別の
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
		t.Fatalf("Update()のエラー = %v", err)
	}
	if updated {
		t.Fatal("別作品へ移された行へのUpdate() = true、期待値 = false")
	}

	stored := readDBUpdateEpisode(t, tx, episodeID)
	if stored.title.Valid && stored.title.String == title {
		t.Errorf("episodes.title = %q、期待値 = 送信前の値のまま", stored.title.String)
	}
	if stored.sortNumber != 100 {
		t.Errorf("episodes.sort_number = %d、期待値 = 100", stored.sortNumber)
	}
}

// dbArchiveEpisodeRowは非公開・再公開テストがエピソードの初期値として与える写像と状態
// タイムスタンプを保持する。AnimeIDにより、正常系はArchive / Unarchiveが更新した行の写像を
// 返すことを検証できる。
type dbArchiveEpisodeRow struct {
	workID        model.WorkID
	animeID       sql.NullInt64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
}

// insertDBArchiveWorkは非公開・再公開・削除テストの親作品を、3者が動かすカウンター
// キャッシュ付きで、updated_atを1日前にして挿入する。3者が行うupdated_atの更新を観測
// できるようにするため。
func insertDBArchiveWork(t *testing.T, tx *sql.Tx, episodesCount int32) model.WorkID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO works (title, media, episodes_count, created_at, updated_at)
		VALUES ($1, 1, $2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day') RETURNING id`,
		"非公開対象の作品", episodesCount,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

// insertDBArchiveEpisodeは非公開テストが非公開にする、または再公開テストが公開に戻す
// エピソードを挿入する。タイムスタンプはinsertDBUpdateEpisodeが述べる理由により、Goの時計では
// なくDBの1時間前を使う。NOW() はトランザクション開始時刻のため、テストプロセスが打刻した
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
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// storedDBArchiveEpisodeは非公開・再公開テストが読み戻すエピソードの状態。両方向が書く
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

// storedDBArchiveWorkは非公開・再公開・削除テストが読み戻す親作品の状態。Rails APIが
// 配信するカウンターキャッシュと、Railsがbelongs_to :work, touch: trueで進めるタイム
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

// countDBEpisodeActivitiesは、あるエピソードについてRails側の管理画面が変更履歴として
// 読む行を、actionを問わず数える。Railsの非公開・再公開はいずれも素の
// update(unpublished_at:) で1件も記録しないため、どちらの方向も件数を変えてはならない。
func countDBEpisodeActivities(t *testing.T, tx *sql.Tx, id model.EpisodeID) int {
	t.Helper()
	var count int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM db_activities WHERE trackable_type = 'Episode' AND trackable_id = $1`,
		int64(id),
	).Scan(&count); err != nil {
		t.Fatalf("DB活動履歴の読み込みに失敗: %v", err)
	}
	return count
}

func TestEpisodeRepository_Archive(t *testing.T) {
	t.Parallel()

	// 公開中のエピソードが非公開になり、版もそれと併せて進む。親作品にはRailsの非公開が
	// 持つ2つの副作用が現れる。カウンターキャッシュはもう数えない行を失い、作品のupdated_atは
	// 更新される。
	// 変更履歴は記録されない。save_and_create_activity! を通らないupdate(unpublished_at:) と
	// 揃えるため。
	t.Run("正常系: 公開中のエピソードを非公開にし、作品のカウンターと更新時刻を動かす", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		var animeID int64
		if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&animeID); err != nil {
			t.Fatalf("animeの挿入に失敗: %v", err)
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
			t.Fatalf("Archive()のエラー = %v", err)
		}
		if result == nil {
			t.Fatal("Archive()の戻り値 = nil、期待値 = 結果あり")
		}
		if result.AnimeID == nil || *result.AnimeID != model.AnimeID(animeID) {
			t.Errorf("Archive().AnimeID = %v、期待値 = %d", result.AnimeID, animeID)
		}

		stored := readDBArchiveEpisode(t, tx, episodeID)
		if !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻")
		}
		if !stored.updatedAt.Valid || !stored.updatedAt.Time.After(before.updatedAt.Time) {
			t.Errorf("episodes.updated_at = %v、期待値 = %vより後", stored.updatedAt, before.updatedAt.Time)
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 2 {
			t.Errorf("works.episodes_count = %d、期待値 = 2", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v、期待値 = %vより後", work.updatedAt, workBefore.updatedAt)
		}
		if count := countDBEpisodeActivities(t, tx, episodeID); count != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0件", count)
		}
	})

	// 他者が先に非公開にした後の確認ページからの送信は、公開中の行を見つけない。書き込まず
	// それを報告することで、1回の遷移に対してカウンターが2度減算されるのを防ぐ。
	t.Run("非公開済みのエピソードはnilを返し、カウンターを動かさない", func(t *testing.T) {
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
			t.Fatalf("Archive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("非公開済みのエピソードへのArchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 削除済みのエピソードは確認ページの対象外のため、その送信も拒否する。何も表示しない行に
	// unpublished_atを打たないようにするため。
	t.Run("削除済みのエピソードはnilを返す", func(t *testing.T) {
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
			t.Fatalf("Archive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済みのエピソードへのArchive()の戻り値 = 結果あり、期待値 = nil")
		}
	})

	// 論理削除済みの親作品は、公開中のepisode行が残っていても管理画面の一覧対象外。
	// 再公開方向と同じく、episodeの更新前に拒否し、状態とカウンターの両方を変えないことを検証する。
	t.Run("削除済み作品のエピソードはnilを返し、非公開にしない", func(t *testing.T) {
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
			t.Fatalf("Archive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済み作品のエピソードへのArchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_atに値が入った、期待値 = NULLのまま")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 確認ページと送信の間に別作品へ移されたエピソードは、そのページが数えていた作品にもう
	// 属していない。したがってエピソードもどちらの作品も書かない。
	t.Run("別作品へ移されたエピソードはnilを返し、元の作品のカウンターを動かさない", func(t *testing.T) {
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
			t.Fatalf("Archive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("別作品へ移された行へのArchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, originalWorkID); work.episodesCount != 3 {
			t.Errorf("移動元のworks.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
		if work := readDBArchiveWork(t, tx, movedWorkID); work.episodesCount != 1 {
			t.Errorf("移動先のworks.episodes_count = %d、期待値 = 1", work.episodesCount)
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_atに値が入った、期待値 = NULLのまま")
		}
	})
}

func TestEpisodeRepository_Unarchive(t *testing.T) {
	t.Parallel()

	// 非公開のエピソードが再び公開になり、版もそれと併せて進む。親作品にはRailsの再公開が
	// 持つ2つの副作用が現れる。カウンターキャッシュは再び数える行を取り戻し、作品の
	// updated_atは更新される。変更履歴は記録されない。save_and_create_activity! を通らない
	// update(unpublished_at: nil) と揃えるため。
	t.Run("正常系: 非公開のエピソードを公開に戻し、作品のカウンターと更新時刻を動かす", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 2)
		var animeID int64
		if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('archived') RETURNING id`).Scan(&animeID); err != nil {
			t.Fatalf("animeの挿入に失敗: %v", err)
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
			t.Fatalf("Unarchive()のエラー = %v", err)
		}
		if result == nil {
			t.Fatal("Unarchive() = nil、期待値 = 結果")
		}
		if result.AnimeID == nil || *result.AnimeID != model.AnimeID(animeID) {
			t.Errorf("Unarchive().AnimeID = %v、期待値 = %d", result.AnimeID, animeID)
		}

		stored := readDBArchiveEpisode(t, tx, episodeID)
		if stored.unpublishedAt.Valid {
			t.Errorf("episodes.unpublished_at = %v、期待値 = NULL", stored.unpublishedAt.Time)
		}
		if !stored.updatedAt.Valid || !stored.updatedAt.Time.After(before.updatedAt.Time) {
			t.Errorf("episodes.updated_at = %v、期待値 = %vより後", stored.updatedAt, before.updatedAt.Time)
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v、期待値 = %vより後", work.updatedAt, workBefore.updatedAt)
		}
		if count := countDBEpisodeActivities(t, tx, episodeID); count != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0件", count)
		}
	})

	// 他者が先に再公開した後の一覧からの送信は、非公開の行を見つけない。書き込まずそれを
	// 報告することで、1回の遷移に対してカウンターが2度加算されるのを防ぐ。
	t.Run("公開中のエピソードはnilを返し、カウンターを動かさない", func(t *testing.T) {
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
			t.Fatalf("Unarchive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("公開中のエピソードへのUnarchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 削除済みのエピソードは再公開の送信元である一覧の対象外のため、その送信も拒否する。
	// 何も表示しない行を公開しないようにするため。
	t.Run("削除済みのエピソードはnilを返す", func(t *testing.T) {
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
			t.Fatalf("Unarchive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済みのエピソードへのUnarchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻のまま")
		}
	})

	// 論理削除済みの親作品は、非公開のepisode行が残っていても管理画面の一覧対象外。
	// episodeの更新前に拒否し、状態とカウンターの両方を変えないことを検証する。
	t.Run("削除済み作品のエピソードはnilを返し、再公開しない", func(t *testing.T) {
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
			t.Fatalf("Unarchive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済み作品のエピソードへのUnarchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻のまま")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 一覧と送信の間に別作品へ移されたエピソードは、その一覧が数えていた作品にもう属して
	// いない。したがってエピソードもどちらの作品も書かない。
	t.Run("別作品へ移されたエピソードはnilを返し、元の作品のカウンターを動かさない", func(t *testing.T) {
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
			t.Fatalf("Unarchive()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("別作品へ移された行へのUnarchive()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, originalWorkID); work.episodesCount != 3 {
			t.Errorf("移動元のworks.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
		if work := readDBArchiveWork(t, tx, movedWorkID); work.episodesCount != 1 {
			t.Errorf("移動先のworks.episodes_count = %d、期待値 = 1", work.episodesCount)
		}
		if stored := readDBArchiveEpisode(t, tx, episodeID); !stored.unpublishedAt.Valid {
			t.Error("episodes.unpublished_atがNULLになった、期待値 = 非公開の時刻のまま")
		}
	})
}

func TestEpisodeRepository_GetForArchiveByID(t *testing.T) {
	t.Parallel()

	// ローダーは確認ページが表示するもの (エピソードの話数とタイトル、親作品のタイトルと
	// no_episodes) と、送信が検査する状態タイムスタンプを射影する。animeの写像は代わりに
	// ArchiveDBEpisodeが更新した行から得る。
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
		})

		got, err := repo.GetForArchiveByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForArchiveByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForArchiveByID() = nil、期待値 = エピソード")
		}
		if got.Episode.Number == nil || *got.Episode.Number != "第2話" {
			t.Errorf("Episode.Number = %v、期待値 = %q", got.Episode.Number, "第2話")
		}
		if got.Episode.Title == nil || *got.Episode.Title != "もう、お婿にいけません" {
			t.Errorf("Episode.Title = %v、期待値 = %q", got.Episode.Title, "もう、お婿にいけません")
		}

		// 状態カラムは呼び出し側がstatusを導出できるよう運ばれる。公開中のエピソードは
		// どちらも持たない。
		if got.Episode.DerivedStatus() != model.EpisodeStatusPublished {
			t.Errorf("Episode.DerivedStatus() = %q、期待値 = %q", got.Episode.DerivedStatus(), model.EpisodeStatusPublished)
		}
		if got.Work.ID != workID {
			t.Errorf("Work.ID = %d、期待値 = %d", int64(got.Work.ID), int64(workID))
		}
		if got.Work.Title != "非公開確認の作品" {
			t.Errorf("Work.Title = %q、期待値 = %q", got.Work.Title, "非公開確認の作品")
		}
		if !got.Work.NoEpisodes {
			t.Error("Work.NoEpisodes = false、期待値 = true")
		}
	})

	// 非公開のエピソードも返す。状態はタイムスタンプで報告し、その扱いは呼び出し側が決める。
	// これにより同じローダーを後の再公開の経路でも使える。
	t.Run("非公開のエピソードはarchivedとして返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:        workID,
			sortNumber:    100,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		got, err := repo.GetForArchiveByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForArchiveByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForArchiveByID() = nil、期待値 = エピソード")
		}
		if got.Episode.DerivedStatus() != model.EpisodeStatusArchived {
			t.Errorf("Episode.DerivedStatus() = %q、期待値 = %q", got.Episode.DerivedStatus(), model.EpisodeStatusArchived)
		}
	})

	// 除外条件は編集フォームのローダーと揃える。非公開の確認ページに到達できるエピソードを、
	// 編集フォームに到達できるエピソードと一致させるため。
	t.Run("削除済みのエピソードと削除済み作品のエピソードはnilを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		keptWorkID := insertDBListWork(t, tx)
		deletedEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     keptWorkID,
			sortNumber: 100,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		deletedWorkID := testutil.NewWorkBuilder(t, tx).WithDeletedAt(time.Now()).Build()
		orphanEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     deletedWorkID,
			sortNumber: 100,
		})

		for name, id := range map[string]model.EpisodeID{
			"削除済みのエピソード":   deletedEpisodeID,
			"削除済み作品のエピソード": orphanEpisodeID,
			"存在しないエピソード":   model.EpisodeID(-1),
		} {
			got, err := repo.GetForArchiveByID(context.Background(), id)
			if err != nil {
				t.Fatalf("%s: GetForArchiveByID()のエラー = %v", name, err)
			}
			if got != nil {
				t.Errorf("%s: GetForArchiveByID() = %+v、期待値 = nil", name, got)
			}
		}
	})
}

// dbDeleteEpisodeRowは削除テストがエピソードの初期値として与える写像・状態タイムスタンプ・
// 直前エピソードのポインタを保持する。AnimeIDにより、正常系はDeleteが更新した行の写像を返す
// ことを検証できる。PrevEpisodeIDは、ある行を別の行の後続行にする (ステートメントがクリアする
// 対象)。
type dbDeleteEpisodeRow struct {
	workID        model.WorkID
	animeID       sql.NullInt64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
	prevEpisodeID sql.NullInt64
}

// insertDBDeleteEpisodeは削除テストが削除するエピソード、またはその周辺の行を挿入する。
// タイムスタンプはinsertDBUpdateEpisodeが述べる理由により、Goの時計ではなくDBの1時間前を
// 使う。
func insertDBDeleteEpisode(t *testing.T, tx *sql.Tx, in dbDeleteEpisodeRow) model.EpisodeID {
	t.Helper()
	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (
			work_id, number, sort_number, title, anime_id, unpublished_at, deleted_at,
			prev_episode_id, created_at, updated_at
		) VALUES (
			$1, '第1話', 100, '教えてティーチャー', $2, $3, $4,
			$5, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour'
		) RETURNING id`,
		int64(in.workID), in.animeID, in.unpublishedAt, in.deletedAt, in.prevEpisodeID,
	).Scan(&id); err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// storedDBDeleteEpisodeは削除テストが読み戻すエピソードの状態。削除が書く状態カラム、
// それと併せて進める版、そして削除するエピソードを名乗る行に対してクリアする直前エピソードの
// ポインタ。
type storedDBDeleteEpisode struct {
	deletedAt     sql.NullTime
	updatedAt     sql.NullTime
	prevEpisodeID sql.NullInt64
}

func readDBDeleteEpisode(t *testing.T, tx *sql.Tx, id model.EpisodeID) storedDBDeleteEpisode {
	t.Helper()
	var row storedDBDeleteEpisode
	if err := tx.QueryRow(
		`SELECT deleted_at, updated_at, prev_episode_id FROM episodes WHERE id = $1`, int64(id),
	).Scan(&row.deletedAt, &row.updatedAt, &row.prevEpisodeID); err != nil {
		t.Fatalf("削除後のエピソードの読み込みに失敗: %v", err)
	}
	return row
}

func TestEpisodeRepository_Delete(t *testing.T) {
	t.Parallel()

	// 公開中のエピソードがソフトデリートされ、版もそれと併せて進む。親作品にはRailsの削除
	// が持つ2つの副作用が現れる。カウンターキャッシュはもう数えない行を失い、作品のupdated_atは
	// 更新される。変更履歴は記録されない。destroy_in_batchesもsave_and_create_activity! を
	// 通らないため。
	t.Run("正常系: 公開中のエピソードを削除し、作品のカウンターと更新時刻を動かす", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		var animeID int64
		if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&animeID); err != nil {
			t.Fatalf("animeの挿入に失敗: %v", err)
		}
		episodeID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:  workID,
			animeID: sql.NullInt64{Int64: animeID, Valid: true},
		})
		before := readDBDeleteEpisode(t, tx, episodeID)
		workBefore := readDBArchiveWork(t, tx, workID)

		result, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}
		if result == nil {
			t.Fatal("Delete()の戻り値 = nil、期待値 = 結果あり")
		}
		if result.AnimeID == nil || *result.AnimeID != model.AnimeID(animeID) {
			t.Errorf("Delete().AnimeID = %v、期待値 = %d", result.AnimeID, animeID)
		}

		stored := readDBDeleteEpisode(t, tx, episodeID)
		if !stored.deletedAt.Valid {
			t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
		}
		if !stored.updatedAt.Valid || !stored.updatedAt.Time.After(before.updatedAt.Time) {
			t.Errorf("episodes.updated_at = %v、期待値 = %vより後", stored.updatedAt, before.updatedAt.Time)
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 2 {
			t.Errorf("works.episodes_count = %d、期待値 = 2", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v、期待値 = %vより後", work.updatedAt, workBefore.updatedAt)
		}
		if count := countDBEpisodeActivities(t, tx, episodeID); count != 0 {
			t.Errorf("DB活動履歴 = %d件、期待値 = 0件", count)
		}
	})

	// 非公開のエピソードは非公開の時点で既にカウンターから外れているため、削除で2度目の
	// 減算をしてはならない。作品のupdated_atの更新は行う。Railsの削除は、エピソードの公開状態に
	// 関わらず作品のupdated_atを更新するため。
	t.Run("非公開のエピソードの削除ではカウンターを減らさず、作品のupdated_atは更新する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})
		workBefore := readDBArchiveWork(t, tx, workID)

		result, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}
		if result == nil {
			t.Fatal("Delete()の戻り値 = nil、期待値 = 結果あり")
		}
		if stored := readDBDeleteEpisode(t, tx, episodeID); !stored.deletedAt.Valid {
			t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
		}

		work := readDBArchiveWork(t, tx, workID)
		if work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
		if !work.updatedAt.After(workBefore.updatedAt) {
			t.Errorf("works.updated_at = %v、期待値 = %vより後", work.updatedAt, workBefore.updatedAt)
		}
	})

	// Goの削除は行を残すため、削除するエピソードを直前として名乗る未削除の行は、公開側が
	// 見せてはならないものを指し続けることになる。それらをすべてクリアする。再公開されるとポインタ
	// を表に戻してしまう非公開の行も含む。一方、削除済みの後続行と、別の行を名乗る行には触れない。
	// 後続行を削除するエピソード自身の直前行へ張り替えることはしない (Railsの削除も行わない)。
	// 後続行の版も動かさない。
	t.Run("自分を指す未削除の行のポインタをクリアし、削除済みの行と他を指す行は触らない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 5)
		predecessorID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{workID: workID})
		targetID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			prevEpisodeID: sql.NullInt64{Int64: int64(predecessorID), Valid: true},
		})
		targetPointer := sql.NullInt64{Int64: int64(targetID), Valid: true}
		publishedFollowerID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			prevEpisodeID: targetPointer,
		})
		archivedFollowerID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
			prevEpisodeID: targetPointer,
		})
		deletedFollowerID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			deletedAt:     sql.NullTime{Time: time.Now(), Valid: true},
			prevEpisodeID: targetPointer,
		})
		otherFollowerID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:        workID,
			prevEpisodeID: sql.NullInt64{Int64: int64(predecessorID), Valid: true},
		})
		otherFollowerBefore := readDBDeleteEpisode(t, tx, otherFollowerID)
		publishedFollowerBefore := readDBDeleteEpisode(t, tx, publishedFollowerID)

		if _, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     targetID,
			WorkID: workID,
		}); err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}

		for name, id := range map[string]model.EpisodeID{
			"公開中の後続行": publishedFollowerID,
			"非公開の後続行": archivedFollowerID,
		} {
			if stored := readDBDeleteEpisode(t, tx, id); stored.prevEpisodeID.Valid {
				t.Errorf("%sのprev_episode_id = %d、期待値 = NULL", name, stored.prevEpisodeID.Int64)
			}
		}
		if stored := readDBDeleteEpisode(t, tx, deletedFollowerID); stored.prevEpisodeID.Int64 != int64(targetID) {
			t.Errorf("削除済みの後続行のprev_episode_id = %v、期待値 = %dのまま", stored.prevEpisodeID, int64(targetID))
		}
		if stored := readDBDeleteEpisode(t, tx, otherFollowerID); stored.prevEpisodeID.Int64 != int64(predecessorID) {
			t.Errorf("他を指す行のprev_episode_id = %v、期待値 = %dのまま", stored.prevEpisodeID, int64(predecessorID))
		}
		if stored := readDBDeleteEpisode(t, tx, otherFollowerID); !stored.updatedAt.Time.Equal(otherFollowerBefore.updatedAt.Time) {
			t.Errorf("他を指す行のupdated_at = %v、期待値 = %vのまま", stored.updatedAt, otherFollowerBefore.updatedAt)
		}
		if stored := readDBDeleteEpisode(t, tx, publishedFollowerID); !stored.updatedAt.Time.Equal(publishedFollowerBefore.updatedAt.Time) {
			t.Errorf("クリアされた後続行のupdated_at = %v、期待値 = %vのまま", stored.updatedAt, publishedFollowerBefore.updatedAt)
		}
		// 削除したエピソード自身のポインタは残す。削除済みの行は何も表示せず、クリアすると
		// それが置かれていた並びが失われるため。
		if stored := readDBDeleteEpisode(t, tx, targetID); stored.prevEpisodeID.Int64 != int64(predecessorID) {
			t.Errorf("削除した行のprev_episode_id = %v、期待値 = %dのまま", stored.prevEpisodeID, int64(predecessorID))
		}
	})

	// 他者が先に削除した後の一覧からの送信は、未削除の行を見つけない。書き込まずそれを報告
	// することで、1回の遷移に対してカウンターが2度減算されるのを防ぐ。
	t.Run("削除済みのエピソードはnilを返し、カウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{
			workID:    workID,
			deletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		result, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済みのエピソードへのDelete()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 論理削除済みの親作品は、episode行が残っていても管理画面の一覧対象外。非公開・再公開の
	// 方向と同じく拒否し、状態とカウンターの両方を変えないことを検証する。
	t.Run("削除済み作品のエピソードはnilを返し、削除しない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBArchiveWork(t, tx, 3)
		episodeID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{workID: workID})
		if _, err := tx.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
			t.Fatalf("親作品の削除に失敗: %v", err)
		}

		result, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     episodeID,
			WorkID: workID,
		})
		if err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("削除済み作品のエピソードへのDelete()の戻り値 = 結果あり、期待値 = nil")
		}
		if stored := readDBDeleteEpisode(t, tx, episodeID); stored.deletedAt.Valid {
			t.Error("episodes.deleted_atに値が入った、期待値 = NULLのまま")
		}
		if work := readDBArchiveWork(t, tx, workID); work.episodesCount != 3 {
			t.Errorf("works.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
	})

	// 一覧と送信の間に別作品へ移されたエピソードは、その一覧が数えていた作品にもう属して
	// いない。したがってエピソードもどちらの作品も書かない。
	t.Run("別作品へ移されたエピソードはnilを返し、元の作品のカウンターを動かさない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		originalWorkID := insertDBArchiveWork(t, tx, 3)
		movedWorkID := insertDBArchiveWork(t, tx, 1)
		episodeID := insertDBDeleteEpisode(t, tx, dbDeleteEpisodeRow{workID: movedWorkID})

		result, err := repo.Delete(context.Background(), repository.DeleteEpisodeParams{
			ID:     episodeID,
			WorkID: originalWorkID,
		})
		if err != nil {
			t.Fatalf("Delete()のエラー = %v", err)
		}
		if result != nil {
			t.Fatal("別作品へ移された行へのDelete()の戻り値 = 結果あり、期待値 = nil")
		}
		if work := readDBArchiveWork(t, tx, originalWorkID); work.episodesCount != 3 {
			t.Errorf("移動元のworks.episodes_count = %d、期待値 = 3", work.episodesCount)
		}
		if work := readDBArchiveWork(t, tx, movedWorkID); work.episodesCount != 1 {
			t.Errorf("移動先のworks.episodes_count = %d、期待値 = 1", work.episodesCount)
		}
		if stored := readDBDeleteEpisode(t, tx, episodeID); stored.deletedAt.Valid {
			t.Error("episodes.deleted_atに値が入った、期待値 = NULLのまま")
		}
	})
}

func TestEpisodeRepository_GetForDeleteByID(t *testing.T) {
	t.Parallel()

	// ローダーが射影するのは書き込みが必要とするものだけ。対象のidと、削除を束縛する親作品。
	// animeの写像は代わりにDeleteDBEpisodeが更新した行から得る。
	t.Run("正常系: idと所属作品を射影する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     workID,
			sortNumber: 100,
		})

		got, err := repo.GetForDeleteByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForDeleteByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForDeleteByID() = nil、期待値 = エピソード")
		}
		if got.ID != episodeID {
			t.Errorf("Episode.ID = %d、期待値 = %d", int64(got.ID), int64(episodeID))
		}
		if got.WorkID != workID {
			t.Errorf("Episode.WorkID = %d、期待値 = %d", int64(got.WorkID), int64(workID))
		}
	})

	// 非公開のエピソードも削除できるため、ローダーはそれも返す。非公開エンドポイントが送信を
	// 公開中の行に限るのとは異なる。
	t.Run("非公開のエピソードも返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		workID := insertDBListWork(t, tx)
		episodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:        workID,
			sortNumber:    100,
			unpublishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		got, err := repo.GetForDeleteByID(context.Background(), episodeID)
		if err != nil {
			t.Fatalf("GetForDeleteByID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("GetForDeleteByID() = nil、期待値 = エピソード")
		}
	})

	// 除外条件は非公開のローダーと揃える。削除が届くエピソードを、一覧が削除の操作を出せる
	// エピソードと一致させるため。
	t.Run("削除済みのエピソードと削除済み作品のエピソードはnilを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		repo := repository.NewEpisodeRepository(query.New(db).WithTx(tx))

		keptWorkID := insertDBListWork(t, tx)
		deletedEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     keptWorkID,
			sortNumber: 100,
			deletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		})

		deletedWorkID := testutil.NewWorkBuilder(t, tx).WithDeletedAt(time.Now()).Build()
		orphanEpisodeID := insertDBListEpisode(t, tx, dbListEpisodeRow{
			workID:     deletedWorkID,
			sortNumber: 100,
		})

		for name, id := range map[string]model.EpisodeID{
			"削除済みのエピソード":   deletedEpisodeID,
			"削除済み作品のエピソード": orphanEpisodeID,
			"存在しないエピソード":   model.EpisodeID(-1),
		} {
			got, err := repo.GetForDeleteByID(context.Background(), id)
			if err != nil {
				t.Fatalf("%s: GetForDeleteByID()のエラー = %v", name, err)
			}
			if got != nil {
				t.Errorf("%s: GetForDeleteByID() = %+v、期待値 = nil", name, got)
			}
		}
	})
}
