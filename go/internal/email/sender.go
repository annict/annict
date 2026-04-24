// Package email はメール送信機能を提供します
package email

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/resend/resend-go/v2"
)

// Sender はメール送信を行うインターフェース
type Sender interface {
	// Send はメールを送信する
	Send(ctx context.Context, input SendInput) error
}

// SendInput はメール送信の入力
type SendInput struct {
	To       string          // 送信先メールアドレス
	Subject  string          // 件名
	HTMLBody templ.Component // メール本文（HTML形式）
	TextBody templ.Component // メール本文（テキスト形式、nilの場合はHTMLのみ）
}

// ResendSender はResend APIを使用してメールを送信する
type ResendSender struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

// NewResendSender は新しいResendSenderを作成する
func NewResendSender(apiKey, fromEmail, fromName string) *ResendSender {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client := resend.NewCustomClient(httpClient, apiKey)
	return &ResendSender{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// from は "Name <email>" または "email" 形式の From アドレスを返す。
// fromName が空文字の場合はアドレスのみを返す（display-name 省略）。
func (s *ResendSender) from() string {
	if s.fromName != "" {
		return fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}
	return s.fromEmail
}

// Send はメールを送信する
func (s *ResendSender) Send(ctx context.Context, input SendInput) error {
	var htmlBuf bytes.Buffer
	if err := input.HTMLBody.Render(ctx, &htmlBuf); err != nil {
		return fmt.Errorf("HTMLテンプレートのレンダリングに失敗しました: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    s.from(),
		To:      []string{input.To},
		Subject: input.Subject,
		Html:    htmlBuf.String(),
	}

	if input.TextBody != nil {
		var textBuf bytes.Buffer
		if err := input.TextBody.Render(ctx, &textBuf); err != nil {
			return fmt.Errorf("テキストテンプレートのレンダリングに失敗しました: %w", err)
		}
		params.Text = textBuf.String()
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("メール送信に失敗しました: %w", err)
	}

	return nil
}

// NoopSender はメールを送信しないダミー実装（テスト用）
type NoopSender struct {
	// SentEmails は送信されたメールを記録する（テスト用）
	SentEmails []SendInput
}

// NewNoopSender は新しいNoopSenderを作成する
func NewNoopSender() *NoopSender {
	return &NoopSender{
		SentEmails: make([]SendInput, 0),
	}
}

// Send はメールを送信せず、記録のみ行う
func (s *NoopSender) Send(_ context.Context, input SendInput) error {
	s.SentEmails = append(s.SentEmails, input)
	return nil
}

// Reset は送信記録をクリアする
func (s *NoopSender) Reset() {
	s.SentEmails = make([]SendInput, 0)
}
