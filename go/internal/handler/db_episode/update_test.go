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
	"github.com/annict/annict/go/internal/validator"
)

// insertUpdateTargetEpisode inserts the episode an update submit edits, committed to the shared
// pool rather than to the test transaction: the update usecase opens its own transaction and
// would not see an episode that is still uncommitted. Its parent work's cleanup removes it.
//
// [Ja] insertUpdateTargetEpisode は更新の送信が編集するエピソードを、テスト用トランザクションでは
// なく共有プールにコミットして挿入する。更新 UseCase は自前のトランザクションを開くため、未コミット
// のエピソードは見えないからである。行の後始末は親作品の後始末が行う。
func insertUpdateTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID) model.EpisodeID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, title, created_at, updated_at)
		VALUES ($1, '#1', 100, '編集前のタイトル', NOW(), NOW()) RETURNING id`,
		int64(workID),
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// readUpdateTargetForm returns the values an episode's edit form would open with: the stored
// title and the version the hidden field carries.
//
// [Ja] readUpdateTargetForm はエピソードの編集フォームが開く値を返す。保存済みのタイトルと、
// hidden が運ぶ版。
func readUpdateTargetForm(t *testing.T, db *sql.DB, episodeID model.EpisodeID) (title string, version string) {
	t.Helper()

	var storedTitle sql.NullString
	var updatedAt sql.NullTime
	if err := db.QueryRow(`SELECT title, updated_at FROM episodes WHERE id = $1`, int64(episodeID)).
		Scan(&storedTitle, &updatedAt); err != nil {
		t.Fatalf("エピソードの読み込みに失敗: %v", err)
	}
	if !updatedAt.Valid {
		return storedTitle.String, validator.FormNullVersion
	}

	return storedTitle.String, updatedAt.Time.UTC().Format(validator.FormVersionLayout)
}

// updateFormValues returns a submit that passes every check, so each test states only the field
// it is about.
//
// [Ja] updateFormValues はすべての検査を通る送信を返す。各テストが、対象のフィールドだけを
// 述べられるようにするため。
func updateFormValues(version string) url.Values {
	form := url.Values{}
	form.Set("number", "第2話")
	form.Set("raw_number", "2.5")
	form.Set("sort_number", "250")
	form.Set("title", "もう、お婿にいけません")
	form.Set("title_en", "No Longer Marriageable")
	form.Set("updated_at", version)

	return form
}

// newUpdateFormRequest builds a PATCH request submitting the edit form for an episode. It
// carries no route context, so it is what the tests that go through a real chi router use; the
// router fills the URL parameter in itself.
//
// [Ja] newUpdateFormRequest はあるエピソードの編集フォームを送信する PATCH リクエストを組み立てる。
// ルートコンテキストを持たないため、実際の chi ルーターを通すテストはこちらを使う (URL パラメータは
// ルーター自身が埋める)。
func newUpdateFormRequest(episodeID model.EpisodeID, form url.Values) *http.Request {
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/db/episodes/%d", int64(episodeID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// newUpdateRequest builds the same request with the id URL parameter chi would have extracted
// from the route pattern, for the tests that call the handler directly.
//
// [Ja] newUpdateRequest は同じリクエストを、chi がルートパターンから取り出す id の URL パラメータ
// 付きで組み立てる。ハンドラーを直接呼ぶテスト向け。
func newUpdateRequest(episodeID model.EpisodeID, form url.Values) *http.Request {
	req := newUpdateFormRequest(episodeID, form)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", fmt.Sprintf("%d", int64(episodeID)))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	user := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	_, version := readUpdateTargetForm(t, db, episodeID)

	rr := httptest.NewRecorder()
	handler.Update(rr, withCreateTestUser(newUpdateRequest(episodeID, updateFormValues(version)), user))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("status code: got %v want %v", status, http.StatusSeeOther)
	}

	// A successful submit lands on the work's episode list, where the edited row sits among
	// the others.
	//
	// [Ja] 送信が成功したら、編集後の行が他の行と並ぶその作品のエピソード一覧に着地する。
	wantLocation := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != wantLocation {
		t.Errorf("Location = %q, want %q", location, wantLocation)
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "もう、お婿にいけません" {
		t.Errorf("title = %q, want %q", title, "もう、お婿にいけません")
	}
}

// TestUpdate_ValidationError covers a submit with a bad field: the form comes back with what was
// typed and the error naming the field, so the editor corrects it instead of retyping the rest.
//
// [Ja] TestUpdate_ValidationError は不正なフィールドを含む送信を検証する。入力した内容と、
// 対象のフィールドを名指しするエラーを伴ってフォームが返り、編集者が残りを入力し直さずに手直し
// できる。
func TestUpdate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	user := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	_, version := readUpdateTargetForm(t, db, episodeID)
	form := updateFormValues(version)
	form.Set("sort_number", "")

	rr := httptest.NewRecorder()
	handler.Update(rr, withCreateTestUser(newUpdateRequest(episodeID, form), user))

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Fatalf("status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	expectedContents := []string{
		// The re-rendered page is the edit form, so og:url names its own GET path and not the
		// PATCH endpoint.
		//
		// [Ja] 再描画するのは編集フォームなので、og:url は PATCH 先ではなくそのページ自身の
		// GET パスを指す。
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/episodes/%d/edit">`, int64(episodeID)),
		fmt.Sprintf(`action="/db/episodes/%d"`, int64(episodeID)),
		`role="alert"`,
		// The submitted values are echoed back into their fields, including the version, so
		// the corrected submit is still made against the version the editor read.
		//
		// [Ja] 送信された値は各欄に書き戻される。版も含まれるため、手直し後の送信も編集者が
		// 読んだ版に対して行われる。
		`value="第2話"`,
		"もう、お婿にいけません",
		fmt.Sprintf(`name="updated_at" value="%s"`, version),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "編集前のタイトル" {
		t.Errorf("title = %q, want %q (却下された送信は行を書かない)", title, "編集前のタイトル")
	}
}

