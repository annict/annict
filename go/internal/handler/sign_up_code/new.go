package sign_up_code

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	signUpCodePage "github.com/annict/annict/go/internal/templates/pages/sign_up_code"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New は新規登録確認コード入力フォームを表示します (GET /sign_up/code)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値取得エラー", "error", err, "key", "sign_up_email")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// メールアドレスがセッションにない場合は /sign_up にリダイレクト
	if email == "" {
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	h.renderNewForm(w, r, http.StatusOK, nil, email)
}

// renderNewForm は新規登録確認コード入力フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderNewForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, email string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_up_code_title")
	meta.Description = i18n.T(ctx, "sign_up_code_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_up/code"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := signUpCodePage.NewPageData{
		CSRFToken:  csrfToken,
		FormErrors: formErrors,
		Email:      email,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), signUpCodePage.New(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
