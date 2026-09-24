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

// timeZoneCookieNameはRails版がログイン外のビューワーのタイムゾーン
// (またはログインユーザーでtime_zone未設定の場合) を覚えておくCookieのキー名。
const timeZoneCookieName = "ann_time_zone"

// defaultTimeZoneはログインユーザーのtime_zoneもCookieも無い場合の
// 既定値。Rails版コントローラーの既定値と一致させる。
const defaultTimeZone = "Asia/Tokyo"

// ShowはGET /fragment/@{username}/tracking_heatmapを処理し、
// プロフィールページのStimulus controllerが <turbo-frame id="tracking-heatmap">
// に差し込むヒートマップHTMLフラグメントを返す。
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

// resolveTimeZoneは集計に使うタイムゾーンを選び、Rails版コントローラーと
// 同じ優先順 (ログインユーザーのtime_zoneが空でなく有効なIANA名ならそれ >
// "ann_time_zone" Cookieが空でなく有効なIANA名ならそれ > defaultTimeZone) で
// 決定する。無効なIANA名は次の候補にフォールスルーさせることで、クライアント
// が自由に書き換えられるCookieの不正値でUseCaseが500になるのを防ぐ。
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
