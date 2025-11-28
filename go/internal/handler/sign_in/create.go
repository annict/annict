package sign_in

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
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

	// リクエストDTOを作成
	req := &CreateRequest{
		Email: r.FormValue("email"),
	}

	// backパラメータ付きのリダイレクトURL生成ヘルパー
	signInRedirectURL := func() string {
		if backURL != "" {
			return "/sign_in?back=" + url.QueryEscape(backURL)
		}
		return "/sign_in"
	}

	// フォームバリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		flashManager := session.NewFlashManager(h.sessionMgr)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, signInRedirectURL(), http.StatusSeeOther)
		return
	}

	// Turnstile検証
	turnstileToken := r.FormValue("cf-turnstile-response")
	isValid, err := h.turnstileClient.Verify(ctx, turnstileToken)
	if err != nil || !isValid {
		// ログ記録（エラーの詳細度に応じてログレベルを分ける）
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

		// エラーレスポンスを返す
		formErrors := &session.FormErrors{}
		formErrors.AddFieldError("email", i18n.T(ctx, "turnstile_verification_failed"))
		flashManager := session.NewFlashManager(h.sessionMgr)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, signInRedirectURL(), http.StatusSeeOther)
		return
	}
	slog.InfoContext(ctx, "Turnstile検証成功")

	// メールアドレスでユーザーを検索
	user, err := h.userRepo.GetByEmailForSignIn(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// ユーザーが見つからない場合はエラーを表示
			formErrors := &session.FormErrors{}
			formErrors.AddFieldError("email", i18n.T(ctx, "sign_in_user_not_found"))
			flashManager := session.NewFlashManager(h.sessionMgr)
			if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, signInRedirectURL(), http.StatusSeeOther)
			return
		}
		// その他のエラー
		slog.Error("ユーザーの検索エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// セッションにメールアドレスとユーザーIDを保存
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_email", user.Email); err != nil {
		slog.Error("セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// ユーザーIDは文字列に変換して保存
	userIDStr := fmt.Sprintf("%d", user.ID)
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_in_user_id", userIDStr); err != nil {
		slog.Error("セッションへのユーザーID保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// encrypted_passwordの有無を確認して自動的に分岐
	if user.EncryptedPassword != "" {
		// パスワードが存在する場合 → パスワードログイン画面へリダイレクト
		slog.InfoContext(ctx, "ユーザーはパスワードログインを使用します",
			"user_id", user.ID,
			"email", user.Email,
		)
		redirectURL := "/sign_in/password"
		if backURL != "" {
			redirectURL = "/sign_in/password?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		// パスワードが存在しない場合 → 6桁コードを送信 → 6桁コード入力画面へリダイレクト
		slog.InfoContext(ctx, "ユーザーはメールログインを使用します",
			"user_id", user.ID,
			"email", user.Email,
		)

		// 6桁コードを生成・送信
		_, err := h.sendSignInCodeUC.Execute(ctx, user.ID)
		if err != nil {
			slog.ErrorContext(ctx, "6桁コードの送信に失敗しました",
				"user_id", user.ID,
				"email", user.Email,
				"error", err,
			)
			http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "6桁コードを送信しました",
			"user_id", user.ID,
			"email", user.Email,
		)

		// フラッシュメッセージを設定
		flashManager := session.NewFlashManager(h.sessionMgr)
		message := i18n.T(ctx, "sign_in_code_sent_to", map[string]any{"Email": user.Email})
		if err := flashManager.SetFlash(w, r, session.FlashSuccess, message); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}

		redirectURL := "/sign_in/code"
		if backURL != "" {
			redirectURL = "/sign_in/code?back=" + url.QueryEscape(backURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}
