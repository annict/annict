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

// readDeleteTargetStateは削除が書く状態カラムを返す。削除されたエピソードと、送信が手を
// 触れなかったエピソードをテストが区別できるようにするため。
func readDeleteTargetState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var deletedAt sql.NullTime
	if err := db.QueryRow(`SELECT deleted_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&deletedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return deletedAt
}

// newDeleteRequestはあるエピソードに対する管理者からのDELETEリクエストを組み立てる。
// ルートコンテキストを持たないため、実際のchiルーターを通すテストはこちらを使う (URLパラメータ
// はルーター自身が埋める)。
func newDeleteRequest(episodeID model.EpisodeID) *http.Request {
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/db/episodes/%d", int64(episodeID)), nil)
	admin := &model.User{ID: 1, Role: model.RoleAdmin}

	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, admin))
}

// TestDelete_Successは、送信がエピソードをソフトデリートし、残った行が並ぶその作品の
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, want)
	}
	if deletedAt := readDeleteTargetState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
	}
}

// TestDelete_HTMXは、エピソード一覧の削除ボタンが行う送信に対し、htmxが追ってボタンに
// スワップしてしまうリダイレクトではなくHX-Redirectを返すことを検証する。
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusNoContent)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if redirect := rr.Header().Get("HX-Redirect"); redirect != want {
		t.Errorf("HX-Redirect = %q、期待値 = %q", redirect, want)
	}
	if location := rr.Header().Get("Location"); location != "" {
		t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移する)", location)
	}
	if deletedAt := readDeleteTargetState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
	}
}

// TestDelete_NotFoundは、削除できないエピソードへの送信が、起きなかった書き込みを報告せず
// 404を返すことを検証する。
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
		"数値でないid":    "/db/episodes/abc",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", target, nil)
			req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleAdmin}))
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestDelete_RequiresAdminはHTTP境界でエンドポイントがadminロールによりゲートされ、
// エピソードを非公開にできる編集者は弾かれることを検証する。削除UseCaseはHTTP以外のentry
// pointに対してもロール検査を繰り返す。
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
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
			if deletedAt := readDeleteTargetState(t, db, episodeID); deletedAt.Valid {
				t.Error("拒否された送信がepisodes.deleted_atを立てました")
			}
		})
	}

	// 管理者はミドルウェアを通過し、エピソードを削除する。拒否のケースが「変更されていない」
	// と検証する行を変えるのはこのサブテストだけのため、最後に走らせる。
	t.Run("管理者は通過", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, newDeleteRequest(episodeID))

		if rr.Code != http.StatusSeeOther {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
		}
	})
}

// TestDelete_ForbiddenはHandlerの直接呼び出しでUseCaseの認可失敗を内部エラーとして公開
// せず、403に変換することを検証する。
func TestDelete_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", "/db/episodes/1", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusForbidden)
	}
}
