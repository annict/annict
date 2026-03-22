package db_work

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// TestIndex はDB作品一覧ページのテスト
func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)

	// テストデータを作成
	testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ1").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ2").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

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

// TestIndex_Empty は結果が空の場合のテスト
func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	db, _ := testutil.SetupTestDB(t)

	// REPEATABLE READで並行テストがコミットしたデータを見えなくする
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	// トランザクション開始時点で存在する作品を削除
	if _, err := tx.Exec("DELETE FROM works"); err != nil {
		t.Fatalf("worksの削除に失敗しました: %v", err)
	}

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

	req := httptest.NewRequest("GET", "/db/works", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	// テーブルが表示されないことを確認
	if strings.Contains(body, "<table") {
		t.Error("response should not contain table when no works exist")
	}
}

// TestIndex_WithFilters はフィルタパラメータ付きリクエストのテスト
func TestIndex_WithFilters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

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

	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)

	handler := NewHandler(cfg, db, workRepo, numberFormatRepo, sessionManager)

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
