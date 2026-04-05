package sign_up_username

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Create POST /sign_up/username - ユーザー登録完了
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.Error("フォームパースエラー", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_username_error_parse_form"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	username := r.FormValue("username")

	// ロケールを取得（セッションまたはデフォルト）
	locale := i18n.GetLocale(ctx)

	// ユースケースを実行（バリデーション + ビジネスロジック）
	ucResult, err := h.completeSignUpUC.Execute(ctx, usecase.CompleteSignUpInput{
		Token:    token,
		Username: username,
		Locale:   locale,
	})
	if err != nil {
		// エラーの種類に応じてメッセージを切り替え
		formErrors := session.FormErrors{}
		if usecase.IsUsernameAlreadyExistsError(err) {
			formErrors.AddFieldError("username", i18n.T(ctx, "sign_up_username_error_username_taken"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, fmt.Sprintf("/sign_up/username?token=%s", token), http.StatusSeeOther)
			return
		} else if usecase.IsTokenInvalidError(err) {
			formErrors.AddFieldError("token", i18n.T(ctx, "sign_up_username_error_token_invalid"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}

		slog.ErrorContext(ctx, "ユーザー登録失敗",
			"username", username,
			"ip_address", clientip.GetClientIP(r),
			"error", err,
		)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_up_username_error_server"))
		http.Redirect(w, r, fmt.Sprintf("/sign_up/username?token=%s", token), http.StatusSeeOther)
		return
	}

	// バリデーションエラーの場合
	if ucResult.FormErrors != nil && ucResult.FormErrors.HasErrors() {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *ucResult.FormErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, fmt.Sprintf("/sign_up/username?token=%s", token), http.StatusSeeOther)
		return
	}

	// セッションCookieを設定（session.Managerに委譲）
	h.sessionMgr.SetSessionCookieByPublicID(w, r, ucResult.SessionPublicID)

	// セッションから sign_up_email を削除
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_up_email"); err != nil {
		slog.Warn("セッションから sign_up_email の削除に失敗しました（処理は続行）", "error", err)
	}

	slog.InfoContext(ctx, "ユーザー登録成功",
		"user_id", ucResult.User.ID,
		"username", ucResult.User.Username,
		"email", ucResult.User.Email,
		"ip_address", clientip.GetClientIP(r),
	)

	// 成功メッセージを設定してホームページにリダイレクト
	h.sessionMgr.SetFlash(w, session.FlashSuccess, i18n.T(ctx, "sign_up_username_success"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
