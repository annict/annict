package sign_in_two_factor_recovery

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/usecase"
)

// Createはリカバリーコードを消費し、セッションを発行します (POST /sign_in/two_factor/recovery)
//
// 照合する相手は、第1の要素を確かめたときに書いたセッションの値から読む。フォームの値からは読まない。
// 受け付けなかった送信は、入力したコードを戻して画面を422で描き直す。
// 回数制限は認証アプリのコードの画面と同じカウンターで数え、2つの画面を行き来しても回数が分かれないようにする。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	setNoStore(w)

	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	backURL := r.FormValue("back")
	code := r.FormValue("code")

	userID, ok, err := h.sessionMgr.SignInTwoFactorUserID(ctx, r, time.Now())
	if err != nil {
		slog.ErrorContext(ctx, "ログイン途中の状態の取得に失敗しました", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}
	if !ok {
		h.flashMgr.SetError(w, i18n.T(ctx, "flash_sign_in_two_factor_restart"))
		redirectToSignIn(w, r, backURL)
		return
	}

	if h.isRateLimited(r, userID) {
		ve := model.NewValidationError()
		ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
		h.renderNew(w, r, http.StatusUnprocessableEntity, ve, code, backURL)
		return
	}

	output, err := h.createSignInTwoFactorRecoveryUC.Execute(ctx, usecase.CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: code})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNew(w, r, http.StatusUnprocessableEntity, ve, code, backURL)
			return
		}
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeConflict {
			slog.WarnContext(ctx, "ログイン途中のユーザーが二要素認証を使えなくなったため、ログインをやり直させます", "user_id", userID)
			h.restartSignIn(w, r, ae.UserMsg, backURL)
			return
		}
		slog.ErrorContext(ctx, "リカバリーコードでのログインに失敗しました", "error", err, "user_id", userID)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// ログイン途中の状態は、リクエストのCookieが指す (ログイン前の) セッションに残っているため、新しいCookieを設定する前に消す
	if err := h.sessionMgr.ClearSignInTwoFactor(ctx, r); err != nil {
		slog.WarnContext(ctx, "ログイン途中の状態の削除に失敗しました", "error", err)
	}
	h.sessionMgr.SetSessionCookieByPublicID(w, r, output.PublicID)
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_in_success"))

	slog.InfoContext(ctx, "ログイン成功 (リカバリーコード)", "user_id", output.UserID, "username", output.Username)

	http.Redirect(w, r, redirect.GetSafeRedirectURL(backURL), http.StatusSeeOther)
}

// isRateLimitedは、ログインの二要素認証の試行をIPアドレスとユーザーの単位で数え、上限を超えたときにtrueを返します。
// 制限が無効化されている場合や、判定自体が失敗した場合はリクエストを通します (ほかのログインの回数制限と同じ)。
func (h *Handler) isRateLimited(r *http.Request, userID model.UserID) bool {
	if h.limiter == nil || h.cfg.DisableRateLimit {
		return false
	}

	ctx := r.Context()
	ipAddress := clientip.GetClientIP(r)
	result, err := h.limiter.CheckSignInTwoFactor(ctx, ipAddress, userID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Rate Limitingチェックが失敗しました", "error", err)
		return false
	}
	if result.Allowed {
		return false
	}

	slog.WarnContext(ctx, "リカバリーコードでのログインがRate Limitingにより制限されました",
		"scope", result.Scope,
		"user_id", userID,
		"ip_address", ipAddress,
	)
	return true
}
