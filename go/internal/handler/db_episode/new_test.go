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

// newNewRequestはある作品の一括作成フォームへのGETリクエストを、chiがルートパターン
// から取り出すwork_idのURLパラメータ付きで組み立てる。
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
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		"<title>エピソード登録 | テストアニメ | Annict DB</title>",
		fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/episodes/new">`, int64(workID)),
		// 見出しは親作品を名指しし、共有のサブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		// フォームは行を作品のエピソードコレクションへPOSTする。
		fmt.Sprintf(`action="/db/works/%d/episodes"`, int64(workID)),
		`method="POST"`,
		`name="csrf_token"`,
		`<textarea id="rows" name="rows"`,
		"エピソード登録",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
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
				t.Fatalf("ステータスコード = %d、期待値 = 200", rr.Code)
			}
			body := rr.Body.String()
			// 編集者向けの見出しは管理者向けの見出しの後方一致になるため、要素全体で
			// 照合する。部分一致だけでは管理者向けの文言でも編集者のケースが通ってしまう。
			if !strings.Contains(body, "<h2>"+tt.wantTitle+"</h2>") {
				t.Errorf("手動作成制限の警告の見出しが%qではありません", tt.wantTitle)
			}
			if !strings.Contains(body, "話数分のエピソード") || !strings.Contains(body, tt.wantMessage) {
				t.Errorf("手動作成制限の警告に%qが含まれていません", tt.wantMessage)
			}
			if got := strings.Contains(body, "readonly"); got != tt.wantReadonly {
				t.Errorf("readonly = %v、期待値 = %v", got, tt.wantReadonly)
			}
			if got := strings.Contains(body, "disabled"); got != tt.wantReadonly {
				t.Errorf("disabled = %v、期待値 = %v", got, tt.wantReadonly)
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
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("削除済みの作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.New(rr, newNewRequest(deletedWorkID))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("数値でないwork_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/works/abc/episodes/new", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("work_id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.New(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

// TestNew_RequiresCommitterは一括作成フォームがcommitterロールで保護されていることを
// 検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへリダイレクト)。
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
