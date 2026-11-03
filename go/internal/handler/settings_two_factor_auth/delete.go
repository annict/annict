package settings_two_factor_auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/ratelimit"
	page "github.com/annict/annict/go/internal/templates/pages/settings_two_factor_auth"
	"github.com/annict/annict/go/internal/usecase"
)

const (
	// disableRateLimitは無効化の本人確認を、1人のユーザーが窓の間に試せる回数。
	// パスワードと6桁のコードの総当たりを防ぐため、本人確認が通った試行も含めて数える。
	disableRateLimit = 5
	// disableRateLimitWindowは無効化の本人確認の回数を数える窓の長さ。
	disableRateLimitWindow = 15 * time.Minute
)

// Deleteは本人確認をして二要素認証を無効にします (DELETE /settings/two_factor_auth)
//
// 本人確認を受け付けなかったときは、設定画面を422で描き直す。
// 二要素認証が有効でなくなっていたとき (二重送信など) は、状態が変わったことを知らせて設定画面へ送る。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)
	setNoStore(w)

	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if h.isRateLimited(r, ratelimit.CheckInput{
		Key:    "settings_two_factor_auth:disable:user:" + user.ID.String(),
		Limit:  disableRateLimit,
		Window: disableRateLimitWindow,
	}, user.ID) {
		ve := model.NewValidationError()
		ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
		h.renderShow(w, r, http.StatusUnprocessableEntity, ve)
		return
	}

	output, err := h.disableTwoFactorAuthUC.Execute(ctx, usecase.DisableTwoFactorAuthInput{
		UserID:     user.ID,
		Credential: r.PostFormValue("credential"),
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderShow(w, r, http.StatusUnprocessableEntity, ve)
			return
		}
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeConflict {
			h.flashMgr.SetWarning(w, ae.UserMsg)
			http.Redirect(w, r, page.ShowPath, http.StatusSeeOther)
			return
		}
		slog.ErrorContext(ctx, "二要素認証の無効化に失敗", "error", err, "user_id", user.ID)
		httperror.InternalServerError(w, r)
		return
	}

	slog.InfoContext(ctx, "二要素認証を無効にしました", "user_id", user.ID, "method", output.Method)
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_two_factor_auth_disabled"))
	http.Redirect(w, r, page.ShowPath, http.StatusSeeOther)
}
