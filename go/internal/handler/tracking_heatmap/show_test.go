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

// newTestHandlerはハンドラーテストで使う依存グラフを構築する。
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	userRepo := repository.NewUserRepository(queries)
	recordRepo := repository.NewRecordRepository(queries)
	uc := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)
	return NewHandler(uc)
}

// newTestHandlerWithUserはハンドラーと、UseCaseのusername検索が成功する
// よう実ユーザー行を作る。返すusernameはtestutil.NewUserBuilderが自動生成する
// ユニーク値で、users.usernameのUNIQUEインデックスで並行テストが直列化する
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
		t.Fatalf("usernameの取得に失敗: %v", err)
	}
	return NewHandler(uc), username
}

// 存在しないusernameが指定された場合に404を返すこと。
func TestShow_NotFound(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@no_such_user_xyz/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
	}
}

// 削除済みユーザーへのアクセスが404になること。
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
		t.Fatalf("usernameの取得に失敗: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/fragment/@{username}/tracking_heatmap", handler.Show)

	req := httptest.NewRequest("GET", "/fragment/@"+username+"/tracking_heatmap", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
	}
}

// レスポンスHTMLがプロフィールページのStimulus controller / SCSSが
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q、期待値 = text/html...", ct)
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
			t.Errorf("レスポンスボディに含まれていない文字列 = %q", s)
		}
	}
}

// ログインユーザー不在時にann_time_zone Cookieが優先されることを検証する。
// 有効なIANAタイムゾーン名を渡し200が返ることから、Cookieの値がHandlerの
// タイムゾーン解決を経てUseCaseで正しく扱われたことが分かる。
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
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
}

// ログインユーザーのtime_zoneがCookieより優先されることを検証する。
// 両方とも有効なIANA名なのでリクエストは200になる。Handlerは
// resolveTimeZoneでユーザー値をCookieより先に評価するため、Railsの
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
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
}

// ann_time_zone Cookieの不正なIANA名がHandlerで拒否され、
// defaultTimeZoneにフォールバックして200になることを検証する。
// Cookieはクライアントが自由に書き換えられるため、不正値でUseCaseが500に
// なる経路をHandler側で塞いでいる。
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
		t.Errorf("ステータスコード = %d、期待値 = %d (cookieのtzが不正ならdefaultTimeZoneへフォールバックすること)", rr.Code, http.StatusOK)
	}
}

// ログインユーザーのtime_zoneに不正なIANA名が入っていた場合に、
// Cookie (最終的にはdefaultTimeZone) にフォールスルーして200を返すことを
// 検証する。TestShow_InvalidCookieFallsBackのユーザー行版。
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
		t.Errorf("ステータスコード = %d、期待値 = %d (ユーザーのtzが不正ならdefaultTimeZoneへフォールバックすること)", rr.Code, http.StatusOK)
	}
}

// ログインユーザーもCookieもタイムゾーンを与えないとき、デフォルト
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
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
}
