package password_reset

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	passwordpages "github.com/annict/annict/go/internal/templates/pages/password"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Create はパスワードリセット申請を処理します (POST /password/reset)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	slog.InfoContext(ctx, "Turnstile検証成功")

	// Rate Limiting: IPアドレス単位の制限（5回/時間）
	if h.handleRateLimit(w, r, email, fmt.Sprintf("password_reset:ip:%s", clientip.GetClientIP(r)), "ip", 5) {
		return
	}

	// Rate Limiting: メールアドレス単位の制限（3回/時間）
	if h.handleRateLimit(w, r, email, fmt.Sprintf("password_reset:email:%s", email), "email", 3) {
		return
	}

	// UseCaseを呼び出し（バリデーション + ユーザー検索 + トークン生成）
	output, err := h.createTokenUseCase.Execute(ctx, usecase.CreatePasswordResetTokenInput{
		Email: email,
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}
		slog.ErrorContext(ctx, "パスワードリセットトークンの生成エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if output != nil {
		slog.InfoContext(ctx, "パスワードリセット申請を受け付けました",
			"user_id", output.UserID,
			"ip_address", clientip.GetClientIP(r),
		)
	}

	// 常に成功ページを表示（ユーザーの存在を明かさない）
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_reset_sent_title")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), passwordpages.ResetSent(ctx))
	if err = component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleRateLimit は Rate Limit を判定し、超過時は 422 + フォーム再描画を行う。
// 戻り値が true の場合、呼び出し側は処理を中断する。
func (h *Handler) handleRateLimit(w http.ResponseWriter, r *http.Request, email, key, scope string, limit int) bool {
	if h.limiter == nil || h.cfg.DisableRateLimit {
		return false
	}

	ctx := r.Context()
	allowed, err := h.limiter.Check(ctx, key, limit, 1*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		return false
	}
	if allowed {
		return false
	}

	slog.WarnContext(ctx, "パスワードリセット申請がRate Limitingにより制限されました",
		"scope", scope,
		"ip_address", clientip.GetClientIP(r),
		"email", email,
	)
	ve := model.NewValidationError()
	ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
	h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
	return true
}
