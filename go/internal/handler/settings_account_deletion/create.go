package settings_account_deletion

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// settingsAccountDeletionCompletedPathは退会の完了ページのパス。
const settingsAccountDeletionCompletedPath = "/settings/account/deletion/completed"

// Createは退会を受け付けます (POST /settings/account/deletion)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		slog.WarnContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(ctx)
	inputUsername := r.FormValue("username")
	err := h.createAccountDeletionUC.Execute(ctx, usecase.CreateAccountDeletionInput{
		User:     user,
		Username: inputUsername,
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNew(w, r, http.StatusUnprocessableEntity, ve, inputUsername)
			return
		}

		slog.ErrorContext(ctx, "退会の受付に失敗", "error", err, "user_id", user.ID)
		httperror.InternalServerError(w, r)
		return
	}

	// セッションの破棄に失敗しても、論理削除したユーザーのセッションは他の端末と同じく
	// 未ログインとして扱われるため、完了ページへ進める。
	if err := h.sessionManager.DestroySession(ctx, w, r); err != nil {
		slog.ErrorContext(ctx, "退会後のセッションの破棄に失敗", "error", err, "user_id", user.ID)
	}

	http.Redirect(w, r, settingsAccountDeletionCompletedPath, http.StatusSeeOther)
}
