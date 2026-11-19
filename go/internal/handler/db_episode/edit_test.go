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

// newEditRequestはエピソード編集フォームへのGETリクエストを、chiがルートパターンから
// 取り出すidのURLパラメータと、ルートのミドルウェアを通過したcommitterのユーザー付きで
// 組み立てる。
func newEditRequest(episodeID model.EpisodeID) *http.Request {
	req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/edit", int64(episodeID)), nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", fmt.Sprintf("%d", int64(episodeID)))

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	return withCreateTestUser(req, &model.User{ID: 1, Role: model.RoleEditor})
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<title>エピソード編集 | 第2話 | テストアニメ | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/episodes/%d/edit">`, int64(episodeID)),
		// 見出しは親作品を名指しし、共有のサブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		// フォームはエピソード自身を更新するため、HTMLフォームがPATCHを送るための
		// メソッドオーバーライドを添えてエピソードのパスを宛先にする。
		fmt.Sprintf(`action="/db/episodes/%d"`, int64(episodeID)),
		`name="_method" value="PATCH"`,
		`name="csrf_token"`,
		// 保存済みの値が各欄に開く。
		`value="第2話"`,
		"もう、お婿にいけません",
		"エピソード編集",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// フォームを開いた時点の版は保存済みの行から来るため、hiddenは空ではなく値を運ぶ
	// 必要がある。
	if strings.Contains(body, `name="updated_at" value=""`) {
		t.Error("版のhiddenが空で描画されています")
	}
	if !strings.Contains(body, `name="updated_at" value="`) {
		t.Error("版のhiddenが描画されていません")
	}
}

// TestEdit_DocumentTitleIsLocalizedAndUniquePerEpisodeは、編集者が並べて開きやすい同じ作品の
// 2エピソードの編集ページを検証する。タイトルは画面名で始まりエピソードが続くため、タイトル全体が
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
					t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
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
					t.Errorf("title = %q、期待値 = %q", titles[i], want)
				}
			}

			if titles[0] == titles[1] {
				t.Errorf("同じ作品のエピソード間でtitleが重複しています: %q", titles[0])
			}
		})
	}
}

// TestEdit_DocumentTitleOmitsWorkWithoutNameは、表示できる名前が無い作品のエピソードを
// 検証する。タイトルは空の区切りを描画せず作品を省き、エピソード識別子は残す。タブ・履歴・
// 支援技術でページを区別できる状態を保つため。
func TestEdit_DocumentTitleOmitsWorkWithoutName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// 空白だけのタイトルは表示名として空文字列に畳まれ、ページはフォールバックの分岐を
	// 通る。
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("   ").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Edit(rr, newEditRequest(episodeID))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	want := "<title>エピソード編集 | 第2話 | Annict DB</title>"
	if !strings.Contains(rr.Body.String(), want) {
		t.Errorf("レスポンスに%qが含まれていません", want)
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
		// 作品が失われたエピソードには編集するページが無い。そのページの見出しと
		// サブナビは作品を示すものであるため。
		{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Edit(rr, newEditRequest(tt.episodeID))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
			}
			assertNotFoundPage(t, rr)
		})
	}

	t.Run("数値でないid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/episodes/abc/edit", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.Edit(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

// TestEdit_RequiresCommitterは編集フォームがcommitterロールで保護されていることを
// 検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// TestEdit_ForbiddenWithoutMiddlewareは既存エピソードの編集フォームをUseCaseが拒否し、
// Handlerが403ページに変換することを検証する。
func TestEdit_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("編集フォーム認可境界テスト").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	handler := newTestHandler(t, db, tx)
	r := chi.NewRouter()
	r.Get("/db/episodes/{id}/edit", handler.Edit)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/edit", int64(episodeID)), nil))

	assertForbiddenPage(t, rr)
}

// TestEdit_RevisionsLinksは、フィーチャーフラグdb_edit_historyを有効にしたデバイスの閲覧者にだけ、
// エピソードの履歴へのリンクとサブナビの「履歴」を出すことを検証する。
func TestEdit_RevisionsLinks(t *testing.T) {
	t.Parallel()

	// feature_flagsは (device_token, name) で一意のため、ケースごとに別のトークンを使う。
	for _, tt := range []struct {
		name        string
		deviceToken string
		flagged     bool
	}{
		{name: "フラグを有効にしたデバイス", deviceToken: "db-episode-edit-revisions-flagged", flagged: true},
		{name: "フラグを有効にしていないデバイス", deviceToken: "db-episode-edit-revisions-unflagged", flagged: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴リンクの作品").Build()
			episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()
			if tt.flagged {
				testutil.NewFeatureFlagBuilder(t, tx).WithDeviceToken(tt.deviceToken).WithName(string(model.FeatureFlagDBEditHistory)).Build()
			}

			req := newEditRequest(episodeID)
			req.AddCookie(&http.Cookie{Name: authMiddleware.DeviceTokenCookieName, Value: tt.deviceToken})
			rr := httptest.NewRecorder()
			newTestHandler(t, db, tx).Edit(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
			}
			body := rr.Body.String()
			episodeLink := fmt.Sprintf(`href="/db/episodes/%d/revisions"`, int64(episodeID))
			if got := strings.Contains(body, episodeLink); got != tt.flagged {
				t.Errorf("%sの有無 = %v、期待値 = %v", episodeLink, got, tt.flagged)
			}
			// 見出しの「履歴」には、履歴を表すアイコンを装飾としてテキストの左に付ける。
			if got := strings.Contains(body, episodeLink+` class="btn rounded-full" data-variant="outline" data-size="sm"><svg data-icon="inline-start" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M136,80v43.47`); got != tt.flagged {
				t.Errorf("アイコン付きの履歴へのリンクの有無 = %v、期待値 = %v", got, tt.flagged)
			}
			// 作品の履歴の入口は作品一覧の行と作品の編集画面に置き、作品のサブナビには出さない。
			if workLink := fmt.Sprintf(`href="/db/works/%d/revisions"`, int64(workID)); strings.Contains(body, workLink) {
				t.Errorf("作品のサブナビに%sが出ている", workLink)
			}
		})
	}
}
