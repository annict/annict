package sign_in_code

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Create POST /sign_in/code - 6桁コード検証処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.Error("フォームパースエラー", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_parse_form"))
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスとユーザーIDを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_email", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.Warn("セッションにメールアドレスが存在しません")
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userIDStr, err := h.sessionMgr.GetValue(ctx, r, "sign_in_user_id")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_user_id", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if userIDStr == "" {
		slog.Warn("セッションにユーザーIDが存在しません")
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		slog.Error("ユーザーIDのパースエラー", "user_id_str", userIDStr, "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_server"))
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック（1 分間に 5 回まで）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_in:verify:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 5, 1*time.Minute)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "6桁コード検証がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "rate_limit_exceeded"))
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		}
	}

	// ユースケース呼び出し: バリデーション + 6桁コード検証 + ユーザー情報取得
	result, err := h.verifySignInCodeUC.Execute(ctx, usecase.VerifySignInCodeInput{
		UserID: userID,
		Code:   r.FormValue("code"),
	})
	if err != nil {
		// エラーの種類によって異なるメッセージを表示
		if errors.Is(err, usecase.ErrCodeNotFound) {
			slog.Warn("コードが見つからないか、有効期限が切れています", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_not_found"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else if errors.Is(err, usecase.ErrCodeInvalid) {
			slog.Warn("コードが正しくありません", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_invalid"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else if errors.Is(err, usecase.ErrCodeAttemptsExceeded) {
			slog.Warn("試行回数が上限に達しました", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_attempts_exceeded"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else {
			slog.Error("コード検証エラー", "error", err)
			h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_server"))
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		}
	}

	// バリデーションエラーの場合
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *result.FormErrors); err != nil {
			slog.Error("フォームエラーの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// セッションを作成（usecase層）
	// トランザクション不要なのでnilを渡す
	sessionResult, err := h.createSessionUC.Execute(ctx, nil, userID, result.EncryptedPassword)
	if err != nil {
		slog.Error("セッション作成エラー", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_code_error_create_session"))
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// Cookie設定（session.Managerに委譲）
	h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)

	// ログイン成功のフラッシュメッセージを設定
	h.sessionMgr.SetFlash(w, session.FlashSuccess, i18n.T(ctx, "sign_in_success"))

	// 一時セッション値を削除（sign_in_email と sign_in_user_id）
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_email"); err != nil {
		slog.Warn("一時セッション値の削除に失敗しました", "key", "sign_in_email", "error", err)
	}
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_user_id"); err != nil {
		slog.Warn("一時セッション値の削除に失敗しました", "key", "sign_in_user_id", "error", err)
	}

	slog.Info("ログイン成功（メールログイン）", "user_id", userID, "username", result.Username)

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	backURL := r.FormValue("back")
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
