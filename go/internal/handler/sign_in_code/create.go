package sign_in_code

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/redirect"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
)

// Create POST /sign_in/code - 6桁コード検証処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.Error("フォームパースエラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_parse_form")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// リクエストを構築
	req := &CreateRequest{
		Code: r.FormValue("code"),
	}

	// バリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *formErrors); err != nil {
			slog.Error("フォームエラーの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスとユーザーIDを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_email", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.Warn("セッションにメールアドレスが存在しません")
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userIDStr, err := h.sessionMgr.GetValue(ctx, r, "sign_in_user_id")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_in_user_id", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}
	if userIDStr == "" {
		slog.Warn("セッションにユーザーIDが存在しません")
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_session_expired")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		slog.Error("ユーザーIDのパースエラー", "user_id_str", userIDStr, "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック（1 分間に 5 回まで）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_in:verify:%s", email)
		allowed, err := h.limiter.Check(ctx, emailKey, 5, 1*time.Minute)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "6桁コード検証がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "rate_limit_exceeded")); err != nil {
				slog.Error("フラッシュメッセージの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		}
	}

	// ユースケース呼び出し: 6桁コード検証
	err = h.verifySignInCodeUC.Execute(ctx, userID, req.Code)
	if err != nil {
		// エラーの種類によって異なるメッセージを表示
		if errors.Is(err, usecase.ErrCodeNotFound) {
			slog.Warn("コードが見つからないか、有効期限が切れています", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_not_found"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else if errors.Is(err, usecase.ErrCodeInvalid) {
			slog.Warn("コードが正しくありません", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_invalid"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else if errors.Is(err, usecase.ErrCodeAttemptsExceeded) {
			slog.Warn("試行回数が上限に達しました", "email", email, "user_id", userID)
			formErrors := session.FormErrors{}
			formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_attempts_exceeded"))
			if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
				slog.Error("フォームエラーの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		} else {
			slog.Error("コード検証エラー", "error", err)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
				slog.Error("フラッシュメッセージの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
			return
		}
	}

	// ユーザー情報を取得（encrypted_passwordを取得するため）
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("ユーザーが見つかりません", "user_id", userID)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_user_not_found")); err != nil {
				slog.Error("フラッシュメッセージの設定に失敗", "error", err)
			}
			http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
			return
		}
		slog.Error("ユーザー取得エラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_server")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
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
	// トランザクション不要なのでnilを渡す
	sessionResult, err := h.createSessionUC.Execute(ctx, nil, userID, user.EncryptedPassword, flashJSON)
	if err != nil {
		slog.Error("セッション作成エラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_in_code_error_create_session")); err != nil {
			slog.Error("フラッシュメッセージの設定に失敗", "error", err)
		}
		http.Redirect(w, r, "/sign_in/code", http.StatusSeeOther)
		return
	}

	// Cookie設定（session.Managerに委譲）
	h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)

	// 一時セッション値を削除（sign_in_email と sign_in_user_id）
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_email"); err != nil {
		slog.Warn("一時セッション値の削除に失敗しました", "key", "sign_in_email", "error", err)
	}
	if err := h.sessionMgr.DeleteValue(ctx, r, "sign_in_user_id"); err != nil {
		slog.Warn("一時セッション値の削除に失敗しました", "key", "sign_in_user_id", "error", err)
	}

	slog.Info("ログイン成功（メールログイン）", "user_id", userID, "username", user.Username)

	// ログイン後のリダイレクト先を取得（バリデーション付き）
	backURL := r.FormValue("back")
	redirectTo := redirect.GetSafeRedirectURL(backURL)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
