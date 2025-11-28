package sign_up_code

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
)

// Create POST /sign_up/code - 新規登録確認コード検証処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.Error("フォームパースエラー", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_parse_form")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
		return
	}

	// リクエストを構築
	req := &CreateRequest{
		Code: r.FormValue("code"),
	}

	// バリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, *formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.Error("セッション値の取得エラー", "key", "sign_up_email", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.Warn("セッションにメールアドレスが存在しません")
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_session_expired")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック: コード検証（10 回/時間/IP）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ipKey := fmt.Sprintf("sign_up:verify:%s", clientip.GetClientIP(r))
		allowed, err := h.limiter.Check(ctx, ipKey, 10, 1*time.Hour)
		if err != nil {
			slog.Error("Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "確認コード検証がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "rate_limit_exceeded")); err != nil {
				slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
			return
		}
	}

	// ユースケース呼び出し: 確認コード検証
	err = h.verifySignUpCodeUC.Execute(ctx, email, req.Code)
	if err != nil {
		// セキュリティのため、すべてのエラーで同じメッセージを表示（情報漏洩対策）
		if errors.Is(err, usecase.ErrCodeNotFound) {
			slog.WarnContext(ctx, "確認コード検証失敗: コードが見つからないか、有効期限が切れています",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
		} else if errors.Is(err, usecase.ErrCodeInvalid) {
			slog.WarnContext(ctx, "確認コード検証失敗: コードが正しくありません",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
		} else if errors.Is(err, usecase.ErrCodeAttemptsExceeded) {
			slog.WarnContext(ctx, "確認コード検証失敗: 試行回数が上限に達しました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
		} else {
			slog.ErrorContext(ctx, "コード検証エラー",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
				"error", err,
			)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_server")); err != nil {
				slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
			return
		}

		// すべてのコード検証失敗で同じメッセージを表示
		formErrors := session.FormErrors{}
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_up_code_error_invalid"))
		if err := h.sessionMgr.SetFormErrors(ctx, w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
		return
	}

	// 一時トークンを生成してRedisに保存（次のステップで使用）
	token, err := generateToken()
	if err != nil {
		slog.Error("一時トークンの生成に失敗しました", "error", err)
		if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_server")); err != nil {
			slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
		return
	}

	// Redisに一時トークンを保存（値はメールアドレス、有効期限: 15分）
	if h.redisClient != nil {
		tokenKey := fmt.Sprintf("sign_up_token:%s", token)
		err = h.redisClient.Set(ctx, tokenKey, email, 15*time.Minute).Err()
		if err != nil {
			slog.Error("一時トークンの保存に失敗しました", "error", err)
			if err := h.sessionMgr.SetFlash(ctx, w, r, session.FlashError, i18n.T(ctx, "sign_up_code_error_server")); err != nil {
				slog.ErrorContext(ctx, "フラッシュメッセージの設定に失敗しました", "error", err)
			}
			http.Redirect(w, r, "/sign_up/code", http.StatusSeeOther)
			return
		}
	} else {
		slog.Warn("Redisクライアントが設定されていないため、一時トークンを保存できませんでした")
		// テスト環境ではRedisがない場合があるため、警告のみ出力して続行
	}

	slog.Info("確認コード検証に成功しました", "email", email)

	// ユーザー名設定画面にリダイレクト
	http.Redirect(w, r, fmt.Sprintf("/sign_up/username?token=%s", token), http.StatusSeeOther)
}

// generateToken は32バイトのランダムな一時トークンを生成します
func generateToken() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
