package email

import (
	"context"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/emails/sign_in"
)

// SignInCodeSender はログインコードメールの送信を行う
type SignInCodeSender struct {
	sender Sender
}

// NewSignInCodeSender は新しい SignInCodeSender を作成する
func NewSignInCodeSender(sender Sender) *SignInCodeSender {
	return &SignInCodeSender{sender: sender}
}

// Send はログインコードメールをレンダリングして送信する
func (s *SignInCodeSender) Send(ctx context.Context, to, code, locale string) error {
	ctx = i18n.SetLocale(ctx, locale)
	subject := i18n.T(ctx, "sign_in_code_email_subject")

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody = sign_in.EnHTML(code)
		textBody = sign_in.EnText(code)
	default:
		htmlBody = sign_in.JaHTML(code)
		textBody = sign_in.JaText(code)
	}

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
}
