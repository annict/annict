package password_reset

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/internal/clientip"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/templates/layouts"
	passwordpages "github.com/annict/annict/internal/templates/pages/password"
	"github.com/annict/annict/internal/viewmodel"
)

// Create はパスワードリセット申請を処理します (POST /password/reset)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームをパース
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// リクエストDTOを作成
	req := &Request{
		Email: r.FormValue("email"),
	}

	// フォームバリデーション
	if formErrors := req.Validate(ctx); formErrors != nil {
		flashManager := session.NewFlashManager(h.sessionManager)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
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
		flashManager := session.NewFlashManager(h.sessionManager)
		if err := flashManager.SetFormErrors(w, r, formErrors); err != nil {
			slog.ErrorContext(ctx, "フォームエラーの設定に失敗しました", "error", err)
		}
		http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
		return
	}
	slog.InfoContext(ctx, "Turnstile検証成功")

	// Rate Limiting: IPアドレス単位の制限（5回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		ip := clientip.GetClientIP(r)
		ipKey := fmt.Sprintf("password_reset:ip:%s", ip)
		allowed, err := h.limiter.Check(ctx, ipKey, 5, 1*time.Hour)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "パスワードリセット申請がRate Limitingにより制限されました（IPアドレス単位）",
				"ip_address", ip,
			)
			http.Error(w, i18n.T(ctx, "rate_limit_exceeded"), http.StatusTooManyRequests)
			return
		}
	}

	// Rate Limiting: メールアドレス単位の制限（3回/時間）
	if h.limiter != nil && !h.cfg.DisableRateLimit {
		emailKey := fmt.Sprintf("password_reset:email:%s", req.Email)
		allowed, err := h.limiter.Check(ctx, emailKey, 3, 1*time.Hour)
		if err != nil {
			slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		} else if !allowed {
			slog.WarnContext(ctx, "パスワードリセット申請がRate Limitingにより制限されました（メールアドレス単位）",
				"email", req.Email,
				"ip_address", clientip.GetClientIP(r),
			)
			http.Error(w, i18n.T(ctx, "rate_limit_exceeded"), http.StatusTooManyRequests)
			return
		}
	}

	// ユーザーを検索（存在しない場合もエラーを返さない - セキュリティ対策）
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.ErrorContext(ctx, "ユーザーの検索エラー", "error", err)
	}

	// ユーザーが存在する場合のみトークンを生成
	if err == nil && user.ID > 0 {
		result, err := h.createTokenUseCase.Execute(ctx, user.ID)
		if err != nil {
			slog.ErrorContext(ctx, "パスワードリセットトークンの生成エラー", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "パスワードリセット申請を受け付けました",
			"user_id", result.UserID,
			"email", user.Email,
			"ip_address", clientip.GetClientIP(r),
		)
	}

	// 常に成功ページを表示（ユーザーの存在を明かさない）
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_reset_sent_title")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, nil, h.cfg.GetAssetVersion(), passwordpages.ResetSent(ctx))
	if err = component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
