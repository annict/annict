package email

import (
	"context"
	"fmt"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/emails/two_factor_auth_notification"
)

// TwoFactorAuthNotificationSenderは、二要素認証の設定の変更を本人に知らせるメールの送信を行う
type TwoFactorAuthNotificationSender struct {
	sender Sender
}

// NewTwoFactorAuthNotificationSenderは新しいTwoFactorAuthNotificationSenderを作成する
func NewTwoFactorAuthNotificationSender(sender Sender) *TwoFactorAuthNotificationSender {
	return &TwoFactorAuthNotificationSender{sender: sender}
}

// Sendは通知の種類とロケールに応じた件名と本文で、通知メールをレンダリングして送信する
func (s *TwoFactorAuthNotificationSender) Send(ctx context.Context, to, locale string, event model.TwoFactorAuthNotificationEvent) error {
	var kind two_factor_auth_notification.Kind
	var subjectKey string
	switch event {
	case model.TwoFactorAuthNotificationEventEnabled:
		kind, subjectKey = two_factor_auth_notification.KindEnabled, "two_factor_auth_enabled_email_subject"
	case model.TwoFactorAuthNotificationEventDisabled:
		kind, subjectKey = two_factor_auth_notification.KindDisabled, "two_factor_auth_disabled_email_subject"
	case model.TwoFactorAuthNotificationEventDisabledByAdmin:
		kind, subjectKey = two_factor_auth_notification.KindDisabledByAdmin, "two_factor_auth_disabled_by_admin_email_subject"
	default:
		return fmt.Errorf("未知の二要素認証の通知の種類です: %q", event)
	}

	ctx = i18n.SetLocale(ctx, locale)
	subject := i18n.T(ctx, subjectKey)

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody = two_factor_auth_notification.EnHTML(kind)
		textBody = two_factor_auth_notification.EnText(kind)
	default:
		htmlBody = two_factor_auth_notification.JaHTML(kind)
		textBody = two_factor_auth_notification.JaText(kind)
	}

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
}
