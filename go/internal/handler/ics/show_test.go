package ics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestShow_UserNotFound ユーザーが見つからない場合は404を返すテスト
func TestShow_UserNotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	handler := NewHandler(cfg, userCalendarRepo)

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

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	handler := NewHandler(cfg, userCalendarRepo)

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

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		Domain: "annict.com",
	}

	userCalendarRepo := repository.NewUserCalendarRepository(queries)
	handler := NewHandler(cfg, userCalendarRepo)

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

	db, tx := testutil.SetupTestDB(t)
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
	handler := NewHandler(cfg, userCalendarRepo)

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

	db, tx := testutil.SetupTestDB(t)
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
	handler := NewHandler(cfg, userCalendarRepo)

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

	db, tx := testutil.SetupTestDB(t)
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
	handler := NewHandler(cfg, userCalendarRepo)

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
