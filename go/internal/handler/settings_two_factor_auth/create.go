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
	// enableRateLimitは有効化のコードの照合を、1人のユーザーが窓の間に試せる回数。
	// 6桁のコードの総当たりを防ぐため、コードが合った試行も含めて数える。
	enableRateLimit = 5
	// enableRateLimitWindowは有効化のコードの照合の回数を数える窓の長さ。
	enableRateLimitWindow = 15 * time.Minute
)

// Createは認証アプリのコードを照合して二要素認証を有効にし、リカバリーコードを表示します (POST /settings/two_factor_auth)
//
// リカバリーコードは保存していないため、リダイレクトせずにこの応答として一度だけ描く。
// コードの誤りでは、同じ秘密鍵のまま登録画面を描き直す。認証アプリへ登録し直させないため。
// 登録の途中の設定が無いとき (二重送信で既に有効にした・照合の後に別の画面で秘密鍵が差し替わった) は、設定画面へ送る。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)
	setNoStore(w)

	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームパースエラー", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	code := r.PostFormValue("code")

	if h.isRateLimited(r, ratelimit.CheckInput{
		Key:    "settings_two_factor_auth:enable:user:" + user.ID.String(),
		Limit:  enableRateLimit,
		Window: enableRateLimitWindow,
	}, user.ID) {
		ve := model.NewValidationError()
		ve.AddGlobal(i18n.T(ctx, "rate_limit_exceeded"))
		h.rerenderNew(w, r, ve, code)
		return
	}

	output, err := h.enableTwoFactorAuthUC.Execute(ctx, usecase.EnableTwoFactorAuthInput{UserID: user.ID, Code: code})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.rerenderNew(w, r, ve, code)
			return
		}
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeConflict {
			h.flashMgr.SetWarning(w, ae.UserMsg)
			http.Redirect(w, r, page.ShowPath, http.StatusSeeOther)
			return
		}
		slog.ErrorContext(ctx, "二要素認証の有効化に失敗", "error", err, "user_id", user.ID)
		httperror.InternalServerError(w, r)
		return
	}

	slog.InfoContext(ctx, "二要素認証を有効にしました", "user_id", user.ID)
	h.renderPage(w, r, http.StatusOK, "settings_two_factor_auth_create_title", page.ShowPath, page.Create(page.CreatePageData{
		RecoveryCodes: output.RecoveryCodes,
	}))
}

// rerenderNewは、受け付けなかったコードを入力欄に戻し、登録の途中の秘密鍵のまま登録画面を422で描き直します。
// 登録の途中の設定が無くなっていたときは、設定画面へ送る。
func (h *Handler) rerenderNew(w http.ResponseWriter, r *http.Request, formErrors *model.ValidationError, inputCode string) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)

	pending, err := h.getPendingTwoFactorAuthUC.Execute(ctx, usecase.GetPendingTwoFactorAuthInput{User: user})
	if err != nil {
		slog.ErrorContext(ctx, "登録の途中の二要素認証の設定の取得に失敗", "error", err, "user_id", user.ID)
		httperror.InternalServerError(w, r)
		return
	}
	if pending.Setup == nil {
		http.Redirect(w, r, page.ShowPath, http.StatusSeeOther)
		return
	}

	h.renderNew(w, r, http.StatusUnprocessableEntity, pending.Setup, formErrors, inputCode)
}
