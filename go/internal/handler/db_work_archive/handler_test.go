package db_work_archive

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)

	getDBWorkArchiveNewUC := usecase.NewGetDBWorkArchiveNewUsecase(workRepo)
	archiveWorkUC := usecase.NewArchiveWorkUsecase(db, workRepo, animeRepo)
	unarchiveWorkUC := usecase.NewUnarchiveWorkUsecase(db, workRepo, animeRepo)

	return NewHandler(cfg, sessionManager, testutil.NewTestFlashManager(), getDBWorkArchiveNewUC, archiveWorkUC, unarchiveWorkUC)
}

// assertNotFoundPage asserts that a 404 is served as the shared error page rather than as the
// one-line plain text http.Error used to return, so a reader who follows a stale link lands on
// a page that says what happened and offers a way back.
//
// [Ja] assertNotFoundPage は 404 が、以前 http.Error が返していた 1 行のプレーンテキストでは
// なく共通のエラーページとして配信されることを検証する。古いリンクを辿った読み手が、何が
// 起きたかを述べ戻る導線を持つページに着地するようにするため。
func assertNotFoundPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>ページが見つかりません | Annict</title>",
		"ページが見つかりません",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("404 レスポンスに %q が含まれていません", expected)
		}
	}
}

// TestNew verifies the archive-confirmation page renders the confirm form for a published
// work.
//
// [Ja] TestNew は非公開確認ページが公開中の作品に対して確認フォームを描画することを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開確認作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/archive/new", int64(workID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/works/%d/archive"`, int64(workID)),
		`method="POST"`,
		"csrf_token",
		"非公開確認作品",
		"<title>作品を非公開にする | Annict DB</title>",
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

