// Package sign_up はsign_up機能を提供します
package sign_up

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
		// ログ記録（エラーの詳細度に応じてログレベルを分ける）
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

		// エラーレスポンスを返す
		ve := model.NewValidationError()
		ve.AddField("email", i18n.T(ctx, "turnstile_verification_failed"))
		if err := h.sessionMgr.SetValidationError(ctx, w, r, *ve); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	slog.InfoContext(ctx, "Turnstile検証成功")

	// Rate Limiting チェック: IP単位（5回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ipKey := fmt.Sprintf("sign_up:ip:%s", clientip.GetClientIP(r))
		allowed, err := h.limiter.Check(ctx, ipKey, 5, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "新規登録申請がRate Limitingにより制限されました（IP単位）",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddField("email", i18n.T(ctx, "rate_limit_exceeded"))
			if err := h.sessionMgr.SetValidationError(ctx, w, r, *ve); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
	}

	// Rate Limiting チェック: メールアドレス単位（3回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_up:email:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "新規登録申請がRate Limitingにより制限されました(メールアドレス単位)",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddField("email", i18n.T(ctx, "rate_limit_exceeded"))
			if err := h.sessionMgr.SetValidationError(ctx, w, r, *ve); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
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
			if err := h.sessionMgr.SetValidationError(ctx, w, r, *ve); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
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
		slog.Error("セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	// フラッシュメッセージを設定
	message := i18n.T(ctx, "sign_up_code_sent_to", map[string]any{"Email": email})
	h.sessionMgr.SetFlash(w, session.FlashSuccess, message)

	// 確認コード入力画面にリダイレクト
	http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
}
