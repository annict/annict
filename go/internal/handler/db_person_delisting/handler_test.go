package db_person_delisting

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

// newTestHandlerはテスト用トランザクション上にHandlerを組み立てる。人物の掲載停止・解除の
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
		usecase.NewGetDBPersonDelistingNewUsecase(personRepo),
		usecase.NewGetDBPersonUndelistingNewUsecase(personRepo),
		usecase.NewDelistPersonUsecase(personRepo),
		usecase.NewUndelistPersonUsecase(personRepo),
	)
}

// newTestRouterは本パッケージの3つのルートを、ミドルウェアを通さずに登録する。
func newTestRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Get("/db/people/{id}/delisting/new", handler.New)
	r.Post("/db/people/{id}/delisting", handler.Create)
	r.Delete("/db/people/{id}/delisting", handler.Delete)
	return r
}

// adminFormRequestは管理者からのフォームの送信を組み立てる。ルートが要求しUseCaseでも検査を
// 繰り返すため、画面そのものを確かめるテストはミドルウェアと同じようにユーザーを載せる必要がある。
func adminFormRequest(t *testing.T, method, target string, form url.Values) *http.Request {
	t.Helper()

	return userFormRequest(t, method, target, form, &model.User{ID: 1, Role: model.RoleAdmin})
}

// userFormRequestはuserからのフォームの送信を組み立てる。
func userFormRequest(t *testing.T, method, target string, form url.Values, user *model.User) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// readStatusは人物の状態と、Rails版への橋渡しのunpublished_atが立っているかを返す。
func readStatus(t *testing.T, tx *sql.Tx, id model.PersonID) (model.PersonStatus, bool) {
	t.Helper()

	var status model.PersonStatus
	var unpublished bool
	if err := tx.QueryRow(`SELECT status, unpublished_at IS NOT NULL FROM people WHERE id = $1`, int64(id)).Scan(&status, &unpublished); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	return status, unpublished
}

