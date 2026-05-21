package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetTrackingHeatmapUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("異常系: 存在しないユーザーは ResourceNotFound を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		_, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: "no_such_user_xyz",
			TimeZone: "Asia/Tokyo",
			Now:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		})
		assertNotFoundAppError(t, err)
	})

	t.Run("異常系: 削除済みユーザーは ResourceNotFound を返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		if _, err := tx.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(userID)); err != nil {
			t.Fatalf("ユーザーの論理削除に失敗: %v", err)
		}

		var username string
		if err := tx.QueryRow("SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
			t.Fatalf("username 取得に失敗: %v", err)
		}

		_, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		})
		assertNotFoundAppError(t, err)
	})

	t.Run("正常系: 記録 0 件でも date_from から今日まで連続したセルを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
		out, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// date_from is (today - 150d) snapped to Sunday. Asia/Tokyo
		// today derived from 2026-04-01 12:00 UTC is 2026-04-01. -150d
		// gives 2025-11-02 (Sunday), so date_from == that day. Total cells
		// from 2025-11-02 through 2026-04-01 inclusive = 151.
		//
		// [Ja] date_from は (今日 - 150日) を直近の日曜日に丸めた日。
		// today (Asia/Tokyo) は now (UTC 12:00) を JST に変換した 2026-04-01。
		// today - 150 日 = 2025-11-02 (日曜)。日曜なので日曜丸めは同日。
		// 結果セル数は 2025-11-02 〜 2026-04-01 の 151 日。
		if len(out.Cells) != 151 {
			t.Errorf("len(Cells) = %d, want 151", len(out.Cells))
		}
		if out.Cells[0].Date != "2025-11-02" {
			t.Errorf("Cells[0].Date = %q, want 2025-11-02", out.Cells[0].Date)
		}
		if out.Cells[len(out.Cells)-1].Date != "2026-04-01" {
			t.Errorf("Cells[last].Date = %q, want 2026-04-01", out.Cells[len(out.Cells)-1].Date)
		}
		for _, c := range out.Cells {
			if c.Count != 0 || c.LeveledCount != 0 {
				t.Errorf("expected zero cell, got %+v", c)
				break
			}
		}
	})

	t.Run("正常系: 密度レベルの境界 (1, 4, 7, 10 件) を正しく分類する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("failed to load Asia/Tokyo: %v", err)
		}
		// Insert different counts on different days and check the
		// resulting density level for each.
		//
		// [Ja] 異なる日に異なる件数の記録を作り、レベル割り当てを検証する。
		type spec struct {
			date   time.Time
			count  int
			level  int
			dateID string
		}
		// today is 2026-04-01 (JST). Spread the records across recent days.
		// [Ja] today は 2026-04-01 (JST)。直近日付に複数件配置する。
		specs := []spec{
			{date: time.Date(2026, 3, 25, 10, 0, 0, 0, jst), count: 1, level: 1, dateID: "2026-03-25"},
			{date: time.Date(2026, 3, 26, 10, 0, 0, 0, jst), count: 4, level: 2, dateID: "2026-03-26"},
			{date: time.Date(2026, 3, 27, 10, 0, 0, 0, jst), count: 7, level: 3, dateID: "2026-03-27"},
			{date: time.Date(2026, 3, 28, 10, 0, 0, 0, jst), count: 10, level: 4, dateID: "2026-03-28"},
		}
		for _, s := range specs {
			for i := 0; i < s.count; i++ {
				insertRecordForTest(t, tx, userID, workID, s.date.Add(time.Duration(i)*time.Minute).UTC())
			}
		}

		now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
		out, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		byDate := map[string]TrackingHeatmapCell{}
		for _, c := range out.Cells {
			byDate[c.Date] = c
		}
		for _, s := range specs {
			cell, ok := byDate[s.dateID]
			if !ok {
				t.Errorf("date %s missing from cells", s.dateID)
				continue
			}
			if cell.Count != s.count {
				t.Errorf("date %s count = %d, want %d", s.dateID, cell.Count, s.count)
			}
			if cell.LeveledCount != s.level {
				t.Errorf("date %s level = %d, want %d", s.dateID, cell.LeveledCount, s.level)
			}
		}
	})

	t.Run("正常系: 150 日境界の日 (date_from) は含まれ、その前日は含まれない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("failed to load Asia/Tokyo: %v", err)
		}
		// today=2026-04-01 (JST), date_from=2025-11-02 (Sunday).
		// One record on date_from, one on the day before.
		//
		// [Ja] today = 2026-04-01 (JST), date_from = 2025-11-02 (日曜)。
		// 境界日と前日に 1 件ずつ作成。
		onBoundary := time.Date(2025, 11, 2, 12, 0, 0, 0, jst).UTC()
		beforeBoundary := time.Date(2025, 11, 1, 12, 0, 0, 0, jst).UTC()
		insertRecordForTest(t, tx, userID, workID, onBoundary)
		insertRecordForTest(t, tx, userID, workID, beforeBoundary)

		now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
		out, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		byDate := map[string]TrackingHeatmapCell{}
		for _, c := range out.Cells {
			byDate[c.Date] = c
		}
		if cell, ok := byDate["2025-11-02"]; !ok || cell.Count != 1 {
			t.Errorf("2025-11-02 cell = %+v, want count=1", cell)
		}
		if _, ok := byDate["2025-11-01"]; ok {
			t.Error("2025-11-01 must not be included in cells")
		}
	})

	t.Run("正常系: タイムゾーンに応じて日付バケットが変わる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		// 2026-03-31 22:00 UTC == 2026-03-31 in America/New_York and
		// 2026-04-01 in Asia/Tokyo. Verify both buckets.
		//
		// [Ja] 2026-03-31 22:00 UTC は America/New_York (UTC-4) では 2026-03-31
		// 18:00、Asia/Tokyo (UTC+9) では 2026-04-01 07:00 に当たる。
		watchedAt := time.Date(2026, 3, 31, 22, 0, 0, 0, time.UTC)
		insertRecordForTest(t, tx, userID, workID, watchedAt)

		now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

		outJST, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("Asia/Tokyo execute failed: %v", err)
		}
		if cnt := countFor(outJST.Cells, "2026-04-01"); cnt != 1 {
			t.Errorf("Asia/Tokyo 2026-04-01 count = %d, want 1", cnt)
		}

		outNY, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "America/New_York",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("America/New_York execute failed: %v", err)
		}
		if cnt := countFor(outNY.Cells, "2026-03-31"); cnt != 1 {
			t.Errorf("America/New_York 2026-03-31 count = %d, want 1", cnt)
		}
	})
}

