package db_episode

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// readDeleteTargetState returns the state column the delete writes, so a test can tell a deleted
// episode from one the submit left alone.
//
// [Ja] readDeleteTargetState は削除が書く状態カラムを返す。削除されたエピソードと、送信が手を
// 触れなかったエピソードをテストが区別できるようにするため。
func readDeleteTargetState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var deletedAt sql.NullTime
	if err := db.QueryRow(`SELECT deleted_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&deletedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return deletedAt
}

// newDeleteRequest builds a DELETE request from an administrator for an episode. It carries no
// route context, so it is what the tests that go through a real chi router use; the router fills
// the URL parameter in itself.
//
// [Ja] newDeleteRequest はあるエピソードに対する管理者からの DELETE リクエストを組み立てる。
// ルートコンテキストを持たないため、実際の chi ルーターを通すテストはこちらを使う (URL パラメータ
// はルーター自身が埋める)。
func newDeleteRequest(episodeID model.EpisodeID) *http.Request {
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/db/episodes/%d", int64(episodeID)), nil)
	admin := &model.User{ID: 1, Role: model.RoleAdmin}

	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, admin))
}

// TestDelete_Success verifies the submit soft-deletes the episode and lands on the work's episode
// list, where the remaining rows are.
//
// [Ja] TestDelete_Success は、送信がエピソードをソフトデリートし、残った行が並ぶその作品の
// エピソード一覧に着地することを検証する。
func TestDelete_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, newDeleteRequest(episodeID))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("Location = %q, want %q", location, want)
	}
	if deletedAt := readDeleteTargetState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL, want 削除の時刻")
	}
}

// TestDelete_HTMX verifies the submit the episode list's delete button makes gets HX-Redirect
// instead of a redirect htmx would follow and swap into the button.
//
// [Ja] TestDelete_HTMX は、エピソード一覧の削除ボタンが行う送信に対し、htmx が追ってボタンに
// スワップしてしまうリダイレクトではなく HX-Redirect を返すことを検証する。
func TestDelete_HTMX(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}", handler.Delete)

	req := newDeleteRequest(episodeID)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if redirect := rr.Header().Get("HX-Redirect"); redirect != want {
		t.Errorf("HX-Redirect = %q, want %q", redirect, want)
	}
	if location := rr.Header().Get("Location"); location != "" {
		t.Errorf("Location = %q, want 空 (htmx は HX-Redirect で遷移する)", location)
	}
	if deletedAt := readDeleteTargetState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL, want 削除の時刻")
	}
}

// TestDelete_NotFound verifies a submit for an episode that cannot be deleted returns 404 rather
// than reporting a write that did not happen.
//
// [Ja] TestDelete_NotFound は、削除できないエピソードへの送信が、起きなかった書き込みを報告せず
// 404 を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertCreateTargetWork(t, db)
	deletedID := insertUpdateTargetEpisode(t, db, workID)
	if _, err := db.Exec(`UPDATE episodes SET deleted_at = NOW() WHERE id = $1`, int64(deletedID)); err != nil {
		t.Fatalf("エピソードの削除に失敗: %v", err)
	}
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}", handler.Delete)

	for name, target := range map[string]string{
		"削除済みのエピソード": fmt.Sprintf("/db/episodes/%d", int64(deletedID)),
		"存在しないエピソード": "/db/episodes/999999999",
		"数値でない id":   "/db/episodes/abc",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", target, nil)
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleAdmin}))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("status = %d, want %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestDelete_RequiresAdmin verifies the endpoint is gated by the admin role at the HTTP boundary
// and that an editor, who may archive an episode, is turned away. The delete usecase repeats the
// role check for non-HTTP entry points.
//
// [Ja] TestDelete_RequiresAdmin は HTTP 境界でエンドポイントが admin ロールによりゲートされ、
// エピソードを非公開にできる編集者は弾かれることを検証する。削除 UseCase は HTTP 以外の entry
// point に対してもロール検査を繰り返す。
func TestDelete_RequiresAdmin(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertCreateTargetWork(t, db)
	episodeID := insertUpdateTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAdmin)
		r.Delete("/db/episodes/{id}", handler.Delete)
	})

	users := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "編集者は403", user: &model.User{ID: 1, Role: model.RoleEditor}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range users {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/db/episodes/%d", int64(episodeID)), nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if deletedAt := readDeleteTargetState(t, db, episodeID); deletedAt.Valid {
				t.Error("拒否された送信が episodes.deleted_at を立てました")
			}
		})
	}

	// An administrator passes the middleware and deletes the episode. This subtest runs last
	// because it is the one that changes the row the rejections assert is untouched.
	//
	// [Ja] 管理者はミドルウェアを通過し、エピソードを削除する。拒否のケースが「変更されていない」
	// と検証する行を変えるのはこのサブテストだけのため、最後に走らせる。
	t.Run("管理者は通過", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, newDeleteRequest(episodeID))

		if rr.Code != http.StatusSeeOther {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusSeeOther)
		}
	})
}

// TestDelete_Forbidden verifies a direct Handler invocation maps the UseCase authorization failure
// to 403 instead of exposing it as an internal error.
//
// [Ja] TestDelete_Forbidden は Handler の直接呼び出しで UseCase の認可失敗を内部エラーとして公開
// せず、403 に変換することを検証する。
func TestDelete_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", "/db/episodes/1", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("status = %d, want %d", status, http.StatusForbidden)
	}
}
