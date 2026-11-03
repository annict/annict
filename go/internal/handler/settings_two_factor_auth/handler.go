// Package settings_two_factor_authは二要素認証の設定画面のHTTPハンドラーを定義する。
package settings_two_factor_auth

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Handlerは二要素認証の設定画面のHTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                       *config.Config
	sessionManager            *session.Manager
	flashMgr                  *session.FlashManager
	imageHelper               *image.Helper
	limiter                   *ratelimit.Limiter
	getTwoFactorAuthStatusUC  *usecase.GetTwoFactorAuthStatusUsecase
	prepareTwoFactorAuthUC    *usecase.PrepareTwoFactorAuthUsecase
	getPendingTwoFactorAuthUC *usecase.GetPendingTwoFactorAuthUsecase
	enableTwoFactorAuthUC     *usecase.EnableTwoFactorAuthUsecase
	disableTwoFactorAuthUC    *usecase.DisableTwoFactorAuthUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	imageHelper *image.Helper,
	limiter *ratelimit.Limiter,
	getTwoFactorAuthStatusUC *usecase.GetTwoFactorAuthStatusUsecase,
	prepareTwoFactorAuthUC *usecase.PrepareTwoFactorAuthUsecase,
	getPendingTwoFactorAuthUC *usecase.GetPendingTwoFactorAuthUsecase,
	enableTwoFactorAuthUC *usecase.EnableTwoFactorAuthUsecase,
	disableTwoFactorAuthUC *usecase.DisableTwoFactorAuthUsecase,
) *Handler {
	return &Handler{
		cfg:                       cfg,
		sessionManager:            sessionManager,
		flashMgr:                  flashMgr,
		imageHelper:               imageHelper,
		limiter:                   limiter,
		getTwoFactorAuthStatusUC:  getTwoFactorAuthStatusUC,
		prepareTwoFactorAuthUC:    prepareTwoFactorAuthUC,
		getPendingTwoFactorAuthUC: getPendingTwoFactorAuthUC,
		enableTwoFactorAuthUC:     enableTwoFactorAuthUC,
		disableTwoFactorAuthUC:    disableTwoFactorAuthUC,
	}
}

// setNoStoreは、応答をブラウザや中継のキャッシュに残さないよう指示する。
// 二要素認証の画面は秘密鍵やリカバリーコードを含むため、戻る操作や共用の端末で読み返されないようにする。
func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store")
}

// isRateLimitedは、inputのキーで本人確認の試行の回数を数え、上限を超えたときにtrueを返します。
// 制限が無効化されている場合や、判定自体が失敗した場合はリクエストを通します。
func (h *Handler) isRateLimited(r *http.Request, input ratelimit.CheckInput, userID model.UserID) bool {
	if h.limiter == nil || h.cfg.DisableRateLimit {
		return false
	}

	ctx := r.Context()
	result, err := h.limiter.Check(ctx, input)
	if err != nil {
		slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err, "scope", "user")
		return false
	}
	if result.Allowed {
		return false
	}

	slog.WarnContext(ctx, "二要素認証の本人確認がRate Limitingにより制限されました", "user_id", userID, "key", input.Key)
	return true
}

// renderPageは二要素認証の画面をレイアウトに入れて、指定したステータスで描画する。
// titleKeyは文書のタイトルの翻訳キー、pathはog:urlに出す代表のパス。
func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, status int, titleKey string, path string, content templ.Component) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, path)
	meta.SetTitle(ctx, titleKey)

	component := layouts.Default(
		ctx,
		meta,
		viewmodel.NewUserForSidebar(user, h.imageHelper),
		viewmodel.NewSeasons(h.cfg),
		h.cfg.GetAssetVersion(),
		content,
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	setNoStore(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "二要素認証の画面のレスポンスの書き込みに失敗", "error", err)
	}
}
