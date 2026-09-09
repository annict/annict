package db_work

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// TestDelete_Success verifies soft-deleting a work redirects to the work list.
//
// [Ja] TestDelete_Success は作品のソフトデリートが作品一覧へリダイレクトすることを検証する。
func TestDelete_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("削除対象作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, fmt.Sprintf("/db/works/%d", int64(workID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("handler returned wrong redirect location: got %v want /db/works", location)
	}
}

// TestDelete_NotFound verifies deleting a nonexistent work returns 404.
//
// [Ja] TestDelete_NotFound は存在しない作品の削除が 404 を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/999999999"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_NotFoundForDeletedWork verifies deleting an already soft-deleted work returns
// 404 (Rails scope Work.without_deleted).
//
// [Ja] TestDelete_NotFoundForDeletedWork は、すでにソフトデリート済みの作品の削除が 404 を
// 返すことを検証する (Rails の scope Work.without_deleted)。
func TestDelete_NotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("既に削除済み").WithDeletedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, fmt.Sprintf("/db/works/%d", int64(workID))))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_InvalidID verifies a non-numeric id returns 404.
//
// [Ja] TestDelete_InvalidID は数値でない id で 404 を返すことを検証する。
func TestDelete_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/abc"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_HTMXRedirect verifies that an htmx-issued delete (HX-Request) responds with 204
// and an HX-Redirect header to the work list instead of the plain 303 redirect, so htmx
// navigates rather than swapping the followed list page into the clicked button.
//
// [Ja] TestDelete_HTMXRedirect は htmx が発行する削除 (HX-Request) が素の 303 ではなく
// 204 と作品一覧への HX-Redirect ヘッダーを返すことを検証する。htmx が押したボタンに一覧を
// スワップせず遷移するようにするため。
func TestDelete_HTMXRedirect(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("htmx削除対象").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	req := deleteRequest(t, fmt.Sprintf("/db/works/%d", int64(workID)))
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/db/works" {
		t.Errorf("HX-Redirect = %q, want /db/works", got)
	}
	if loc := rr.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, want empty (htmx navigates via HX-Redirect)", loc)
	}
}

// TestDelete_RequiresAdmin verifies the delete route is admin-only: an unauthenticated
// request is redirected to sign-in, a regular user and an editor get 403, and an admin
// passes through to a successful delete (303 to the work list). Unlike the committer-gated
// write endpoints, an editor is rejected here (ADR 0009: deletion is admin-only). The full
// role matrix of RequireAdmin itself is covered by TestRequireAdmin in the middleware
// package.
//
// [Ja] TestDelete_RequiresAdmin は削除ルートが admin 専用であることを検証する (未認証は
// サインインへリダイレクト、一般ユーザーと編集者は 403、admin は削除成功で作品一覧へ 303)。
// committer でゲートされる書き込みエンドポイントと異なり、ここでは編集者も弾かれる
// (ADR 0009: 削除は admin 専用)。RequireAdmin 自体のロール判定の網羅は middleware パッケージの
// TestRequireAdmin で担保する。
func TestDelete_RequiresAdmin(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("削除認可テスト").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireAdmin).Delete("/db/works/{id}", handler.Delete)

	target := fmt.Sprintf("/db/works/%d", int64(workID))
	tests := []struct {
		name         string
		user         *model.User
		wantStatus   int
		wantLocation string
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者は403", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusForbidden},
		{name: "管理者は削除成功でリダイレクト", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantStatus: http.StatusSeeOther, wantLocation: "/db/works"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", target, nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantLocation != "" && rr.Header().Get("Location") != tt.wantLocation {
				t.Errorf("location = %q, want %q", rr.Header().Get("Location"), tt.wantLocation)
			}
		})
	}
}

// deleteRequest builds a delete request from an administrator, the role the route requires and
// the usecase repeats.
//
// [Ja] deleteRequest は管理者からの削除リクエストを組み立てる。ルートが要求し UseCase でも
// 繰り返すロールに合わせるため。
func deleteRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("DELETE", target, nil)
	user := &model.User{ID: 1, Role: model.RoleAdmin}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestDelete_ForbiddenWithoutMiddleware verifies the handler preserves the authorization
// boundary even when it is invoked without the route middleware.
//
// [Ja] TestDelete_ForbiddenWithoutMiddleware はルートミドルウェアを通さず Handler を呼んでも、
// 認可境界が維持され 403 を返すことを検証する。
func TestDelete_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可境界テスト").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", fmt.Sprintf("/db/works/%d", int64(workID)), nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

// TestDelete_ReturnsToSubmittedListing verifies the delete lands on the listing the confirmation
// screen submitted, and falls back to the work list when the value names something outside the
// Annict DB admin UI, so a crafted return_to cannot send the reader off-site.
//
// [Ja] TestDelete_ReturnsToSubmittedListing は、削除が確認画面の送信した一覧に着地し、Annict DB
// 管理画面の外を指す値では作品一覧にフォールバックすることを検証する。細工した return_to で
// 読み手をサイト外へ送れないようにするため。
func TestDelete_ReturnsToSubmittedListing(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Delete("/db/works/{id}", handler.Delete)

	tests := []struct {
		name         string
		returnTo     string
		wantLocation string
	}{
		{name: "検索結果に戻る", returnTo: "/db/search?q=%E6%A4%9C%E7%B4%A2", wantLocation: "/db/search?q=%E6%A4%9C%E7%B4%A2"},
		{name: "空のときは作品一覧", returnTo: "", wantLocation: "/db/works"},
		{name: "Annict DB の外は作品一覧", returnTo: "/settings", wantLocation: "/db/works"},
		{name: "外部 URL は作品一覧", returnTo: "https://example.com/", wantLocation: "/db/works"},
		{name: "プロトコル相対 URL は作品一覧", returnTo: "//example.com/db/works", wantLocation: "/db/works"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workID := testutil.NewWorkBuilder(t, tx).WithTitle("戻り先テスト").WithMedia(1).Build()
			form := url.Values{"_method": {"DELETE"}, "return_to": {tt.returnTo}}
			req := httptest.NewRequest("POST", fmt.Sprintf("/db/works/%d", int64(workID)), strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleAdmin}))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("Location = %q, want %q", got, tt.wantLocation)
			}
		})
	}
}
