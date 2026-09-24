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

// TestDelete_Successは作品のソフトデリートが作品一覧へリダイレクトすることを検証する。
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("リダイレクト先 = %v、期待値 = /db/works", location)
	}
}

// TestDelete_NotFoundは存在しない作品の削除が404を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/999999999"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_NotFoundForDeletedWorkは、すでにソフトデリート済みの作品の削除が404を
// 返すことを検証する (Railsのscope Work.without_deleted)。
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_InvalidIDは数値でないidで404を返すことを検証する。
func TestDelete_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/abc"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_HTMXRedirectはhtmxが発行する削除 (HX-Request) が素の303ではなく
// 204と作品一覧へのHX-Redirectヘッダーを返すことを検証する。htmxが押したボタンに一覧を
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusNoContent)
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/db/works" {
		t.Errorf("HX-Redirect = %q、期待値 = /db/works", got)
	}
	if loc := rr.Header().Get("Location"); loc != "" {
		t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移するため)", loc)
	}
}

// TestDelete_RequiresAdminは削除ルートがadmin専用であることを検証する (未認証は
// サインインへリダイレクト、一般ユーザーと編集者は403、adminは削除成功で作品一覧へ303)。
// committerでゲートされる書き込みエンドポイントと異なり、ここでは編集者も弾かれる
// (ADR 0009: 削除はadmin専用)。RequireAdmin自体のロール判定の網羅はmiddlewareパッケージの
// TestRequireAdminで担保する。
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
			if tt.wantLocation != "" && rr.Header().Get("Location") != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", rr.Header().Get("Location"), tt.wantLocation)
			}
		})
	}
}

// deleteRequestは管理者からの削除リクエストを組み立てる。ルートが要求しUseCaseでも
// 繰り返すロールに合わせるため。
func deleteRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("DELETE", target, nil)
	user := &model.User{ID: 1, Role: model.RoleAdmin}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestDelete_ForbiddenWithoutMiddlewareはルートミドルウェアを通さずHandlerを呼んでも、
// 認可境界が維持され403を返すことを検証する。
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}

// TestDelete_ReturnsToSubmittedListingは、削除が確認画面の送信した一覧に着地し、Annict DB
// 管理画面の外を指す値では作品一覧にフォールバックすることを検証する。細工したreturn_toで
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
		{name: "Annict DBの外は作品一覧", returnTo: "/settings", wantLocation: "/db/works"},
		{name: "外部URLは作品一覧", returnTo: "https://example.com/", wantLocation: "/db/works"},
		{name: "プロトコル相対URLは作品一覧", returnTo: "//example.com/db/works", wantLocation: "/db/works"},
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
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", got, tt.wantLocation)
			}
		})
	}
}
