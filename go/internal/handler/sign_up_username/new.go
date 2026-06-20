package sign_up_username

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
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
		slog.WarnContext(ctx, "トークンが指定されていません")
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_username_error_token_missing"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Redisからトークンを検証してメールアドレスを取得
	var email string
	if h.redisClient != nil {
		tokenKey := fmt.Sprintf("sign_up_token:%s", token)
		result, err := h.redisClient.Get(ctx, tokenKey).Result()
		if err != nil {
			slog.WarnContext(ctx, "トークンが無効か期限切れです", "error", err)
			h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_username_error_token_invalid"))
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
		email = result
	} else {
		slog.WarnContext(ctx, "Redisクライアントが設定されていないため、トークン検証をスキップします")
	}

	h.renderNewForm(w, r, http.StatusOK, nil, token, email, "")
}

// renderNewForm はユーザー名設定フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderNewForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, token string, email string, username string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_up_username_heading")
	meta.Description = i18n.T(ctx, "sign_up_username_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_up/username"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := sign_up_username.NewPageData{
		CSRFToken:  csrfToken,
		FormErrors: formErrors,
		Token:      token,
		Email:      email,
		Username:   username,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), sign_up_username.New(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
	}
}
