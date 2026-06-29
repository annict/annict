package db_work

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// TestEdit verifies the work edit form renders with its existing values pre-filled.
//
// [Ja] TestEdit はDB作品編集フォームが既存値を埋めて描画されることを検証する。
func TestEdit(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("編集対象アニメ").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	if _, err := tx.Exec(`
		UPDATE works SET
			official_site_url = 'https://example.dev/anime',
			synopsis = 'あらすじテキスト',
			twitter_username = 'anime_official',
			media = 1
		WHERE id = $1
	`, int64(workID)); err != nil {
		t.Fatalf("works のフィールド設定に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/works/%d"`, int64(workID)),
		`name="_method"`,
		`value="PATCH"`,
		"csrf_token",
		`value="編集対象アニメ"`,                   // タイトルが初期値として埋まる
		`value="https://example.dev/anime"`, // 公式サイトURLが埋まる
		"あらすじテキスト",                          // あらすじが textarea に埋まる
		`value="anime_official"`,            // Twitterユーザー名が埋まる
		`value="2024" selected`,             // シーズン年が選択済み
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

// TestEdit_NotFound verifies a nonexistent work ID returns 404.
//
// [Ja] TestEdit_NotFound は存在しない作品IDで404を返すことを検証する。
func TestEdit_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := httptest.NewRequest("GET", "/db/works/999999999/edit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestEdit_InvalidID verifies a non-numeric ID returns 404.
//
// [Ja] TestEdit_InvalidID は数値でないIDで404を返すことを検証する。
func TestEdit_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := httptest.NewRequest("GET", "/db/works/not-a-number/edit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestEdit_RequiresCommitter verifies the edit form route is protected by the
// committer role (committer gets 200, a regular user 403, and an unauthenticated
// request is redirected to sign-in).
//
// [Ja] TestEdit_RequiresCommitter は編集フォームのルートが committer ロールで保護されている
// ことを検証する (committer は 200、一般ユーザーは 403、未認証はサインインへリダイレクト)。
func TestEdit_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{id}/edit", handler.Edit)

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
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil)
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
