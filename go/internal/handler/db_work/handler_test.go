package db_work

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)

	listDbWorksUC := usecase.NewListDbWorksUsecase(workRepo)
	getDbWorkFormOptionsUC := usecase.NewGetDbWorkFormOptionsUsecase(numberFormatRepo)
	getDbWorkEditUC := usecase.NewGetDbWorkEditUsecase(workRepo, numberFormatRepo)
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, validator.NewDbWorkCreateValidator())

	return NewHandler(cfg, sessionManager, testutil.NewTestFlashManager(), listDbWorksUC, getDbWorkFormOptionsUC, getDbWorkEditUC, createWorkUC)
}

// TestIndex はDB作品一覧ページのテスト
func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// テストデータを作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ1").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ2").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"テストアニメ1",
		"テストアニメ2",
		"2024",
		"<table",
		"<thead",
		"<tbody",
		// Each row links to its edit form via DBWorkEditPath.
		//
		// [Ja] 各行が DBWorkEditPath 経由で編集フォームへリンクする。
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content-type: got %v want %v", ct, expectedContentType)
	}
}

// TestIndex_WithFilters はフィルタパラメータ付きリクエストのテスト
func TestIndex_WithFilters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works?filter_no_episodes=1&filter_no_image=1&page=1", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	// フィルタのチェックボックスがチェックされていることを確認
	if !strings.Contains(body, `checked`) {
		t.Error("response should contain checked checkboxes for active filters")
	}
}

// TestNew はDB作品新規作成フォームのテスト
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works/new", nil)
	rr := httptest.NewRecorder()

	handler.New(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<form",
		`action="/db/works"`,
		`method="POST"`,
		"csrf_token",
		`name="title"`,
		`name="media"`,
		`name="season_year"`,
		`name="season_name"`,
		`name="number_format_id"`,
		`name="no_episodes"`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content-type: got %v want %v", ct, expectedContentType)
	}
}
