package ics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
