package sign_in_two_factor

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/usecase"
)

// Newは認証アプリのコードの入力画面を表示します (GET /sign_in/two_factor)
//
// ログイン途中の状態が無い・期限切れ・ユーザーが退会した・二要素認証を無効にしたときは、
// メールアドレスの入力からやり直させる。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setNoStore(w)
	backURL := r.URL.Query().Get("back")

	// すでにログイン済みの場合はホームにリダイレクト
	currentUser, err := h.sessionMgr.GetCurrentUser(ctx, r)
	if err != nil {
		slog.ErrorContext(ctx, "セッション取得エラー", "error", err)
	}
	if currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, ok, err := h.sessionMgr.SignInTwoFactorUserID(ctx, r, time.Now())
	if err != nil {
		slog.ErrorContext(ctx, "ログイン途中の状態の取得に失敗しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirectToSignIn(w, r, backURL)
		return
	}

	status, err := h.getSignInTwoFactorStatusUC.Execute(ctx, usecase.GetSignInTwoFactorStatusInput{UserID: userID})
	if err != nil {
		slog.ErrorContext(ctx, "ログイン途中の二要素認証の状態の取得に失敗しました", "error", err, "user_id", userID)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}
	if !status.Available {
		if err := h.sessionMgr.ClearSignInTwoFactor(ctx, r); err != nil {
			slog.WarnContext(ctx, "ログイン途中の状態の削除に失敗しました", "error", err)
		}
		h.flashMgr.SetError(w, i18n.T(ctx, "flash_sign_in_two_factor_restart"))
		redirectToSignIn(w, r, backURL)
		return
	}

	h.renderNew(w, r, http.StatusOK, nil, "", backURL)
}
