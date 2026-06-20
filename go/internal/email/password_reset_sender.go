package email

import (
	"context"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/emails/password_reset"
)

// PasswordResetSender はパスワードリセットメールの送信を行う
type PasswordResetSender struct {
	sender Sender
}

// NewPasswordResetSender は新しい PasswordResetSender を作成する
func NewPasswordResetSender(sender Sender) *PasswordResetSender {
	return &PasswordResetSender{sender: sender}
}

// Send はパスワードリセットメールをレンダリングして送信する
func (s *PasswordResetSender) Send(ctx context.Context, to, resetURL, locale string) error {
	ctx = i18n.SetLocale(ctx, locale)
	subject := i18n.T(ctx, "password_reset_email_subject")

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody = password_reset.EnHTML(resetURL)
		textBody = password_reset.EnText(resetURL)
	default:
		htmlBody = password_reset.JaHTML(resetURL)
		textBody = password_reset.JaText(resetURL)
	}

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
}
