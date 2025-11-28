// Package mail はmail機能を提供します
package mail

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/resend/resend-go/v2"
)

// MailSender はメール送信のインターフェース
type MailSender interface {
	SendMultipartEmail(ctx context.Context, to, subject, text, html string) error
}

// ResendClient はResend APIを使用したメール送信クライアント
type ResendClient struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

// NewResendClient は新しいResendClientを作成します
func NewResendClient(apiKey, fromEmail, fromName string) *ResendClient {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client := resend.NewCustomClient(httpClient, apiKey)
	return &ResendClient{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// from はFromアドレスを生成します
func (r *ResendClient) from() string {
	if r.fromName != "" {
		return fmt.Sprintf("%s <%s>", r.fromName, r.fromEmail)
	}
	return r.fromEmail
}

// SendMultipartEmail はテキストとHTMLの両方を含むメールを送信します
func (r *ResendClient) SendMultipartEmail(ctx context.Context, to, subject, text, html string) error {
	params := &resend.SendEmailRequest{
		From:    r.from(),
		To:      []string{to},
		Subject: subject,
		Text:    text,
		Html:    html,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