// TestNewは掲載停止確認ページが掲載を停止できる人物に対して、人物を名指しし、保存済みの対応記録を
// 初期値にした確認フォームを描画することを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithName("掲載停止確認の人物").WithDelistMessage("以前の対応記録").Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "GET", fmt.Sprintf("/db/people/%d/delisting/new", int64(personID)), nil))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %v、期待値 = text/html; charset=utf-8", ct)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		fmt.Sprintf(`action="/db/people/%d/delisting"`, int64(personID)),
		`method="POST"`,
		"csrf_token",
		`name="updated_at"`,
		`name="return_to" value="/db/people"`,
		"「掲載停止確認の人物」の掲載を停止しますか？",
		">以前の対応記録</textarea>",
		"<title>人物掲載停止 | 掲載停止確認の人物 | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/delisting/new">`, int64(personID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_NotFoundは、存在しない人物・数値でないid・掲載を停止できない状態の人物で404を返す
// ことを検証する。
func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	for _, target := range []string{
		"/db/people/999999999/delisting/new",
		"/db/people/abc/delisting/new",
		fmt.Sprintf("/db/people/%d/delisting/new", int64(delistedID)),
		fmt.Sprintf("/db/people/%d/delisting/new", int64(deletedID)),
	} {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, adminFormRequest(t, "GET", target, nil))

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
		{name: "外部URLは人物一覧", query: "?return_to=https%3A%2F%2Fexample.com%2F", wantHref: "/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, adminFormRequest(t, "GET", fmt.Sprintf("/db/people/%d/delisting/new%s", int64(personID), tt.query), nil))

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

// TestCreateは公開中の人物の掲載を停止して対応記録を書き、確認画面が送信した一覧 (無ければ人物
// 一覧) へリダイレクトすることを検証する。
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
			req := adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
				"updated_at":     {testutil.PersonVersion(t, tx, int64(personID))},
				"return_to":      {tt.returnTo},
				"delist_message": {"本人から申出"},
			})

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", got, tt.wantLocation)
			}
			if status, unpublished := readStatus(t, tx, personID); status != model.PersonStatusDelisted || !unpublished {
				t.Errorf("状態 = %q、unpublished_atあり = %v、期待値 = delisted、true", status, unpublished)
			}

			var message sql.NullString
			if err := tx.QueryRow(`SELECT delist_message FROM people WHERE id = $1`, int64(personID)).Scan(&message); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if message.String != "本人から申出" {
				t.Errorf("delist_message = %+v、期待値 = 本人から申出", message)
			}
		})
	}
}

// TestCreate_NotFoundForDelistedPersonは、すでに掲載停止の人物への送信が404を返すことを検証する。
// 古い確認画面からの送信が対応記録と時刻を打ち直さないようにするため。
func TestCreate_NotFoundForDelistedPerson(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"updated_at": {testutil.PersonVersion(t, tx, int64(personID))},
	}))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
}

// TestCreate_Conflictは確認後に人物名が変わった送信を409で拒否し、現在の人物名と保存済みの
// 対応記録を確認画面に示して、送信した対応記録を入力欄に残すことを検証する。
func TestCreate_Conflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithName("変更前の人物").Build()
	version := testutil.PersonVersion(t, tx, int64(personID))
	if _, err := tx.Exec(`UPDATE people SET name = '変更後の人物', updated_at = updated_at + interval '1 second' WHERE id = $1`, int64(personID)); err != nil {
		t.Fatalf("人物の更新に失敗: %v", err)
	}
	r := newTestRouter(newTestHandler(t, db, tx))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"updated_at": {version}, "delist_message": {"古い名義への対応記録"},
	}))
	if rr.Code != http.StatusConflict {
		t.Fatalf("ステータスコード = %d、期待値 = 409", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "変更後の人物") || !strings.Contains(body, "他の編集者によって") {
		t.Errorf("競合画面に現在の人物と競合の案内がありません: %s", body)
	}
	if !strings.Contains(body, ">古い名義への対応記録</textarea>") {
		t.Errorf("競合画面の入力欄に送信した対応記録が残っていません: %s", body)
	}
	if !strings.Contains(body, "保存済みの対応記録はありません。") {
		t.Errorf("競合画面に保存済みの対応記録 (未記録) が示されていません: %s", body)
	}
	if status, unpublished := readStatus(t, tx, personID); status != model.PersonStatusPublished || unpublished {
		t.Errorf("競合後の状態 = %q、unpublished_atあり = %v", status, unpublished)
	}
}

// TestCreate_MissingVersionは版を持たない掲載停止の送信を422で拒否し、確認画面を開き直すよう
// 案内して、送信した対応記録を入力欄に残すことを検証する。
func TestCreate_MissingVersion(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()
	r := newTestRouter(newTestHandler(t, db, tx))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"updated_at": {""}, "delist_message": {"版の無い送信の対応記録"},
	}))
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = 422", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<p role="alert" class="text-destructive">編集中のデータを特定できませんでした。`) {
		t.Errorf("確認画面に版の欠落の案内がありません: %s", body)
	}
	if !strings.Contains(body, ">版の無い送信の対応記録</textarea>") {
		t.Errorf("確認画面の入力欄に送信した対応記録が残っていません: %s", body)
	}
	if status, unpublished := readStatus(t, tx, personID); status != model.PersonStatusPublished || unpublished {
		t.Errorf("拒否後の状態 = %q、unpublished_atあり = %v", status, unpublished)
	}
}

// TestDelete_MissingVersionは版を持たない解除の送信を422で拒否し、確認画面を開き直すよう案内する
// ことを検証する。
func TestDelete_MissingVersion(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"_method": {"DELETE"}, "updated_at": {""},
	}))
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = 422", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, `<p role="alert" class="text-destructive">編集中のデータを特定できませんでした。`) {
		t.Errorf("確認画面に版の欠落の案内がありません: %s", body)
	}
	if status, _ := readStatus(t, tx, personID); status != model.PersonStatusDelisted {
		t.Errorf("拒否後の状態 = %q、期待値 = delisted", status)
	}
}

// TestCreate_StateChangedAfterConfirmationは、別の管理者が先に掲載停止した人物への古い
// 確認画面の送信を404で拒否し、対応記録を書き換えないことを検証する。
func TestCreate_StateChangedAfterConfirmation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()
	version := testutil.PersonVersion(t, tx, int64(personID))
	if _, err := tx.Exec(`UPDATE people SET status = 'delisted', delist_message = '先の対応', unpublished_at = NOW() AT TIME ZONE 'UTC', updated_at = updated_at + interval '1 second' WHERE id = $1`, int64(personID)); err != nil {
		t.Fatalf("先の掲載停止に失敗: %v", err)
	}
	r := newTestRouter(newTestHandler(t, db, tx))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"updated_at": {version}, "delist_message": {"後から送った対応"},
	}))
	if rr.Code != http.StatusNotFound {
		t.Errorf("ステータスコード = %d、期待値 = 404", rr.Code)
	}
	var message string
	if err := tx.QueryRow(`SELECT delist_message FROM people WHERE id = $1`, int64(personID)).Scan(&message); err != nil {
		t.Fatalf("対応記録の取得に失敗: %v", err)
	}
	if message != "先の対応" {
		t.Errorf("対応記録 = %q、期待値 = 先の対応", message)
	}
}

