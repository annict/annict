package db_work_archive

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// assertNotFoundPageは404が、以前http.Errorが返していた1行のプレーンテキストでは
// なく共通のエラーページとして配信されることを検証する。古いリンクを辿った読み手が、何が
// 起きたかを述べ戻る導線を持つページに着地するようにするため。
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

// TestNewは非公開確認ページが公開中の作品に対して確認フォームを描画することを検証する。
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/works/%d/archive"`, int64(workID)),
		`method="POST"`,
		"csrf_token",
		"非公開確認作品",
		"<title>作品非公開 | 非公開確認作品 | Annict DB</title>",
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
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, want) {
				t.Errorf("レスポンスに含まれていない文字列 = %q", want)
			}
		})
	}
}

// TestNew_NotFoundForArchivedWorkは、現在公開中でない (すでにアーカイブ済みの) 作品に
// 対して確認ページが404を返すことを検証する。
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
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, "/db/works/abc/archive/new"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestNew_RequiresCommitterは確認ルートがcommitterロールで保護されていることを検証
// する (committerは処理続行、一般ユーザーは403、未認証はリダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{id}/archive/new", handler.New)

	assertRequiresCommitter(t, r, "GET", fmt.Sprintf("/db/works/%d/archive/new", int64(workID)))
}

// TestCreate_Successは公開中の作品の非公開が作品一覧へリダイレクトすることを検証する。
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("リダイレクト先 = %v、期待値 = /db/works", location)
	}
}

// TestCreate_NotFoundは存在しない作品の非公開が404を返すことを検証する。
func TestCreate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/works/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, postRequest(t, "/db/works/999999999/archive"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestCreate_RequiresCommitterは非公開ルートがcommitterロールで保護されていることを
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

// assertRequiresCommitterは未認証・一般ユーザー・編集者で指定ルートを叩き、
// RequireCommitterの判定表を検証する。編集者では非エラーのステータスを許容する
// (committerはミドルウェアを通過する。ハンドラー自体の結果は状態次第で、成功 / not-foundの
// テストで担保する)。adminもcommitterだがここでは回さない。RequireCommitter自体の
// admin / editor / userの網羅はmiddlewareパッケージのTestRequireCommitterで担保する。
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}

	// committer (編集者) はミドルウェアを通過し、403やサインインへのリダイレクトで弾かれ
	// ないこと。成功したPOSTも303リダイレクトするため、ステータスだけでなく「リダイレクト先が
	// サインインページでないこと」でミドルウェア通過を判定する。
	t.Run("編集者は通過", func(t *testing.T) {
		req := httptest.NewRequest(method, target, nil)
		req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code == http.StatusForbidden {
			t.Errorf("編集者が403で拒否された")
		}
		if strings.HasPrefix(rr.Header().Get("Location"), "/sign_in") {
			t.Errorf("編集者のリダイレクト先 = %s、期待値 = /sign_inで始まらないこと", rr.Header().Get("Location"))
		}
	})
}

// TestDelete_Successはアーカイブ済みの作品の再公開が作品一覧へリダイレクトすることを
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/db/works" {
		t.Errorf("リダイレクト先 = %v、期待値 = /db/works", location)
	}
}

// TestDelete_NotFoundは存在しない作品の再公開が404を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(t, "/db/works/999999999/archive"))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestDelete_RequiresCommitterは再公開ルートがcommitterロールで保護されていることを
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

// TestDelete_HTMXRedirectはhtmxが発行する再公開 (HX-Request) が素の303ではなく
// 204と作品一覧へのHX-Redirectヘッダーを返すことを検証する。htmxが押したボタンに一覧を
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusNoContent)
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/db/works" {
		t.Errorf("HX-Redirect = %q、期待値 = /db/works", got)
	}
	if loc := rr.Header().Get("Location"); loc != "" {
		t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移するため)", loc)
	}
}

// committerRequestはcommitterロールである編集者からのリクエストを組み立てる。ルートが
// 要求しUseCaseでも検査を繰り返すため、画面そのものを確かめるテストはミドルウェアと同じように
// ユーザーを載せる必要がある。
func committerRequest(t *testing.T, method string, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	user := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

func getRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return committerRequest(t, "GET", target)
}

func postRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return committerRequest(t, "POST", target)
}

func deleteRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return committerRequest(t, "DELETE", target)
}

