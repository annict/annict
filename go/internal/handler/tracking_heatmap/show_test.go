package tracking_heatmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// newTestHandler wires the dependency graph used by the handler tests.
//
// [Ja] newTestHandler はハンドラーテストで使う依存グラフを構築する。
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	userRepo := repository.NewUserRepository(queries)
	recordRepo := repository.NewRecordRepository(queries)
	uc := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)
	return NewHandler(uc)
}

// newTestHandlerWithUser builds a handler plus a real user row so the use
// case's username lookup succeeds. The returned username is the auto-generated
// unique value produced by testutil.NewUserBuilder, which keeps parallel tests
// from serializing on the users.username UNIQUE index.
//
// [Ja] newTestHandlerWithUser はハンドラーと、UseCase の username 検索が成功する
// よう実ユーザー行を作る。返す username は testutil.NewUserBuilder が自動生成する
// ユニーク値で、users.username の UNIQUE インデックスで並行テストが直列化する
// のを避ける。
func newTestHandlerWithUser(t *testing.T) (*Handler, string) {
	t.Helper()
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	userRepo := repository.NewUserRepository(queries)
	recordRepo := repository.NewRecordRepository(queries)
	uc := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)

	userID := testutil.NewUserBuilder(t, tx).Build()
	var username string
	if err := tx.QueryRow("SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
		t.Fatalf("username の取得に失敗: %v", err)
	}
	return NewHandler(uc), username
}

// TestShow_NotFound ensures the handler returns 404 when the requested
// username does not exist.
//
// [Ja] 存在しない username が指定された場合に 404 を返すこと。
func TestShow_NotFound(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@no_such_user_xyz/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// TestShow_DeletedUser ensures the handler returns 404 when the target user
// has been soft-deleted.
//
// [Ja] 削除済みユーザーへのアクセスが 404 になること。
func TestShow_DeletedUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	userRepo := repository.NewUserRepository(queries)
	recordRepo := repository.NewRecordRepository(queries)
	uc := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)
	handler := NewHandler(uc)

	userID := testutil.NewUserBuilder(t, tx).Build()
	if _, err := tx.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(userID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}
	var username string
	if err := tx.QueryRow("SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
		t.Fatalf("username の取得に失敗: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// TestShow_Success returns the fragment HTML with the expected wrapper,
// classes, and attributes the profile page's Stimulus controller and SCSS
// expect.
//
// [Ja] レスポンス HTML がプロフィールページの Stimulus controller / SCSS が
// 期待するラッパー・クラス・属性を含むこと。
func TestShow_Success(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html...", ct)
	}

	body := rr.Body.String()
	wantSubstrings := []string{
		`<turbo-frame id="tracking-heatmap">`,
		`class="c-tracking-heatmap"`,
		`c-tracking-heatmap__day`,
		`c-tracking-heatmap__density-`,
		`data-bs-toggle="tooltip"`,
		`data-bs-placement="top"`,
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(body, s) {
			t.Errorf("response body missing %q", s)
		}
	}
}

// TestShow_TimeZoneCookieRespected verifies that the ann_time_zone cookie is
// honored when there is no signed-in user. We pass a valid IANA name and
// assert a 200 response: the cookie value flowed through the handler's time
// zone resolution and the use case loaded the location successfully.
//
// [Ja] ログインユーザー不在時に ann_time_zone Cookie が優先されることを検証する。
// 有効な IANA タイムゾーン名を渡し 200 が返ることから、Cookie の値が Handler の
// タイムゾーン解決を経て UseCase で正しく扱われたことが分かる。
func TestShow_TimeZoneCookieRespected(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	req.AddCookie(&http.Cookie{Name: timeZoneCookieName, Value: "Europe/London"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// TestShow_LoggedInUserTimeZonePrecedence verifies that the signed-in user's
// time_zone takes precedence over the cookie. Both values are valid IANA
// names, so the request should succeed; the test relies on the handler
// preferring user.TimeZone over the cookie in resolveTimeZone, matching
// Rails' "current_user&.time_zone.presence || cookies[...].presence" order.
//
// [Ja] ログインユーザーの time_zone が Cookie より優先されることを検証する。
// 両方とも有効な IANA 名なのでリクエストは 200 になる。Handler は
// resolveTimeZone でユーザー値を Cookie より先に評価するため、Rails の
// "current_user&.time_zone.presence || cookies[...].presence" と同じ優先順
// で動くことを担保している。
func TestShow_LoggedInUserTimeZonePrecedence(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	req.AddCookie(&http.Cookie{Name: timeZoneCookieName, Value: "Asia/Tokyo"})

	user := &model.User{ID: model.UserID(0), Username: "ctx_user", TimeZone: "Europe/London"}
	ctx := context.WithValue(req.Context(), authMiddleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// TestShow_InvalidCookieFallsBack verifies that a malformed IANA name in the
// ann_time_zone cookie is rejected by the handler (which a client can set
// freely) and the request still succeeds by falling back to defaultTimeZone.
//
// [Ja] ann_time_zone Cookie の不正な IANA 名が Handler で拒否され、
// defaultTimeZone にフォールバックして 200 になることを検証する。
// Cookie はクライアントが自由に書き換えられるため、不正値で UseCase が 500 に
// なる経路を Handler 側で塞いでいる。
func TestShow_InvalidCookieFallsBack(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	req.AddCookie(&http.Cookie{Name: timeZoneCookieName, Value: "Not/A_Real_Zone_xyz"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (invalid cookie tz should fall back to defaultTimeZone)", rr.Code, http.StatusOK)
	}
}

// TestShow_InvalidUserTimeZoneFallsBack verifies that a malformed IANA name
// on the signed-in user falls through to the cookie (and ultimately
// defaultTimeZone), instead of bubbling up as a 500. This is the user-row
// counterpart of TestShow_InvalidCookieFallsBack.
//
// [Ja] ログインユーザーの time_zone に不正な IANA 名が入っていた場合に、
// Cookie (最終的には defaultTimeZone) にフォールスルーして 200 を返すことを
// 検証する。TestShow_InvalidCookieFallsBack のユーザー行版。
func TestShow_InvalidUserTimeZoneFallsBack(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)

	user := &model.User{ID: model.UserID(0), Username: "ctx_user", TimeZone: "Not/A_Real_Zone_xyz"}
	ctx := context.WithValue(req.Context(), authMiddleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (invalid user tz should fall back to defaultTimeZone)", rr.Code, http.StatusOK)
	}
}

// TestShow_DefaultTimeZoneFallback verifies that when neither the
// signed-in user nor the cookie supplies a time zone, the request still
// succeeds (resolving to "Asia/Tokyo").
//
// [Ja] ログインユーザーも Cookie もタイムゾーンを与えないとき、デフォルト
// ("Asia/Tokyo") にフォールバックしてリクエストが成功すること。
func TestShow_DefaultTimeZoneFallback(t *testing.T) {
	t.Parallel()
	handler, username := newTestHandlerWithUser(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
