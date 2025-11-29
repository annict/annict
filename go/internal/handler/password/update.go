package password

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// Update はパスワードを更新します (PATCH /password)
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// リクエストDTOを作成
	req := &Request{
		Token:                r.FormValue("token"),
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
	}

	// フォームバリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		flashManager := session.NewFlashManager(h.sessionManager)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/password/edit?token="+req.Token, http.StatusSeeOther)
		return
	}

	// パスワード強度チェック
	if err := auth.ValidatePasswordStrength(ctx, req.Password); err != nil {
		flashManager := session.NewFlashManager(h.sessionManager)
		formErrors := &session.FormErrors{}
		formErrors.AddFieldError("password", err.Error())
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/password/edit?token="+req.Token, http.StatusSeeOther)
		return
	}

	// UseCaseを使ってパスワードを更新
	result, err := h.updatePasswordUseCase.Execute(ctx, req.Token, req.Password)
	if err != nil {
		// トークンが無効な場合
		if err.Error() == "invalid token" {
			slog.WarnContext(ctx, "パスワード更新時に無効なリセットトークン",
				"ip_address", clientip.GetClientIP(r),
			)
			h.renderInvalidTokenError(w, r)
			return
		}

		// その他のエラー
		slog.ErrorContext(ctx, "パスワード更新エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 監査ログ
	slog.InfoContext(ctx, "パスワード更新が完了しました",
		"user_id", result.UserID,
		"ip_address", clientip.GetClientIP(r),
	)

	// フラッシュメッセージを設定
	flashManager := session.NewFlashManager(h.sessionManager)
	if err := flashManager.SetFlash(w, r, session.FlashSuccess, i18n.T(ctx, "password_reset_success")); err != nil {
		slog.ErrorContext(ctx, "フラッシュメッセージの設定エラー", "error", err)
		// エラーでも続行（ユーザー体験を損ねない）
	}

	// ログインページにリダイレクト
	http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
}
