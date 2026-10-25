package db_person_undelisting

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

func newTestRouter(t *testing.T, db *sql.DB, tx *sql.Tx) chi.Router {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)
	handler := NewHandler(cfg, sessionManager, usecase.NewGetDBPersonUndelistingNewUsecase(repository.NewPersonRepository(queries)))

	r := chi.NewRouter()
	r.Get("/db/people/{id}/undelisting/new", handler.New)
	return r
}

// userRequestはuserからのGETリクエストを組み立てる。
func userRequest(t *testing.T, target string, user *model.User) *http.Request {
	t.Helper()

	req := httptest.NewRequest("GET", target, nil)
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// TestNewは掲載停止解除確認ページが、掲載停止の人物に対して人物を名指しし、対応記録を示す確認
// フォームを描画することを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).
		WithName("掲載停止解除確認の人物").
		WithStatus(model.PersonStatusDelisted).
		WithDelistMessage("本人から申出").
		Build()
	r := newTestRouter(t, db, tx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, userRequest(t, fmt.Sprintf("/db/people/%d/undelisting/new", int64(personID)), &model.User{ID: 1, Role: model.RoleAdmin}))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		fmt.Sprintf(`action="/db/people/%d/delisting"`, int64(personID)),
		`name="_method" value="DELETE"`,
		"csrf_token",
		`name="return_to" value="/db/people"`,
		"「掲載停止解除確認の人物」の掲載停止を解除しますか？",
		"本人から申出",
		"<title>人物掲載停止解除 | 掲載停止解除確認の人物 | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/undelisting/new">`, int64(personID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_NotFoundは、存在しない人物・数値でないid・掲載停止でない人物で404を返すことを検証する。
func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	r := newTestRouter(t, db, tx)

	targets := []string{
		"/db/people/999999999/undelisting/new",
		"/db/people/abc/undelisting/new",
	}
	for _, b := range []*testutil.PersonBuilder{
		testutil.NewPersonBuilder(t, tx),
		testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived),
		testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted),
	} {
		targets = append(targets, fmt.Sprintf("/db/people/%d/undelisting/new", int64(b.Build())))
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
// フォールバックすることを検証する。
func TestNew_CarriesReturnToThroughTheConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	r := newTestRouter(t, db, tx)

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
			target := fmt.Sprintf("/db/people/%d/undelisting/new", int64(personID)) + tt.query
			r.ServeHTTP(rr, userRequest(t, target, &model.User{ID: 1, Role: model.RoleAdmin}))

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

// TestNew_Forbiddenは、ルートのミドルウェアを通さずにHandlerを呼んでも、管理者でないユーザー
// (編集者を含む) を403で拒否し、対応記録を見せないことを検証する。
func TestNew_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).
		WithStatus(model.PersonStatusDelisted).
		WithDelistMessage("管理者だけの対応記録").
		Build()
	r := newTestRouter(t, db, tx)

	for _, user := range []*model.User{{ID: 1, Role: model.RoleUser}, {ID: 1, Role: model.RoleEditor}} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, userRequest(t, fmt.Sprintf("/db/people/%d/undelisting/new", int64(personID)), user))

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("ロール %v: ステータスコード = %v、期待値 = %v", user.Role, status, http.StatusForbidden)
		}
		if strings.Contains(rr.Body.String(), "管理者だけの対応記録") {
			t.Errorf("ロール %v: 拒否したレスポンスに対応記録が含まれています", user.Role)
		}
	}
}
