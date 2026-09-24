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

// insertUpdateTargetEpisodeは更新の送信が編集するエピソードを、テスト用トランザクションでは
// なく共有プールにコミットして挿入する。更新UseCaseは自前のトランザクションを開くため、未コミット
// のエピソードは見えないからである。行の後始末は親作品の後始末が行う。
func insertUpdateTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID) model.EpisodeID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, title, created_at, updated_at)
		VALUES ($1, '#1', 100, '編集前のタイトル', NOW(), NOW()) RETURNING id`,
		int64(workID),
	).Scan(&id); err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// readUpdateTargetFormはエピソードの編集フォームが開く値を返す。保存済みのタイトルと、
// hiddenが運ぶ版。
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

// updateFormValuesはすべての検査を通る送信を返す。各テストが、対象のフィールドだけを
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

// newUpdateFormRequestはあるエピソードの編集フォームを送信するPATCHリクエストを組み立てる。
// ルートコンテキストを持たないため、実際のchiルーターを通すテストはこちらを使う (URLパラメータは
// ルーター自身が埋める)。
func newUpdateFormRequest(episodeID model.EpisodeID, form url.Values) *http.Request {
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/db/episodes/%d", int64(episodeID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// newUpdateRequestは同じリクエストを、chiがルートパターンから取り出すidのURLパラメータ
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusSeeOther)
	}

	// 送信が成功したら、編集後の行が他の行と並ぶその作品のエピソード一覧に着地する。
	wantLocation := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != wantLocation {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, wantLocation)
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "もう、お婿にいけません" {
		t.Errorf("title = %q、期待値 = %q", title, "もう、お婿にいけません")
	}
}

// TestUpdate_ValidationErrorは不正なフィールドを含む送信を検証する。入力した内容と、
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
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusUnprocessableEntity)
	}

	body := rr.Body.String()
	expectedContents := []string{
		// 再描画するのは編集フォームなので、og:urlはPATCH先ではなくそのページ自身の
		// GETパスを指す。
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/episodes/%d/edit">`, int64(episodeID)),
		fmt.Sprintf(`action="/db/episodes/%d"`, int64(episodeID)),
		`role="alert"`,
		// 送信された値は各欄に書き戻される。版も含まれるため、手直し後の送信も編集者が
		// 読んだ版に対して行われる。
		`value="第2話"`,
		"もう、お婿にいけません",
		fmt.Sprintf(`name="updated_at" value="%s"`, version),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "編集前のタイトル" {
		t.Errorf("title = %q、期待値 = %q (却下された送信は行を書かない)", title, "編集前のタイトル")
	}
}

// TestUpdate_Conflictは2人の編集者が同じフォームから送信する場合を検証する。2人目には
// 競合を述べたフォームが返り、1人目の値はそのまま残る。
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
		t.Fatalf("1件目のステータスコード = %v、期待値 = %v", firstRR.Code, http.StatusSeeOther)
	}

	second := updateFormValues(version)
	second.Set("title", "後から届いたタイトル")
	secondRR := httptest.NewRecorder()
	handler.Update(secondRR, withCreateTestUser(newUpdateRequest(episodeID, second), user))

	if secondRR.Code != http.StatusConflict {
		t.Fatalf("2件目のステータスコード = %v、期待値 = %v", secondRR.Code, http.StatusConflict)
	}

	body := secondRR.Body.String()
	expectedContents := []string{
		"他の編集者によって更新された",
		// 自動マージは行わないため、却下された送信は保存済みの値と並べて返される。編集者が
		// 両者を見比べて選べるようにするため。
		"後から届いたタイトル",
		"先に保存したタイトル",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	title, storedVersion := readUpdateTargetForm(t, db, episodeID)
	if title != "先に保存したタイトル" {
		t.Errorf("title = %q、期待値 = %q (後の送信は上書きしない)", title, "先に保存したタイトル")
	}
	// 保存済みの行を示された編集者は、自分の値を残すと決められる。フォームはその値の版を
	// 運ぶため、ここからの送信は表示された内容だけを上書きし、その後に届く書き込みには依然として
	// 負ける。
	if !strings.Contains(body, fmt.Sprintf(`name="updated_at" value="%s"`, storedVersion)) {
		t.Error("競合時のフォームが保存済みの版を運んでいません")
	}
	if strings.Contains(body, fmt.Sprintf(`name="updated_at" value="%s"`, version)) {
		t.Error("競合時のフォームが古い版を運んでいます")
	}
}

// TestUpdate_Busyは、再試行の間ずっと必要な行ロックを取れなかった送信を検証する。誰も
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

	// そのエピソードへのRailsの書き込みと同じ形で、更新が再試行する時間より長く対象行を
	// 保持する。
	holdTx, err := db.Begin()
	if err != nil {
		t.Fatalf("ロック保持トランザクションのBegin()に失敗: %v", err)
	}
	defer func() { _ = holdTx.Rollback() }()
	if _, err := holdTx.Exec("UPDATE episodes SET title = title WHERE id = $1", int64(episodeID)); err != nil {
		t.Fatalf("対象行のロック取得に失敗: %v", err)
	}

	form := updateFormValues(version)
	rr := httptest.NewRecorder()
	handler.Update(rr, withCreateTestUser(newUpdateRequest(episodeID, form), user))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusServiceUnavailable)
	}
	if got := rr.Header().Get("Retry-After"); got != "1" {
		t.Errorf("Retry-After = %q、期待値 = %q", got, "1")
	}

	body := rr.Body.String()
	expectedContents := []string{
		"そのままもう一度送信してください",
		// 送信された値は書き戻され、再送信で打ち直しが要らないようにする。版も何も動かして
		// いないため、送信が前提としたものがそのまま返る。
		"もう、お婿にいけません",
		fmt.Sprintf(`name="updated_at" value="%s"`, version),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
	if strings.Contains(body, "他の編集者によって更新された") {
		t.Error("ロックを取れなかった送信を、他者による更新として説明しています")
	}

	title, _ := readUpdateTargetForm(t, db, episodeID)
	if title != "編集前のタイトル" {
		t.Errorf("title = %q、期待値 = %q (適用されなかった送信は行を書かない)", title, "編集前のタイトル")
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestUpdate_RequiresCommitterは更新の送信がcommitterロールで保護されていることを検証
// する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
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
				t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