// TestDelete_Conflictは確認後に対応記録が変わった解除を409で拒否し、現在の記録を示す。
func TestDelete_Conflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).WithDelistMessage("以前の記録").Build()
	version := testutil.PersonVersion(t, tx, int64(personID))
	if _, err := tx.Exec(`UPDATE people SET delist_message = '更新された記録', updated_at = updated_at + interval '1 second' WHERE id = $1`, int64(personID)); err != nil {
		t.Fatalf("対応記録の更新に失敗: %v", err)
	}
	r := newTestRouter(newTestHandler(t, db, tx))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"_method": {"DELETE"}, "updated_at": {version},
	}))
	if rr.Code != http.StatusConflict {
		t.Fatalf("ステータスコード = %d、期待値 = 409", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "更新された記録") || !strings.Contains(body, "他の編集者によって") {
		t.Errorf("競合画面に現在の対応記録と競合の案内がありません: %s", body)
	}
	if status, _ := readStatus(t, tx, personID); status != model.PersonStatusDelisted {
		t.Errorf("競合後の状態 = %q、期待値 = delisted", status)
	}
}

// TestDeleteは掲載停止の人物の掲載停止を解除し、確認画面が送信した一覧へリダイレクトすることを
// 検証する。
func TestDelete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	req := adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"updated_at": {testutil.PersonVersion(t, tx, int64(personID))},
		"_method":    {"DELETE"},
		"return_to":  {"/db/search?q=test"},
	})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if got := rr.Header().Get("Location"); got != "/db/search?q=test" {
		t.Errorf("リダイレクト先 = %q、期待値 = /db/search?q=test", got)
	}
	if status, unpublished := readStatus(t, tx, personID); status != model.PersonStatusPublished || unpublished {
		t.Errorf("状態 = %q、unpublished_atあり = %v、期待値 = published、false", status, unpublished)
	}
}

// TestDelete_NotFoundForPublishedPersonは、公開中の人物への掲載停止の解除の送信が404を返すことを
// 検証する。
func TestDelete_NotFoundForPublishedPerson(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, adminFormRequest(t, "POST", fmt.Sprintf("/db/people/%d/delisting", int64(personID)), url.Values{
		"_method": {"DELETE"}, "updated_at": {testutil.PersonVersion(t, tx, int64(personID))},
	}))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
}

// TestForbiddenWithoutMiddlewareは、ルートミドルウェアを通さずにHandlerを呼んでも、管理者でない
// ユーザー (編集者を含む) を403で拒否し、人物を書き換えないことを検証する。
func TestForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	publishedID := testutil.NewPersonBuilder(t, tx).Build()
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	r := newTestRouter(newTestHandler(t, db, tx))

	tests := []struct {
		method string
		target string
	}{
		{method: "GET", target: fmt.Sprintf("/db/people/%d/delisting/new", int64(publishedID))},
		{method: "POST", target: fmt.Sprintf("/db/people/%d/delisting", int64(publishedID))},
		{method: "DELETE", target: fmt.Sprintf("/db/people/%d/delisting", int64(delistedID))},
	}

	for _, user := range []*model.User{{ID: 1, Role: model.RoleUser}, {ID: 1, Role: model.RoleEditor}} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v %s %s", user.Role, tt.method, tt.target), func(t *testing.T) {
				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, userFormRequest(t, tt.method, tt.target, nil, user))

				if status := rr.Code; status != http.StatusForbidden {
					t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
				}
			})
		}
	}

	if status, _ := readStatus(t, tx, publishedID); status != model.PersonStatusPublished {
		t.Errorf("公開中の人物の状態 = %q、期待値 = published", status)
	}
	if status, _ := readStatus(t, tx, delistedID); status != model.PersonStatusDelisted {
		t.Errorf("掲載停止の人物の状態 = %q、期待値 = delisted", status)
	}
}
