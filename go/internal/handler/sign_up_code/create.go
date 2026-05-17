package sign_up_code

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// codeErrors は usecase のコード検証で既知のエラー一覧。
// ユーザー向けメッセージはセキュリティのため全ケースで共通化している（情報漏洩対策）ため、
// メッセージキーを持たず単純なエラーのスライスとして定義する。
var codeErrors = []error{
	usecase.ErrCodeNotFound,
	usecase.ErrCodeInvalid,
	usecase.ErrCodeAttemptsExceeded,
}

// Create POST /sign_up/code - 新規登録確認コード検証処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータを取得
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値の取得エラー", "key", "sign_up_email", "error", err)
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_code_error_server"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}
	if email == "" {
		slog.WarnContext(ctx, "セッションにメールアドレスが存在しません")
		h.flashMgr.SetError(w, i18n.T(ctx, "sign_up_code_error_session_expired"))
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Rate Limiting チェック: コード検証（10 回/時間/IP）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ipKey := fmt.Sprintf("sign_up:verify:%s", clientip.GetClientIP(r))
		allowed, err := h.limiter.Check(ctx, ipKey, 10, 1*time.Hour)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "確認コード検証がRate Limitingにより制限されました",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
			)
			ve := model.NewValidationError()
			ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}
	}

	// ユースケース呼び出し: バリデーション + 確認コード検証
	_, err = h.verifySignUpCodeUC.Execute(ctx, usecase.VerifySignUpCodeInput{
		Email: email,
		Code:  r.FormValue("code"),
	})
	if err != nil {
		// バリデーションエラー
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewForm(w, r, http.StatusUnprocessableEntity, ve, email)
			return
		}

		// セキュリティのため、すべての既知エラーで同じメッセージを表示（情報漏洩対策）
		// errors.Is で排他的にマッチするため、順序は意味を持たない
		matched := false
		for _, ce := range codeErrors {
			if errors.Is(err, ce) {
				slog.WarnContext(ctx, "確認コード検証失敗",
					"reason", ce.Error(),
					"email", email,
					"ip_address", clientip.GetClientIP(r),
				)
				matched = true
				break
			}
		}
		if !matched {
			slog.ErrorContext(ctx, "コード検証エラー",
				"email", email,
				"ip_address", clientip.GetClientIP(r),
				"error", err,
			)
			http.Error(w, i18n.T(ctx, "sign_up_code_error_server"), http.StatusInternalServerError)
			return
		}

		// すべての既知エラーで同じメッセージを表示（情報漏洩対策）
		codeErr := model.NewValidationError()
		codeErr.AddField("code", i18n.T(ctx, "sign_up_code_error_invalid"))
		h.renderNewForm(w, r, http.StatusUnprocessableEntity, codeErr, email)
		return
	}

	// 一時トークンを生成してRedisに保存（次のステップで使用）
	token, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "一時トークンの生成に失敗しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_up_code_error_server"), http.StatusInternalServerError)
		return
	}

	// Redisに一時トークンを保存（値はメールアドレス、有効期限: 15分）
	if h.redisClient != nil {
		tokenKey := fmt.Sprintf("sign_up_token:%s", token)
		err = h.redisClient.Set(ctx, tokenKey, email, 15*time.Minute).Err()
		if err != nil {
			slog.ErrorContext(ctx, "一時トークンの保存に失敗しました", "error", err)
			http.Error(w, i18n.T(ctx, "sign_up_code_error_server"), http.StatusInternalServerError)
			return
		}
	} else {
		slog.WarnContext(ctx, "Redisクライアントが設定されていないため、一時トークンを保存できませんでした")
		// テスト環境ではRedisがない場合があるため、警告のみ出力して続行
	}

	slog.InfoContext(ctx, "確認コード検証に成功しました", "email", email)

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
