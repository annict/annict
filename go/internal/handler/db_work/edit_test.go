package db_work

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// TestEditはDB作品編集フォームが既存値を埋めて描画されることを検証する。
func TestEdit(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("編集対象アニメ").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	if _, err := tx.Exec(`
		UPDATE works SET
			official_site_url = 'https://example.dev/anime',
			synopsis = 'あらすじテキスト',
			twitter_username = 'anime_official',
			media = 1
		WHERE id = $1
	`, int64(workID)); err != nil {
		t.Fatalf("worksのフィールド設定に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := withCommitterUser(httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// DBのページはブラウザのタブで公開画面と区別できるよう " | Annict DB" の
		// タイトルサフィックスを持つ。
		"<title>作品編集 | 編集対象アニメ | Annict DB</title>",
		"<form",
		fmt.Sprintf(`action="/db/works/%d"`, int64(workID)),
		`name="_method"`,
		`value="PATCH"`,
		"csrf_token",
		`value="編集対象アニメ"`,                   // タイトルが初期値として埋まる
		`value="https://example.dev/anime"`, // 公式サイトURLが埋まる
		"あらすじテキスト",                          // あらすじがtextareaに埋まる
		`value="anime_official"`,            // Twitterユーザー名が埋まる
		`value="2024" selected`,             // シーズン年が選択済み
		fmt.Sprintf(`href="/works/%d"`, int64(workID)),
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Content-Type = %v、期待値 = %v", ct, expectedContentType)
	}
}

// TestEdit_NotFoundは存在しない作品IDで404を返すことを検証する。
func TestEdit_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := withCommitterUser(httptest.NewRequest("GET", "/db/works/999999999/edit", nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestEdit_InvalidIDは数値でないIDで404を返すことを検証する。
func TestEdit_InvalidID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := httptest.NewRequest("GET", "/db/works/not-a-number/edit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
	}
	assertNotFoundPage(t, rr)
}

// TestEdit_RequiresCommitterは編集フォームのルートがcommitterロールで保護されている
// ことを検証する (committerは200、一般ユーザーは403、未認証はサインインへリダイレクト)。
func TestEdit_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト作品").Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{id}/edit", handler.Edit)

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
			name:       "管理者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleAdmin},
			wantStatus: http.StatusOK,
		},
		{
			name:       "編集者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil)
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

// TestEdit_ForbiddenWithoutMiddlewareはUseCaseの認可失敗が403ページになることを検証する。
func TestEdit_ForbiddenWithoutMiddleware(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("編集フォーム認可境界テスト").Build()
	handler := newTestHandler(t, db, tx)
	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil))

	assertForbiddenPage(t, rr)
}

// TestEdit_DocumentTitleWithoutWorkNameは、タイトルが空白文字だけの作品では文書タイトルが
// 画面名だけになることを検証する。見出しがフォールバックする名前と同じものになる。
func TestEdit_DocumentTitleWithoutWorkName(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle(" \t ").Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/works/{id}/edit", handler.Edit)

	req := withCommitterUser(httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>作品編集 | Annict DB</title>",
		">作品編集</h1>",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestEdit_RevisionsLinkByDeviceTokenは、フィーチャーフラグdb_edit_historyをデバイスで有効にした閲覧者に、
// 履歴の画面をゲートするリバースプロキシと同じく、編集画面の見出しに作品の「履歴」を出すことを検証する。
func TestEdit_RevisionsLinkByDeviceToken(t *testing.T) {
	t.Parallel()

	// feature_flagsは (device_token, name) で一意のため、並行するサブテストが互いを待たないよう
	// ケースごとに別のトークンを使う。Cookieの値に使える文字だけで書く。
	for _, tt := range []struct {
		name        string
		deviceToken string
		flagged     bool
	}{
		{name: "フラグを有効にしたデバイス", deviceToken: "db-work-edit-revisions-flagged", flagged: true},
		{name: "フラグを有効にしていないデバイス", deviceToken: "db-work-edit-revisions-unflagged", flagged: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
			workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴リンク対象").Build()
			deviceToken := tt.deviceToken
			if tt.flagged {
				testutil.NewFeatureFlagBuilder(t, tx).WithDeviceToken(deviceToken).WithName(string(model.FeatureFlagDBEditHistory)).Build()
			}

			r := chi.NewRouter()
			r.Get("/db/works/{id}/edit", newTestHandler(t, db, tx).Edit)

			req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/edit", int64(workID)), nil)
			req.AddCookie(&http.Cookie{Name: authMiddleware.DeviceTokenCookieName, Value: deviceToken})
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: userID, Role: model.RoleEditor}))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			revisionsLink := fmt.Sprintf(`href="/db/works/%d/revisions"`, int64(workID))
			if got := strings.Contains(rr.Body.String(), revisionsLink); got != tt.flagged {
				t.Errorf("履歴へのリンクの有無 = %v、期待値 = %v", got, tt.flagged)
			}
			// 見出しの「履歴」には、履歴を表すアイコンを装飾としてテキストの左に付ける。
			if got := strings.Contains(rr.Body.String(), revisionsLink+` class="btn rounded-full" data-variant="outline" data-size="sm"><svg data-icon="inline-start" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M136,80v43.47`); got != tt.flagged {
				t.Errorf("アイコン付きの履歴へのリンクの有無 = %v、期待値 = %v", got, tt.flagged)
			}
		})
	}
}
