package email

import (
	"context"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/emails/sign_up"
)

// SignUpCodeSender は新規登録確認コードメールの送信を行う
type SignUpCodeSender struct {
	sender Sender
}

// NewSignUpCodeSender は新しい SignUpCodeSender を作成する
func NewSignUpCodeSender(sender Sender) *SignUpCodeSender {
	return &SignUpCodeSender{sender: sender}
}

// Send は新規登録確認コードメールをレンダリングして送信する
func (s *SignUpCodeSender) Send(ctx context.Context, to, code, locale string) error {
	ctx = i18n.SetLocale(ctx, locale)
	subject := i18n.T(ctx, "sign_up_code_email_subject")

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody = sign_up.EnHTML(code)
		textBody = sign_up.EnText(code)
	default:
		htmlBody = sign_up.JaHTML(code)
		textBody = sign_up.JaText(code)
	}

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
}
