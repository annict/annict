// Package sign_up はsign_up機能を提供します
package sign_up

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// Create はメールアドレス送信処理と確認コード送信を行います (POST /sign_up)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// リクエストDTOを作成
	req := &CreateRequest{
		Email: r.FormValue("email"),
	}

	// フォームバリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		flashManager := session.NewFlashManager(h.sessionMgr)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
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
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	slog.InfoContext(ctx, "Turnstile検証成功")

	// Rate Limiting チェック: IP単位（5回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ipKey := fmt.Sprintf("sign_up:ip:%s", clientip.GetClientIP(r))
		allowed, err := h.limiter.Check(ctx, ipKey, 5, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "新規登録申請がRate Limitingにより制限されました（IP単位）",
				"email", req.Email,
				"ip_address", clientip.GetClientIP(r),
			)
			formErrors := &session.FormErrors{}
			formErrors.AddFieldError("email", i18n.T(ctx, "rate_limit_exceeded"))
			flashManager := session.NewFlashManager(h.sessionMgr)
			if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
	}

	// Rate Limiting チェック: メールアドレス単位（3回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("sign_up:email:%s", req.Email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "新規登録申請がRate Limitingにより制限されました（メールアドレス単位）",
				"email", req.Email,
				"ip_address", clientip.GetClientIP(r),
			)
			formErrors := &session.FormErrors{}
			formErrors.AddFieldError("email", i18n.T(ctx, "rate_limit_exceeded"))
			flashManager := session.NewFlashManager(h.sessionMgr)
			if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
				slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
			return
		}
	}

	// メールアドレスの重複チェック
	_, err = h.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		// ユーザーが存在する場合はエラーを表示
		formErrors := &session.FormErrors{}
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_already_exists"))
		flashManager := session.NewFlashManager(h.sessionMgr)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	if err != sql.ErrNoRows {
		// その他のエラー
		slog.Error("ユーザーの検索エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	// ユーザーのロケールを取得（デフォルトは日本語）
	locale := i18n.GetLocale(ctx)

	// 新規登録確認コードを生成・送信
	_, err = h.sendSignUpCodeUC.Execute(ctx, req.Email, locale)
	if err != nil {
		slog.ErrorContext(ctx, "新規登録確認コードの送信に失敗しました",
			"email", req.Email,
			"error", err,
		)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "新規登録確認コードを送信しました",
		"email", req.Email,
	)

	// セッションにメールアドレスを保存
	if err := h.sessionMgr.SetValue(ctx, w, r, "sign_up_email", req.Email); err != nil {
		slog.Error("セッションへのメールアドレス保存エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_error_server"), http.StatusInternalServerError)
		return
	}

	// フラッシュメッセージを設定
	flashManager := session.NewFlashManager(h.sessionMgr)
	message := i18n.T(ctx, "sign_up_code_sent_to", map[string]any{"Email": req.Email})
	if err := flashManager.SetFlash(w, r, session.FlashSuccess, message); err != nil {
		slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
	}

	// 確認コード入力画面にリダイレクト
	http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
}
