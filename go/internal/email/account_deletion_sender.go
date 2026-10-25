package email

import (
	"context"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/emails/account_deletion"
)

// AccountDeletionSenderは退会完了メールの送信を行う
type AccountDeletionSender struct {
	sender Sender
}

// NewAccountDeletionSenderは新しいAccountDeletionSenderを作成する
func NewAccountDeletionSender(sender Sender) *AccountDeletionSender {
	return &AccountDeletionSender{sender: sender}
}

// Sendは退会完了メールをレンダリングして送信する
func (s *AccountDeletionSender) Send(ctx context.Context, to, locale string, isStripeSupporter bool) error {
	ctx = i18n.SetLocale(ctx, locale)
	subject := i18n.T(ctx, "account_deletion_email_subject")

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody = account_deletion.EnHTML(isStripeSupporter)
		textBody = account_deletion.EnText(isStripeSupporter)
	default:
		htmlBody = account_deletion.JaHTML(isStripeSupporter)
		textBody = account_deletion.JaText(isStripeSupporter)
	}

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
}
