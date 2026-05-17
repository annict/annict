package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestUserCalendarRepository_GetByUsername(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testcalendar").
		WithLocale("ja").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "テストアニメ", "Test Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		WithTitle("第1話 始まりの予感").
		Build()

	// チャンネルグループとチャンネルを作成
	channelGroupID := createTestChannelGroup(t, tx, "関東")
	channelID := createTestChannel(t, tx, channelGroupID, "TOKYO MX")

	// プログラムを作成
	programID := createTestProgram(t, tx, channelID, workID)

	// スロットを作成（現在時刻の1時間後）
	now := time.Now()
	futureTime := now.Add(1 * time.Hour)
	slotID := createTestSlot(t, tx, channelID, workID, episodeID, programID, futureTime)

	// ステータスを作成（kind=2: watching）
	statusID := createTestStatus(t, tx, userID, workID, 2)

	// ライブラリエントリを作成
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("正常系: ユーザーのカレンダーデータを取得できる", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "testcalendar", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// ユーザー情報の確認
		if calendar.Username != "testcalendar" {
			t.Errorf("Username = %q, want %q", calendar.Username, "testcalendar")
		}
		if calendar.TimeZone != "Asia/Tokyo" {
			t.Errorf("TimeZone = %q, want %q", calendar.TimeZone, "Asia/Tokyo")
		}
		if calendar.Locale != "ja" {
			t.Errorf("Locale = %q, want %q", calendar.Locale, "ja")
		}

		// スロットの確認
		if len(calendar.Slots) != 1 {
			t.Errorf("len(Slots) = %d, want 1", len(calendar.Slots))
		} else {
			slot := calendar.Slots[0]
			if slot.ID != slotID {
				t.Errorf("Slot.ID = %d, want %d", slot.ID, slotID)
			}
			if slot.WorkTitle != "テストアニメ" {
				t.Errorf("Slot.WorkTitle = %q, want %q", slot.WorkTitle, "テストアニメ")
			}
			if slot.WorkTitleEn != "Test Anime" {
				t.Errorf("Slot.WorkTitleEn = %q, want %q", slot.WorkTitleEn, "Test Anime")
			}
			if slot.EpisodeID != episodeID {
				t.Errorf("Slot.EpisodeID = %d, want %d", slot.EpisodeID, episodeID)
			}
			if slot.EpisodeTitle != "第1話 始まりの予感" {
				t.Errorf("Slot.EpisodeTitle = %q, want %q", slot.EpisodeTitle, "第1話 始まりの予感")
			}
			if slot.ChannelName != "TOKYO MX" {
				t.Errorf("Slot.ChannelName = %q, want %q", slot.ChannelName, "TOKYO MX")
			}
		}

		// 作品の確認
		if len(calendar.Works) != 1 {
			t.Errorf("len(Works) = %d, want 1", len(calendar.Works))
		} else {
			work := calendar.Works[0]
			if work.ID != workID {
				t.Errorf("Work.ID = %d, want %d", work.ID, workID)
			}
			if work.Title != "テストアニメ" {
				t.Errorf("Work.Title = %q, want %q", work.Title, "テストアニメ")
			}
			if work.TitleEn != "Test Anime" {
				t.Errorf("Work.TitleEn = %q, want %q", work.TitleEn, "Test Anime")
			}
		}
	})

	t.Run("正常系: 視聴済みエピソードは除外される", func(t *testing.T) {
		// 別のユーザーを作成
		userID2 := testutil.NewUserBuilder(t, tx).
			WithUsername("testcalendar2").
			Build()

		// 同じ作品に対してステータスとライブラリエントリを作成
		statusID2 := createTestStatus(t, tx, userID2, workID, 2)
		// 視聴済みエピソードIDを含める
		createTestLibraryEntry(t, tx, userID2, workID, statusID2, programID, []model.EpisodeID{episodeID})

		calendar, err := repo.GetByUsername(ctx, "testcalendar2", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// 視聴済みエピソードは除外されるため、スロットは0件
		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (watched episodes should be excluded)", len(calendar.Slots))
		}
	})

	t.Run("異常系: 存在しないユーザー", func(t *testing.T) {
		_, err := repo.GetByUsername(ctx, "nonexistent", now)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestUserCalendarRepository_GetByUsername_PastSlots(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testpast").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "過去アニメ", "Past Anime", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		Build()

	// チャンネルを作成
	channelGroupID := createTestChannelGroup(t, tx, "関東2")
	channelID := createTestChannel(t, tx, channelGroupID, "TOKYO MX2")

	// プログラムを作成
	programID := createTestProgram(t, tx, channelID, workID)

	// 過去のスロットを作成（1時間前）
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	createTestSlot(t, tx, channelID, workID, episodeID, programID, pastTime)

	// ステータスとライブラリエントリを作成
	statusID := createTestStatus(t, tx, userID, workID, 2)
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("正常系: 過去のスロットは取得されない", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "testpast", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// 過去のスロットは除外される
		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (past slots should be excluded)", len(calendar.Slots))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_LateNightSlots は深夜帯の放送枠に関するテストです
// Rails版の不具合: Date.today.beginning_of_dayを使用していたため、日本時間の午前0時を過ぎると
// 当日の深夜帯（例: 25時放送 = 翌日01:00）の放送枠が消えてしまっていた
// Go版では現在時刻を基準にフィルタリングすることで修正
func TestUserCalendarRepository_GetByUsername_LateNightSlots(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testlatenight").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "深夜アニメ", "Late Night Anime", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		WithTitle("第1話").
		Build()

	// チャンネルを作成
	channelGroupID := createTestChannelGroup(t, tx, "関東深夜")
	channelID := createTestChannel(t, tx, channelGroupID, "TOKYO MX深夜")

	// プログラムを作成
	programID := createTestProgram(t, tx, channelID, workID)

	// 深夜帯のスロットを作成（日本時間 2025年1月16日 01:00 = UTC 2025年1月15日 16:00）
	slotStartTime := time.Date(2025, 1, 15, 16, 0, 0, 0, time.UTC)
	createTestSlot(t, tx, channelID, workID, episodeID, programID, slotStartTime)

	// ステータスとライブラリエントリを作成
	statusID := createTestStatus(t, tx, userID, workID, 2)
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("正常系: 午前0時を過ぎても深夜帯のスロットが表示される", func(t *testing.T) {
		// 現在時刻を日本時間 0:30 に設定（まだ放送開始前）= UTC 15:30
		// Rails版の不具合では、Date.today.beginning_of_dayを使用していたため、
		// 日付が変わった瞬間に基準が「今日の00:00」になり、前日からのコンテキストが失われていた
		// Go版では現在時刻を基準にするため、深夜1:00の放送枠は正しく表示される
		now := time.Date(2025, 1, 15, 15, 30, 0, 0, time.UTC)

		calendar, err := repo.GetByUsername(ctx, "testlatenight", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// 深夜帯のスロット（まだ放送開始前）が含まれているべき
		if len(calendar.Slots) != 1 {
			t.Errorf("len(Slots) = %d, want 1 (深夜帯のスロットが含まれていない)", len(calendar.Slots))
		}

		if len(calendar.Slots) > 0 {
			slot := calendar.Slots[0]
			if slot.EpisodeID != episodeID {
				t.Errorf("Slot.EpisodeID = %d, want %d", slot.EpisodeID, episodeID)
			}
		}
	})

	t.Run("正常系: 放送開始後はスロットが表示されない", func(t *testing.T) {
		// 現在時刻を日本時間 1:30 に設定（放送開始後）= UTC 16:30
		now := time.Date(2025, 1, 15, 16, 30, 0, 0, time.UTC)

		calendar, err := repo.GetByUsername(ctx, "testlatenight", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// 放送開始後のスロットは表示されないべき
		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (放送開始後のスロットは表示されないべき)", len(calendar.Slots))
		}
	})
}

// createTestWorkWithStartedOn はstarted_onを設定した作品を作成します
func createTestWorkWithStartedOn(t *testing.T, tx *sql.Tx, title, titleEn string, startedOn time.Time) model.WorkID {
	t.Helper()

	query := `
		INSERT INTO works (
			title, title_kana, title_en, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count, started_on,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		) RETURNING id
	`

	var id int64
	err := tx.QueryRow(
		query,
		title,
		"",
		titleEn,
		0,
		"",
		"",
		2025,
		testutil.SeasonSpring,
		100,
		12,
		startedOn,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		t.Fatalf("作品データの作成に失敗しました: %v", err)
	}

	return model.WorkID(id)
}

// createTestChannelGroup はテスト用チャンネルグループを作成します
func createTestChannelGroup(t *testing.T, tx *sql.Tx, name string) int64 {
	t.Helper()

	query := `
		INSERT INTO channel_groups (name, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, name, 0, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("チャンネルグループデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestChannel はテスト用チャンネルを作成します
func createTestChannel(t *testing.T, tx *sql.Tx, channelGroupID int64, name string) int64 {
	t.Helper()

	query := `
		INSERT INTO channels (channel_group_id, name, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, channelGroupID, name, 0, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("チャンネルデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestProgram はテスト用プログラムを作成します
func createTestProgram(t *testing.T, tx *sql.Tx, channelID int64, workID model.WorkID) int64 {
	t.Helper()

	query := `
		INSERT INTO programs (channel_id, work_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, channelID, int64(workID), time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("プログラムデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestSlot はテスト用スロットを作成します
func createTestSlot(t *testing.T, tx *sql.Tx, channelID int64, workID model.WorkID, episodeID model.EpisodeID, programID int64, startedAt time.Time) model.SlotID {
	t.Helper()

	query := `
		INSERT INTO slots (channel_id, work_id, episode_id, program_id, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, channelID, int64(workID), int64(episodeID), programID, startedAt, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("スロットデータの作成に失敗しました: %v", err)
	}

	return model.SlotID(id)
}

// createTestStatus はテスト用ステータスを作成します
func createTestStatus(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, kind int) int64 {
	t.Helper()

	query := `
		INSERT INTO statuses (user_id, work_id, kind, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, int64(userID), int64(workID), kind, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("ステータスデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestLibraryEntry はテスト用ライブラリエントリを作成します
func createTestLibraryEntry(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, statusID, programID int64, watchedEpisodeIDs []model.EpisodeID) {
	t.Helper()

	// 配列をPostgreSQL形式に変換
	watchedIDsStr := "{"
	for i, id := range watchedEpisodeIDs {
		if i > 0 {
			watchedIDsStr += ","
		}
		watchedIDsStr += fmt.Sprintf("%d", int64(id))
	}
	watchedIDsStr += "}"

	query := `
		INSERT INTO library_entries (user_id, work_id, status_id, program_id, watched_episode_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::bigint[], $6, $7)
	`

	_, err := tx.Exec(query, int64(userID), int64(workID), statusID, programID, watchedIDsStr, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ライブラリエントリデータの作成に失敗しました: %v", err)
	}
}

// TestUserCalendarRepository_GetByUsername_DeletedUser は削除されたユーザーのテストです
func TestUserCalendarRepository_GetByUsername_DeletedUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// 削除されたユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("deleteduser").
		Build()

	// ユーザーを削除（deleted_atを設定）
	_, err := tx.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		t.Fatalf("ユーザーの削除に失敗しました: %v", err)
	}

	ctx := context.Background()

	t.Run("削除されたユーザーにアクセスした場合、sql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := repo.GetByUsername(ctx, "deleteduser", time.Now())
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

// TestUserCalendarRepository_GetByUsername_EmptyLibrary は視聴リストが空の場合のテストです
func TestUserCalendarRepository_GetByUsername_EmptyLibrary(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// ユーザーを作成（ライブラリエントリなし）
	testutil.NewUserBuilder(t, tx).
		WithUsername("emptyuser").
		Build()

	ctx := context.Background()
	now := time.Now()

	t.Run("視聴リストに追加済みのアニメがない場合、空のカレンダーを返す", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "emptyuser", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0", len(calendar.Slots))
		}
		if len(calendar.Works) != 0 {
			t.Errorf("len(Works) = %d, want 0", len(calendar.Works))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_SlotsAfter7Days は8日以降のスロットが除外されることをテストします
func TestUserCalendarRepository_GetByUsername_SlotsAfter7Days(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test7days").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "7日テストアニメ", "7 Days Test Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		Build()

	// チャンネルとプログラムを作成
	channelGroupID := createTestChannelGroup(t, tx, "テストグループ7日")
	channelID := createTestChannel(t, tx, channelGroupID, "テストチャンネル7日")
	programID := createTestProgram(t, tx, channelID, workID)

	now := time.Now()

	// 8日後のスロットを作成（除外されるべき）
	after8Days := now.AddDate(0, 0, 8)
	createTestSlot(t, tx, channelID, workID, episodeID, programID, after8Days)

	// ステータスとライブラリエントリを作成
	statusID := createTestStatus(t, tx, userID, workID, 2)
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("8日以降の放送枠は含まれない", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "test7days", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (8日以降のスロットは除外されるべき)", len(calendar.Slots))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_NoProgramID は番組が設定されていないライブラリエントリのテストです
func TestUserCalendarRepository_GetByUsername_NoProgramID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testnoprog").
		Build()

	// 作品を作成（started_on設定あり）
	workID := createTestWorkWithStartedOn(t, tx, "番組なしアニメ", "No Program Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// ステータスを作成（kind=2: watching）
	statusID := createTestStatus(t, tx, userID, workID, 2)

	// program_idがNULLのライブラリエントリを作成
	query := `
		INSERT INTO library_entries (user_id, work_id, status_id, program_id, watched_episode_ids, created_at, updated_at)
		VALUES ($1, $2, $3, NULL, '{}', $4, $5)
	`
	_, err := tx.Exec(query, userID, workID, statusID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ライブラリエントリデータの作成に失敗しました: %v", err)
	}

	ctx := context.Background()
	now := time.Now()

	t.Run("番組が設定されていないライブラリエントリは無視される", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "testnoprog", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		// スロットは0件であるべき（program_idがないため）
		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (program_idがないライブラリエントリは無視されるべき)", len(calendar.Slots))
		}

		// ただし、作品イベント（started_on）は含まれる
		if len(calendar.Works) != 1 {
			t.Errorf("len(Works) = %d, want 1 (started_onが設定された作品は含まれるべき)", len(calendar.Works))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_DeletedSlot は削除済みスロットが除外されることをテストします
func TestUserCalendarRepository_GetByUsername_DeletedSlot(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testdeletedslot").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "削除スロットアニメ", "Deleted Slot Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		Build()

	// チャンネルとプログラムを作成
	channelGroupID := createTestChannelGroup(t, tx, "削除テストグループ")
	channelID := createTestChannel(t, tx, channelGroupID, "削除テストチャンネル")
	programID := createTestProgram(t, tx, channelID, workID)

	now := time.Now()

	// 未来のスロットを作成
	slotID := createTestSlot(t, tx, channelID, workID, episodeID, programID, now.Add(1*time.Hour))

	// スロットを削除（deleted_atを設定）
	_, err := tx.Exec(`UPDATE slots SET deleted_at = NOW() WHERE id = $1`, int64(slotID))
	if err != nil {
		t.Fatalf("スロットの削除に失敗しました: %v", err)
	}

	// ステータスとライブラリエントリを作成
	statusID := createTestStatus(t, tx, userID, workID, 2)
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("削除済みの放送枠は含まれない", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "testdeletedslot", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		if len(calendar.Slots) != 0 {
			t.Errorf("len(Slots) = %d, want 0 (削除済みスロットは除外されるべき)", len(calendar.Slots))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_WannaWatchStatus はwanna_watchステータスのテストです
func TestUserCalendarRepository_GetByUsername_WannaWatchStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testwannawatch").
		Build()

	// 作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "見たいアニメ", "Wanna Watch Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		Build()

	// チャンネルとプログラムを作成
	channelGroupID := createTestChannelGroup(t, tx, "見たいテストグループ")
	channelID := createTestChannel(t, tx, channelGroupID, "見たいテストチャンネル")
	programID := createTestProgram(t, tx, channelID, workID)

	now := time.Now()

	// 未来のスロットを作成
	createTestSlot(t, tx, channelID, workID, episodeID, programID, now.Add(1*time.Hour))

	// ステータスを作成（kind=1: wanna_watch）
	statusID := createTestStatus(t, tx, userID, workID, 1)
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []model.EpisodeID{})

	ctx := context.Background()

	t.Run("wanna_watchステータスの作品がカレンダーに含まれる", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "testwannawatch", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		if len(calendar.Slots) != 1 {
			t.Errorf("len(Slots) = %d, want 1 (wanna_watchの作品のスロットは含まれるべき)", len(calendar.Slots))
		}
	})
}

// TestUserCalendarRepository_GetByUsername_WorkStartedOnEvent は作品の放送開始日イベントのテストです
func TestUserCalendarRepository_GetByUsername_WorkStartedOnEvent(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserCalendarRepository(queries)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("teststartedon").
		Build()

	// started_onが設定された作品を作成
	workID := createTestWorkWithStartedOn(t, tx, "開始日テストアニメ", "Started On Test Anime", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC))

	// ステータスを作成（kind=2: watching）
	statusID := createTestStatus(t, tx, userID, workID, 2)

	// ライブラリエントリを作成（program_idはNULL）
	query := `
		INSERT INTO library_entries (user_id, work_id, status_id, program_id, watched_episode_ids, created_at, updated_at)
		VALUES ($1, $2, $3, NULL, '{}', $4, $5)
	`
	_, err := tx.Exec(query, userID, workID, statusID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ライブラリエントリデータの作成に失敗しました: %v", err)
	}

	ctx := context.Background()
	now := time.Now()

	t.Run("開始日（started_on）が設定されているアニメがイベントとして含まれる", func(t *testing.T) {
		calendar, err := repo.GetByUsername(ctx, "teststartedon", now)
		if err != nil {
			t.Fatalf("GetByUsername failed: %v", err)
		}

		if len(calendar.Works) != 1 {
			t.Errorf("len(Works) = %d, want 1", len(calendar.Works))
		} else {
			work := calendar.Works[0]
			if work.ID != workID {
				t.Errorf("Work.ID = %d, want %d", work.ID, workID)
			}
			expectedStartedOn := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
			if !work.StartedOn.Equal(expectedStartedOn) {
				t.Errorf("Work.StartedOn = %v, want %v", work.StartedOn, expectedStartedOn)
			}
		}
	})
}
