package tracking_heatmap

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/pages/tracking_heatmap"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// timeZoneCookieName is the cookie key the Rails app uses to remember the
// viewer's preferred time zone when they have not signed in (or when the
// signed-in user has no time_zone set).
//
// [Ja] timeZoneCookieName は Rails 版がログイン外のビューワーのタイムゾーン
// (またはログインユーザーで time_zone 未設定の場合) を覚えておく Cookie のキー名。
const timeZoneCookieName = "ann_time_zone"

// defaultTimeZone is the fallback when neither the signed-in user nor the
// cookie supplies a time zone. Mirrors the Rails controller default.
//
// [Ja] defaultTimeZone はログインユーザーの time_zone も Cookie も無い場合の
// 既定値。Rails 版コントローラーの既定値と一致させる。
const defaultTimeZone = "Asia/Tokyo"

// Show handles GET /fragment/@{username}/tracking_heatmap and returns the
// heatmap HTML fragment that the profile page's Stimulus controller injects
// into <turbo-frame id="tracking-heatmap">.
//
// [Ja] Show は GET /fragment/@{username}/tracking_heatmap を処理し、
// プロフィールページの Stimulus controller が <turbo-frame id="tracking-heatmap">
// に差し込むヒートマップ HTML フラグメントを返す。
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := chi.URLParam(r, "username")
	if username == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	tz := h.resolveTimeZone(r)

	output, err := h.getTrackingHeatmapUC.Execute(ctx, usecase.GetTrackingHeatmapInput{
		Username: username,
		TimeZone: tz,
		Now:      time.Now(),
	})
	if err != nil {
		var ae *model.AppError
		if errors.As(err, &ae) && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "視聴記録ヒートマップの取得に失敗", "error", err, "username", username)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cells := viewmodel.NewTrackingHeatmapCellsFromUsecase(output.Cells)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := tracking_heatmap.Show(cells).Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "視聴記録ヒートマップのレンダリングに失敗", "error", err)
	}
}

// resolveTimeZone selects the time zone used to bucket records into days,
// matching the Rails controller's precedence: signed-in user's time_zone
// (when non-empty and a valid IANA name) > "ann_time_zone" cookie (when
// non-empty and a valid IANA name) > defaultTimeZone. Invalid IANA names
// fall through to the next candidate so a malformed cookie (which a client
// can set freely) cannot force the use case to 500.
//
// [Ja] resolveTimeZone は集計に使うタイムゾーンを選び、Rails 版コントローラーと
// 同じ優先順 (ログインユーザーの time_zone が空でなく有効な IANA 名ならそれ >
// "ann_time_zone" Cookie が空でなく有効な IANA 名ならそれ > defaultTimeZone) で
// 決定する。無効な IANA 名は次の候補にフォールスルーさせることで、クライアント
// が自由に書き換えられる Cookie の不正値で UseCase が 500 になるのを防ぐ。
func (h *Handler) resolveTimeZone(r *http.Request) string {
	if user := authMiddleware.GetUserFromContext(r.Context()); user != nil && user.TimeZone != "" {
		if _, err := time.LoadLocation(user.TimeZone); err == nil {
			return user.TimeZone
		}
	}
	if c, err := r.Cookie(timeZoneCookieName); err == nil && c.Value != "" {
		if _, err := time.LoadLocation(c.Value); err == nil {
			return c.Value
		}
	}
	return defaultTimeZone
}
