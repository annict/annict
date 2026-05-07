package sign_in

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
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

	email := r.FormValue("email")
	backURL := r.FormValue("back")

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
		h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
		return
	}
	slog.DebugContext(ctx, "Turnstile検証成功")

	// UseCaseを呼び出し（バリデーション + ユーザー検索 + コード送信）
	output, err := h.sendSignInCodeUC.Execute(ctx, usecase.SendSignInCodeInput{
		Email: email,
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}
		slog.ErrorContext(ctx, "サインイン処理でエラーが発生しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// セッションにメールアドレスとユーザーIDを保存
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_email", output.Email); err != nil {
		slog.ErrorContext(ctx, "セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	userIDStr := output.UserID.String()
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_user_id", userIDStr); err != nil {
		slog.ErrorContext(ctx, "セッションへのユーザーID保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	if output.HasPassword {
		// パスワードが存在する場合 → パスワードログイン画面へリダイレクト
		redirectURL := "/sign_in/password"
		if backURL != "" {
			redirectURL = "/sign_in/password?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		// パスワードが存在しない場合 → 6桁コード入力画面へリダイレクト
		slog.InfoContext(ctx, "6桁コードを送信しました",
			"user_id", output.UserID,
			"email", output.Email,
		)

		// フラッシュメッセージを設定
		message := i18n.T(ctx, "sign_in_code_sent_to", map[string]any{"Email": output.Email})
		h.flashMgr.SetSuccess(w, message)

		redirectURL := "/sign_in/code"
		if backURL != "" {
			redirectURL = "/sign_in/code?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}