func newTrackingHeatmapUsecaseForTest(queries *query.Queries) *GetTrackingHeatmapUsecase {
	return NewGetTrackingHeatmapUsecase(
		repository.NewUserRepository(queries),
		repository.NewRecordRepository(queries),
	)
}

func assertNotFoundAppError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var ae *model.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("error is not *model.AppError: %v", err)
	}
	if ae.Code != model.AppErrCodeResourceNotFound {
		t.Errorf("AppError.Code = %d, want %d", ae.Code, model.AppErrCodeResourceNotFound)
	}
}

func countFor(cells []TrackingHeatmapCell, date string) int {
	for _, c := range cells {
		if c.Date == date {
			return c.Count
		}
	}
	return -1
}

func lookupUsername(t *testing.T, tx *sql.Tx, userID model.UserID) string {
	t.Helper()
	var username string
	if err := tx.QueryRow("SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
		t.Fatalf("username 取得に失敗: %v", err)
	}
	return username
}

func insertRecordForTest(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, watchedAt time.Time) {
	t.Helper()
	const q = `
		INSERT INTO records (
			user_id, work_id, watched_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`
	now := time.Now()
	if _, err := tx.Exec(q, int64(userID), int64(workID), watchedAt, now, now); err != nil {
		t.Fatalf("レコードの挿入に失敗しました: %v", err)
	}
}
