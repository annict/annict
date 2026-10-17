package db_person

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// insertCreateTestUserは送信者となるユーザーを、一括登録UseCase自身のトランザクション
// から編集履歴を帰属させられるよう共有プールにコミットして挿入する。
func insertCreateTestUser(t *testing.T, db *sql.DB, role int32) *model.User {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("ユーザー作成トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).WithRole(role).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("ユーザー作成トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM db_activities WHERE user_id = $1`, int64(userID)); err != nil {
			t.Errorf("DB活動履歴の後始末に失敗: %v", err)
		}
		testutil.DeleteUser(t, db, userID)
	})

	return &model.User{ID: userID, Role: role}
}

// cleanupCommittedPeopleは、送信がコミットした人物をテスト終了時に削除する。編集履歴は
// 送信者のユーザーの後始末で消えるが、t.Cleanupは登録の逆順に走るため、人物を先に消せる
// よう編集履歴もここで消す。
func cleanupCommittedPeople(t *testing.T, db *sql.DB, names ...string) {
	t.Helper()

	t.Cleanup(func() {
		if _, err := db.Exec(`
			DELETE FROM db_activities
			WHERE trackable_type = 'Person' AND trackable_id IN (SELECT id FROM people WHERE name = ANY($1))
		`, pq.Array(names)); err != nil {
			t.Errorf("人物の編集履歴の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM people WHERE name = ANY($1)`, pq.Array(names)); err != nil {
			t.Errorf("人物の後始末に失敗: %v", err)
		}
	})
}

func countPeopleByNames(t *testing.T, db *sql.DB, names ...string) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM people WHERE name = ANY($1)`, pq.Array(names)).Scan(&count); err != nil {
		t.Fatalf("人物件数の取得に失敗: %v", err)
	}
	return count
}

func newCreateRequest(rows string) *http.Request {
	form := url.Values{}
	form.Set("rows", rows)

	req := httptest.NewRequest("POST", "/db/people", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	user := insertCreateTestUser(t, db, model.RoleEditor)
	first := "一括登録の人物1-" + t.Name()
	second := "一括登録の人物2-" + t.Name()
	cleanupCommittedPeople(t, db, first, second)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Create(rr, withTestUser(newCreateRequest(first+",いっかつ\n"+second), user))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}
	// 送信が成功したら、作成した人物が先頭に並ぶ人物一覧に着地する。
	if location := rr.Header().Get("Location"); location != "/db/people" {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, "/db/people")
	}

	if count := countPeopleByNames(t, db, first, second); count != 2 {
		t.Errorf("作成された人物 = %d件、期待値 = 2", count)
	}
}

// TestCreate_ValidationErrorは不正な行を含む送信を検証する。送信した行と、失敗した行を
// 名指しするエラーを伴ってフォームが返り、1件も登録されない。
func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	user := &model.User{ID: 1, Role: model.RoleEditor}
	valid := "正しい行の人物-" + t.Name()
	cleanupCommittedPeople(t, db, valid)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Create(rr, withTestUser(newCreateRequest(valid+",ただしい\n,よみだけ"), user))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// 再描画するのは一括登録フォームなので、og:urlはPOST先ではなくそのページ自身の
		// GETパスを指す。
		`<meta property="og:url" content="https://test.annict.com/db/people/new">`,
		`action="/db/people"`,
		`role="alert"`,
		"2 行目: 名前を入力してください",
		// 送信された行はtextareaに書き戻される。
		valid + ",ただしい",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	if count := countPeopleByNames(t, db, valid); count != 0 {
		t.Errorf("作成された人物 = %d件、期待値 = 0 (送信全体が失敗するため)", count)
	}
}

// TestCreate_RequiresCommitterは一括登録の送信がcommitterロールで保護されていることを
// 検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
func TestCreate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	editor := insertCreateTestUser(t, db, model.RoleEditor)
	name := "認可テストの人物-" + t.Name()
	cleanupCommittedPeople(t, db, name)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Post("/db/people", handler.Create)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者はアクセス許可 (作成成功でリダイレクト)", user: editor, wantStatus: http.StatusSeeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newCreateRequest(name)
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

	if count := countPeopleByNames(t, db, name); count != 1 {
		t.Errorf("作成された人物 = %d件、期待値 = 1 (編集者の送信だけが通るため)", count)
	}
}

// TestCreate_ForbiddenWithoutMiddlewareは、ルートのミドルウェアを通らない一般ユーザーの
// 送信をUseCaseが拒否し、1件も登録しないことを検証する。
func TestCreate_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	name := "認可境界テストの人物-" + t.Name()
	cleanupCommittedPeople(t, db, name)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Create(rr, withTestUser(newCreateRequest(name), &model.User{ID: 1, Role: model.RoleUser}))

	if rr.Code != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusForbidden)
	}
	if count := countPeopleByNames(t, db, name); count != 0 {
		t.Errorf("作成された人物 = %d件、期待値 = 0", count)
	}
}
