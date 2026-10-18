package db_person

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// updateTestVersionはテストの人物に設定する過去のupdated_atと、送信が運ぶその版。同じ
// トランザクションの中ではNOW() が進まないため、過去の時刻を使う。
var updateTestVersion = time.Date(2026, 10, 1, 9, 8, 7, 654321000, time.UTC)

const updateTestFormVersion = "2026-10-01T09:08:07.654321Z"

// newUpdateRequestは人物編集フォームの送信を、idのURLパラメータと送信者付きで組み立てる。
func newUpdateRequest(personID model.PersonID, user *model.User, form url.Values) *http.Request {
	req := httptest.NewRequest("POST", fmt.Sprintf("/db/people/%d", int64(personID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withPersonIDParam(req, personID.String())

	return withTestUser(req, user)
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
	personID := testutil.NewPersonBuilder(t, tx).WithUpdatedAt(updateTestVersion).Build()
	newName := "更新後の人物-" + t.Name()
	handler := newTestHandler(t, db, tx)

	form := url.Values{}
	form.Set("name", newName)
	form.Set("blood_type", "a")
	form.Set("birthday", "2000-02-29")
	form.Set("updated_at", updateTestFormVersion)

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: userID, Role: model.RoleEditor}, form))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %v、期待値 = %v\n%s", status, http.StatusSeeOther, rr.Body.String())
	}
	// 送信が成功したら同じ人物の編集フォームに着地する (Rails版と同じ)。
	wantLocation := fmt.Sprintf("/db/people/%d/edit", int64(personID))
	if location := rr.Header().Get("Location"); location != wantLocation {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, wantLocation)
	}

	var name, bloodType, birthday string
	if err := tx.QueryRow(`
		SELECT name, blood_type, to_char(birthday, 'YYYY-MM-DD') FROM people WHERE id = $1
	`, int64(personID)).Scan(&name, &bloodType, &birthday); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if name != newName || bloodType != "a" || birthday != "2000-02-29" {
		t.Errorf("人物 = {name: %q, blood_type: %q, birthday: %q}", name, bloodType, birthday)
	}
}

// TestUpdate_ValidationErrorは、不正な送信が422で、送信した値とエラーを伴ってフォームに
// 返ることを検証する。
func TestUpdate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	storedName := "保存済みの人物-" + t.Name()
	personID := testutil.NewPersonBuilder(t, tx).WithName(storedName).WithUpdatedAt(updateTestVersion).Build()
	handler := newTestHandler(t, db, tx)

	form := url.Values{}
	form.Set("name", storedName)
	form.Set("height", "百七十")
	form.Set("url", "example.com")
	form.Set("updated_at", updateTestFormVersion)

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: 1, Role: model.RoleEditor}, form))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	expectedContents := []string{
		// 再描画するのは編集フォームなので、og:urlはPATCH先ではなくそのページ自身の
		// GETパスを指す。
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/people/%d/edit">`, int64(personID)),
		"整数で入力してください",
		"URLの形式が正しくありません",
		`value="百七十"`,
		`value="example.com"`,
		`name="updated_at" value="` + updateTestFormVersion + `"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestUpdate_DelistedNameは、名前を掲載停止の人物と同じ名前に変える送信が422で、掲載停止の
// 文言を伴ってフォームに返り、申出によることを伝えないことを検証する。
func TestUpdate_DelistedName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	personID := testutil.NewPersonBuilder(t, tx).WithUpdatedAt(updateTestVersion).Build()
	delistedName := "掲載停止の人物-" + t.Name()
	testutil.NewPersonBuilder(t, tx).WithName(delistedName).WithStatus(model.PersonStatusDelisted).Build()
	handler := newTestHandler(t, db, tx)

	form := url.Values{}
	form.Set("name", delistedName)
	form.Set("updated_at", updateTestFormVersion)

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: 1, Role: model.RoleEditor}, form))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "この人物は掲載を停止しています。") {
		t.Errorf("掲載停止のエラーが表示されていません: %s", body)
	}
	for _, unexpected := range []string{"この名前の人物は既に登録されています", "申出"} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに%qが含まれています", unexpected)
		}
	}
}

// TestUpdate_Conflictは、フォームを開いた後に他者が書いた人物への送信が409で、送信した
// 値と保存済みの値を並べ、保存済みの版を運ぶフォームに返ることを検証する。
func TestUpdate_Conflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	storedName := "他者が書いた人物-" + t.Name()
	personID := testutil.NewPersonBuilder(t, tx).WithName(storedName).WithUpdatedAt(updateTestVersion).Build()
	handler := newTestHandler(t, db, tx)

	submittedName := "古い版からの人物-" + t.Name()
	form := url.Values{}
	form.Set("name", submittedName)
	form.Set("updated_at", "2026-09-30T00:00:00Z")

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: 1, Role: model.RoleEditor}, form))

	if status := rr.Code; status != http.StatusConflict {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusConflict)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"他の編集者によって更新されたため",
		"現在保存されている内容",
		"<dt>名前</dt>",
		storedName,
		`value="` + submittedName + `"`,
		// 次の送信が、いま見た保存済みの値に対して行われるよう、保存済みの版を運ぶ。
		`name="updated_at" value="` + updateTestFormVersion + `"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestUpdate_NotFoundは、削除済みの人物への送信に404を返すことを検証する。
func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	personID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).WithUpdatedAt(updateTestVersion).Build()
	handler := newTestHandler(t, db, tx)

	form := url.Values{}
	form.Set("name", "削除済みの人物-"+t.Name())
	form.Set("updated_at", updateTestFormVersion)

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: 1, Role: model.RoleEditor}, form))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
}

