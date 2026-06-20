package sign_in_code

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/usecase"
)

// Update PATCH /sign_in/code - 6桁コード再送信処理
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
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
		slog.ErrorContext(ctx, "セッション取得エラー", "error", err)
	}
	if currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスとユーザーIDを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値の取得エラー", "key", "sign_in_email", "error", err)
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.WarnContext(ctx, "セッションにメールアドレスが存在しません")
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_in_code_error_session_expired"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userIDStr, err := h.sessionMgr.GetValue(ctx, r, "sign_in_user_id")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値の取得エラー", "key", "sign_in_user_id", "error", err)
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if userIDStr == "" {
		slog.WarnContext(ctx, "セッションにユーザーIDが存在しません")
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_in_code_error_session_expired"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "ユーザーIDのパースエラー", "user_id_str", userIDStr, "error", err)
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	userID := model.UserID(userIDInt)

	// Rate Limiting チェック（1 分間に 3 回まで）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_in:send:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Minute)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "6桁コード再送信がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
			h.renderShowForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}
	}

	// ユースケース呼び出し: 6桁コードを再送信
	_, err = h.sendSignInCodeUC.Execute(ctx, usecase.SendSignInCodeInput{Email: email})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderShowForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}
		slog.ErrorContext(ctx, "6桁コードの再送信に失敗しました", "user_id", userID, "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_code_error_server"), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "6桁コードを再送信しました", "user_id", userID, "email", email)

	// フラッシュメッセージを設定
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_in_code_resend_success"))

	// /sign_in/code にリダイレクト（backパラメータを引き継ぐ）
	http.Redirect(w, r, buildRedirectURL("/sign_in/code"), http.StatusSeeOther)
}
