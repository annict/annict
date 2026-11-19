package db_person_revision

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
	"github.com/annict/annict/go/internal/image"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func newTestRouter(t *testing.T, db *sql.DB, tx *sql.Tx) chi.Router {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	handler := NewHandler(cfg, image.NewHelper(cfg), usecase.NewGetDBPersonRevisionsUsecase(
		repository.NewPersonRepository(queries),
		repository.NewPersonRevisionRepository(queries),
		repository.NewPrefectureRepository(queries),
	))

	r := chi.NewRouter()
	r.Get("/db/people/{id}/revisions", handler.Index)
	return r
}

func newRequest(target string, user *model.User) *http.Request {
	req := httptest.NewRequest("GET", target, nil)
	if user == nil {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// createDelistedHistoryは、掲載停止と掲載再開を挟んだ人物の版を作る。
func createDelistedHistory(t *testing.T, tx *sql.Tx) model.PersonID {
	t.Helper()
	ctx := context.Background()

	personID := testutil.NewPersonBuilder(t, tx).WithName("履歴の人物").Build()
	repo := repository.NewPersonRevisionRepository(query.New(tx))
	createdAt := time.Date(2026, 11, 8, 3, 4, 0, 0, time.UTC)
	for i, v := range []struct {
		action model.DBEditAction
		name   string
		status model.PersonStatus
	}{
		{model.DBEditActionCreate, "履歴の人物", model.PersonStatusPublished},
		{model.DBEditActionDelist, "履歴の人物", model.PersonStatusDelisted},
		{model.DBEditActionUndelist, "履歴の人物", model.PersonStatusPublished},
		{model.DBEditActionUpdate, "履歴の人物 (改名)", model.PersonStatusPublished},
	} {
		editID := testutil.NewDBEditBuilder(t, tx).WithAction(v.action).WithCreatedAt(createdAt.Add(time.Duration(i) * time.Hour)).Build()
		if err := repo.Create(ctx, editID, personID, model.PersonSnapshot{Name: v.name, Status: v.status}); err != nil {
			t.Fatalf("版の作成に失敗: %v", err)
		}
	}
	return personID
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := createDelistedHistory(t, tx)

	rr := httptest.NewRecorder()
	newTestRouter(t, db, tx).ServeHTTP(rr, newRequest(fmt.Sprintf("/db/people/%d/revisions", int64(personID)), nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"<title>人物の履歴 | 履歴の人物 | Annict DB</title>",
		`<meta name="robots" content="noindex">`,
		`<time datetime="2026-11-08T03:04:00Z"`,
		"履歴の人物 (改名)",
		"作成",
		"更新",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
	// 管理者以外には、掲載を再開したあとも掲載停止・掲載再開の操作と状態の変化を出さない。
	// ログインしていない閲覧者には、編集画面へ戻るリンクも出さない。
	for _, unexpected := range []string{"掲載停止", "掲載再開", "delisted", fmt.Sprintf(`href="/db/people/%d/edit"`, int64(personID)), "編集に戻る"} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに%qが含まれている", unexpected)
		}
	}
}

func TestIndex_Admin(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := createDelistedHistory(t, tx)

	rr := httptest.NewRecorder()
	newTestRouter(t, db, tx).ServeHTTP(rr, newRequest(fmt.Sprintf("/db/people/%d/revisions", int64(personID)), &model.User{ID: 1, Role: model.RoleAdmin}))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	// 「状態」の値は、保存した値ではなく人物一覧の状態のバッジと同じ言葉で示す。
	for _, expected := range []string{"掲載停止", "掲載再開", `<td class="whitespace-pre-wrap break-all">掲載停止</td>`, `<td class="whitespace-pre-wrap break-all">公開</td>`, fmt.Sprintf(`href="/db/people/%d/edit"`, int64(personID)), "編集に戻る"} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
	for _, unexpected := range []string{">delisted<", ">published<"} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに保存した状態の値%qが含まれている", unexpected)
		}
	}
}

func TestIndex_NotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		user   *model.User
		target func(t *testing.T, tx *sql.Tx) string
	}{
		{name: "存在しない人物", target: func(t *testing.T, tx *sql.Tx) string {
			return fmt.Sprintf("/db/people/%d/revisions", int64(1)<<62)
		}},
		{name: "削除した人物", target: func(t *testing.T, tx *sql.Tx) string {
			id := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()
			return fmt.Sprintf("/db/people/%d/revisions", int64(id))
		}},
		{name: "管理者以外への掲載停止中の人物", user: &model.User{ID: 1, Role: model.RoleEditor}, target: func(t *testing.T, tx *sql.Tx) string {
			id := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
			return fmt.Sprintf("/db/people/%d/revisions", int64(id))
		}},
		{name: "数値でないID", target: func(t *testing.T, tx *sql.Tx) string {
			return "/db/people/abc/revisions"
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			rr := httptest.NewRecorder()
			newTestRouter(t, db, tx).ServeHTTP(rr, newRequest(tt.target(t, tx), tt.user))

			if rr.Code != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}