// TestUpdate_ForbiddenWithoutMiddlewareは、ルートのミドルウェアを通らない一般ユーザーの
// 送信をUseCaseが拒否し、書き込まないことを検証する。
func TestUpdate_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	storedName := "認可される人物-" + t.Name()
	personID := testutil.NewPersonBuilder(t, tx).WithName(storedName).WithUpdatedAt(updateTestVersion).Build()
	handler := newTestHandler(t, db, tx)

	form := url.Values{}
	form.Set("name", "書き換えられた人物-"+t.Name())
	form.Set("updated_at", updateTestFormVersion)

	rr := httptest.NewRecorder()
	handler.Update(rr, newUpdateRequest(personID, &model.User{ID: 1, Role: model.RoleUser}, form))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusForbidden)
	}

	var name string
	if err := tx.QueryRow(`SELECT name FROM people WHERE id = $1`, int64(personID)).Scan(&name); err != nil {
		t.Fatalf("人物の取得に失敗: %v", err)
	}
	if name != storedName {
		t.Errorf("name = %q、期待値 = %q (書き込まない)", name, storedName)
	}
}

// TestUpdate_ConcurrentNameConflictは、事前の重複確認を通った後で別の更新が同じ名前を
// 確定しても、500ではなく名前の入力エラーを返すことを検証する。
func TestUpdate_ConcurrentNameConflict(t *testing.T) {
	t.Parallel()

	db, setupTx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, setupTx).WithRole(model.RoleEditor).Build()
	personID := testutil.NewPersonBuilder(t, setupTx).WithUpdatedAt(updateTestVersion).Build()
	otherID := testutil.NewPersonBuilder(t, setupTx).Build()
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("前提データの確定に失敗: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM db_activities WHERE trackable_type = 'Person' AND trackable_id IN ($1, $2)`, int64(personID), int64(otherID)); err != nil {
			t.Errorf("編集履歴の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM people WHERE id IN ($1, $2)`, int64(personID), int64(otherID)); err != nil {
			t.Errorf("人物の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM profiles WHERE user_id = $1`, int64(userID)); err != nil {
			t.Errorf("プロフィールの後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM settings WHERE user_id = $1`, int64(userID)); err != nil {
			t.Errorf("設定の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM email_notifications WHERE user_id = $1`, int64(userID)); err != nil {
			t.Errorf("メール通知設定の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, int64(userID)); err != nil {
			t.Errorf("ユーザーの後始末に失敗: %v", err)
		}
	})

	name := "競合する人物-" + t.Name()
	blocker, err := db.Begin()
	if err != nil {
		t.Fatalf("競合トランザクションの開始に失敗: %v", err)
	}
	defer func() {
		if rollbackErr := blocker.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			t.Errorf("競合トランザクションの取り消しに失敗: %v", rollbackErr)
		}
	}()
	var blockerPID int
	if err := blocker.QueryRow(`SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("接続IDの取得に失敗: %v", err)
	}
	if _, err := blocker.Exec(`UPDATE people SET name = $1 WHERE id = $2`, name, int64(otherID)); err != nil {
		t.Fatalf("競合する名前の設定に失敗: %v", err)
	}

	handler := newTestHandler(t, db, nil)
	form := url.Values{}
	form.Set("name", name)
	form.Set("updated_at", updateTestFormVersion)
	req := newUpdateRequest(personID, &model.User{ID: userID, Role: model.RoleEditor}, form)
	rr := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		handler.Update(rr, req)
		close(done)
	}()

	// 更新文が一意インデックスのロックで待つまで競合トランザクションを保持する。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		var blocked bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_stat_activity a
				WHERE $1 = ANY(pg_blocking_pids(a.pid))
			)
		`, blockerPID).Scan(&blocked)
		if err != nil {
			t.Fatalf("ロック待ちの確認に失敗: %v", err)
		}
		if blocked {
			break
		}
		select {
		case <-done:
			t.Fatalf("ロック待ちになる前に更新が終了: %d、%s", rr.Code, rr.Body.String())
		case <-ctx.Done():
			t.Fatal("更新が一意制約のロック待ちになりませんでした")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := blocker.Commit(); err != nil {
		t.Fatalf("競合する名前の確定に失敗: %v", err)
	}
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("競合後の更新が終了しませんでした")
	}
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d、本文 = %s", rr.Code, http.StatusUnprocessableEntity, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "この名前の人物は既に登録されています") {
		t.Errorf("名前の重複エラーが表示されていません: %s", rr.Body.String())
	}
}
