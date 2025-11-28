package sign_in

import (
	"context"
	"strings"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// CreateRequest はメールアドレス送信フォームのリクエストを表します
type CreateRequest struct {
	Email string
}

// Validate はフォームの形式バリデーションを行います
func (req *CreateRequest) Validate(ctx context.Context) *session.FormErrors {
	errors := &session.FormErrors{}

	if strings.TrimSpace(req.Email) == "" {
		errors.AddFieldError("email", i18n.T(ctx, "sign_in_email_required"))
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}
