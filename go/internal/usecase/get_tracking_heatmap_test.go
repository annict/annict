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

	t.Run("異常系: 存在しないユーザーはResourceNotFoundを返す", func(t *testing.T) {
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

	t.Run("異常系: 削除済みユーザーはResourceNotFoundを返す", func(t *testing.T) {
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
			t.Fatalf("username取得に失敗: %v", err)
		}

		_, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		})
		assertNotFoundAppError(t, err)
	})

	t.Run("正常系: 記録0件でもdate_fromから今日まで連続したセルを返す", func(t *testing.T) {
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
			t.Fatalf("想定外のエラー = %v", err)
		}

		// date_fromは (今日 - 150日) を直近の日曜日に丸めた日。
		// today (Asia/Tokyo) はnow (UTC 12:00) をJSTに変換した2026-04-01。
		// today - 150日 = 2025-11-02 (日曜)。日曜なので日曜丸めは同日。
		// 結果セル数は2025-11-02 〜 2026-04-01の151日。
		if len(out.Cells) != 151 {
			t.Errorf("len(Cells) = %d、期待値 = 151", len(out.Cells))
		}
		if out.Cells[0].Date != "2025-11-02" {
			t.Errorf("Cells[0].Date = %q、期待値 = 2025-11-02", out.Cells[0].Date)
		}
		if out.Cells[len(out.Cells)-1].Date != "2026-04-01" {
			t.Errorf("Cells[last].Date = %q、期待値 = 2026-04-01", out.Cells[len(out.Cells)-1].Date)
		}
		for _, c := range out.Cells {
			if c.Count != 0 || c.LeveledCount != 0 {
				t.Errorf("cell = %+v、期待値 = ゼロ値", c)
				break
			}
		}
	})

	t.Run("正常系: 密度レベルの境界 (1, 4, 7, 10件) を正しく分類する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("Asia/Tokyoの読み込みエラー = %v", err)
		}
		// 異なる日に異なる件数の記録を作り、レベル割り当てを検証する。
		type spec struct {
			date   time.Time
			count  int
			level  int
			dateID string
		}
		// todayは2026-04-01 (JST)。直近日付に複数件配置する。
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
			t.Fatalf("想定外のエラー = %v", err)
		}

		byDate := map[string]TrackingHeatmapCell{}
		for _, c := range out.Cells {
			byDate[c.Date] = c
		}
		for _, s := range specs {
			cell, ok := byDate[s.dateID]
			if !ok {
				t.Errorf("%sの日付がcellsに無い", s.dateID)
				continue
			}
			if cell.Count != s.count {
				t.Errorf("%sのcount = %d、期待値 = %d", s.dateID, cell.Count, s.count)
			}
			if cell.LeveledCount != s.level {
				t.Errorf("%sのlevel = %d、期待値 = %d", s.dateID, cell.LeveledCount, s.level)
			}
		}
	})

	t.Run("正常系: 150日境界の日 (date_from) は含まれ、その前日は含まれない", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := newTrackingHeatmapUsecaseForTest(queries)

		userID := testutil.NewUserBuilder(t, tx).Build()
		workID := testutil.NewWorkBuilder(t, tx).Build()
		username := lookupUsername(t, tx, userID)

		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("Asia/Tokyoの読み込みエラー = %v", err)
		}
		// today = 2026-04-01 (JST), date_from = 2025-11-02 (日曜)。
		// 境界日と前日に1件ずつ作成。
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
			t.Fatalf("想定外のエラー = %v", err)
		}

		byDate := map[string]TrackingHeatmapCell{}
		for _, c := range out.Cells {
			byDate[c.Date] = c
		}
		if cell, ok := byDate["2025-11-02"]; !ok || cell.Count != 1 {
			t.Errorf("2025-11-02のcell = %+v、期待値 = count=1", cell)
		}
		if _, ok := byDate["2025-11-01"]; ok {
			t.Error("2025-11-01がcellsに含まれている")
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

		// 2026-03-31 22:00 UTCはAmerica/New_York (UTC-4) では2026-03-31
		// 18:00、Asia/Tokyo (UTC+9) では2026-04-01 07:00に当たる。
		watchedAt := time.Date(2026, 3, 31, 22, 0, 0, 0, time.UTC)
		insertRecordForTest(t, tx, userID, workID, watchedAt)

		now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

		outJST, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "Asia/Tokyo",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("Asia/TokyoでのExecute()のエラー = %v", err)
		}
		if cnt := countFor(outJST.Cells, "2026-04-01"); cnt != 1 {
			t.Errorf("Asia/Tokyoの2026-04-01のcount = %d、期待値 = 1", cnt)
		}

		outNY, err := uc.Execute(context.Background(), GetTrackingHeatmapInput{
			Username: username,
			TimeZone: "America/New_York",
			Now:      now,
		})
		if err != nil {
			t.Fatalf("America/New_YorkでのExecute()のエラー = %v", err)
		}
		if cnt := countFor(outNY.Cells, "2026-03-31"); cnt != 1 {
			t.Errorf("America/New_Yorkの2026-03-31のcount = %d、期待値 = 1", cnt)
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
		t.Fatal("エラーを期待したが、nilだった")
	}
	var ae *model.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("エラー = %v、期待値 = *model.AppError", err)
	}
	if ae.Code != model.AppErrCodeResourceNotFound {
		t.Errorf("AppError.Code = %d、期待値 = %d", ae.Code, model.AppErrCodeResourceNotFound)
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
		t.Fatalf("username取得に失敗: %v", err)
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
