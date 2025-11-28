package password_reset

import (
	"context"
	"strings"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// Request はパスワードリセット申請フォームのリクエストを表します
type Request struct {
	Email string
}

// Validate はフォームの形式バリデーションを行います
func (req *Request) Validate(ctx context.Context) *session.FormErrors {
	errors := &session.FormErrors{}

	if strings.TrimSpace(req.Email) == "" {
		errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}