// TestNew_DocumentTitleWithoutWorkNameは、タイトルが空白文字だけの作品では文書タイトルが
// 画面名だけになることを検証する。見出しがフォールバックする名前と同じものになる。
func TestNew_DocumentTitleWithoutWorkName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle(" \t ").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(t, fmt.Sprintf("/db/works/%d/archive/new", int64(workID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>作品非公開 | Annict DB</title>",
		">作品非公開</h1>",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_ForbiddenWithoutMiddlewareはルートミドルウェアを通さずHandlerを呼んでも、
// 確認ページの認可境界が維持され403を返すことを検証する。
func TestNew_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("GET", "/db/works/1/archive/new", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}

// TestCreate_ForbiddenWithoutMiddlewareはルートミドルウェアを通さずHandlerを呼んでも、
// 非公開の認可境界が維持され403を返すことを検証する。
func TestCreate_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開認可境界テスト").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/works/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("POST", fmt.Sprintf("/db/works/%d/archive", int64(workID)), nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}

// TestDelete_ForbiddenWithoutMiddlewareはルートミドルウェアを通さずHandlerを呼んでも、
// 再公開の認可境界が維持され403を返すことを検証する。
func TestDelete_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("再公開認可境界テスト").WithMedia(1).WithUnpublishedAt(time.Now()).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/works/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", fmt.Sprintf("/db/works/%d/archive", int64(workID)), nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}

// TestNew_CarriesReturnToThroughTheConfirmationは、確認ページがリンクの名指した一覧を
// キャンセルリンクとフォームの双方へ渡すこと、およびAnnict DB管理画面の外を指す値では作品一覧に
// フォールバックすることを検証する。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("戻り先確認作品").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/archive/new", handler.New)

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
			target := fmt.Sprintf("/db/works/%d/archive/new%s", int64(workID), tt.query)
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

// TestCreate_ReturnsToSubmittedListingは、非公開が確認画面の送信した一覧に着地し、
// Annict DB管理画面の外を指す値では作品一覧にフォールバックすることを検証する。細工した
// return_toで読み手をサイト外へ送れないようにするため。
func TestCreate_ReturnsToSubmittedListing(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Post("/db/works/{id}/archive", handler.Create)

	for _, tt := range returnToCases() {
		t.Run(tt.name, func(t *testing.T) {
			workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開戻り先テスト").WithMedia(1).Build()
			form := url.Values{"return_to": {tt.returnTo}}
			req := httptest.NewRequest("POST", fmt.Sprintf("/db/works/%d/archive", int64(workID)), strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))

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

// TestDelete_ReturnsToSubmittedListingは、再公開が確認画面の送信した一覧に着地すること、
// フォールバックが非公開と同じであることを検証する。
func TestDelete_ReturnsToSubmittedListing(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Delete("/db/works/{id}/archive", handler.Delete)

	for _, tt := range returnToCases() {
		t.Run(tt.name, func(t *testing.T) {
			workID := testutil.NewWorkBuilder(t, tx).
				WithTitle("再公開戻り先テスト").
				WithMedia(1).
				WithUnpublishedAt(time.Now()).
				Build()
			form := url.Values{"_method": {"DELETE"}, "return_to": {tt.returnTo}}
			req := httptest.NewRequest("POST", fmt.Sprintf("/db/works/%d/archive", int64(workID)), strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))

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

// returnToCasesは本パッケージの2つの書き込みエンドポイントが同じ結果を返すreturn_toの
// 集合。Annict DBの一覧は尊重し、それ以外は作品一覧にフォールバックする。
func returnToCases() []struct {
	name         string
	returnTo     string
	wantLocation string
} {
	return []struct {
		name         string
		returnTo     string
		wantLocation string
	}{
		{name: "検索結果に戻る", returnTo: "/db/search?q=test", wantLocation: "/db/search?q=test"},
		{name: "空のときは作品一覧", returnTo: "", wantLocation: "/db/works"},
		{name: "Annict DBの外は作品一覧", returnTo: "/settings", wantLocation: "/db/works"},
		{name: "外部URLは作品一覧", returnTo: "https://example.com/", wantLocation: "/db/works"},
		{name: "プロトコル相対URLは作品一覧", returnTo: "//example.com/db/works", wantLocation: "/db/works"},
	}
}
