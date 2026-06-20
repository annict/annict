// Package sign_up はsign_up機能を提供します
package sign_up

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Create はメールアドレス送信処理と確認コード送信を行います (POST /sign_up)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	// Turnstile検証
	turnstileToken := r.FormValue("cf-turnstile-response")
	isValid, err := h.turnstileClient.Verify(ctx, turnstileToken)
	if err != nil || !isValid {
		if err != nil {
			slog.ErrorContext(ctx, "Turnstile検証エラー",
				"error", err,
				"token", turnstileToken,
			)
		} else {
			slog.WarnContext(ctx, "Turnstile検証失敗",
				"token", turnstileToken,
			)
		}

		ve := model.NewValidationError()
		ve.AddGlobal(i18n.T(ctx, "turnstile_verification_failed"))
		h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
		return
	}
	slog.DebugContext(ctx, "Turnstile検証成功")

	// Rate Limiting チェック: IP単位（5回/時間）→ メールアドレス単位（3回/時間）
	if blocked := h.enforceSignUpRateLimit(ctx, w, r, "sign_up:ip:"+clientip.GetClientIP(r), 5, "IP単位", email); blocked {
		return
	}
	if blocked := h.enforceSignUpRateLimit(ctx, w, r, "sign_up:email:"+email, 3, "メールアドレス単位", email); blocked {
		return
	}

	// ユーザーのロケールを取得（デフォルトは日本語）
	locale := i18n.GetLocale(ctx)

	// UseCaseを呼び出し（バリデーション + メールアドレス重複チェック + 確認コード生成・送信）
	_, err = h.sendSignUpCodeUC.Execute(ctx, usecase.SendSignUpCodeInput{
		Email:  email,
		Locale: locale,
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}
		slog.ErrorContext(ctx, "新規登録確認コードの送信に失敗しました",
			"email", email,
			"error", err,
		)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "新規登録確認コードを送信しました",
		"email", email,
	)

	// セッションにメールアドレスを保存
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_up_email", email); err != nil {
		slog.ErrorContext(ctx, "セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	// フラッシュメッセージを設定
	message := i18n.T(ctx, "sign_up_code_sent_to", map[string]any{"Email": email})
	h.flashMgr.SetSuccess(w, message)

	// 確認コード入力画面にリダイレクト
	http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
}

// enforceSignUpRateLimit は Rate Limit に引っかかった場合に 422 でフォームを再描画し true を返します。
// チェックが無効化されている、またはチェック自体のエラー時はリクエストを通過させます（true は返しません）。
func (h *Handler) enforceSignUpRateLimit(ctx context.Context, w http.ResponseWriter, r *http.Request, key string, limit int, scopeLabel string, email string) bool {
	if h.limiter == nil || h.cfg.DisableRateLimit {
		return false
	}

	allowed, err := h.limiter.Check(ctx, key, limit, 1*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err, "scope", scopeLabel)
		return false
	}
	if allowed {
		return false
	}

	slog.WarnContext(ctx, "新規登録申請がRate Limitingにより制限されました",
		"scope", scopeLabel,
		"email", email,
		"ip_address", clientip.GetClientIP(r),
	)
	ve := model.NewValidationError()
	ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
	h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
	return true
}
