package sign_in

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateValidator はメールアドレス送信フォームのバリデーションを行います
type CreateValidator struct {
	Email string
}

// Validate はフォームの形式バリデーションを行います
func (v *CreateValidator) Validate(ctx context.Context) *session.FormErrors {
	errors := &session.FormErrors{}

	if strings.TrimSpace(v.Email) == "" {
		errors.AddFieldError("email", i18n.T(ctx, "sign_in_email_required"))
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}
