package db_person

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// withPersonIDParamはchiがルートパターンから取り出すidのURLパラメータをリクエストに足す。
func withPersonIDParam(req *http.Request, id string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

// newEditRequestは人物編集フォームへのGETリクエストを、idのURLパラメータと、ルートの
// ミドルウェアを通過したcommitterのユーザー付きで組み立てる。
func newEditRequest(personID model.PersonID) *http.Request {
	req := httptest.NewRequest("GET", fmt.Sprintf("/db/people/%d/edit", int64(personID)), nil)
	req = withPersonIDParam(req, personID.String())

	return withTestUser(req, &model.User{ID: 1, Role: model.RoleEditor})
}

func TestEdit(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	prefectureID := testutil.CreateTestPrefecture(t, tx, "埼玉県")
	name := "編集される人物-" + t.Name()
	personID := testutil.NewPersonBuilder(t, tx).
		WithName(name).
		WithNameKana("へんしゅうされるじんぶつ").
		WithUpdatedAt(time.Date(2026, 10, 15, 1, 2, 3, 123456000, time.UTC)).
		Build()
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Edit(rr, newEditRequest(personID))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		fmt.Sprintf("<title>人物編集 | %s | Annict DB</title>", name),
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/edit">`, int64(personID)),
		fmt.Sprintf(`action="/db/people/%d"`, int64(personID)),
		`value="へんしゅうされるじんぶつ"`,
		// 版はtimestamp without time zoneの値を、精度を落とさずUTCの往復書式で運ぶ。
		`name="updated_at" value="2026-10-15T01:02:03.123456Z"`,
		// 出身地の選択肢は登録済みの出身地から作る。
		fmt.Sprintf(`<option value="%d">埼玉県</option>`, int64(prefectureID)),
		`<option value="male">男性</option>`,
		`<option value="ab">AB型</option>`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestEdit_ArchivedPersonは、アーカイブの人物も編集できることを検証する。
func TestEdit_ArchivedPerson(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Edit(rr, newEditRequest(personID))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}
}

// TestEdit_NotFoundは、編集フォームを開けない人物に404を返すことを検証する。
func TestEdit_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()
	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name string
		id   string
	}{
		{name: "掲載停止の人物", id: delistedID.String()},
		{name: "削除済みの人物", id: deletedID.String()},
		{name: "存在しない人物", id: "999999999"},
		{name: "数値でないID", id: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/people/"+tt.id+"/edit", nil)
			req = withTestUser(withPersonIDParam(req, tt.id), &model.User{ID: 1, Role: model.RoleEditor})

			rr := httptest.NewRecorder()
			handler.Edit(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
			}
		})
	}
}

// TestEdit_RequiresCommitterは、編集フォームがcommitterロールで保護されていることを
// 検証する (committerは表示、一般ユーザーは403、未認証はサインインへリダイレクト)。
func TestEdit_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	personID := testutil.NewPersonBuilder(t, tx).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/people/{id}/edit", handler.Edit)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者は表示", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/people/%d/edit", int64(personID)), nil)
			if tt.user != nil {
				req = withTestUser(req, tt.user)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// TestEdit_ForbiddenWithoutMiddlewareは、ルートのミドルウェアを通らない一般ユーザーの
// リクエストをUseCaseが拒否することを検証する。
func TestEdit_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	personID := testutil.NewPersonBuilder(t, tx).Build()
	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", fmt.Sprintf("/db/people/%d/edit", int64(personID)), nil)
	req = withTestUser(withPersonIDParam(req, personID.String()), &model.User{ID: 1, Role: model.RoleUser})

	rr := httptest.NewRecorder()
	handler.Edit(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}
}
