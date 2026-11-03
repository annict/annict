package settings_two_factor_auth

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/qrcode"
	page "github.com/annict/annict/go/internal/templates/pages/settings_two_factor_auth"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Newは認証アプリへ登録する新しい秘密鍵を作り、登録画面を表示します (GET /settings/two_factor_auth/new)
//
// 既に有効にしているときは、秘密鍵を差し替えずに設定画面へ送る。
// GETで行を書き換えることになるが、ログインしたあとのno-storeの画面でプリフェッチやクローラーが届かず、
// 書き換えるのは有効にする前の秘密鍵だけのため許容する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)
	setNoStore(w)

	output, err := h.prepareTwoFactorAuthUC.Execute(ctx, usecase.PrepareTwoFactorAuthInput{User: user})
	if err != nil {
		slog.ErrorContext(ctx, "二要素認証の登録の準備に失敗", "error", err, "user_id", user.ID)
		httperror.InternalServerError(w, r)
		return
	}
	if output.Setup == nil {
		http.Redirect(w, r, page.ShowPath, http.StatusSeeOther)
		return
	}

	h.renderNew(w, r, http.StatusOK, output.Setup, nil, "")
}

// renderNewは登録画面を指定したステータスで描画します。
// 表示 (200) と、コードを受け付けなかったときの再描画 (422) で共有する。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, status int, setup *usecase.TwoFactorAuthSetup, formErrors *model.ValidationError, inputCode string) {
	ctx := r.Context()

	qrCode, err := qrcode.Encode(setup.OTPAuthURL)
	if err != nil {
		slog.ErrorContext(ctx, "二要素認証の登録のQRコードの作成に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.renderPage(w, r, status, "settings_two_factor_auth_new_title", page.NewPath, page.New(page.NewPageData{
		CSRFToken:  middleware.GetCSRFToken(r, h.sessionManager),
		Secret:     setup.Secret,
		OTPAuthURL: setup.OTPAuthURL,
		QRCode:     qrCode,
		FormErrors: viewmodel.NewFormErrors(formErrors),
		InputCode:  inputCode,
	}))
}
