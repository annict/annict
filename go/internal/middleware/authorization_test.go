package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
)

func newUserWithRole(role int32) *model.User {
	return &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     role,
	}
}

func setUserContext(r *http.Request, user *model.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
	return r.WithContext(ctx)
}

// assertForbiddenPage asserts that a 403 is served as the shared error page rather than as the
// one-line plain text http.Error used to return. This is the 403 a viewer without the role
// actually receives on a /db page, since the route middleware refuses before any handler runs.
//
// [Ja] assertForbiddenPage は 403 が、以前 http.Error が返していた 1 行のプレーンテキストでは
// なく共通のエラーページとして配信されることを検証する。/db の画面で権限の無い閲覧者が実際に
// 受け取る 403 はこれで、ルートのミドルウェアがハンドラーに入る前に拒否するため。
func assertForbiddenPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>アクセスできません | Annict</title>",
		"この操作を行う権限がありません。",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("403 レスポンスに %q が含まれていません", expected)
		}
	}
}

// assertForbiddenIsNotRedirected asserts that a plain (non-HTMX) request keeps the response the
// shared 403 page has answered with since it replaced http.Error: the document itself, with no
// navigation instruction for the browser to act on.
//
// [Ja] assertForbiddenIsNotRedirected は、通常 (非 HTMX) のリクエストへの応答が、共通の 403
// ページが http.Error を置き換えて以降返しているもののままであることを検証する。すなわち文書
// そのものを返し、ブラウザに解釈させる遷移の指示は付けない。
func assertForbiddenIsNotRedirected(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertForbiddenPage(t, rr)
}

func TestIsAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
		want bool
	}{
		{
			name: "nilユーザーはfalse",
			user: nil,
			want: false,
		},
		{
			name: "一般ユーザーはfalse",
			user: newUserWithRole(middleware.RoleUser),
			want: false,
		},
		{
			name: "管理者はtrue",
			user: newUserWithRole(middleware.RoleAdmin),
			want: true,
		},
		{
			name: "編集者はfalse",
			user: newUserWithRole(middleware.RoleEditor),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := middleware.IsAdmin(tt.user); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEditor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
		want bool
	}{
		{
			name: "nilユーザーはfalse",
			user: nil,
			want: false,
		},
		{
			name: "一般ユーザーはfalse",
			user: newUserWithRole(middleware.RoleUser),
			want: false,
		},
		{
			name: "管理者はfalse",
			user: newUserWithRole(middleware.RoleAdmin),
			want: false,
		},
		{
			name: "編集者はtrue",
			user: newUserWithRole(middleware.RoleEditor),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := middleware.IsEditor(tt.user); got != tt.want {
				t.Errorf("IsEditor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCommitter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
		want bool
	}{
		{
			name: "nilユーザーはfalse",
			user: nil,
			want: false,
		},
		{
			name: "一般ユーザーはfalse",
			user: newUserWithRole(middleware.RoleUser),
			want: false,
		},
		{
			name: "管理者はtrue",
			user: newUserWithRole(middleware.RoleAdmin),
			want: true,
		},
		{
			name: "編集者はtrue",
			user: newUserWithRole(middleware.RoleEditor),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := middleware.IsCommitter(tt.user); got != tt.want {
				t.Errorf("IsCommitter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireCommitter(t *testing.T) {
	t.Parallel()

	// 後続ハンドラー（ミドルウェアを通過した場合に実行される）
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{
			name:       "未認証の場合はログインページにリダイレクト",
			user:       nil,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "一般ユーザーは403",
			user:       newUserWithRole(middleware.RoleUser),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "管理者はアクセス許可",
			user:       newUserWithRole(middleware.RoleAdmin),
			wantStatus: http.StatusOK,
		},
		{
			name:       "編集者はアクセス許可",
			user:       newUserWithRole(middleware.RoleEditor),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/db/works", nil)
			if tt.user != nil {
				req = setUserContext(req, tt.user)
			}
			rr := httptest.NewRecorder()

			middleware.RequireCommitter(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RequireCommitter() status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusForbidden {
				assertForbiddenIsNotRedirected(t, rr)
			}

			// 未認証の場合はリダイレクト先を確認
			if tt.user == nil {
				location := rr.Header().Get("Location")
				if location == "" {
					t.Error("未認証の場合はLocationヘッダーが必要")
				}
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	t.Parallel()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{
			name:       "未認証の場合はログインページにリダイレクト",
			user:       nil,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "一般ユーザーは403",
			user:       newUserWithRole(middleware.RoleUser),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "管理者はアクセス許可",
			user:       newUserWithRole(middleware.RoleAdmin),
			wantStatus: http.StatusOK,
		},
		{
			name:       "編集者は403",
			user:       newUserWithRole(middleware.RoleEditor),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/db/works/1", nil)
			if tt.user != nil {
				req = setUserContext(req, tt.user)
			}
			rr := httptest.NewRecorder()

			middleware.RequireAdmin(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RequireAdmin() status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusForbidden {
				assertForbiddenIsNotRedirected(t, rr)
			}

			if tt.user == nil {
				location := rr.Header().Get("Location")
				if location == "" {
					t.Error("未認証の場合はLocationヘッダーが必要")
				}
			}
		})
	}
}

// TestRequireRole_HTMXRequestIsRedirectedToForbiddenPage fixes that a refused HTMX request is sent
// to the full-page 403. The DB lists issue their archive and delete actions with hx-delete and no
// hx-target, so htmx would otherwise swap the whole 403 document into the button that was pressed.
// A viewer reaches this by keeping the list open until their role is withdrawn: the buttons are
// only rendered for those who hold the role, and an expired session is caught earlier by the CSRF
// middleware.
//
// [Ja] TestRequireRole_HTMXRequestIsRedirectedToForbiddenPage は、拒否された HTMX リクエストが
// 全画面の 403 へ送られることを固定する。DB 一覧の非公開・削除は hx-target を指定していない
// hx-delete で発行するため、そのままでは 403 の文書全体が押したボタンの中にスワップされる。
// 閲覧者がここに到達するのは、一覧を開いたままロールが外れた場合に限られる (ボタンはロールを
// 持つ閲覧者にしか描画されず、セッション切れは先に CSRF ミドルウェアが捕まえる)。
func TestRequireRole_HTMXRequestIsRedirectedToForbiddenPage(t *testing.T) {
	t.Parallel()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		middleware func(http.Handler) http.Handler
		user       *model.User
		path       string
	}{
		{
			name:       "RequireCommitter",
			middleware: middleware.RequireCommitter,
			user:       newUserWithRole(middleware.RoleUser),
			path:       "/db/episodes/1/archive",
		},
		{
			name:       "RequireAdmin",
			middleware: middleware.RequireAdmin,
			user:       newUserWithRole(middleware.RoleEditor),
			path:       "/db/episodes/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			req.Header.Set("HX-Request", "true")
			req = setUserContext(req, tt.user)
			rr := httptest.NewRecorder()

			tt.middleware(nextHandler).ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
			}
			if got := rr.Header().Get("HX-Redirect"); got != httperror.ForbiddenPath {
				t.Errorf("HX-Redirect = %q, want %q", got, httperror.ForbiddenPath)
			}
			assertForbiddenPage(t, rr)
		})
	}
}
