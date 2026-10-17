package db_person_unarchive

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
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

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)

	return NewHandler(cfg, sessionManager, usecase.NewGetDBPersonUnarchiveNewUsecase(repository.NewPersonRepository(queries)))
}

// userRequestはuserからのGETリクエストを組み立てる。userがnilなら未認証のリクエストになる。
func userRequest(t *testing.T, target string, user *model.User) *http.Request {
	t.Helper()
	req := httptest.NewRequest("GET", target, nil)
	if user == nil {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestNewはアーカイブ解除確認ページが、アーカイブ解除できる人物に対して人物を名指しする確認フォームを描画する
// ことを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithName("アーカイブ解除確認の人物").WithStatus(model.PersonStatusArchived).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/people/{id}/unarchive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, userRequest(t, fmt.Sprintf("/db/people/%d/unarchive/new", int64(personID)), &model.User{ID: 1, Role: model.RoleEditor}))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		fmt.Sprintf(`action="/db/people/%d/archive"`, int64(personID)),
		`name="_method" value="DELETE"`,
		"csrf_token",
		`name="return_to" value="/db/people"`,
		"「アーカイブ解除確認の人物」のアーカイブを解除しますか？",
		"<title>人物アーカイブ解除 | アーカイブ解除確認の人物 | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/unarchive/new">`, int64(personID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_NotFoundは、存在しない人物・数値でないid・アーカイブ解除できない状態の人物で404を返す
// ことを検証する。
func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/people/{id}/unarchive/new", handler.New)

	targets := []string{
		"/db/people/999999999/unarchive/new",
		"/db/people/abc/unarchive/new",
	}
	for _, b := range []*testutil.PersonBuilder{
		testutil.NewPersonBuilder(t, tx),
		testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted),
		testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted),
	} {
		targets = append(targets, fmt.Sprintf("/db/people/%d/unarchive/new", int64(b.Build())))
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, userRequest(t, target, &model.User{ID: 1, Role: model.RoleAdmin}))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
			}
		})
	}
}

// TestNew_CarriesReturnToThroughTheConfirmationは、確認ページがリンクの名指した一覧を
// キャンセルリンクとフォームの双方へ渡し、Annict DB管理画面の外を指す値では人物一覧に
// フォールバックすることを検証する。細工したreturn_toで読み手をサイト外へ送れないようにするため。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/people/{id}/unarchive/new", handler.New)

	tests := []struct {
		name     string
		query    string
		wantHref string
	}{
		{name: "検索結果を持ち回る", query: "?return_to=%2Fdb%2Fsearch%3Fq%3Dtest", wantHref: "/db/search?q=test"},
		{name: "外部URLは人物一覧", query: "?return_to=https%3A%2F%2Fexample.com%2F", wantHref: "/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			target := fmt.Sprintf("/db/people/%d/unarchive/new", int64(personID)) + tt.query
			r.ServeHTTP(rr, userRequest(t, target, &model.User{ID: 1, Role: model.RoleEditor}))

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

// TestNew_Forbiddenは、ルートのミドルウェアを通さずにHandlerを呼んでも、操作する権限の無い
// ユーザーを403で拒否することを検証する。
func TestNew_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/people/{id}/unarchive/new", handler.New)

	for _, user := range []*model.User{{ID: 1, Role: model.RoleUser}} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, userRequest(t, fmt.Sprintf("/db/people/%d/unarchive/new", int64(personID)), user))

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("ロール %v: ステータスコード = %v、期待値 = %v", user.Role, status, http.StatusForbidden)
		}
	}
}
