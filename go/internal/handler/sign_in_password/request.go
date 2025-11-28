package sign_in_password

import (
	"context"
	"strings"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// Request サインインリクエストのデータ
type Request struct {
	Password string
}

// Validate リクエストのバリデーション
func (req *Request) Validate(ctx context.Context) *session.FormErrors {
	errors := &session.FormErrors{}

	if strings.TrimSpace(req.Password) == "" {
		errors.AddFieldError("password", i18n.T(ctx, "sign_in_error_password_required"))
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}
