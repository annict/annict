package db_work

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// validUpdateForm returns a form that passes validation for the work update endpoint.
//
// [Ja] validUpdateForm は作品更新エンドポイントの検証を通過するフォームを返す。
func validUpdateForm(title string) url.Values {
	form := url.Values{}
	form.Set("title", title)
	form.Set("media", "1") // tv
	return form
}

func patchRequest(t *testing.T, workID int64, form url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/db/works/%d", workID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// TestUpdate_ValidationError verifies that invalid input re-renders the edit form with
// a 422 and the submitted values preserved. Validation runs before the work is loaded,
// so the endpoint responds 422 without needing the work to exist.
//
// [Ja] TestUpdate_ValidationError は不正入力で編集フォームが 422 で再描画され、送信値が
// 保持されることを検証する。検証は work 読み込みより前に走るため、work の存在なしに 422 を
// 返す。
func TestUpdate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	form := url.Values{}
	form.Set("title", "") // required
	form.Set("media", "")
	form.Set("title_kana", "ほぞんされるかな")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, 123, form))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		`action="/db/works/123"`,
		`name="_method"`,
		`value="PATCH"`,
		"text-red-600",     // エラーメッセージのスタイル
		`value="ほぞんされるかな"`, // 入力値が保持される
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestUpdate_Success verifies a valid update redirects back to the edit page.
//
// [Ja] TestUpdate_Success は有効な更新が編集ページへリダイレクトすることを検証する。
func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("更新前作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, int64(workID), validUpdateForm("更新後作品")))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != fmt.Sprintf("/db/works/%d/edit", int64(workID)) {
		t.Errorf("handler returned wrong redirect location: got %v", location)
	}
}

// TestUpdate_NotFound verifies a valid update for a nonexistent work returns 404.
//
// [Ja] TestUpdate_NotFound は存在しない作品への有効な更新が 404 を返すことを検証する。
func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, 999999999, validUpdateForm("存在しない作品の更新")))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestUpdate_InvalidID verifies a non-numeric ID returns 404.
//
// [Ja] TestUpdate_InvalidID は数値でないIDで404を返すことを検証する。
func TestUpdate_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	req := httptest.NewRequest("PATCH", "/db/works/not-a-number", strings.NewReader(validUpdateForm("x").Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestUpdate_RequiresCommitter verifies the update route is protected by the committer
// role (committer proceeds, a regular user 403, an unauthenticated request is
// redirected to sign-in).
//
// [Ja] TestUpdate_RequiresCommitter は更新ルートが committer ロールで保護されている
// ことを検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへ
// リダイレクト)。
func TestUpdate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Patch("/db/works/{id}", handler.Update)

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
			name:       "編集者はアクセス許可 (更新成功でリダイレクト)",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := patchRequest(t, int64(workID), validUpdateForm("更新後作品"))
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
