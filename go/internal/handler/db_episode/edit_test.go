package db_episode

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// newEditRequest builds a GET request for an episode's edit form with the id URL parameter
// chi would have extracted from the route pattern.
//
// [Ja] newEditRequest はエピソード編集フォームへの GET リクエストを、chi がルートパターンから
// 取り出す id の URL パラメータ付きで組み立てる。
func newEditRequest(episodeID model.EpisodeID) *http.Request {
	req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/edit", int64(episodeID)), nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", fmt.Sprintf("%d", int64(episodeID)))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestEdit(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第2話").
		WithTitle("もう、お婿にいけません").
		Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Edit(rr, newEditRequest(episodeID))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<title>エピソード編集 | 第2話 | テストアニメ | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/episodes/%d/edit">`, int64(episodeID)),
		// The heading names the parent work, and the shared subnav links back to its form.
		//
		// [Ja] 見出しは親作品を名指しし、共有のサブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		// The form updates the episode itself, so it targets the episode path with the
		// method override an HTML form needs to send a PATCH.
		//
		// [Ja] フォームはエピソード自身を更新するため、HTML フォームが PATCH を送るための
		// メソッドオーバーライドを添えてエピソードのパスを宛先にする。
		fmt.Sprintf(`action="/db/episodes/%d"`, int64(episodeID)),
		`name="_method" value="PATCH"`,
		`name="csrf_token"`,
		// The stored values open in their fields.
		//
		// [Ja] 保存済みの値が各欄に開く。
		`value="第2話"`,
		"もう、お婿にいけません",
		"エピソード編集",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	// The version the form was opened against comes from the stored row, so the hidden field
	// has to carry a value rather than be rendered empty.
	//
	// [Ja] フォームを開いた時点の版は保存済みの行から来るため、hidden は空ではなく値を運ぶ
	// 必要がある。
	if strings.Contains(body, `name="updated_at" value=""`) {
		t.Error("版の hidden が空で描画されています")
	}
	if !strings.Contains(body, `name="updated_at" value="`) {
		t.Error("版の hidden が描画されていません")
	}
}

// TestEdit_DocumentTitleIsLocalizedAndUniquePerEpisode covers the two episodes of one work whose
// edit pages an editor is most likely to keep open side by side. The page name opens the title and
// the episode follows it, so a tab too narrow for the whole title still shows which page it is and
// which episode it belongs to.
//
// [Ja] TestEdit_DocumentTitleIsLocalizedAndUniquePerEpisode は、編集者が並べて開きやすい同じ作品の
// 2 エピソードの編集ページを検証する。タイトルは画面名で始まりエピソードが続くため、タイトル全体が
// 収まらない幅のタブでも、どの画面でどのエピソードのものかが読める。
func TestEdit_DocumentTitleIsLocalizedAndUniquePerEpisode(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("Test Anime").Build()
	episodeIDs := []model.EpisodeID{
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("#1").Build(),
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("#2").Build(),
	}
	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name     string
		locale   string
		pageName string
	}{
		{name: "日本語", locale: "ja", pageName: "エピソード編集"},
		{name: "英語", locale: "en", pageName: "Edit Episode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			titles := make([]string, len(episodeIDs))
			for i, episodeID := range episodeIDs {
				req := newEditRequest(episodeID)
				req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
				rr := httptest.NewRecorder()
				handler.Edit(rr, req)

				if status := rr.Code; status != http.StatusOK {
					t.Fatalf("status code: got %v want %v", status, http.StatusOK)
				}

				body := rr.Body.String()
				start := strings.Index(body, "<title>")
				end := strings.Index(body, "</title>")
				if start < 0 || end < start {
					t.Fatalf("文書タイトルが描画されていません: %s", body)
				}
				titles[i] = body[start : end+len("</title>")]

				want := fmt.Sprintf(
					"<title>%s | #%d | Test Anime | Annict DB</title>",
					tt.pageName,
					i+1,
				)
				if titles[i] != want {
					t.Errorf("title = %q, want %q", titles[i], want)
				}
			}

			if titles[0] == titles[1] {
				t.Errorf("同じ作品のエピソード間で title が重複しています: %q", titles[0])
			}
		})
	}
}

// TestEdit_DocumentTitleOmitsWorkWithoutName covers an episode whose work has no name to show.
// The title drops the work rather than rendering an empty segment, and keeps the episode
// identifier so the page stays distinguishable in tabs, history, and assistive technology.
//
// [Ja] TestEdit_DocumentTitleOmitsWorkWithoutName は、表示できる名前が無い作品のエピソードを
// 検証する。タイトルは空の区切りを描画せず作品を省き、エピソード識別子は残す。タブ・履歴・
// 支援技術でページを区別できる状態を保つため。
func TestEdit_DocumentTitleOmitsWorkWithoutName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// A title of nothing but whitespace collapses to the empty display name, which is what
	// makes the page take the fallback branch.
	//
	// [Ja] 空白だけのタイトルは表示名として空文字列に畳まれ、ページはフォールバックの分岐を
	// 通る。
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("   ").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Edit(rr, newEditRequest(episodeID))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("status code: got %v want %v", status, http.StatusOK)
	}

	want := "<title>エピソード編集 | 第2話 | Annict DB</title>"
	if !strings.Contains(rr.Body.String(), want) {
		t.Errorf("レスポンスに %q が含まれていません", want)
	}
}

func TestEdit_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	deletedEpisodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithDeletedAt(time.Now()).Build()

	deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済みアニメ").WithDeletedAt(time.Now()).Build()
	episodeOfDeletedWorkID := testutil.NewEpisodeBuilder(t, tx, deletedWorkID).Build()

	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name      string
		episodeID model.EpisodeID
	}{
		{name: "存在しないエピソード", episodeID: model.EpisodeID(999999999)},
		{name: "削除済みのエピソード", episodeID: deletedEpisodeID},
		// An episode whose work is gone has no page to be edited on: the heading and subnav
		// of that page describe the work.
		//
		// [Ja] 作品が失われたエピソードには編集するページが無い。そのページの見出しと
		// サブナビは作品を示すものであるため。
		{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Edit(rr, newEditRequest(tt.episodeID))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
			}
			assertNotFoundPage(t, rr)
		})
	}

	t.Run("数値でない id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/episodes/abc/edit", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.Edit(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

// TestEdit_RequiresCommitter verifies the edit form is protected by the committer role
// (committer proceeds, a regular user 403, an unauthenticated request is redirected to
// sign-in).
//
// [Ja] TestEdit_RequiresCommitter は編集フォームが committer ロールで保護されていることを
// 検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへリダイレクト)。
func TestEdit_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テストアニメ").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/episodes/{id}/edit", handler.Edit)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者はアクセス許可", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/edit", int64(episodeID)), nil)
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
