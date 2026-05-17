package ics

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// TestShow_UserNotFound ユーザーが見つからない場合は404を返すテスト
func TestShow_UserNotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	// 存在しないユーザーでリクエスト
	req := httptest.NewRequest("GET", "/@nonexistent_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("存在しないユーザーの場合: expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestShow_EmptyUsername usernameが空の場合は404を返すテスト
func TestShow_EmptyUsername(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// /ics エンドポイントでusernameパラメータなしでリクエスト
	req := httptest.NewRequest("GET", "/ics", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("usernameが空の場合: expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestShow_QueryParam クエリパラメータでusernameを指定するテスト
func TestShow_QueryParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// /ics?username=nonexistent でリクエスト（存在しないユーザー）
	req := httptest.NewRequest("GET", "/ics?username=nonexistent_user", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("クエリパラメータで存在しないユーザーの場合: expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestShow_Success 正常系のテスト（ユーザーが存在する場合）
func TestShow_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	testutil.NewUserBuilder(t, tx).
		WithUsername("ics_test_user").
		WithEmail("ics_test@example.com").
		Build()

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_test_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("正常系のステータスコード: expected %d, got %d", http.StatusOK, rr.Code)
	}

	// Content-Typeの確認
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/calendar; charset=utf-8" {
		t.Errorf("Content-Type: expected %q, got %q", "text/calendar; charset=utf-8", contentType)
	}

	// Content-Dispositionの確認
	contentDisposition := rr.Header().Get("Content-Disposition")
	if contentDisposition != `attachment; filename="annict.ics"` {
		t.Errorf("Content-Disposition: expected %q, got %q", `attachment; filename="annict.ics"`, contentDisposition)
	}

	// レスポンスボディにiCalendarヘッダーが含まれていることを確認
	body := rr.Body.String()
	if !strings.Contains(body, "BEGIN:VCALENDAR") {
		t.Error("レスポンスボディにBEGIN:VCALENDARが含まれていない")
	}
	if !strings.Contains(body, "END:VCALENDAR") {
		t.Error("レスポンスボディにEND:VCALENDARが含まれていない")
	}
	if !strings.Contains(body, "X-WR-CALNAME:Annict@ics_test_user") {
		t.Error("レスポンスボディにX-WR-CALNAME:Annict@ics_test_userが含まれていない")
	}
}

// TestShow_QueryParamSuccess クエリパラメータでユーザーが存在する場合のテスト
func TestShow_QueryParamSuccess(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	testutil.NewUserBuilder(t, tx).
		WithUsername("ics_query_user").
		WithEmail("ics_query@example.com").
		Build()

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成（/icsエンドポイント用）
	r := chi.NewRouter()
	r.Get("/ics", handler.Show)

	req := httptest.NewRequest("GET", "/ics?username=ics_query_user", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("クエリパラメータ正常系のステータスコード: expected %d, got %d", http.StatusOK, rr.Code)
	}

	// Content-Typeの確認
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/calendar; charset=utf-8" {
		t.Errorf("Content-Type: expected %q, got %q", "text/calendar; charset=utf-8", contentType)
	}
}

// TestShow_EpisodeNumberFormatting エピソード番号がそのまま出力されるテスト
func TestShow_EpisodeNumberFormatting(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("ics_episode_test").
		WithEmail("ics_episode@example.com").
		Build()

	// 作品を作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ").
		Build()

	// チャンネルを作成（チャンネルグループは自動作成）
	channelID := testutil.NewChannelBuilder(t, tx).
		WithName("TOKYO MX").
		Build()

	// プログラムを作成
	programID := testutil.NewProgramBuilder(t, tx).
		WithChannelID(channelID).
		WithWorkID(workID).
		Build()

	// エピソードを作成（#付きの番号）
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("#4").
		WithTitle("サブタイトル").
		Build()

	// 放送枠を作成（現在時刻から1時間後に放送開始）
	slotStartTime := time.Now().Add(1 * time.Hour)
	testutil.NewSlotBuilder(t, tx).
		WithWorkID(workID).
		WithEpisodeID(episodeID).
		WithChannelID(channelID).
		WithProgramID(programID).
		WithStartedAt(slotStartTime).
		Build()

	// ライブラリエントリを作成（視聴中）
	testutil.NewLibraryEntryBuilder(t, tx).
		WithUserID(userID).
		WithWorkID(workID).
		WithProgramID(programID).
		WithStatus("watching").
		Build()

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_episode_test/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード: expected %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	// エピソード番号がそのまま出力されていることを確認
	if !strings.Contains(body, "#4") {
		t.Error("レスポンスにエピソード番号(#4)が含まれていない")
	}

	// ##4（二重ハッシュ）が含まれていないことを確認
	if strings.Contains(body, "##4") {
		t.Error("レスポンスに二重ハッシュ(##4)が含まれている")
	}
}

// TestShow_DeletedUser 削除されたユーザーの場合は404を返すテスト
func TestShow_DeletedUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("ics_deleted_user").
		WithEmail("ics_deleted@example.com").
		Build()

	// ユーザーを削除（deleted_atを設定）
	_, err := tx.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		t.Fatalf("ユーザーの削除に失敗しました: %v", err)
	}

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_deleted_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("削除されたユーザーの場合: expected %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestShow_EmptyCalendar 視聴リストが空の場合は空のカレンダーを返すテスト
func TestShow_EmptyCalendar(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// ユーザーを作成（ライブラリエントリなし）
	testutil.NewUserBuilder(t, tx).
		WithUsername("ics_empty_user").
		WithEmail("ics_empty@example.com").
		Build()

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_empty_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("空のカレンダーの場合: expected %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	// カレンダーヘッダーが含まれていることを確認
	if !strings.Contains(body, "BEGIN:VCALENDAR") {
		t.Error("レスポンスボディにBEGIN:VCALENDARが含まれていない")
	}
	if !strings.Contains(body, "X-WR-CALNAME:Annict@ics_empty_user") {
		t.Error("レスポンスボディにX-WR-CALNAMEが含まれていない")
	}

	// イベントが含まれていないことを確認
	if strings.Contains(body, "BEGIN:VEVENT") {
		t.Error("空のカレンダーにVEVENTが含まれている")
	}
}

// TestShow_WorkStartedOnEvent 開始日が設定された作品がイベントとして含まれるテスト
func TestShow_WorkStartedOnEvent(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("ics_startedon_user").
		WithEmail("ics_startedon@example.com").
		Build()

	// started_onが設定された作品を作成
	startedOn := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	workID := createWorkWithStartedOn(t, tx, "開始日テストアニメ", startedOn)

	// ステータスを作成（kind=2: watching）
	statusID := createStatus(t, tx, userID, workID, 2)

	// ライブラリエントリを作成（program_idなし）
	_, err := tx.Exec(`
		INSERT INTO library_entries (user_id, work_id, status_id, program_id, watched_episode_ids, created_at, updated_at)
		VALUES ($1, $2, $3, NULL, '{}', NOW(), NOW())
	`, int64(userID), int64(workID), statusID)
	if err != nil {
		t.Fatalf("ライブラリエントリの作成に失敗しました: %v", err)
	}

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_startedon_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード: expected %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	// 作品イベントが含まれていることを確認
	if !strings.Contains(body, "BEGIN:VEVENT") {
		t.Error("レスポンスにVEVENTが含まれていない")
	}
	if !strings.Contains(body, "開始日テストアニメ") {
		t.Error("レスポンスに作品タイトルが含まれていない")
	}
	// 終日イベント（VALUE=DATE）であることを確認
	if !strings.Contains(body, "VALUE=DATE") {
		t.Error("レスポンスにVALUE=DATE（終日イベント）が含まれていない")
	}
}

// TestShow_WannaWatchStatus wanna_watchステータスの作品がカレンダーに含まれるテスト
func TestShow_WannaWatchStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("ics_wannawatch_user").
		WithEmail("ics_wannawatch@example.com").
		Build()

	// 作品を作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("見たいアニメ").
		Build()

	// チャンネルを作成
	channelID := testutil.NewChannelBuilder(t, tx).
		WithName("テストチャンネル見たい").
		Build()

	// プログラムを作成
	programID := testutil.NewProgramBuilder(t, tx).
		WithChannelID(channelID).
		WithWorkID(workID).
		Build()

	// エピソードを作成
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		WithTitle("第1話").
		Build()

	// 放送枠を作成（現在時刻から1時間後に放送開始）
	slotStartTime := time.Now().Add(1 * time.Hour)
	testutil.NewSlotBuilder(t, tx).
		WithWorkID(workID).
		WithEpisodeID(episodeID).
		WithChannelID(channelID).
		WithProgramID(programID).
		WithStartedAt(slotStartTime).
		Build()

	// ライブラリエントリを作成（wanna_watch = 見たい）
	testutil.NewLibraryEntryBuilder(t, tx).
		WithUserID(userID).
		WithWorkID(workID).
		WithProgramID(programID).
		WithStatus("wanna_watch").
		Build()

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	handler := NewHandler(cfg, getUserCalendarUC)

	// chiルーターを作成
	r := chi.NewRouter()
	r.Get("/@{username}/ics", handler.Show)

	req := httptest.NewRequest("GET", "/@ics_wannawatch_user/ics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// ステータスコードの確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード: expected %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	// イベントが含まれていることを確認
	if !strings.Contains(body, "BEGIN:VEVENT") {
		t.Error("wanna_watchの作品のイベントが含まれていない")
	}
	if !strings.Contains(body, "見たいアニメ") {
		t.Error("レスポンスに作品タイトルが含まれていない")
	}
}

// createWorkWithStartedOn はstarted_onを設定した作品を作成するヘルパー
func createWorkWithStartedOn(t *testing.T, tx *sql.Tx, title string, startedOn time.Time) model.WorkID {
	t.Helper()

	query := `
		INSERT INTO works (
			title, title_kana, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count, started_on,
			created_at, updated_at
		) VALUES (
			$1, '', 0, '', '', 2025, 2, 0, 0, $2, NOW(), NOW()
		) RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, title, startedOn).Scan(&id)
	if err != nil {
		t.Fatalf("作品データの作成に失敗しました: %v", err)
	}

	return model.WorkID(id)
}

// createStatus はテスト用ステータスを作成するヘルパー
func createStatus(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, kind int) int64 {
	t.Helper()

	query := `
		INSERT INTO statuses (user_id, work_id, kind, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(query, int64(userID), int64(workID), kind).Scan(&id)
	if err != nil {
		t.Fatalf("ステータスデータの作成に失敗しました: %v", err)
	}

	return id
}
