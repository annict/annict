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
			t.Fatalf("想定外のエラー = %v", err)
		}
		if len(counts) != 0 {
			t.Errorf("len(counts) = %d、期待値 = 0", len(counts))
		}
	})

	t.Run("正常系: 日別件数を返し、削除済みレコードは除外する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// 2026-04-01 (Asia/Tokyo) に2件、2026-04-02 (Asia/Tokyo) に1件、
		// 削除済みを1件混ぜる。
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("Asia/Tokyoの読み込みエラー = %v", err)
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
			t.Fatalf("想定外のエラー = %v", err)
		}

		got := map[string]int64{}
		for _, c := range counts {
			got[c.Day.Format("2006-01-02")] = c.Count
		}
		if got["2026-04-01"] != 2 {
			t.Errorf("2026-04-01のcount = %d、期待値 = 2", got["2026-04-01"])
		}
		if got["2026-04-02"] != 1 {
			t.Errorf("2026-04-02のcount = %d、期待値 = 1", got["2026-04-02"])
		}
		if len(got) != 2 {
			t.Errorf("結果の日付 = %v、期待値 = 2026-04-01と2026-04-02の2件のみ", got)
		}
	})

	t.Run("正常系: タイムゾーンに応じて日付バケットが変わる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// 2026-04-01 23:00 UTCはAsia/Tokyo (UTC+9) では2026-04-02 08:00に
		// 当たる。タイムゾーン引数を切り替えると日付バケットが変わることを
		// 確認する。
		watchedAt := time.Date(2026, 4, 1, 23, 0, 0, 0, time.UTC)
		insertTestRecord(t, tx, userID, workID, watchedAt, sql.NullTime{})

		dateFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

		utcCounts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "UTC")
		if err != nil {
			t.Fatalf("UTCでの集計のエラー = %v", err)
		}
		if len(utcCounts) != 1 || utcCounts[0].Day.Format("2006-01-02") != "2026-04-01" {
			t.Errorf("UTCでの日別集計結果 = %+v、期待値 = 2026-04-01", utcCounts)
		}

		jstCounts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "Asia/Tokyo")
		if err != nil {
			t.Fatalf("Asia/Tokyoでの集計のエラー = %v", err)
		}
		if len(jstCounts) != 1 || jstCounts[0].Day.Format("2006-01-02") != "2026-04-02" {
			t.Errorf("Asia/Tokyoでの日別集計結果 = %+v、期待値 = 2026-04-02", jstCounts)
		}
	})

	t.Run("正常系: date_fromより前のレコードは除外する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewRecordRepository(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()

		// dateFrom当日と前日に1件ずつ作成。dateFromより前は除外される
		// ことを確認する。
		dateFrom := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		insertTestRecord(t, tx, userID, workID, dateFrom, sql.NullTime{})
		insertTestRecord(t, tx, userID, workID, dateFrom.AddDate(0, 0, -1), sql.NullTime{})

		counts, err := repo.AggregateDailyCountsByUserID(context.Background(), userID, dateFrom, "UTC")
		if err != nil {
			t.Fatalf("想定外のエラー = %v", err)
		}
		if len(counts) != 1 || counts[0].Day.Format("2006-01-02") != "2026-04-01" {
			t.Errorf("counts = %+v、期待値 = 2026-04-01のみ", counts)
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
		t.Fatalf("想定外のエラー = %v", err)
	}
	if len(counts) != 1 {
		t.Errorf("len(counts) = %d、期待値 = 1", len(counts))
	}
}

// insertTestRecordはテスト用にrecords行を1件挿入する。
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

func TestRecordRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	hiddenWorkID := testutil.NewWorkBuilder(t, tx).Build()
	if _, err := tx.ExecContext(ctx,
		"UPDATE works SET records_count = 10 WHERE id IN ($1, $2)",
		int64(workID), int64(hiddenWorkID),
	); err != nil {
		t.Fatalf("作品の記録数の設定に失敗: %v", err)
	}

	watchedAt := time.Now()
	insertTestRecord(t, tx, userID, workID, watchedAt, sql.NullTime{})
	insertTestRecord(t, tx, userID, workID, watchedAt, sql.NullTime{})
	// 論理削除された記録は記録数に含まれていないため、減算しない。
	insertTestRecord(t, tx, userID, hiddenWorkID, watchedAt, sql.NullTime{Time: watchedAt, Valid: true})
	insertTestRecord(t, tx, otherUserID, workID, watchedAt, sql.NullTime{})

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{2, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 2)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "records", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーの記録の件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "records", "user_id = $1", int64(otherUserID)); got != 1 {
		t.Errorf("他のユーザーの記録の件数 = %d、期待値 = 1", got)
	}

	for workID, want := range map[model.WorkID]int32{
		workID:       8,
		hiddenWorkID: 10,
	} {
		var recordsCount int32
		if err := tx.QueryRowContext(ctx, "SELECT records_count FROM works WHERE id = $1", int64(workID)).Scan(&recordsCount); err != nil {
			t.Fatalf("作品の記録数の取得に失敗: %v", err)
		}
		if recordsCount != want {
			t.Errorf("作品 %d の記録数 = %d、期待値 = %d", workID, recordsCount, want)
		}
	}
}
