package sign_up_username

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/sign_up_username"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New GET /sign_up/username - ユーザー名設定フォーム表示
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// クエリパラメータからトークンを取得
	token := r.URL.Query().Get("token")
	if token == "" {
		slog.Warn("トークンが指定されていません")
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_username_error_token_missing"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Redisからトークンを検証してメールアドレスを取得
	var email string
	if h.redisClient != nil {
		tokenKey := fmt.Sprintf("sign_up_token:%s", token)
		result, err := h.redisClient.Get(ctx, tokenKey).Result()
		if err != nil {
			slog.Warn("トークンが無効か期限切れです", "error", err)
			h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_username_error_token_invalid"))
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
		email = result
	} else {
		slog.Warn("Redisクライアントが設定されていないため、トークン検証をスキップします")
	}

	// Flashメッセージを取得
	flash := h.sessionMgr.GetFlash(w, r)
	formErrors, _ := h.sessionMgr.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_up_username_heading")
	meta.Description = i18n.T(ctx, "sign_up_username_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_up/username"

	// CSRFトークンを取得
	csrfToken := middleware.GetCSRFToken(r, h.sessionMgr)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), sign_up_username.New(ctx, token, email, csrfToken, formErrors))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
