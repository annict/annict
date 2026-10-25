package db_person

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// newDeleteRequestは人物の削除の送信を、idのURLパラメータと送信者付きで組み立てる。削除
// 確認画面のフォームはPOSTをメソッドオーバーライドでDELETEとして運ぶため、本文のreturn_toが
// 読まれるよう、ミドルウェアを通す前のPOSTとして組み立てる (net/httpはDELETEの本文を
// フォームとして解析しない)。
func newDeleteRequest(id string, user *model.User, form url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/db/people/"+id, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withPersonIDParam(req, id)

	return withTestUser(req, user)
}

// isDeletedは人物のstatusがdeletedで、Rails版への橋渡しのdeleted_atが立っているかを返す。
func isDeleted(t *testing.T, tx *sql.Tx, id model.PersonID) bool {
	t.Helper()
	var deleted bool
	if err := tx.QueryRow(`SELECT status = 'deleted' AND deleted_at IS NOT NULL FROM people WHERE id = $1`, int64(id)).Scan(&deleted); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	return deleted
}

// TestDeleteは、管理者が公開中・アーカイブの人物を論理削除し、確認画面が送信した一覧 (無ければ
// 人物一覧) へリダイレクトすることを検証する。
func TestDelete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)
	admin := &model.User{ID: 1, Role: model.RoleAdmin}

	tests := []struct {
		name         string
		archived     bool
		returnTo     string
		wantLocation string
	}{
		{name: "公開中の人物を削除して人物一覧に戻る", wantLocation: "/db/people"},
		{name: "アーカイブの人物を削除して検索結果に戻る", archived: true, returnTo: "/db/search?q=test", wantLocation: "/db/search?q=test"},
		{name: "外部URLは人物一覧", returnTo: "https://example.com/", wantLocation: "/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testutil.NewPersonBuilder(t, tx)
			if tt.archived {
				builder = builder.WithStatus(model.PersonStatusArchived)
			}
			personID := builder.Build()

			rr := httptest.NewRecorder()
			handler.Delete(rr, newDeleteRequest(personID.String(), admin, url.Values{"return_to": {tt.returnTo}}))

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", got, tt.wantLocation)
			}
			if !isDeleted(t, tx, personID) {
				t.Error("人物が削除されていません")
			}
		})
	}
}

// TestDelete_HTMXRedirectは、htmxの送信 (人物一覧の削除ボタン) が素の303ではなく、204と
// 人物一覧へのHX-Redirectを返すことを検証する。htmxが押したボタンに一覧をスワップせず
// 遷移するようにするため。
func TestDelete_HTMXRedirect(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)
	personID := testutil.NewPersonBuilder(t, tx).Build()

	req := newDeleteRequest(personID.String(), &model.User{ID: 1, Role: model.RoleAdmin}, url.Values{})
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	handler.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNoContent)
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/db/people" {
		t.Errorf("HX-Redirect = %q、期待値 = /db/people", got)
	}
	if loc := rr.Header().Get("Location"); loc != "" {
		t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移するため)", loc)
	}
}

// TestDelete_NotFoundは、存在しない人物・数値でないid・掲載停止・削除済みの人物への送信が
// 404を返すことを検証する。古い送信がdeleted_atを打ち直さないようにするため。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()

	for _, id := range []string{"999999999", "abc", delistedID.String(), deletedID.String()} {
		t.Run(id, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Delete(rr, newDeleteRequest(id, &model.User{ID: 1, Role: model.RoleAdmin}, url.Values{}))

			if rr.Code != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

// TestDelete_Forbiddenは、ルートのミドルウェアを通さずにHandlerを呼んでも、管理者でない
// ユーザー (編集者を含む) を403で拒否し、人物を削除しないことを検証する。
func TestDelete_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)
	personID := testutil.NewPersonBuilder(t, tx).Build()

	for _, role := range []int32{model.RoleUser, model.RoleEditor} {
		t.Run(fmt.Sprint(role), func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Delete(rr, newDeleteRequest(personID.String(), &model.User{ID: 1, Role: role}, url.Values{}))

			if rr.Code != http.StatusForbidden {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusForbidden)
			}
		})
	}

	if isDeleted(t, tx, personID) {
		t.Error("拒否した送信が人物を削除しました")
	}
}
