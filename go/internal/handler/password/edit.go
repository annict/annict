package password

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	errorpages "github.com/annict/annict/go/internal/templates/pages/errors"
	passwordpages "github.com/annict/annict/go/internal/templates/pages/password"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Edit は新しいパスワード入力フォームを表示します (GET /password/edit)
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")

	if token == "" {
		h.renderInvalidTokenError(w, r)
		return
	}

	// Rate Limiting: トークン検証の制限（10回/時間/IP）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ip := clientip.GetClientIP(r)
		tokenVerifyKey := fmt.Sprintf("password_reset:token_verify:ip:%s", ip)
		allowed, err := h.limiter.Check(ctx, tokenVerifyKey, 10, 1*time.Hour)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "トークン検証がRate Limitingにより制限されました",
				"ip_address", ip,
			)
			http.Error(w, i18n.T(ctx, "rate_limit_exceeded"), http.StatusTooManyRequests)
			return
		}
	}

	// UseCaseでトークンの有効性を検証
	result, err := h.getPasswordResetTokenUC.Execute(ctx, usecase.GetPasswordResetTokenInput{
		Token: token,
	})
	if err != nil {
		slog.ErrorContext(ctx, "パスワードリセットトークンの検証エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !result.Valid {
		slog.WarnContext(ctx, "無効なパスワードリセットトークンによるアクセス",
			"ip_address", clientip.GetClientIP(r),
		)
		h.renderInvalidTokenError(w, r)
		return
	}

	h.renderEditForm(w, r, http.StatusOK, nil, token)
}

// renderEditForm は新しいパスワード入力フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderEditForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, token string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_edit_title")
	meta.OGURL = h.cfg.AppURL() + "/password/edit"

	csrfToken := middleware.GetCSRFToken(r, h.sessionMgr)

	data := passwordpages.EditPageData{
		CSRFToken:  csrfToken,
		Token:      token,
		FormErrors: formErrors,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), passwordpages.Edit(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}

// renderInvalidTokenError は無効なトークンエラーを表示します
func (h *Handler) renderInvalidTokenError(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.Title = i18n.T(ctx, "password_reset_token_invalid")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	backLink := &errorpages.BackLink{
		URL:  "/password/reset",
		Text: i18n.T(ctx, "password_reset_back_to_sign_in"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), errorpages.Error(ctx, i18n.T(ctx, "password_reset_token_invalid"), i18n.T(ctx, "password_reset_token_invalid_message"), backLink))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
