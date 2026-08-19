package db_work

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// insertUpdateTargetWork inserts a work the update endpoint can write. The row is committed
// rather than held in the test transaction because the update usecase opens a transaction of
// its own, which cannot see what an outer transaction has not committed: a version matched
// against a row it cannot see finds nothing and comes back as a conflict.
//
// [Ja] insertUpdateTargetWork は更新エンドポイントが書き込める作品を挿入する。行をテスト
// トランザクションに留めずコミットするのは、更新 UseCase が自前のトランザクションを開き、外側の
// トランザクションが未コミットのものを見られないため。見えない行に対する版の照合は何にも一致せず、
// 競合として返る。
func insertUpdateTargetWork(t *testing.T, db *sql.DB, title string) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(
		`INSERT INTO works (title, media, season_year, season_name, created_at, updated_at)
		 VALUES ($1, 1, 2024, 1, NOW(), NOW()) RETURNING id`,
		title,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM works WHERE id = $1`, id); err != nil {
			t.Logf("works の後始末に失敗: %v", err)
		}
	})

	return model.WorkID(id)
}

// readWorkTitle returns the stored title of a work, which tells whether a submit was applied.
//
// [Ja] readWorkTitle は作品の保存済みタイトルを返す。送信が適用されたかどうかを表す。
func readWorkTitle(t *testing.T, db *sql.DB, workID model.WorkID) string {
	t.Helper()

	var title string
	if err := db.QueryRow(`SELECT title FROM works WHERE id = $1`, int64(workID)).Scan(&title); err != nil {
		t.Fatalf("作品タイトルの読み込みに失敗: %v", err)
	}

	return title
}

// readWorkVersion returns the version a work's edit form would carry, which a submit has to
// state to be accepted.
//
// [Ja] readWorkVersion は作品の編集フォームが運ぶ版を返す。送信が受理されるには、この版を名乗る
// 必要がある。
func readWorkVersion(t *testing.T, db *sql.DB, workID model.WorkID) string {
	t.Helper()

	var updatedAt sql.NullTime
	if err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&updatedAt); err != nil {
		t.Fatalf("作品の版の読み込みに失敗: %v", err)
	}
	if !updatedAt.Valid {
		return validator.FormNullVersion
	}

	return updatedAt.Time.UTC().Format(validator.FormVersionLayout)
}

// validUpdateForm returns a form that passes validation for the work update endpoint, stating
// the version the submit is made against.
//
// [Ja] validUpdateForm は作品更新エンドポイントの検証を通過するフォームを、送信が前提とする版
// とともに返す。
func validUpdateForm(title, version string) url.Values {
	form := url.Values{}
	form.Set("title", title)
	form.Set("media", "1") // tv
	form.Set("updated_at", version)
	return form
}

func patchRequest(t *testing.T, workID int64, form url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/db/works/%d", workID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// TestUpdate_ValidationError verifies that invalid input re-renders the edit form with
// a 422 and the submitted values preserved. Validation runs before the work is loaded,
// so the endpoint responds 422 without needing the work to exist.
//
// [Ja] TestUpdate_ValidationError は不正入力で編集フォームが 422 で再描画され、送信値が
// 保持されることを検証する。検証は work 読み込みより前に走るため、work の存在なしに 422 を
// 返す。
func TestUpdate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	form := url.Values{}
	form.Set("title", "") // required
	form.Set("media", "")
	form.Set("title_kana", "ほぞんされるかな")
	form.Set("updated_at", validator.FormNullVersion)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, 123, form))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<title>作品を編集 | Annict DB</title>",
		"<form",
		`action="/db/works/123"`,
		`name="_method"`,
		`value="PATCH"`,
		`role="alert"`,
		`value="ほぞんされるかな"`, // 入力値が保持される
		// The re-rendered page is the edit form, so og:url names its own GET path and not the
		// PATCH endpoint (which has no GET route).
		//
		// [Ja] 再描画するのは編集フォームなので、og:url は PATCH 先ではなくそのページ自身の
		// GET パスを指す (PATCH 先に GET のルートは無い)。
		`<meta property="og:url" content="https://test.annict.com/db/works/123/edit">`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestUpdate_Success verifies a valid update redirects back to the edit page.
//
// [Ja] TestUpdate_Success は有効な更新が編集ページへリダイレクトすることを検証する。
func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertUpdateTargetWork(t, db, "更新前作品_"+t.Name())
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	newTitle := "更新後作品_" + t.Name()
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, int64(workID), validUpdateForm(newTitle, readWorkVersion(t, db, workID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != fmt.Sprintf("/db/works/%d/edit", int64(workID)) {
		t.Errorf("handler returned wrong redirect location: got %v", location)
	}
	if title := readWorkTitle(t, db, workID); title != newTitle {
		t.Errorf("title = %q, want %q", title, newTitle)
	}
}

// TestUpdate_VersionConflict verifies a submit made against a stale version is refused with a
// 409 that shows the stored values and hands the form the stored version, so the editor can
// compare the two and, having seen what they would overwrite, submit again.
//
// [Ja] TestUpdate_VersionConflict は、古い版に対する送信が 409 で却下され、保存済みの値を表示し、
// フォームに保存済みの版を渡すことを検証する。編集者が両者を見比べ、何を上書きするのかを見たうえで
// 送信し直せるようにするため。
func TestUpdate_VersionConflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertUpdateTargetWork(t, db, "衝突前作品_"+t.Name())
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	// A version an hour behind the stored one is what an editor's form carries once someone
	// else has written the work since it was opened.
	//
	// [Ja] 保存済みより 1 時間古い版は、フォームを開いた後に他者がその作品を書いたときに編集者の
	// フォームが運ぶものである。
	staleVersion := time.Now().Add(-time.Hour).UTC().Format(validator.FormVersionLayout)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, int64(workID), validUpdateForm("衝突後タイトル", staleVersion)))

	if status := rr.Code; status != http.StatusConflict {
		t.Fatalf("status = %d, want %d", status, http.StatusConflict)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"現在保存されている内容",
		"このデータは他の編集者によって更新されたため",
		// The stored title is listed as a value the next submit would overwrite, while the
		// input keeps what was typed.
		//
		// [Ja] 保存済みのタイトルは、次の送信が上書きする値として並ぶ。入力欄は入力された内容を
		// 保つ。
		"<dt>タイトル</dt>",
		"衝突前作品_" + t.Name(),
		`value="衝突後タイトル"`,
		// The form now carries the stored version, so submitting again applies to the row the
		// page has just shown.
		//
		// [Ja] フォームは保存済みの版を運ぶようになるため、送信し直すとページがいま示した行に
		// 適用される。
		fmt.Sprintf(`name="updated_at" value="%s"`, readWorkVersion(t, db, workID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("409 レスポンスに %q が含まれていません", expected)
		}
	}

	if title := readWorkTitle(t, db, workID); title != "衝突前作品_"+t.Name() {
		t.Errorf("title = %q, want 衝突前作品_%s (古い版の送信は上書きしない)", title, t.Name())
	}
}

// TestUpdate_VersionMissing verifies a submit that states no version is refused instead of
// being applied to whatever the row holds now.
//
// [Ja] TestUpdate_VersionMissing は、版を示さない送信が、その時点の行の内容に適用されずに却下
// されることを検証する。
func TestUpdate_VersionMissing(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertUpdateTargetWork(t, db, "版なし送信作品_"+t.Name())
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	form := validUpdateForm("版なし更新後", "")
	form.Del("updated_at")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, int64(workID), form))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if body := rr.Body.String(); !strings.Contains(body, "編集中のデータを特定できませんでした") {
		t.Error("422 レスポンスに版の欠落を述べるメッセージが含まれていません")
	}

	if title := readWorkTitle(t, db, workID); title != "版なし送信作品_"+t.Name() {
		t.Errorf("title = %q, want 版なし送信作品_%s (版を示さない送信は適用されない)", title, t.Name())
	}
}

// TestUpdate_NotFound verifies a valid update for a nonexistent work returns 404.
//
// [Ja] TestUpdate_NotFound は存在しない作品への有効な更新が 404 を返すことを検証する。
func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, patchRequest(t, 999999999, validUpdateForm("存在しない作品の更新", validator.FormNullVersion)))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestUpdate_InvalidID verifies a non-numeric ID returns 404.
//
// [Ja] TestUpdate_InvalidID は数値でないIDで404を返すことを検証する。
func TestUpdate_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Patch("/db/works/{id}", handler.Update)

	req := httptest.NewRequest("PATCH", "/db/works/not-a-number", strings.NewReader(validUpdateForm("x", validator.FormNullVersion).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestUpdate_RequiresCommitter verifies the update route is protected by the committer
// role (committer proceeds, a regular user 403, an unauthenticated request is
// redirected to sign-in).
//
// [Ja] TestUpdate_RequiresCommitter は更新ルートが committer ロールで保護されている
// ことを検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへ
// リダイレクト)。
func TestUpdate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertUpdateTargetWork(t, db, "認可テスト作品_"+t.Name())
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Patch("/db/works/{id}", handler.Update)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{
			name:       "未認証はサインインへリダイレクト",
			user:       nil,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "一般ユーザーは403",
			user:       &model.User{ID: 1, Role: model.RoleUser},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "編集者はアクセス許可 (更新成功でリダイレクト)",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := patchRequest(t, int64(workID), validUpdateForm("更新後作品_"+t.Name(), readWorkVersion(t, db, workID)))
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
