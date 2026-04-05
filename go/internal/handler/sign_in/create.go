package sign_in

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Create はメールアドレス送信処理とログイン方法自動判定を行います (POST /sign_in)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// backパラメータを取得（ログイン後のリダイレクト先）
	backURL := r.FormValue("back")

	// backパラメータ付きのリダイレクトURL生成ヘルパー
	signInRedirectURL := func() string {
		if backURL != "" {
			return "/sign_in?back=" + url.QueryEscape(backURL)
		}
		return "/sign_in"
	}

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

		formErrors := &session.FormErrors{}
		formErrors.AddFieldError("email", i18n.T(ctx, "turnstile_verification_failed"))
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, signInRedirectURL(), http.StatusSeeOther)
		return
	}
	slog.InfoContext(ctx, "Turnstile検証成功")

	// UseCaseを呼び出し（バリデーション + ユーザー検索 + コード送信）
	result, err := h.sendSignInCodeUC.Execute(ctx, usecase.SendSignInCodeInput{
		Email: r.FormValue("email"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "サインイン処理でエラーが発生しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *result.FormErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, signInRedirectURL(), http.StatusSeeOther)
		return
	}

	// セッションにメールアドレスとユーザーIDを保存
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_email", result.Email); err != nil {
		slog.Error("セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	userIDStr := fmt.Sprintf("%d", result.UserID)
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_user_id", userIDStr); err != nil {
		slog.Error("セッションへのユーザーID保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	if result.HasPassword {
		// パスワードが存在する場合 → パスワードログイン画面へリダイレクト
		redirectURL := "/sign_in/password"
		if backURL != "" {
			redirectURL = "/sign_in/password?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		// パスワードが存在しない場合 → 6桁コード入力画面へリダイレクト
		slog.InfoContext(ctx, "6桁コードを送信しました",
			"user_id", result.UserID,
			"email", result.Email,
		)

		// フラッシュメッセージを設定
		message := i18n.T(ctx, "sign_in_code_sent_to", map[string]any{"Email": result.Email})
		h.sessionMgr.SetFlash(w, session.FlashSuccess, message)

		redirectURL := "/sign_in/code"
		if backURL != "" {
			redirectURL = "/sign_in/code?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}
