package sign_up_code

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Update PATCH /sign_up/code - 新規登録確認コード再送信処理
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_up_email", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_code_error_server"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.Warn("セッションにメールアドレスが存在しません")
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_code_error_session_expired"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック: メールアドレス単位（3 回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_up:send:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "確認コード再送信がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "rate_limit_exceeded"))
			http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
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
			if err := h.sessionMgr.SetValidationError(ctx, w, r, *ve); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
			return
		}
		slog.Error("確認コードの再送信に失敗しました", "email", email, "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_code_error_server"))
		http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
		return
	}

	slog.Info("確認コードを再送信しました", "email", email)

	// フラッシュメッセージを設定
	h.sessionMgr.SetFlash(w, session.FlashSuccess, i18n.T(ctx, "sign_up_code_resend_success"))

	// /sign_up/code にリダイレクト
	http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
}
