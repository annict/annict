package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestUserCalendarRepository_GetByUsername(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
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
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []int64{})

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
		createTestLibraryEntry(t, tx, userID2, workID, statusID2, programID, []int64{episodeID})

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
	db, tx := testutil.SetupTestDB(t)
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
	createTestLibraryEntry(t, tx, userID, workID, statusID, programID, []int64{})

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

// createTestWorkWithStartedOn はstarted_onを設定した作品を作成します
func createTestWorkWithStartedOn(t *testing.T, tx *sql.Tx, title, titleEn string, startedOn time.Time) int64 {
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

	return id
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
func createTestProgram(t *testing.T, tx *sql.Tx, channelID, workID int64) int64 {
	t.Helper()

	query := `
		INSERT INTO programs (channel_id, work_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, channelID, workID, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("プログラムデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestSlot はテスト用スロットを作成します
func createTestSlot(t *testing.T, tx *sql.Tx, channelID, workID, episodeID, programID int64, startedAt time.Time) int64 {
	t.Helper()

	query := `
		INSERT INTO slots (channel_id, work_id, episode_id, program_id, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, channelID, workID, episodeID, programID, startedAt, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("スロットデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestStatus はテスト用ステータスを作成します
func createTestStatus(t *testing.T, tx *sql.Tx, userID, workID int64, kind int) int64 {
	t.Helper()

	query := `
		INSERT INTO statuses (user_id, work_id, kind, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, userID, workID, kind, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		t.Fatalf("ステータスデータの作成に失敗しました: %v", err)
	}

	return id
}

// createTestLibraryEntry はテスト用ライブラリエントリを作成します
func createTestLibraryEntry(t *testing.T, tx *sql.Tx, userID, workID, statusID, programID int64, watchedEpisodeIDs []int64) {
	t.Helper()

	// 配列をPostgreSQL形式に変換
	watchedIDsStr := "{"
	for i, id := range watchedEpisodeIDs {
		if i > 0 {
			watchedIDsStr += ","
		}
		watchedIDsStr += string(rune('0' + id%10)) // 簡易的な変換
	}
	watchedIDsStr += "}"

	query := `
		INSERT INTO library_entries (user_id, work_id, status_id, program_id, watched_episode_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::bigint[], $6, $7)
	`

	_, err := tx.Exec(query, userID, workID, statusID, programID, watchedIDsStr, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ライブラリエントリデータの作成に失敗しました: %v", err)
	}
}
