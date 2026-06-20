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
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/usecase"
)

// codeErrorMap は usecase のコード検証エラーと i18n メッセージキーの対応表。
// ログメッセージは err.Error() を使用するため重複を定義しない。
// errors.Is で排他的にマッチするため、順序は意味を持たない。
var codeErrorMap = []struct {
	usecaseErr error
	msgKey     string
}{
	{usecase.ErrCodeNotFound, "sign_in_code_error_code_not_found"},
	{usecase.ErrCodeInvalid, "sign_in_code_error_code_invalid"},
	{usecase.ErrCodeAttemptsExceeded, "sign_in_code_error_attempts_exceeded"},
}

// Create POST /sign_in/code - 6桁コード検証処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	backURL := r.FormValue("back")

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

	// Rate Limiting チェック（1 分間に 5 回まで）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_in:verify:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 5, 1*time.Minute)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "6桁コード検証がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
			h.renderShowForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}
	}

	// ユースケース呼び出し: バリデーション + 6桁コード検証 + ユーザー情報取得
	output, err := h.verifySignInCodeUC.Execute(ctx, usecase.VerifySignInCodeInput{
		UserID: userID,
		Code:   r.FormValue("code"),
	})
	if err != nil {
		// バリデーションエラー
		if ve := model.AsValidationError(err); ve != nil {
			h.renderShowForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}

		// エラーの種類によって異なるメッセージを表示
		for _, ce := range codeErrorMap {
			if errors.Is(err, ce.usecaseErr) {
				slog.WarnContext(ctx, "ログインコード検証失敗",
					"reason", ce.usecaseErr.Error(),
					"email", email,
					"user_id", userID,
				)
				codeErr := model.NewValidationError()
				codeErr.AddField("code", i18n.T(ctx, ce.msgKey))
				h.renderShowForm(w, r, http.StatusUnprocessableEntity, codeErr, email, backURL)
				return
			}
		}

		slog.ErrorContext(ctx, "コード検証エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_code_error_server"), http.StatusInternalServerError)
		return
	}

	// セッションを作成（usecase層）
	// トランザクション不要なのでnilを渡す
	sessionResult, err := h.createSessionUC.Execute(ctx, nil, userID, output.EncryptedPassword)
	if err != nil {
		slog.ErrorContext(ctx, "セッション作成エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_code_error_create_session"), http.StatusInternalServerError)
		return
	}

	// Cookie設定（session.Managerに委譲）
	h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)

	// ログイン成功のフラッシュメッセージを設定
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_in_success"))

	// 一時セッション値を削除（sign_in_email と sign_in_user_id）
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_email"); err != nil {
		slog.WarnContext(ctx, "一時セッション値の削除に失敗しました", "key", "sign_in_email", "error", err)
	}
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_user_id"); err != nil {
		slog.WarnContext(ctx, "一時セッション値の削除に失敗しました", "key", "sign_in_user_id", "error", err)
	}

	slog.InfoContext(ctx, "ログイン成功（メールログイン）", "user_id", userID, "username", output.Username)

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
