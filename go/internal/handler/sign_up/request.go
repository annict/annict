package sign_up

import (
	"context"
	"net/mail"
	"strings"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// CreateRequest は新規登録フォームのリクエストを表します
type CreateRequest struct {
	Email string
}

// Validate はリクエストのバリデーションを行います
func (req *CreateRequest) Validate(ctx context.Context) *session.FormErrors {
	formErrors := &session.FormErrors{}

	// メールアドレスの必須チェック
	if strings.TrimSpace(req.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_required"))
		return formErrors
	}

	// メールアドレスの形式チェック
	if _, err := mail.ParseAddress(req.Email); err != nil {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_invalid"))
		return formErrors
	}

	return nil
}