// TestUpdate_Conflict covers two editors submitting from the same form: the second one gets the
// form back with the conflict stated, and the first one's values stay in place.
//
// [Ja] TestUpdate_Conflict は 2 人の編集者が同じフォームから送信する場合を検証する。2 人目には
// 競合を述べたフォームが返り、1 人目の値はそのまま残る。
func TestUpdate_Conflict(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	user := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	_, version := readUpdateTargetForm(t, db, episodeID)

	first := updateFormValues(version)
	first.Set("title", "先に保存したタイトル")
	firstRR := httptest.NewRecorder()
	handler.Update(firstRR, withCreateTestUser(newUpdateRequest(episodeID, first), user))
	if firstRR.Code != http.StatusSeeOther {
		t.Fatalf("1 件目の status code: got %v want %v", firstRR.Code, http.StatusSeeOther)
	}

	second := updateFormValues(version)
	second.Set("title", "後から届いたタイトル")
	secondRR := httptest.NewRecorder()
	handler.Update(secondRR, withCreateTestUser(newUpdateRequest(episodeID, second), user))

	if secondRR.Code != http.StatusConflict {
		t.Fatalf("2 件目の status code: got %v want %v", secondRR.Code, http.StatusConflict)
	}

	body := secondRR.Body.String()
	expectedContents := []string{
		"他の編集者によって更新された",
		// Nothing is merged automatically, so the rejected submit is handed back next to the
		// stored values for the editor to compare and decide between.
		//
		// [Ja] 自動マージは行わないため、却下された送信は保存済みの値と並べて返される。編集者が
		// 両者を見比べて選べるようにするため。
		"後から届いたタイトル",
		"先に保存したタイトル",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	title, storedVersion := readUpdateTargetForm(t, db, episodeID)
	if title != "先に保存したタイトル" {
		t.Errorf("title = %q, want %q (後の送信は上書きしない)", title, "先に保存したタイトル")
	}
	// Having been shown the stored row, the editor can decide to keep their own values: the
	// form carries the version those values belong to, so a submit made from here overwrites
	// exactly what was displayed and still loses to a write that arrives afterwards.
	//
	// [Ja] 保存済みの行を示された編集者は、自分の値を残すと決められる。フォームはその値の版を
	// 運ぶため、ここからの送信は表示された内容だけを上書きし、その後に届く書き込みには依然として
	// 負ける。
	if !strings.Contains(body, fmt.Sprintf(`name="updated_at" value="%s"`, storedVersion)) {
		t.Error("競合時のフォームが保存済みの版を運んでいません")
	}
	if strings.Contains(body, fmt.Sprintf(`name="updated_at" value="%s"`, version)) {
		t.Error("競合時のフォームが古い版を運んでいます")
	}
}

// TestUpdate_Busy covers a submit that could not take a row lock it needed for as long as it
// retried. No one wrote the episode, so this must not come back as the conflict above: the
// editor is told to send the same submit again, and the form keeps the version it was made
// against so that resend still matches.
//
// [Ja] TestUpdate_Busy は、再試行の間ずっと必要な行ロックを取れなかった送信を検証する。誰も
// そのエピソードを書いていないため、上の競合として返してはならない。編集者には同じ送信をもう
// 一度送るよう伝え、フォームは送信が前提とした版を保つことで、その再送信が一致するようにする。
func TestUpdate_Busy(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	user := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	_, version := readUpdateTargetForm(t, db, episodeID)

	// Hold the target row the way any Rails write to that episode does, for longer than the
	// update is willing to retry.
	//
	// [Ja] そのエピソードへの Rails の書き込みと同じ形で、更新が再試行する時間より長く対象行を
	// 保持する。
	holdTx, err := db.Begin()
	if err != nil {
		t.Fatalf("ロック保持トランザクションの Begin() に失敗: %v", err)
	}
	defer func() { _ = holdTx.Rollback() }()
	if _, err := holdTx.Exec("UPDATE episodes SET title = title WHERE id = $1", int64(episodeID)); err != nil {
		t.Fatalf("対象行のロック取得に失敗: %v", err)
	}

	form := updateFormValues(version)
	rr := httptest.NewRecorder()
	handler.Update(rr, withCreateTestUser(newUpdateRequest(episodeID, form), user))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code: got %v want %v", rr.Code, http.StatusServiceUnavailable)
	}
	if got := rr.Header().Get("Retry-After"); got != "1" {
		t.Errorf("Retry-After = %q, want %q", got, "1")
	}

	body := rr.Body.String()
	expectedContents := []string{
		"そのままもう一度送信してください",
		// The submitted values come back so the resend costs no retyping, and the version they
		// were made against comes with them because nothing has moved it.
		//
		// [Ja] 送信された値は書き戻され、再送信で打ち直しが要らないようにする。版も何も動かして
		// いないため、送信が前提としたものがそのまま返る。
		"もう、お婿にいけません",
		fmt.Sprintf(`name="updated_at" value="%s"`, version),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}
	if strings.Contains(body, "他の編集者によって更新された") {
		t.Error("ロックを取れなかった送信を、他者による更新として説明しています")
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "編集前のタイトル" {
		t.Errorf("title = %q, want %q (適用されなかった送信は行を書かない)", title, "編集前のタイトル")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)
	user := &model.User{ID: 1, Role: model.RoleEditor}

	rr := httptest.NewRecorder()
	form := updateFormValues(validator.FormNullVersion)
	handler.Update(rr, withCreateTestUser(newUpdateRequest(model.EpisodeID(999999999), form), user))

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestUpdate_RequiresCommitter verifies the update submit is protected by the committer role
// (committer proceeds, a regular user 403, an unauthenticated request is redirected to sign-in).
//
// [Ja] TestUpdate_RequiresCommitter は更新の送信が committer ロールで保護されていることを検証
// する (committer は処理続行、一般ユーザーは 403、未認証はサインインへリダイレクト)。
func TestUpdate_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	editor := insertCreateTestUser(t, db, model.RoleEditor)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Patch("/db/episodes/{id}", handler.Update)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者はアクセス許可 (更新成功でリダイレクト)", user: editor, wantStatus: http.StatusSeeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, version := readUpdateTargetForm(t, db, episodeID)
			req := newUpdateFormRequest(episodeID, updateFormValues(version))
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got %v want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
