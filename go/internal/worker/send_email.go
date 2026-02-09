package worker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/mail"
	password_reset "github.com/annict/annict/go/internal/templates/emails/password_reset"
	sign_in "github.com/annict/annict/go/internal/templates/emails/sign_in"
	sign_up "github.com/annict/annict/go/internal/templates/emails/sign_up"
)

// SendEmailArgs はメール送信ジョブの引数です
type SendEmailArgs struct {
	To       string `json:"to"`        // 送信先メールアドレス
	Subject  string `json:"subject"`   // 件名
	HTMLBody string `json:"html_body"` // メール本文（HTML形式、レンダリング済み）
	TextBody string `json:"text_body"` // メール本文（テキスト形式、レンダリング済み）
}

// Kind はジョブの種類を返します
func (SendEmailArgs) Kind() string {
	return "send_email"
}

// InsertOpts はジョブ挿入時のデフォルトオプションを返します
func (SendEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 5,
	}
}

// SendEmailWorker はメール送信ワーカーです
type SendEmailWorker struct {
	river.WorkerDefaults[SendEmailArgs]
	mailSender mail.Sender
}

// NewSendEmailWorker は新しいSendEmailWorkerを作成します
func NewSendEmailWorker(mailSender mail.Sender) *SendEmailWorker {
	return &SendEmailWorker{
		mailSender: mailSender,
	}
}

// Work はメールを送信します
func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[SendEmailArgs]) error {
	slog.InfoContext(ctx, "メール送信ジョブを開始します",
		"to", job.Args.To,
		"subject", job.Args.Subject,
	)

	if job.Args.To == "" {
		slog.ErrorContext(ctx, "メールアドレスが空です")
		return fmt.Errorf("メールアドレスが空です")
	}

	err := w.mailSender.SendRaw(ctx, mail.SendRawInput{
		To:       job.Args.To,
		Subject:  job.Args.Subject,
		HTMLBody: job.Args.HTMLBody,
		TextBody: job.Args.TextBody,
	})
	if err != nil {
		slog.ErrorContext(ctx, "メール送信に失敗しました",
			"to", job.Args.To,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "メール送信が完了しました",
		"to", job.Args.To,
	)

	return nil
}

// normalizeLocale はロケールを正規化します（"en"以外は日本語として扱う）
func normalizeLocale(locale string) string {
	if locale == "en" {
		return "en"
	}
	return "ja"
}

// renderComponent はtemplコンポーネントを文字列にレンダリングします
func renderComponent(ctx context.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// BuildSignUpCodeEmail は新規登録確認コード送信メールのSendEmailArgsを構築します
func BuildSignUpCodeEmail(ctx context.Context, email, code, locale string) (*SendEmailArgs, error) {
	locale = normalizeLocale(locale)
	ctx = i18n.SetLocale(ctx, locale)

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody, textBody = sign_up.EnHTML(code), sign_up.EnText(code)
	default:
		htmlBody, textBody = sign_up.JaHTML(code), sign_up.JaText(code)
	}

	htmlStr, err := renderComponent(ctx, htmlBody)
	if err != nil {
		return nil, fmt.Errorf("HTMLテンプレートのレンダリングに失敗: %w", err)
	}

	textStr, err := renderComponent(ctx, textBody)
	if err != nil {
		return nil, fmt.Errorf("テキストテンプレートのレンダリングに失敗: %w", err)
	}

	subject := i18n.T(ctx, "sign_up_code_email_subject")

	return &SendEmailArgs{
		To:       email,
		Subject:  subject,
		HTMLBody: htmlStr,
		TextBody: textStr,
	}, nil
}

// BuildSignInCodeEmail はログインコード送信メールのSendEmailArgsを構築します
func BuildSignInCodeEmail(ctx context.Context, email, code, locale string) (*SendEmailArgs, error) {
	locale = normalizeLocale(locale)
	ctx = i18n.SetLocale(ctx, locale)

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody, textBody = sign_in.EnHTML(code), sign_in.EnText(code)
	default:
		htmlBody, textBody = sign_in.JaHTML(code), sign_in.JaText(code)
	}

	htmlStr, err := renderComponent(ctx, htmlBody)
	if err != nil {
		return nil, fmt.Errorf("HTMLテンプレートのレンダリングに失敗: %w", err)
	}

	textStr, err := renderComponent(ctx, textBody)
	if err != nil {
		return nil, fmt.Errorf("テキストテンプレートのレンダリングに失敗: %w", err)
	}

	subject := i18n.T(ctx, "sign_in_code_email_subject")

	return &SendEmailArgs{
		To:       email,
		Subject:  subject,
		HTMLBody: htmlStr,
		TextBody: textStr,
	}, nil
}

// BuildPasswordResetEmail はパスワードリセットメールのSendEmailArgsを構築します
func BuildPasswordResetEmail(ctx context.Context, email, resetURL, locale string) (*SendEmailArgs, error) {
	locale = normalizeLocale(locale)
	ctx = i18n.SetLocale(ctx, locale)

	var htmlBody, textBody templ.Component
	switch locale {
	case "en":
		htmlBody, textBody = password_reset.EnHTML(resetURL), password_reset.EnText(resetURL)
	default:
		htmlBody, textBody = password_reset.JaHTML(resetURL), password_reset.JaText(resetURL)
	}

	htmlStr, err := renderComponent(ctx, htmlBody)
	if err != nil {
		return nil, fmt.Errorf("HTMLテンプレートのレンダリングに失敗: %w", err)
	}

	textStr, err := renderComponent(ctx, textBody)
	if err != nil {
		return nil, fmt.Errorf("テキストテンプレートのレンダリングに失敗: %w", err)
	}

	subject := i18n.T(ctx, "password_reset_email_subject")

	return &SendEmailArgs{
		To:       email,
		Subject:  subject,
		HTMLBody: htmlStr,
		TextBody: textStr,
	}, nil
}
