// Package mail はメール送信機能を提供します
package mail

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/resend/resend-go/v2"
)

// SendInput はメール送信の入力
type SendInput struct {
	To       string          // 送信先メールアドレス
	Subject  string          // 件名
	HTMLBody templ.Component // メール本文（HTML形式）
	TextBody templ.Component // メール本文（テキスト形式、nilの場合はHTMLのみ）
}

// Sender はメール送信のインターフェース
type Sender interface {
	Send(ctx context.Context, input SendInput) error
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

// Send はメールを送信します
func (r *ResendClient) Send(ctx context.Context, input SendInput) error {
	var htmlBuf bytes.Buffer
	if err := input.HTMLBody.Render(ctx, &htmlBuf); err != nil {
		return fmt.Errorf("HTMLテンプレートのレンダリングに失敗: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    r.from(),
		To:      []string{input.To},
		Subject: input.Subject,
		Html:    htmlBuf.String(),
	}

	if input.TextBody != nil {
		var textBuf bytes.Buffer
		if err := input.TextBody.Render(ctx, &textBuf); err != nil {
			return fmt.Errorf("テキストテンプレートのレンダリングに失敗: %w", err)
		}
		params.Text = textBuf.String()
	}

	_, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	return nil
}
