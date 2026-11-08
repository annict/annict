package sign_in_two_factor

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

// Createは認証アプリのコードを照合し、合っていればセッションを発行します (POST /sign_in/two_factor)
//
// 照合する相手は、第1の要素を確かめたときに書いたセッションの値から読む。フォームの値からは読まない。
// 受け付けなかった送信は、入力したコードを戻して画面を422で描き直す。
// 第1の要素を確かめたあとに退会した・二要素認証を無効にしたときは、メールアドレスの入力からやり直させる。
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

	output, err := h.createSignInTwoFactorUC.Execute(ctx, usecase.CreateSignInTwoFactorInput{UserID: userID, Code: code})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNew(w, r, http.StatusUnprocessableEntity, ve, code, backURL)
			return
		}
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeConflict {
			slog.WarnContext(ctx, "ログイン途中のユーザーが二要素認証を使えなくなったため、ログインをやり直させます", "user_id", userID)
			if err := h.sessionMgr.ClearSignInTwoFactor(ctx, r); err != nil {
				slog.WarnContext(ctx, "ログイン途中の状態の削除に失敗しました", "error", err)
			}
			h.flashMgr.SetError(w, ae.UserMsg)
			redirectToSignIn(w, r, backURL)
			return
		}
		slog.ErrorContext(ctx, "認証アプリのコードの照合に失敗しました", "error", err, "user_id", userID)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// トランザクション不要なのでnilを渡す
	sessionResult, err := h.createSessionUC.Execute(ctx, nil, output.UserID, output.EncryptedPassword)
	if err != nil {
		// コードは合ったが、セッションを作れなかった。黙って未ログインのままにせず500にする。
		slog.ErrorContext(ctx, "セッション作成エラー", "error", err, "user_id", userID)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	// ログイン途中の状態は、リクエストのCookieが指す (ログイン前の) セッションに残っているため、新しいCookieを設定する前に消す
	if err := h.sessionMgr.ClearSignInTwoFactor(ctx, r); err != nil {
		slog.WarnContext(ctx, "ログイン途中の状態の削除に失敗しました", "error", err)
	}
	h.sessionMgr.SetSessionCookieByPublicID(w, r, sessionResult.PublicID)
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "sign_in_success"))

	slog.InfoContext(ctx, "ログイン成功 (二要素認証)", "user_id", output.UserID, "username", output.Username)

	// redirect.GetSafeRedirectURLで戻り先を同じサイトのパスに限っているため、gosec G710は誤検知として抑制する。
	//nolint:gosec // G710
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

	slog.WarnContext(ctx, "ログインの二要素認証がRate Limitingにより制限されました",
		"scope", result.Scope,
		"user_id", userID,
		"ip_address", ipAddress,
	)
	return true
}
