package db_work

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
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

	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}

	getDBWorksUC := usecase.NewGetDBWorksUsecase(workRepo)
	getDBWorkFormOptionsUC := usecase.NewGetDBWorkFormOptionsUsecase(numberFormatRepo)
	getDBWorkEditUC := usecase.NewGetDBWorkEditUsecase(workRepo, numberFormatRepo)
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator())
	updateWorkUC := usecase.NewUpdateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator())

	return NewHandler(cfg, sessionManager, testutil.NewTestFlashManager(), testutil.NewTestImageHelper(), getDBWorksUC, getDBWorkFormOptionsUC, getDBWorkEditUC, createWorkUC, updateWorkUC)
}

// TestIndex はDB作品一覧ページのテスト
func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// テストデータを作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ1").
		WithTitleKana("てすとあにめいち").
		WithTitleEn("Test Anime 1").
		WithMedia(1).
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
		// The ID column links to the work's public page in a new tab.
		//
		// [Ja] ID 列は作品の公開ページを新しいタブで開くリンクになる。
		fmt.Sprintf(`href="/works/%d"`, int64(workID)),
		`target="_blank"`,
		// The ID link exposes an aria-label announcing that it opens in a new tab.
		//
		// [Ja] ID リンクは新しいタブで開くことを知らせる aria-label を持つ。
		fmt.Sprintf(`aria-label="作品 %d を新しいタブで開く"`, int64(workID)),
		// The title cell shows the kana / English titles; the media column shows the media label.
		//
		// [Ja] タイトルセルにふりがな・英語タイトルを、メディア列にメディア名を表示する。
		"てすとあにめいち",
		"Test Anime 1",
		"TV",
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

// TestIndex_WithSeasonAndSlotFilters はリリース時期の複数選択・放送予定未登録フィルタのテスト
func TestIndex_WithSeasonAndSlotFilters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	spring2024 := testutil.NewWorkBuilder(t, tx).
		WithTitle("2024春アニメ").
		WithSeason(2024, testutil.SeasonSpring).
		Build()
	testutil.NewWorkBuilder(t, tx).
		WithTitle("2024夏アニメ").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works?season_slugs=2024-spring&filter_no_slots=1", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	// Only the work in the selected season is shown; the other season is filtered out.
	//
	// [Ja] 選択したシーズンの作品のみが表示され、別シーズンは除外される。
	if !strings.Contains(body, "2024春アニメ") {
		t.Error("選択したシーズンに一致する作品が表示されるべき")
	}
	if strings.Contains(body, "2024夏アニメ") {
		t.Error("選択外のシーズンの作品は除外されるべき")
	}
	// The matched work's edit link pins the match precisely.
	//
	// [Ja] 一致した作品の編集リンクで一致を厳密に確認する。
	if !strings.Contains(body, fmt.Sprintf(`href="/db/works/%d/edit"`, int64(spring2024))) {
		t.Error("一致した作品の編集リンクが表示されるべき")
	}
	// The selected season option is re-rendered in its selected state.
	//
	// [Ja] 選択したシーズンオプションが selected 状態で再描画される。
	if !strings.Contains(body, `<option value="2024-spring" selected>`) {
		t.Error("選択したシーズンオプションが selected で再描画されるべき")
	}
	// The no-slots checkbox itself renders in its checked state.
	//
	// [Ja] 放送予定未登録チェックボックス自体が checked 状態で描画される。
	if !strings.Contains(body, `<input type="checkbox" name="filter_no_slots" value="1" checked>`) {
		t.Error("放送予定未登録フィルタのチェックボックスが checked で描画されるべき")
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

// TestNew_RequiresCommitter verifies the new-work form route is protected by the
// committer role (committer proceeds, a regular user 403, an unauthenticated request is
// redirected to sign-in).
//
// [Ja] TestNew_RequiresCommitter は新規作成フォームのルートが committer ロールで保護されて
// いることを検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへ
// リダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/new", handler.New)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{
			name:       "未認証はサインインへリダイレクト",
			user:       nil,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "一般ユーザーは403",
			user:       &model.User{ID: 1, Role: model.RoleUser},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "管理者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleAdmin},
			wantStatus: http.StatusOK,
		},
		{
			name:       "編集者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/works/new", nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
