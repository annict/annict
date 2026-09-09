package db_work_unarchive

import (
	"context"
	"database/sql"
	"fmt"
	"html"
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

	return NewHandler(cfg, sessionManager, usecase.NewGetDBWorkUnarchiveNewUsecase(workRepo))
}

// assertNotFoundPage asserts that a 404 is served as the shared error page, so a reader who
// follows a stale link lands on a page that says what happened and offers a way back.
//
// [Ja] assertNotFoundPage は 404 が共通のエラーページとして配信されることを検証する。古い
// リンクを辿った読み手が、何が起きたかを述べ戻る導線を持つページに着地するようにするため。
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

// TestNew verifies the publish-confirmation page renders the confirm form for an archived
// work.
//
// [Ja] TestNew は公開確認ページがアーカイブ済みの作品に対して確認フォームを描画することを
// 検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("公開確認作品").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/unarchive/new", int64(workID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/works/%d/archive"`, int64(workID)),
		`method="POST"`,
		`name="_method" value="DELETE"`,
		"csrf_token",
		"公開確認作品",
		"<title>作品公開 | 公開確認作品 | Annict DB</title>",
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
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("代表 URL 確認作品").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	want := fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/unarchive/new">`, int64(workID))

	for _, target := range []string{
		fmt.Sprintf("/db/works/%d/unarchive/new", int64(workID)),
		fmt.Sprintf("/db/works/000%d/unarchive/new", int64(workID)),
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

// TestNew_NotFoundForPublishedWork verifies the confirmation page returns 404 for a work that
// is not currently archived, so the screen never offers to publish an already published work.
//
// [Ja] TestNew_NotFoundForPublishedWork は、現在アーカイブ済みでない作品に対して確認ページが
// 404 を返すことを検証する。すでに公開中の作品の公開を画面が提案しないようにするため。
func TestNew_NotFoundForPublishedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("既に公開中").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/unarchive/new", int64(workID))))

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
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, "/db/works/abc/unarchive/new"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_DocumentTitleWithoutWorkName verifies that a work whose title is only whitespace
// leaves the document title as the page name alone, the same name the heading falls back to.
//
// [Ja] TestNew_DocumentTitleWithoutWorkName は、タイトルが空白文字だけの作品では文書タイトルが
// 画面名だけになることを検証する。見出しがフォールバックする名前と同じものになる。
func TestNew_DocumentTitleWithoutWorkName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle(" \t ").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/unarchive/new", int64(workID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>作品公開 | Annict DB</title>",
		">作品公開</h1>",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestNew_ForbiddenWithoutMiddleware verifies the handler preserves the authorization
// boundary even when it is invoked without the route middleware.
//
// [Ja] TestNew_ForbiddenWithoutMiddleware はルートミドルウェアを通さず Handler を呼んでも、
// 認可境界が維持され 403 を返すことを検証する。
func TestNew_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("GET", "/db/works/1/unarchive/new", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

// TestNew_RequiresCommitter verifies the confirmation route is protected by the committer
// role (an unauthenticated request is redirected, a regular user 403, a committer proceeds),
// so only whoever may submit the publish can reach the screen offering it. The full role
// matrix of RequireCommitter itself is covered by TestRequireCommitter in the middleware
// package.
//
// [Ja] TestNew_RequiresCommitter は確認ルートが committer ロールで保護されていることを検証
// する (未認証はリダイレクト、一般ユーザーは 403、committer は処理続行)。公開を送信できる者
// だけが、それを提案する画面に到達するようにするため。RequireCommitter 自体のロール判定の
// 網羅は middleware パッケージの TestRequireCommitter で担保する。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{id}/unarchive/new", handler.New)

	target := fmt.Sprintf("/db/works/%d/unarchive/new", int64(workID))
	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者は確認画面を表示", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", target, nil)
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

func getRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("GET", target, nil)
	user := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestNew_CarriesReturnToThroughTheConfirmation verifies the confirmation page hands the listing
// the link named to both its cancel link and its form, and falls back to the work list when the
// value names something outside the Annict DB admin UI, so a crafted return_to cannot send the
// reader off-site.
//
// [Ja] TestNew_CarriesReturnToThroughTheConfirmation は、確認ページがリンクの名指した一覧を
// キャンセルリンクとフォームの双方へ渡すこと、および Annict DB 管理画面の外を指す値では作品一覧に
// フォールバックすることを検証する。細工した return_to で読み手をサイト外へ送れないようにするため。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("戻り先確認作品").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/unarchive/new", handler.New)

	tests := []struct {
		name     string
		query    string
		wantHref string
	}{
		{name: "検索結果を持ち回る", query: "?return_to=%2Fdb%2Fsearch%3Fq%3Dtest", wantHref: "/db/search?q=test"},
		{name: "指定なしは作品一覧", query: "", wantHref: "/db/works"},
		{name: "Annict DB の外は作品一覧", query: "?return_to=%2Fsettings", wantHref: "/db/works"},
		{name: "外部 URL は作品一覧", query: "?return_to=https%3A%2F%2Fexample.com%2F", wantHref: "/db/works"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			target := fmt.Sprintf("/db/works/%d/unarchive/new%s", int64(workID), tt.query)
			r.ServeHTTP(rr, getRequest(t, target))

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, expected := range []string{
				fmt.Sprintf(`<a href="%s"`, html.EscapeString(tt.wantHref)),
				fmt.Sprintf(`name="return_to" value="%s"`, html.EscapeString(tt.wantHref)),
			} {
				if !strings.Contains(body, expected) {
					t.Errorf("response doesn't contain expected string: %q", expected)
				}
			}
		})
	}
}
