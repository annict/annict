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

// deleteWorkFromEditRedirectはcreateハンドラーがリダイレクトした編集ページ
// (/db/works/{id}/edit) の作品を削除する。CreateはUseCase自前のトランザクションで新規
// workをコミットするため、行はテストのロールバックされるトランザクションの外に残り、並行
// テストを隔離するには削除が要る。locationが編集パスでない場合は何もしない。
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

// TestCreate_ValidationErrorはバリデーションエラー時にフォームが再表示されることをテスト
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// フォームが再表示されていることを確認
	expectedContents := []string{
		"<title>作品登録 | Annict DB</title>",
		"<form",
		`action="/db/works"`,
		`method="POST"`,
		`role="alert"`,
		// 再描画するのは新規作成フォームなので、og:urlはPOST先ではなくそのページ自身の
		// GETパスを指す (POST先はGETでは作品一覧を返す)。
		`<meta property="og:url" content="https://test.annict.com/db/works/new">`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestCreate_ValidationError_PreservesFormValuesはバリデーションエラー時にフォーム値が保持されることをテスト
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	// 入力値が保持されていることを確認
	if !strings.Contains(body, "テスト作品") {
		t.Error("レスポンスがtitleの値を保持していない")
	}
	if !strings.Contains(body, "てすとさくひん") {
		t.Error("レスポンスがtitle_kanaの値を保持していない")
	}
}

// TestCreate_Successは正常に作品が作成されることをテスト
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}

	// createハンドラーは新規作品の編集ページへリダイレクトする (Railsのcreate
	// アクションdb_edit_work_pathやUpdateハンドラーと同じ)。
	if !strings.HasPrefix(location, "/db/works/") || !strings.HasSuffix(location, "/edit") {
		t.Errorf("リダイレクト先 = %v、期待値 = /db/works/で始まり/editで終わるURL", location)
	}
}

// TestCreate_RequiresCommitterは作品作成ルートがcommitterロールで保護されている
// ことを検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへ
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

			// committer経路はUseCase自前のトランザクションで新規workをコミットするため、
			// 並行テストへ影響しないよう削除する (TestCreate_Successと同じ)。
			deleteWorkFromEditRedirect(db, rr.Header().Get("Location"))

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
