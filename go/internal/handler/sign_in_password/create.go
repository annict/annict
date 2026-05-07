package sign_in_password

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/usecase"
)

// Create POST /sign_in/password - パスワードログイン処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッションからメールアドレスの取得に失敗しました", "error", err)
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// メールアドレスがない場合は /sign_in にリダイレクト
	if email == "" {
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	backURL := r.FormValue("back")

	// UseCase を実行
	output, err := h.authenticateByPasswordUC.Execute(ctx, usecase.AuthenticateByPasswordInput{
		Email:    email,
		Password: r.FormValue("password"),
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email, backURL)
			return
		}
		slog.ErrorContext(ctx, "パスワード認証に失敗しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// ログイン処理前にセッションからサインイン用の一時データを削除
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_email"); err != nil {
		slog.WarnContext(ctx, "セッションからsign_in_emailの削除に失敗しました", "error", err)
	}
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_user_id"); err != nil {
		slog.WarnContext(ctx, "セッションからsign_in_user_idの削除に失敗しました", "error", err)
	}

	// Cookieを設定
	h.sessionMgr.SetSessionCookieByPublicID(w, r, output.PublicID)

	// ログイン成功のフラッシュメッセージを設定
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_in_success"))

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
