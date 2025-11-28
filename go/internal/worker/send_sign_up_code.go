package worker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/mail"
	sign_up "github.com/annict/annict/internal/templates/emails/sign_up"
	"github.com/riverqueue/river"
)

// SendSignUpCodeArgs は新規登録確認コード送信ジョブの引数です
type SendSignUpCodeArgs struct {
	Email  string `json:"email"`  // メールアドレス
	Code   string `json:"code"`   // 平文コード（メール本文に含めるため）
	Locale string `json:"locale"` // ロケール（ja/en）
}

// Kind はジョブの種類を返します
func (SendSignUpCodeArgs) Kind() string {
	return "send_sign_up_code"
}

// SendSignUpCodeWorker は新規登録確認コード送信ワーカーです
type SendSignUpCodeWorker struct {
	river.WorkerDefaults[SendSignUpCodeArgs]
	mailClient mail.MailSender
	cfg        *config.Config
}

// NewSendSignUpCodeWorker は新しいSendSignUpCodeWorkerを作成します
func NewSendSignUpCodeWorker(mailClient mail.MailSender, cfg *config.Config) *SendSignUpCodeWorker {
	return &SendSignUpCodeWorker{
		mailClient: mailClient,
		cfg:        cfg,
	}
}

// renderSignUpCodeTemplate はtemplコンポーネントを使ってメールをレンダリングします
func renderSignUpCodeTemplate(ctx context.Context, locale, format, code string) (string, error) {
	var buf bytes.Buffer
	var err error

	// ロケールとフォーマットに応じてtemplコンポーネントを選択
	switch {
	case locale == "ja" && format == "text":
		err = sign_up.JaText(code).Render(ctx, &buf)
	case locale == "ja" && format == "html":
		err = sign_up.JaHTML(code).Render(ctx, &buf)
	case locale == "en" && format == "text":
		err = sign_up.EnText(code).Render(ctx, &buf)
	case locale == "en" && format == "html":
		err = sign_up.EnHTML(code).Render(ctx, &buf)
	default:
		// デフォルトは日本語
		if format == "html" {
			err = sign_up.JaHTML(code).Render(ctx, &buf)
		} else {
			err = sign_up.JaText(code).Render(ctx, &buf)
		}
	}

	if err != nil {
		return "", fmt.Errorf("テンプレートのレンダリングに失敗: %w", err)
	}

	return buf.String(), nil
}

// Work は新規登録確認コードを送信します
func (w *SendSignUpCodeWorker) Work(ctx context.Context, job *river.Job[SendSignUpCodeArgs]) error {
	slog.InfoContext(ctx, "新規登録確認コード送信ジョブを開始します",
		"email", job.Args.Email,
		"locale", job.Args.Locale,
	)

	// メールアドレスの検証
	if job.Args.Email == "" {
		slog.ErrorContext(ctx, "メールアドレスが空です")
		return fmt.Errorf("メールアドレスが空です")
	}

	// ロケールの正規化（デフォルトは日本語）
	locale := job.Args.Locale
	if locale == "" {
		locale = "ja"
	}
	// "en"以外は日本語として扱う
	if locale != "en" {
		locale = "ja"
	}

	// テキストメール本文をレンダリング
	textBody, err := renderSignUpCodeTemplate(ctx, locale, "text", job.Args.Code)
	if err != nil {
		slog.ErrorContext(ctx, "テキストメールのレンダリングに失敗しました",
			"email", job.Args.Email,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("テキストメールのレンダリングに失敗: %w", err)
	}

	// HTMLメール本文をレンダリング
	htmlBody, err := renderSignUpCodeTemplate(ctx, locale, "html", job.Args.Code)
	if err != nil {
		slog.ErrorContext(ctx, "HTMLメールのレンダリングに失敗しました",
			"email", job.Args.Email,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("HTMLメールのレンダリングに失敗: %w", err)
	}

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "sign_up_code_email_subject")

	// メール送信
	err = w.mailClient.SendMultipartEmail(ctx, job.Args.Email, subject, textBody, htmlBody)
	if err != nil {
		slog.ErrorContext(ctx, "メール送信に失敗しました",
			"email", job.Args.Email,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "新規登録確認コードを送信しました",
		"email", job.Args.Email,
		"locale", locale,
	)

	return nil
}
