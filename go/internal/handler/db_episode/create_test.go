package db_episode

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// createTargetWorkSeasonYearは、本テスト群がコミットする作品を、作品一覧が全体に対して
// 数える「シーズン未設定」の集合から外すためのもの。行はテストが終わるまで共有テストDBに
// 残り、同時に走る他パッケージからも見えるため。
const createTargetWorkSeasonYear = 1903

// insertCreateTargetWorkは一括作成の送信先となる作品を、テスト用トランザクションでは
// なく共有プールにコミットして挿入する。作成UseCaseは自前のトランザクションを開くため、
// 未コミットの作品は見えないからである。作品は未マッピング (works.anime_idがNULL) のままに
// して送信がepisodesの行だけを書くようにする (animes / anime_classificationsへの両書きは
// UseCaseのテストで検証済み)。テスト終了時の後始末では、作品とその作品への送信が作成した行の
// 削除を試みる。他パッケージがエピソードへの参照をコミット済みの場合は、失敗をログに残し、行は
// 次回のテストDBリセットまで残す。
func insertCreateTargetWork(t *testing.T, db *sql.DB) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(
		`INSERT INTO works (title, media, season_year, season_name, created_at, updated_at) VALUES ($1, 1, $2, 1, NOW(), NOW()) RETURNING id`,
		"一括作成テストアニメ_"+t.Name(), createTargetWorkSeasonYear,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
	}

	t.Cleanup(func() { deleteCreateTargetWork(t, db, id) })

	return model.WorkID(id)
}

// deleteCreateTargetWorkはinsertCreateTargetWorkが挿入した作品と、その作品への送信が
// 作った行の削除を試みる。失敗した文はテストを落とさずログに残す。internal/usecase/seedは
// DB全体からエピソードをランダムに選び、それを参照するepisode_records / activitiesを
// コミットするため、両パッケージが並行して走る間はこの削除が外部キーに阻まれることが正当に
// 起こりうる。ログに残すことで、完了できなかった後始末を、通常の並行実行をflakyな失敗に
// 変えずに観測できる。
func deleteCreateTargetWork(t *testing.T, db *sql.DB, workID int64) {
	t.Helper()

	statements := []string{
		`DELETE FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`,
		`DELETE FROM episodes WHERE work_id = $1`,
		`DELETE FROM works WHERE id = $1`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement, workID); err != nil {
			t.Logf("作品の後始末に失敗 (%s): %v", statement, err)
		}
	}
}

// insertCreateTestUserは送信者となるユーザーを、作成UseCase自身のトランザクションから
// 行を帰属させられるよう共有プールにコミットして挿入する。送信が記録した活動履歴がユーザーを
// 参照するため、ユーザー自身の行より先に消す。
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

func withCreateTestUser(req *http.Request, user *model.User) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// newCreateFormRequestはある作品に行を送信するPOSTリクエストを組み立てる。ルート
// コンテキストを持たないため、実際のchiルーターを通すテストはこちらを使う (URLパラメータは
// ルーター自身が埋める)。
func newCreateFormRequest(workID model.WorkID, rows string) *http.Request {
	form := url.Values{}
	form.Set("rows", rows)

	req := httptest.NewRequest("POST", fmt.Sprintf("/db/works/%d/episodes", int64(workID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// newCreateRequestは同じリクエストを、chiがルートパターンから取り出すwork_idのURL
// パラメータ付きで組み立てる。ハンドラーを直接呼ぶテスト向け。
func newCreateRequest(workID model.WorkID, rows string) *http.Request {
	req := newCreateFormRequest(workID, rows)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("work_id", fmt.Sprintf("%d", int64(workID)))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	user := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Create(rr, withCreateTestUser(newCreateRequest(workID, "#1,1,はじまり\n#2,2,つづき"), user))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}

	// 送信が成功したら、作成された行が並ぶその作品のエピソード一覧に着地する。
	wantLocation := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != wantLocation {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, wantLocation)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("エピソード件数の取得に失敗: %v", err)
	}
	if count != 2 {
		t.Errorf("作成されたエピソード = %d件、期待値 = 2", count)
	}
}

// TestCreate_ValidationErrorは不正な行を含む送信を検証する。送信した行と、失敗した行を
// 名指しするエラーを伴ってフォームが返り、編集者が全行を入力し直さずに手直しできる。
func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	user := &model.User{ID: 1, Role: model.RoleEditor}
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Create(rr, withCreateTestUser(newCreateRequest(workID, "#1,1,はじまり\n#2,いち,つづき"), user))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// 再描画するのは一括作成フォームなので、og:urlはPOST先ではなくそのページ自身の
		// GETパスを指す (POST先はGETではエピソード一覧を返す)。
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/episodes/new">`, int64(workID)),
		fmt.Sprintf(`action="/db/works/%d/episodes"`, int64(workID)),
		`role="alert"`,
		"2 行目",
		// 送信された行はtextareaに書き戻される。
		"#1,1,はじまり",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("エピソード件数の取得に失敗: %v", err)
	}
	if count != 0 {
		t.Errorf("作成されたエピソード = %d件、期待値 = 0 (送信全体が失敗するため)", count)
	}
}

func TestCreate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	user := &model.User{ID: 1, Role: model.RoleEditor}
	handler.Create(rr, withCreateTestUser(newCreateRequest(model.WorkID(999999999), "#1,1,はじまり"), user))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

func TestCreate_ManualCreationRestriction(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertCreateTargetWork(t, db)
	if _, err := db.Exec(`UPDATE works SET manual_episodes_count = 1 WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("予定話数の更新に失敗: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO episodes (work_id, number, sort_number, created_at, updated_at)
		VALUES ($1, '#1', 100, NOW(), NOW())
	`, int64(workID)); err != nil {
		t.Fatalf("既存エピソードの作成に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)
	editor := insertCreateTestUser(t, db, model.RoleEditor)
	admin := insertCreateTestUser(t, db, model.RoleAdmin)

	editorRR := httptest.NewRecorder()
	handler.Create(editorRR, withCreateTestUser(newCreateRequest(workID, "#2,2,つづき"), editor))
	if editorRR.Code != http.StatusUnprocessableEntity {
		t.Fatalf("編集者のステータスコード = %d、期待値 = 422", editorRR.Code)
	}
	if body := editorRR.Body.String(); !strings.Contains(body, "話数分のエピソード") ||
		!strings.Contains(body, "readonly") || !strings.Contains(body, "disabled") {
		t.Errorf("編集者向け制限フォームに警告・無効化属性が揃っていません")
	}

	adminRR := httptest.NewRecorder()
	handler.Create(adminRR, withCreateTestUser(newCreateRequest(workID, "#2,2,つづき"), admin))
	if adminRR.Code != http.StatusSeeOther {
		t.Errorf("管理者のステータスコード = %d、期待値 = 303", adminRR.Code)
	}
}

// TestCreate_RequiresCommitterは一括作成の送信がcommitterロールで保護されていることを
// 検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
func TestCreate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	editor := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Post("/db/works/{work_id}/episodes", handler.Create)

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
			req := newCreateFormRequest(workID, "#1,1,はじまり")
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
