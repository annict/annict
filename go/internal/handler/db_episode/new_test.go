package db_episode

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

// newNewRequest builds a GET request for a work's bulk-create form with the work_id URL
// parameter chi would have extracted from the route pattern.
//
// [Ja] newNewRequest はある作品の一括作成フォームへの GET リクエストを、chi がルートパターン
// から取り出す work_id の URL パラメータ付きで組み立てる。
func newNewRequest(workID model.WorkID) *http.Request {
	req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/episodes/new", int64(workID)), nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("work_id", fmt.Sprintf("%d", int64(workID)))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.New(rr, newNewRequest(workID))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<title>エピソード登録 | テストアニメ | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/episodes/new">`, int64(workID)),
		// The heading names the parent work, and the shared subnav links back to its form.
		//
		// [Ja] 見出しは親作品を名指しし、共有のサブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		// The form posts the lines to the work's episode collection.
		//
		// [Ja] フォームは行を作品のエピソードコレクションへ POST する。
		fmt.Sprintf(`action="/db/works/%d/episodes"`, int64(workID)),
		`method="POST"`,
		`name="csrf_token"`,
		`<textarea id="rows" name="rows"`,
		"エピソード登録",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}
}

func TestNew_ManualCreationRestriction(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("予定話数到達アニメ").
		WithManualEpisodesCount(1).
		Build()
	testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()
	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name         string
		user         *model.User
		wantTitle    string
		wantMessage  string
		wantReadonly bool
	}{
		{
			name:         "編集者には警告してフォームを無効化する",
			user:         &model.User{ID: 1, Role: model.RoleEditor},
			wantTitle:    "手動登録できません",
			wantMessage:  "新規登録はできません",
			wantReadonly: true,
		},
		{
			name:         "管理者には警告するがフォームを有効に保つ",
			user:         &model.User{ID: 1, Role: model.RoleAdmin},
			wantTitle:    "通常は手動登録できません",
			wantMessage:  "管理者は手動でも登録できますが",
			wantReadonly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.New(rr, withCreateTestUser(newNewRequest(workID), tt.user))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rr.Code)
			}
			body := rr.Body.String()
			// The editor heading is a suffix of the admin one, so the whole element is matched:
			// substring alone would let the admin wording satisfy the editor case.
			//
			// [Ja] 編集者向けの見出しは管理者向けの見出しの後方一致になるため、要素全体で
			// 照合する。部分一致だけでは管理者向けの文言でも編集者のケースが通ってしまう。
			if !strings.Contains(body, "<h2>"+tt.wantTitle+"</h2>") {
				t.Errorf("手動作成制限の警告の見出しが %q ではありません", tt.wantTitle)
			}
			if !strings.Contains(body, "話数分のエピソード") || !strings.Contains(body, tt.wantMessage) {
				t.Errorf("手動作成制限の警告に %q が含まれていません", tt.wantMessage)
			}
			if got := strings.Contains(body, "readonly"); got != tt.wantReadonly {
				t.Errorf("readonly = %v, want %v", got, tt.wantReadonly)
			}
			if got := strings.Contains(body, "disabled"); got != tt.wantReadonly {
				t.Errorf("disabled = %v, want %v", got, tt.wantReadonly)
			}
		})
	}
}

func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済みアニメ").Build()
	if _, err := tx.Exec("UPDATE works SET deleted_at = NOW() WHERE id = $1", int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)

	t.Run("存在しない作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.New(rr, newNewRequest(model.WorkID(999999999)))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("削除済みの作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.New(rr, newNewRequest(deletedWorkID))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("数値でない work_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/works/abc/episodes/new", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("work_id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.New(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

// TestNew_RequiresCommitter verifies the bulk-create form is protected by the committer role
// (committer proceeds, a regular user 403, an unauthenticated request is redirected to
// sign-in).
//
// [Ja] TestNew_RequiresCommitter は一括作成フォームが committer ロールで保護されていることを
// 検証する (committer は処理続行、一般ユーザーは 403、未認証はサインインへリダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テストアニメ").Build()
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/{work_id}/episodes/new", handler.New)

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
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/episodes/new", int64(workID)), nil)
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
