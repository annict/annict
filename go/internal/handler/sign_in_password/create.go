package sign_in_password

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/redirect"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/validator"
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
		slog.Error("フォームパースエラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_error_parse_form")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// リクエストを構築
	req := &Request{
		Password: r.FormValue("password"),
	}

	// バリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *formErrors); err != nil {
			slog.Error("フォームエラーの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// メールアドレスでユーザーを取得
	user, err := h.userRepo.GetByEmailOrUsername(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Info("ユーザーが見つかりません", "email", email)
			formErrors := session.FormErrors{}
			formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
			return
		}
		slog.Error("ユーザー取得エラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_error_server")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// NOT NULL制約があるフィールドの検証（fail-fast）
	// Note: user.CreatedAtはtime.Time型（NOT NULL）のため検証不要
	if err := validator.ValidateNotNullTime(user.UpdatedAt, "updated_at", user.ID); err != nil {
		slog.ErrorContext(ctx, "データベース制約違反を検出しました（NOT NULL制約）",
			"table", "users",
			"field", "updated_at",
			"user_id", user.ID,
			"error", err,
		)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// パスワードを検証
	if err := auth.CheckPassword(user.EncryptedPassword, req.Password); err != nil {
		slog.Info("パスワードが一致しません", "email", email)
		formErrors := session.FormErrors{}
		formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
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

	// flashメッセージをJSON形式で作成（Railsと互換性のある形式）
	flashMsg := session.Flash{
		Type:    session.FlashSuccess,
		Message: i18n.T(ctx, "sign_in_success"),
	}
	flashJSON, err := flashMsg.ToJSON()
	if err != nil {
		slog.Error("flashメッセージのエンコードエラー", "error", err)
		flashJSON = "" // エラー時は空文字列
	}

	// セッションを作成（usecase層、flashメッセージを含める）
	// encryptedPasswordを渡してauthenticatable_saltを生成させる
	// トランザクション不要なのでnilを渡す
	sessionResult, err := h.createSessionUC.Execute(ctx, nil, user.ID, user.EncryptedPassword, flashJSON)
	if err != nil {
		slog.Error("セッション作成エラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_error_create_session")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
		return
	}

	// Cookieを設定
	h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)

	slog.Info("ログイン成功", "user_id", user.ID, "username", user.Username)

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	backURL := r.FormValue("back")
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
