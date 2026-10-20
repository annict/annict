package db_person_archive

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

// newTestHandlerはテスト用トランザクション上にHandlerを組み立てる。人物のアーカイブ・解除の
// UseCaseは自前のトランザクションを開かず1文の更新で書き込むため、書き込みはテスト終了時に
// 巻き戻る。
func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)
	personRepo := repository.NewPersonRepository(queries)

	return NewHandler(
		cfg,
		sessionManager,
		testutil.NewTestFlashManager(),
		usecase.NewGetDBPersonArchiveNewUsecase(personRepo),
		usecase.NewArchivePersonUsecase(personRepo),
		usecase.NewUnarchivePersonUsecase(personRepo),
	)
}

// newTestRouterは本パッケージの3つのルートを、ミドルウェアを通さずに登録する。
func newTestRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Get("/db/people/{id}/archive/new", handler.New)
	r.Post("/db/people/{id}/archive", handler.Create)
	r.Delete("/db/people/{id}/archive", handler.Delete)
	return r
}

// committerRequestはcommitterロールである編集者からのリクエストを組み立てる。ルートが
// 要求しUseCaseでも検査を繰り返すため、画面そのものを確かめるテストはミドルウェアと同じように
// ユーザーを載せる必要がある。
func committerRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	user := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// isArchivedは人物のstatusがarchivedかを返す。
func isArchived(t *testing.T, tx *sql.Tx, id model.PersonID) bool {
	t.Helper()
	var archived bool
	if err := tx.QueryRow(`SELECT status = 'archived' FROM people WHERE id = $1`, int64(id)).Scan(&archived); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	return archived
}

// TestNewはアーカイブ確認ページが公開中の人物に対して、人物を名指しする確認フォームを描画する
// ことを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithName("非公開確認の人物").Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, committerRequest(t, "GET", fmt.Sprintf("/db/people/%d/archive/new", int64(personID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %v、期待値 = text/html; charset=utf-8", ct)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		fmt.Sprintf(`action="/db/people/%d/archive"`, int64(personID)),
		`method="POST"`,
		"csrf_token",
		"「非公開確認の人物」を非公開にしますか？",
		"<title>人物非公開 | 非公開確認の人物 | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/archive/new">`, int64(personID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_NotFoundは、存在しない人物・数値でないid・公開中でない人物で404を返すことを
// 検証する。
func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	for _, target := range []string{
		"/db/people/999999999/archive/new",
		"/db/people/abc/archive/new",
		fmt.Sprintf("/db/people/%d/archive/new", int64(archivedID)),
		fmt.Sprintf("/db/people/%d/archive/new", int64(deletedID)),
	} {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, committerRequest(t, "GET", target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
			}
		})
	}
}

// TestNew_CarriesReturnToThroughTheConfirmationは、確認ページがリンクの名指した一覧を
// キャンセルリンクとフォームの双方へ渡し、Annict DB管理画面の外を指す値では人物一覧に
// フォールバックすることを検証する。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	tests := []struct {
		name     string
		query    string
		wantHref string
	}{
		{name: "検索結果を持ち回る", query: "?return_to=%2Fdb%2Fsearch%3Fq%3Dtest", wantHref: "/db/search?q=test"},
		{name: "指定なしは人物一覧", query: "", wantHref: "/db/people"},
		{name: "外部URLは人物一覧", query: "?return_to=https%3A%2F%2Fexample.com%2F", wantHref: "/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, committerRequest(t, "GET", fmt.Sprintf("/db/people/%d/archive/new%s", int64(personID), tt.query)))

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

// TestCreateは公開中の人物をアーカイブし、確認画面が送信した一覧 (無ければ人物一覧) へ
// リダイレクトすることを検証する。
func TestCreate(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	r := newTestRouter(newTestHandler(t, db, tx))

	tests := []struct {
		name         string
		returnTo     string
		wantLocation string
	}{
		{name: "検索結果に戻る", returnTo: "/db/search?q=test", wantLocation: "/db/search?q=test"},
		{name: "空のときは人物一覧", returnTo: "", wantLocation: "/db/people"},
		{name: "外部URLは人物一覧", returnTo: "https://example.com/", wantLocation: "/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			personID := testutil.NewPersonBuilder(t, tx).Build()
			req := committerFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/archive", int64(personID)), url.Values{"return_to": {tt.returnTo}})

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", got, tt.wantLocation)
			}
			if !isArchived(t, tx, personID) {
				t.Error("人物がアーカイブされていません")
			}
		})
	}
}

// TestCreate_NotFoundForArchivedPersonは、すでにアーカイブの人物への送信が404を返すことを
// 検証する。古い確認画面からの送信が状態を打ち直さないようにするため。
func TestCreate_NotFoundForArchivedPerson(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, committerRequest(t, "POST", fmt.Sprintf("/db/people/%d/archive", int64(personID))))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
}

// TestDeleteはアーカイブの人物のアーカイブを解除し、人物一覧へリダイレクトすることを検証する。
// htmxの送信 (人物一覧のアーカイブ解除ボタン) には、押したボタンに一覧をスワップさせないよう、204と
// HX-Redirectを返す。
func TestDelete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	r := newTestRouter(newTestHandler(t, db, tx))

	t.Run("フォームの送信は303", func(t *testing.T) {
		personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
		req := committerFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/archive", int64(personID)), url.Values{"_method": {"DELETE"}, "return_to": {"/db/search?q=test"}})

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
		}
		if got := rr.Header().Get("Location"); got != "/db/search?q=test" {
			t.Errorf("リダイレクト先 = %q、期待値 = /db/search?q=test", got)
		}
		if isArchived(t, tx, personID) {
			t.Error("人物のアーカイブが解除されていません")
		}
	})

	t.Run("htmxの送信は204とHX-Redirect", func(t *testing.T) {
		personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
		req := committerRequest(t, "DELETE", fmt.Sprintf("/db/people/%d/archive", int64(personID)))
		req.Header.Set("HX-Request", "true")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNoContent)
		}
		if got := rr.Header().Get("HX-Redirect"); got != "/db/people" {
			t.Errorf("HX-Redirect = %q、期待値 = /db/people", got)
		}
		if loc := rr.Header().Get("Location"); loc != "" {
			t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移するため)", loc)
		}
	})
}

// TestDelete_NotFoundForPublishedPersonは、公開中の人物へのアーカイブ解除の送信が404を返すことを
// 検証する。
func TestDelete_NotFoundForPublishedPerson(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, committerRequest(t, "DELETE", fmt.Sprintf("/db/people/%d/archive", int64(personID))))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
}

// TestForbiddenWithoutMiddlewareは、ルートミドルウェアを通さずにHandlerを呼んでも、
// 一般ユーザーを403で拒否し、人物を書き換えないことを検証する。
func TestForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	publishedID := testutil.NewPersonBuilder(t, tx).Build()
	archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	tests := []struct {
		method string
		target string
	}{
		{method: "GET", target: fmt.Sprintf("/db/people/%d/archive/new", int64(publishedID))},
		{method: "POST", target: fmt.Sprintf("/db/people/%d/archive", int64(publishedID))},
		{method: "DELETE", target: fmt.Sprintf("/db/people/%d/archive", int64(archivedID))},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleUser}))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusForbidden {
				t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
			}
		})
	}

	if isArchived(t, tx, publishedID) || !isArchived(t, tx, archivedID) {
		t.Error("拒否したリクエストが人物の状態を書き換えました")
	}
}

// committerFormRequestは編集者からの、フォームを本文に持つリクエストを組み立てる。
func committerFormRequest(t *testing.T, method, target string, form url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	user := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}
