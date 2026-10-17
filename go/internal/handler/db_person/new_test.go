package db_person

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

func withTestUser(req *http.Request, user *model.User) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.New(rr, withTestUser(httptest.NewRequest("GET", "/db/people/new", nil), &model.User{ID: 1, Role: model.RoleEditor}))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<title>人物登録 | Annict DB</title>",
		`<meta property="og:url" content="https://test.annict.com/db/people/new">`,
		`action="/db/people"`,
		`method="POST"`,
		`name="csrf_token"`,
		`<textarea id="rows" name="rows"`,
		`aria-describedby="rows-description"`,
		"徳川家康,とくがわいえやす",
		`href="/db/people"`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestNew_RequiresCommitterは一括登録フォームがcommitterロールで保護されていることを
// 検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/people/new", handler.New)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者はアクセス許可", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusOK},
		{name: "管理者はアクセス許可", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/people/new", nil)
			if tt.user != nil {
				req = withTestUser(req, tt.user)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// TestNew_ForbiddenWithoutMiddlewareは、ルートのミドルウェアを通らない一般ユーザーの
// リクエストをUseCaseが拒否し、Handlerが403ページに変換することを検証する。
func TestNew_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.New(rr, withTestUser(httptest.NewRequest("GET", "/db/people/new", nil), &model.User{ID: 1, Role: model.RoleUser}))

	if rr.Code != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusForbidden)
	}
	if !strings.Contains(rr.Body.String(), "この操作を行う権限がありません。") {
		t.Error("403ページが描画されていません")
	}
}
