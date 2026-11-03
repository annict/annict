// Package sign_in_two_factorは、ログインで認証アプリのコードを入力する画面のHTTPハンドラーを定義する。
package sign_in_two_factor

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/templates/layouts"
	page "github.com/annict/annict/go/internal/templates/pages/sign_in_two_factor"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Handlerは、ログインで認証アプリのコードを入力する画面のHTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                        *config.Config
	sessionMgr                 *session.Manager
	flashMgr                   *session.FlashManager
	limiter                    *ratelimit.Limiter
	getSignInTwoFactorStatusUC *usecase.GetSignInTwoFactorStatusUsecase
	createSignInTwoFactorUC    *usecase.CreateSignInTwoFactorUsecase
	createSessionUC            *usecase.CreateSessionUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	flashMgr *session.FlashManager,
	limiter *ratelimit.Limiter,
	getSignInTwoFactorStatusUC *usecase.GetSignInTwoFactorStatusUsecase,
	createSignInTwoFactorUC *usecase.CreateSignInTwoFactorUsecase,
	createSessionUC *usecase.CreateSessionUsecase,
) *Handler {
	return &Handler{
		cfg:                        cfg,
		sessionMgr:                 sessionMgr,
		flashMgr:                   flashMgr,
		limiter:                    limiter,
		getSignInTwoFactorStatusUC: getSignInTwoFactorStatusUC,
		createSignInTwoFactorUC:    createSignInTwoFactorUC,
		createSessionUC:            createSessionUC,
	}
}

// setNoStoreは、応答をブラウザや中継のキャッシュに残さないよう指示する。
// ログイン途中の画面を、戻る操作や共用の端末で読み返されないようにする。
func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store")
}

// redirectToSignInは、ログイン後の戻り先を引き継いでメールアドレスの入力からやり直させる。
func redirectToSignIn(w http.ResponseWriter, r *http.Request, backURL string) {
	redirectURL := "/sign_in"
	if backURL != "" {
		redirectURL = "/sign_in?back=" + url.QueryEscape(backURL)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// renderNewは認証アプリのコードの入力画面を指定したステータスで描画します。
// 表示 (200) と、送信を受け付けなかったときの再描画 (422) で共有する。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, inputCode string, backURL string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, page.NewPath)
	meta.SetTitle(ctx, "sign_in_two_factor_new_title")
	meta.Description = i18n.T(ctx, "sign_in_two_factor_new_description")

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), page.New(page.NewPageData{
		CSRFToken:  middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr),
		FormErrors: viewmodel.NewFormErrors(formErrors),
		InputCode:  inputCode,
		BackURL:    backURL,
	}))
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
		http.Error(w, i18n.T(ctx, "sign_in_error_server"), http.StatusInternalServerError)
		return
	}

	setNoStore(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "認証アプリのコードの入力画面のレスポンスの書き込みに失敗", "error", err)
	}
}
