package sign_in_password

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
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
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_error_parse_form"))
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// UseCase を実行
	result, err := h.authenticateByPasswordUC.Execute(ctx, usecase.AuthenticateByPasswordInput{
		Email:    email,
		Password: r.FormValue("password"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "パスワード認証に失敗しました", "error", err)
		h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_error_server"))
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// バリデーションエラーがある場合はフォームを再表示
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *result.FormErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
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
	h.sessionMgr.SetSessionCookieByPublicID(w, r, result.PublicID)

	// ログイン成功のフラッシュメッセージを設定
	h.sessionMgr.SetFlash(w, session.FlashSuccess, i18n.T(ctx, "sign_in_success"))

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	backURL := r.FormValue("back")
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
