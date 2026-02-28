package db_work

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreate_ValidationError はバリデーションエラー時にフォームが再表示されることをテスト
func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

	// タイトルとメディアが空のリクエスト
	form := url.Values{}
	form.Set("title", "")
	form.Set("media", "")
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 422 Unprocessable Entityが返ることを確認
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// フォームが再表示されていることを確認
	expectedContents := []string{
		"<form",
		`action="/db/works"`,
		`method="POST"`,
		"text-red-600", // エラーメッセージのスタイル
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestCreate_ValidationError_PreservesFormValues はバリデーションエラー時にフォーム値が保持されることをテスト
func TestCreate_ValidationError_PreservesFormValues(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

	// タイトルはあるがメディアが空のリクエスト
	form := url.Values{}
	form.Set("title", "テスト作品")
	form.Set("media", "")
	form.Set("title_kana", "てすとさくひん")
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// 入力値が保持されていることを確認
	if !strings.Contains(body, "テスト作品") {
		t.Error("response doesn't preserve title value")
	}
	if !strings.Contains(body, "てすとさくひん") {
		t.Error("response doesn't preserve title_kana value")
	}
}

// TestCreate_Success は正常に作品が作成されることをテスト
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

	// 有効なフォームデータ
	form := url.Values{}
	form.Set("title", "新しいアニメ作品")
	form.Set("media", "1") // tv
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 303 See Otherでリダイレクトされることを確認
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// リダイレクト先が /db/works?highlight=XXX であることを確認
	location := rr.Header().Get("Location")
	if !strings.HasPrefix(location, "/db/works?highlight=") {
		t.Errorf("handler returned wrong redirect location: got %v", location)
	}
}
