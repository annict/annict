package sign_in_code

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/redirect"
	"github.com/annict/annict/internal/session"
)

// Update PATCH /sign_in/code - 6桁コード再送信処理
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータをパース
	if err := r.ParseForm(); err != nil {
		slog.Error("フォームパースエラー", "error", err)
	}

	// backパラメータを取得（リダイレクト時に引き継ぐ）
	backURL := r.FormValue("back")

	// backパラメータ付きのリダイレクト先URLを構築するヘルパー関数
	buildRedirectURL := func(basePath string) string {
		if redirect.ValidateBackURL(backURL) {
			return basePath + "?back=" + url.QueryEscape(backURL)
		}
		return basePath
	}

	// すでにログイン済みの場合はホームにリダイレクト
	currentUser, err := h.sessionMgr.GetCurrentUser(ctx, r)
	if err != nil {
		slog.Error("セッション取得エラー", "error", err)
	}
	if currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスとユーザーIDを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_email", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.Warn("セッションにメールアドレスが存在しません")
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userIDStr, err := h.sessionMgr.GetValue(ctx, r, "sign_in_user_id")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_user_id", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if userIDStr == "" {
		slog.Warn("セッションにユーザーIDが存在しません")
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		slog.Error("ユーザーIDのパースエラー", "user_id_str", userIDStr, "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック（1 分間に 3 回まで）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_in:send:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Minute)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "6桁コード再送信がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "rate_limit_exceeded")); err != nil {
				slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, buildRedirectURL("/sign_in/code"), http.StatusSeeOther)
			return
		}
	}

	// ユースケース呼び出し: 6桁コードを再送信
	_, err = h.sendSignInCodeUC.Execute(ctx, userID)
	if err != nil {
		slog.Error("6桁コードの再送信に失敗しました", "user_id", userID, "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, buildRedirectURL("/sign_in/code"), http.StatusSeeOther)
		return
	}

	slog.Info("6桁コードを再送信しました", "user_id", userID, "email", email)

	// フラッシュメッセージを設定
	if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashSuccess, i18n.T(ctx, "sign_in_code_resend_success")); err != nil {
		slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
	}

	// /sign_in/code にリダイレクト（backパラメータを引き継ぐ）
	http.Redirect(w, r, buildRedirectURL("/sign_in/code"), http.StatusSeeOther)
}
