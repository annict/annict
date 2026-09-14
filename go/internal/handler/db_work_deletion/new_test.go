package db_work_deletion

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

	return NewHandler(cfg, sessionManager, usecase.NewGetDBWorkDeletionNewUsecase(workRepo))
}

// assertNotFoundPageは404が共通のエラーページとして配信されることを検証する。古い
// リンクを辿った読み手が、何が起きたかを述べ戻る導線を持つページに着地するようにするため。
func assertNotFoundPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>ページが見つかりません | Annict</title>",
		"ページが見つかりません",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("404レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestNewは削除確認ページが未削除の作品に対して確認フォームを描画することを、公開中と
// アーカイブ済みの双方で検証する (削除はどちらも受け付ける)。
func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.WorkBuilder) *testutil.WorkBuilder
	}{
		{
			name: "公開中",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b
			},
		},
		{
			name: "アーカイブ済み",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b.WithUnpublishedAt(time.Now())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			workID := tt.prepare(testutil.NewWorkBuilder(t, tx).WithTitle("削除確認作品").WithMedia(1)).Build()
			handler := newTestHandler(t, db, tx)

			r := chi.NewRouter()
			r.Get("/db/works/{id}/deletion/new", handler.New)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/deletion/new", int64(workID))))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}

			body := rr.Body.String()
			expectedContents := []string{
				"<form",
				fmt.Sprintf(`action="/db/works/%d"`, int64(workID)),
				`method="POST"`,
				`name="_method" value="DELETE"`,
				"csrf_token",
				"削除確認作品",
				"<title>作品削除 | 削除確認作品 | Annict DB</title>",
			}
			for _, expected := range expectedContents {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}

			expectedContentType := "text/html; charset=utf-8"
			if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("Content-Type = %v、期待値 = %v", ct, expectedContentType)
			}
		})
	}
}

// TestNew_OGURLはog:urlがパース済みの作品IDから組み立てたページ自身のGETパスに
// なることを検証する。IDを先頭ゼロ付きで書いたリンクでも、そのページの代表URLは1つに
// 揃う。
func TestNew_OGURL(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("代表URL確認作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	want := fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/deletion/new">`, int64(workID))

	for _, target := range []string{
		fmt.Sprintf("/db/works/%d/deletion/new", int64(workID)),
		fmt.Sprintf("/db/works/000%d/deletion/new", int64(workID)),
	} {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(t, target))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, want) {
				t.Errorf("レスポンスに含まれていない文字列 = %q", want)
			}
		})
	}
}

// TestNew_NotFoundForDeletedWorkは、すでに削除済みの作品に対して確認ページが404を
// 返すことを検証する。古いリンクが、すでに失われた作品の削除を提案せずnot foundのページに
// 着地するようにするため。
func TestNew_NotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("既に削除済み").WithDeletedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/deletion/new", int64(workID))))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_InvalidIDは数値でないidで404を返すことを検証する。
func TestNew_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, "/db/works/abc/deletion/new"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_DocumentTitleWithoutWorkNameは、タイトルが空白文字だけの作品では文書タイトルが
// 画面名だけになることを検証する。見出しがフォールバックする名前と同じものになる。
func TestNew_DocumentTitleWithoutWorkName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle(" \t ").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/deletion/new", int64(workID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>作品削除 | Annict DB</title>",
		">作品削除</h1>",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_ForbiddenWithoutMiddlewareはルートミドルウェアを通さずHandlerを呼んでも、
// 認可境界が維持され403を返すことを検証する。
func TestNew_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("GET", "/db/works/1/deletion/new", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}

// TestNew_RequiresAdminは確認ルートが、送信先の削除と同じくadmin専用であることを
// 検証する (ADR 0009: 削除はadmin専用)。編集者が、自分では送信できない画面に到達しない
// ようにするため (未認証はリダイレクト、一般ユーザーと編集者は403、adminは確認画面を表示)。
// RequireAdmin自体のロール判定の網羅はmiddlewareパッケージのTestRequireAdminで担保する。
func TestNew_RequiresAdmin(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireAdmin).Get("/db/works/{id}/deletion/new", handler.New)

	target := fmt.Sprintf("/db/works/%d/deletion/new", int64(workID))
	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者は403", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusForbidden},
		{name: "管理者は確認画面を表示", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantStatus: http.StatusOK},
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func getRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("GET", target, nil)
	user := &model.User{ID: 1, Role: model.RoleAdmin}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestNew_CarriesReturnToThroughTheConfirmationは、確認ページがリンクの名指した一覧を
// キャンセルリンクとフォームの双方へ渡すこと、およびAnnict DB管理画面の外を指す値では作品一覧に
// フォールバックすることを検証する。細工したreturn_toで読み手をサイト外へ送れないようにするため。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("戻り先確認作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/deletion/new", handler.New)

	tests := []struct {
		name     string
		query    string
		wantHref string
	}{
		{name: "検索結果を持ち回る", query: "?return_to=%2Fdb%2Fsearch%3Fq%3Dtest", wantHref: "/db/search?q=test"},
		{name: "指定なしは作品一覧", query: "", wantHref: "/db/works"},
		{name: "Annict DBの外は作品一覧", query: "?return_to=%2Fsettings", wantHref: "/db/works"},
		{name: "外部URLは作品一覧", query: "?return_to=https%3A%2F%2Fexample.com%2F", wantHref: "/db/works"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			target := fmt.Sprintf("/db/works/%d/deletion/new%s", int64(workID), tt.query)
			r.ServeHTTP(rr, getRequest(t, target))

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, expected := range []string{
				fmt.Sprintf(`<a href="%s"`, html.EscapeString(tt.wantHref)),
				fmt.Sprintf(`name="return_to" value="%s"`, html.EscapeString(tt.wantHref)),
			} {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
		})
	}
}
