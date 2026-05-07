package sign_up_code

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Update PATCH /sign_up/code - 新規登録確認コード再送信処理
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値の取得エラー", "key", "sign_up_email", "error", err)
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_code_error_server"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.WarnContext(ctx, "セッションにメールアドレスが存在しません")
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_code_error_session_expired"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック: メールアドレス単位（3 回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_up:send:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Hour)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "確認コード再送信がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}
	}

	// ロケールを取得（セッションまたはデフォルト）
	locale := i18n.GetLocale(ctx)

	// ユースケース呼び出し: 確認コードを再送信
	_, err = h.sendSignUpCodeUC.Execute(ctx, usecase.SendSignUpCodeInput{
		Email:  email,
		Locale: locale,
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			slog.WarnContext(ctx, "確認コード再送信のバリデーションエラー", "email", email)
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}
		slog.ErrorContext(ctx, "確認コードの再送信に失敗しました", "email", email, "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_code_error_server"), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "確認コードを再送信しました", "email", email)

	// フラッシュメッセージを設定
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_up_code_resend_success"))

	// /sign_up/code にリダイレクト
	http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
}
