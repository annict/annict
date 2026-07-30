package db_work

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// deleteWorkFromEditRedirect deletes the work whose edit page the create handler
// redirected to (/db/works/{id}/edit). Create commits the new work through the usecase's
// own transaction, so the row escapes the test's rolled-back transaction and must be
// cleaned up to keep the parallel test suite isolated. It is a no-op when the location is
// not an edit path.
//
// [Ja] deleteWorkFromEditRedirect は create ハンドラーがリダイレクトした編集ページ
// (/db/works/{id}/edit) の作品を削除する。Create は UseCase 自前のトランザクションで新規
// work をコミットするため、行はテストのロールバックされるトランザクションの外に残り、並行
// テストを隔離するには削除が要る。location が編集パスでない場合は何もしない。
func deleteWorkFromEditRedirect(db *sql.DB, location string) {
	const prefix, suffix = "/db/works/", "/edit"
	if !strings.HasPrefix(location, prefix) || !strings.HasSuffix(location, suffix) {
		return
	}
	idStr := strings.TrimSuffix(strings.TrimPrefix(location, prefix), suffix)
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		_, _ = db.Exec("DELETE FROM works WHERE id = $1", id)
	}
}

// TestCreate_ValidationError はバリデーションエラー時にフォームが再表示されることをテスト
func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	// タイトルとメディアが空のリクエスト
	form := url.Values{}
	form.Set("title", "")
	form.Set("media", "")
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 422 Unprocessable Entityが返ることを確認
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// フォームが再表示されていることを確認
	expectedContents := []string{
		"<title>作品登録 | Annict DB</title>",
		"<form",
		`action="/db/works"`,
		`method="POST"`,
		`role="alert"`,
		// The re-rendered page is the new-work form, so og:url names its own GET path and not
		// the POST endpoint (which serves the work list on GET).
		//
		// [Ja] 再描画するのは新規作成フォームなので、og:url は POST 先ではなくそのページ自身の
		// GET パスを指す (POST 先は GET では作品一覧を返す)。
		`<meta property="og:url" content="https://test.annict.com/db/works/new">`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestCreate_ValidationError_PreservesFormValues はバリデーションエラー時にフォーム値が保持されることをテスト
func TestCreate_ValidationError_PreservesFormValues(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	// タイトルはあるがメディアが空のリクエスト
	form := url.Values{}
	form.Set("title", "テスト作品")
	form.Set("media", "")
	form.Set("title_kana", "てすとさくひん")
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// 入力値が保持されていることを確認
	if !strings.Contains(body, "テスト作品") {
		t.Error("response doesn't preserve title value")
	}
	if !strings.Contains(body, "てすとさくひん") {
		t.Error("response doesn't preserve title_kana value")
	}
}

// TestCreate_Success は正常に作品が作成されることをテスト
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	// 有効なフォームデータ
	form := url.Values{}
	form.Set("title", "新しいアニメ作品")
	form.Set("media", "1") // tv
	req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// ユースケースが独自のトランザクションでコミットするため、作成された作品を即座に削除する。
	// 並行テストへの影響を最小化するため、アサーション前に同期的にクリーンアップする。
	location := rr.Header().Get("Location")
	deleteWorkFromEditRedirect(db, location)

	// 303 See Otherでリダイレクトされることを確認
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// The create handler redirects to the new work's edit page, mirroring the Rails create
	// action (db_edit_work_path) and the Update handler.
	//
	// [Ja] create ハンドラーは新規作品の編集ページへリダイレクトする (Rails の create
	// アクション db_edit_work_path や Update ハンドラーと同じ)。
	if !strings.HasPrefix(location, "/db/works/") || !strings.HasSuffix(location, "/edit") {
		t.Errorf("handler returned wrong redirect location: got %v", location)
	}
}

// TestCreate_RequiresCommitter verifies the work creation route is protected by the
// committer role (committer proceeds, a regular user 403, an unauthenticated request is
// redirected to sign-in).
//
// [Ja] TestCreate_RequiresCommitter は作品作成ルートが committer ロールで保護されている
// ことを検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへ
// リダイレクト)。
func TestCreate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Post("/db/works", handler.Create)

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
			name:       "編集者はアクセス許可 (作成成功でリダイレクト)",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Set("title", "認可テスト作成作品")
			form.Set("media", "1") // tv
			req := httptest.NewRequest("POST", "/db/works", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// The committer path commits a new work through the usecase's own transaction;
			// delete it so the parallel test suite is unaffected (mirrors TestCreate_Success).
			//
			// [Ja] committer 経路は UseCase 自前のトランザクションで新規 work をコミットするため、
			// 並行テストへ影響しないよう削除する (TestCreate_Success と同じ)。
			deleteWorkFromEditRedirect(db, rr.Header().Get("Location"))

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
