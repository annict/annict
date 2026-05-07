package password

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Update はパスワードを更新します (PATCH /password)
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")

	output, err := h.updatePasswordResetUC.Execute(ctx, usecase.UpdatePasswordResetInput{
		Token:                token,
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
	})
	if err != nil {
		// バリデーションエラー → フォーム再描画 (422)
		if ve := model.AsValidationError(err); ve != nil {
			h.renderEditForm(w, r, http.StatusUnprocessableEntity, ve, token)
			return
		}

		// トークンが無効な場合 → 専用エラーページ
		if errors.Is(err, usecase.ErrInvalidPasswordResetToken) {
			slog.WarnContext(ctx, "パスワード更新時に無効なリセットトークン",
				"ip_address", clientip.GetClientIP(r),
			)
			h.renderInvalidTokenError(w, r)
			return
		}

		slog.ErrorContext(ctx, "パスワード更新エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "パスワード更新が完了しました",
		"user_id", output.UserID,
		"ip_address", clientip.GetClientIP(r),
	)

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "password_reset_success"))

	http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
}