// TestNew_OGURL verifies that og:url names the page's own GET path built from the parsed work
// ID, so that a link spelling the ID with leading zeros still declares the one representative
// URL of that page.
//
// [Ja] TestNew_OGURL は og:url がパース済みの作品 ID から組み立てたページ自身の GET パスに
// なることを検証する。ID を先頭ゼロ付きで書いたリンクでも、そのページの代表 URL は 1 つに
// 揃う。
func TestNew_OGURL(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("代表 URL 確認作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	want := fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/archive/new">`, int64(workID))

	for _, target := range []string{
		fmt.Sprintf("/db/works/%d/archive/new", int64(workID)),
		fmt.Sprintf("/db/works/000%d/archive/new", int64(workID)),
	} {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(t, target))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, want) {
				t.Errorf("response doesn't contain expected string: %q", want)
			}
		})
	}
}

// TestNew_NotFoundForArchivedWork verifies the confirmation page returns 404 for a work
// that is not currently published (already archived).
//
// [Ja] TestNew_NotFoundForArchivedWork は、現在公開中でない (すでにアーカイブ済みの) 作品に
// 対して確認ページが 404 を返すことを検証する。
func TestNew_NotFoundForArchivedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("既に非公開").WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/archive/new", int64(workID))))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_InvalidID verifies a non-numeric id returns 404.
//
// [Ja] TestNew_InvalidID は数値でない id で 404 を返すことを検証する。
func TestNew_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, "/db/works/abc/archive/new"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_RequiresCommitter verifies the confirmation route is protected by the committer
// role (committer proceeds, a regular user 403, an unauthenticated request is redirected).
//
// [Ja] TestNew_RequiresCommitter は確認ルートが committer ロールで保護されていることを検証
// する (committer は処理続行、一般ユーザーは 403、未認証はリダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{id}/archive/new", handler.New)

	assertRequiresCommitter(t, r, "GET", fmt.Sprintf("/db/works/%d/archive/new", int64(workID)))
}

// TestCreate_Success verifies archiving a published work redirects to the work list.
//
// [Ja] TestCreate_Success は公開中の作品の非公開が作品一覧へリダイレクトすることを検証する。
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開対象作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/works/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, postRequest(t, fmt.Sprintf("/db/works/%d/archive", int64(workID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("handler returned wrong redirect location: got %v want /db/works", location)
	}
}

// TestCreate_NotFound verifies archiving a nonexistent work returns 404.
//
// [Ja] TestCreate_NotFound は存在しない作品の非公開が 404 を返すことを検証する。
func TestCreate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/works/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, postRequest(t, "/db/works/999999999/archive"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestCreate_RequiresCommitter verifies the archive route is protected by the committer
// role.
//
// [Ja] TestCreate_RequiresCommitter は非公開ルートが committer ロールで保護されていることを
// 検証する。
func TestCreate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開認可テスト").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Post("/db/works/{id}/archive", handler.Create)

	assertRequiresCommitter(t, r, "POST", fmt.Sprintf("/db/works/%d/archive", int64(workID)))
}

// assertRequiresCommitter drives the given route with an unauthenticated request, a regular
// user and an editor, asserting the RequireCommitter matrix. It accepts any non-error status
// for the editor (committers pass the middleware; the handler's own outcome depends on state
// and is covered by the success / not-found tests). Admin is also a committer but is not
// exercised here; the full admin / editor / user matrix of RequireCommitter itself is covered
// by TestRequireCommitter in the middleware package.
//
// [Ja] assertRequiresCommitter は未認証・一般ユーザー・編集者で指定ルートを叩き、
// RequireCommitter の判定表を検証する。編集者では非エラーのステータスを許容する
// (committer はミドルウェアを通過する。ハンドラー自体の結果は状態次第で、成功 / not-found の
// テストで担保する)。admin も committer だがここでは回さない。RequireCommitter 自体の
// admin / editor / user の網羅は middleware パッケージの TestRequireCommitter で担保する。
func assertRequiresCommitter(t *testing.T, r chi.Router, method, target string) {
	t.Helper()

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(method, target, nil)
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

	// A committer (editor) passes the middleware and must not be rejected with a 403 or a
	// sign-in redirect. A successful POST also redirects (303), so the middleware pass is
	// detected by the redirect NOT targeting the sign-in page rather than by the status
	// alone.
	//
	// [Ja] committer (編集者) はミドルウェアを通過し、403 やサインインへのリダイレクトで弾かれ
	// ないこと。成功した POST も 303 リダイレクトするため、ステータスだけでなく「リダイレクト先が
	// サインインページでないこと」でミドルウェア通過を判定する。
	t.Run("編集者は通過", func(t *testing.T) {
		req := httptest.NewRequest(method, target, nil)
		req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code == http.StatusForbidden {
			t.Errorf("committer was rejected with 403")
		}
		if strings.HasPrefix(rr.Header().Get("Location"), "/sign_in") {
			t.Errorf("committer was redirected to sign-in: %s", rr.Header().Get("Location"))
		}
	})
}

// TestDelete_Success verifies re-publishing an archived work redirects to the work list.
//
// [Ja] TestDelete_Success はアーカイブ済みの作品の再公開が作品一覧へリダイレクトすることを
// 検証する。
func TestDelete_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("再公開対象作品").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, fmt.Sprintf("/db/works/%d/archive", int64(workID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("handler returned wrong redirect location: got %v want /db/works", location)
	}
}

// TestDelete_NotFound verifies re-publishing a nonexistent work returns 404.
//
// [Ja] TestDelete_NotFound は存在しない作品の再公開が 404 を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/999999999/archive"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_RequiresCommitter verifies the un-archive route is protected by the committer
// role.
//
// [Ja] TestDelete_RequiresCommitter は再公開ルートが committer ロールで保護されていることを
// 検証する。
func TestDelete_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("再公開認可テスト").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Delete("/db/works/{id}/archive", handler.Delete)

	assertRequiresCommitter(t, r, "DELETE", fmt.Sprintf("/db/works/%d/archive", int64(workID)))
}

// TestDelete_HTMXRedirect verifies that an htmx-issued un-archive (HX-Request) responds with
// 204 and an HX-Redirect header to the work list instead of the plain 303 redirect, so htmx
// navigates rather than swapping the followed list page into the clicked button.
//
// [Ja] TestDelete_HTMXRedirect は htmx が発行する再公開 (HX-Request) が素の 303 ではなく
// 204 と作品一覧への HX-Redirect ヘッダーを返すことを検証する。htmx が押したボタンに一覧を
// スワップせず遷移するようにするため。
func TestDelete_HTMXRedirect(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("htmx再公開対象").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}/archive", handler.Delete)

	req := deleteRequest(t, fmt.Sprintf("/db/works/%d/archive", int64(workID)))
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

func getRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return httptest.NewRequest("GET", target, nil)
}

func postRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return httptest.NewRequest("POST", target, nil)
}

func deleteRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return httptest.NewRequest("DELETE", target, nil)
}
