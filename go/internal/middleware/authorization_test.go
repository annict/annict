package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

			if tt.user == nil {
				location := rr.Header().Get("Location")
				if location == "" {
					t.Error("未認証の場合はLocationヘッダーが必要")
				}
			}
		})
	}
}
