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

// SendRawInput はレンダリング済みメール送信の入力
type SendRawInput struct {
	To       string // 送信先メールアドレス
	Subject  string // 件名
	HTMLBody string // メール本文（HTML形式、レンダリング済み）
	TextBody string // メール本文（テキスト形式、レンダリング済み、空の場合はHTMLのみ）
}

// Sender はメール送信のインターフェース
type Sender interface {
	Send(ctx context.Context, input SendInput) error
	SendRaw(ctx context.Context, input SendRawInput) error
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

// SendRaw はレンダリング済みの文字列でメールを送信します
func (r *ResendClient) SendRaw(ctx context.Context, input SendRawInput) error {
	params := &resend.SendEmailRequest{
		From:    r.from(),
		To:      []string{input.To},
		Subject: input.Subject,
		Html:    input.HTMLBody,
	}

	if input.TextBody != "" {
		params.Text = input.TextBody
	}

	_, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	return nil
}
