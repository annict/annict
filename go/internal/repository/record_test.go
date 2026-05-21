package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestRecordRepository_AggregateDailyCountsByUserID(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 記録がないユーザーは空スライスを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()

		dateFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		counts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "Asia/Tokyo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(counts) != 0 {
			t.Errorf("len(counts) = %d, want 0", len(counts))
		}
	})

	t.Run("正常系: 日別件数を返し、削除済みレコードは除外する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// Insert 2 records on 2026-04-01 (Asia/Tokyo), 1 on 2026-04-02
		// (Asia/Tokyo), plus one soft-deleted record that should be ignored.
		//
		// [Ja] 2026-04-01 (Asia/Tokyo) に 2 件、2026-04-02 (Asia/Tokyo) に 1 件、
		// 削除済みを 1 件混ぜる。
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("failed to load Asia/Tokyo: %v", err)
		}
		watched1a := time.Date(2026, 4, 1, 10, 0, 0, 0, jst).UTC()
		watched1b := time.Date(2026, 4, 1, 23, 0, 0, 0, jst).UTC()
		watched2 := time.Date(2026, 4, 2, 8, 0, 0, 0, jst).UTC()
		watchedDel := time.Date(2026, 4, 1, 12, 0, 0, 0, jst).UTC()

		insertTestRecord(t, tx, userID, workID, watched1a, sql.NullTime{})
		insertTestRecord(t, tx, userID, workID, watched1b, sql.NullTime{})
		insertTestRecord(t, tx, userID, workID, watched2, sql.NullTime{})
		insertTestRecord(t, tx, userID, workID, watchedDel, sql.NullTime{Time: time.Now(), Valid: true})

		dateFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		counts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "Asia/Tokyo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := map[string]int64{}
		for _, c := range counts {
			got[c.Day.Format("2006-01-02")] = c.Count
		}
		if got["2026-04-01"] != 2 {
			t.Errorf("2026-04-01 count = %d, want 2", got["2026-04-01"])
		}
		if got["2026-04-02"] != 1 {
			t.Errorf("2026-04-02 count = %d, want 1", got["2026-04-02"])
		}
		if len(got) != 2 {
			t.Errorf("unexpected days in result: %v", got)
		}
	})

	t.Run("正常系: タイムゾーンに応じて日付バケットが変わる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// 2026-04-01 23:00 UTC == 2026-04-02 08:00 Asia/Tokyo. Verify
		// that switching the time-zone argument changes the day bucket.
		//
		// [Ja] 2026-04-01 23:00 UTC は Asia/Tokyo (UTC+9) では 2026-04-02 08:00 に
		// 当たる。タイムゾーン引数を切り替えると日付バケットが変わることを
		// 確認する。
		watchedAt := time.Date(2026, 4, 1, 23, 0, 0, 0, time.UTC)
		insertTestRecord(t, tx, userID, workID, watchedAt, sql.NullTime{})

		dateFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

		utcCounts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "UTC")
		if err != nil {
			t.Fatalf("UTC aggregate failed: %v", err)
		}
		if len(utcCounts) != 1 || utcCounts[0].Day.Format("2006-01-02") != "2026-04-01" {
			t.Errorf("UTC bucket = %+v, want 2026-04-01", utcCounts)
		}

		jstCounts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "Asia/Tokyo")
		if err != nil {
			t.Fatalf("Asia/Tokyo aggregate failed: %v", err)
		}
		if len(jstCounts) != 1 || jstCounts[0].Day.Format("2006-01-02") != "2026-04-02" {
			t.Errorf("Asia/Tokyo bucket = %+v, want 2026-04-02", jstCounts)
		}
	})

	t.Run("正常系: date_from より前のレコードは除外する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// Insert one record on date_from and one on the prior day; the
		// prior day must be excluded.
		//
		// [Ja] dateFrom 当日と前日に 1 件ずつ作成。dateFrom より前は除外される
		// ことを確認する。
		dateFrom := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		insertTestRecord(t, tx, userID, workID, dateFrom, sql.NullTime{})
		insertTestRecord(t, tx, userID, workID, dateFrom.AddDate(0, 0, -1), sql.NullTime{})

		counts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "UTC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(counts) != 1 || counts[0].Day.Format("2006-01-02") != "2026-04-01" {
			t.Errorf("counts = %+v, want exactly 2026-04-01", counts)
		}
	})
}

func TestRecordRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewRecordRepository(queries).WithTx(tx)

	userID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()

	watchedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	insertTestRecord(t, tx, userID, workID, watchedAt, sql.NullTime{})

	dateFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	counts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(counts) != 1 {
		t.Errorf("len(counts) = %d, want 1", len(counts))
	}
}

// insertTestRecord inserts a row into `records` for testing.
// [Ja] insertTestRecord はテスト用に records 行を 1 件挿入する。
func insertTestRecord(
	t *testing.T,
	tx *sql.Tx,
	userID model.UserID,
	workID model.WorkID,
	watchedAt time.Time,
	deletedAt sql.NullTime,
) {
	t.Helper()
	const q = `
		INSERT INTO records (
			user_id, work_id, watched_at, deleted_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`
	now := time.Now()
	if _, err := tx.Exec(q, int64(userID), int64(workID), watchedAt, deletedAt, now, now); err != nil {
		t.Fatalf("レコードの挿入に失敗しました: %v", err)
	}
}
