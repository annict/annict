package password

import (
	"context"
	"strings"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// Request はパスワード更新フォームのリクエストを表します
type Request struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// Validate はフォームの形式バリデーションを行います
func (req *Request) Validate(ctx context.Context) *session.FormErrors {
	errors := &session.FormErrors{}

	if strings.TrimSpace(req.Token) == "" {
		errors.AddFieldError("token", i18n.T(ctx, "password_reset_token_invalid"))
	}

	if strings.TrimSpace(req.Password) == "" {
		errors.AddFieldError("password", i18n.T(ctx, "password_reset_password_required"))
	}

	if strings.TrimSpace(req.PasswordConfirmation) == "" {
		errors.AddFieldError("password_confirmation", i18n.T(ctx, "password_reset_password_confirmation_required"))
	}

	if strings.TrimSpace(req.Password) != "" && strings.TrimSpace(req.PasswordConfirmation) != "" && req.Password != req.PasswordConfirmation {
		errors.AddFieldError("password_confirmation", i18n.T(ctx, "password_reset_password_mismatch"))
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}
