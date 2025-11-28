package password

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/password_reset"
	"github.com/annict/annict/internal/templates/layouts"
	errorpages "github.com/annict/annict/internal/templates/pages/errors"
	passwordpages "github.com/annict/annict/internal/templates/pages/password"
	"github.com/annict/annict/internal/viewmodel"
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
			// エラーでも続行
		} else if !allowed {
			slog.WarnContext(ctx, "トークン検証がRate Limitingにより制限されました",
				"ip_address", ip,
			)
			http.Error(w, i18n.T(ctx, "rate_limit_exceeded"), http.StatusTooManyRequests)
			return
		}
	}

	// トークンをハッシュ化してデータベースから検索
	tokenDigest := password_reset.HashToken(token)
	resetToken, err := h.passwordResetTokenRepository.GetByDigest(ctx, tokenDigest)
	if err == sql.ErrNoRows {
		// トークンが見つからない（セキュアなメッセージ）
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "not_found",
			"ip_address", clientip.GetClientIP(r),
		)
		h.renderInvalidTokenError(w, r)
		return
	} else if err != nil {
		slog.ErrorContext(ctx, "パスワードリセットトークンのクエリエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// トークンの有効性をチェック
	if resetToken.UsedAt.Valid {
		// 使用済み
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "used",
			"ip_address", clientip.GetClientIP(r),
		)
		h.renderInvalidTokenError(w, r)
		return
	}

	if time.Now().After(resetToken.ExpiresAt) {
		// 有効期限切れ
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "expired",
			"ip_address", clientip.GetClientIP(r),
		)
		h.renderInvalidTokenError(w, r)
		return
	}

	// セッションからフラッシュメッセージとフォームエラーを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)
	formErrors, _ := h.sessionManager.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_edit_title")
	meta.OGURL = h.cfg.AppURL() + "/password/edit"

	// CSRFトークンを取得
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), passwordpages.Edit(ctx, formErrors, csrfToken, token))
	if err = component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// renderInvalidTokenError は無効なトークンエラーを表示します
func (h *Handler) renderInvalidTokenError(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.Title = i18n.T(ctx, "password_reset_token_invalid")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	// バックリンクを作成
	backLink := &errorpages.BackLink{
		URL:  "/password/reset",
		Text: i18n.T(ctx, "password_reset_back_to_sign_in"),
	}

	// テンプレートをレンダリング
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, nil, h.cfg.GetAssetVersion(), errorpages.Error(ctx, i18n.T(ctx, "password_reset_token_invalid"), i18n.T(ctx, "password_reset_token_invalid_message"), backLink))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
	}
}
