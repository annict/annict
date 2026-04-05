package password

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
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

	token := r.FormValue("token")

	// UseCaseを使ってバリデーション・パスワード更新を実行
	result, err := h.updatePasswordResetUC.Execute(ctx, usecase.UpdatePasswordResetInput{
		Token:                token,
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
	})
	if err != nil {
		// トークンが無効な場合
		if errors.Is(err, usecase.ErrInvalidPasswordResetToken) {
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

	// バリデーションエラーの場合
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		if err := h.sessionManager.SetFormErrors(ctx, w, r, *result.FormErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/password/edit?token="+token, http.StatusSeeOther)
		return
	}

	// 監査ログ
	slog.InfoContext(ctx, "パスワード更新が完了しました",
		"user_id", result.UserID,
		"ip_address", clientip.GetClientIP(r),
	)

	// フラッシュメッセージを設定
	h.sessionManager.SetFlash(w, session.FlashSuccess, i18n.T(ctx, "password_reset_success"))

	// ログインページにリダイレクト
	http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
}
