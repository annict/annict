package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/mail"
	sign_up "github.com/annict/annict/go/internal/templates/emails/sign_up"
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
	mailSender mail.Sender
	cfg        *config.Config
}

// NewSendSignUpCodeWorker は新しいSendSignUpCodeWorkerを作成します
func NewSendSignUpCodeWorker(mailSender mail.Sender, cfg *config.Config) *SendSignUpCodeWorker {
	return &SendSignUpCodeWorker{
		mailSender: mailSender,
		cfg:        cfg,
	}
}

// signUpCodeTemplates はロケールに応じたメールテンプレートを返します
func signUpCodeTemplates(locale, code string) (htmlBody, textBody templ.Component) {
	switch locale {
	case "en":
		return sign_up.EnHTML(code), sign_up.EnText(code)
	default:
		return sign_up.JaHTML(code), sign_up.JaText(code)
	}
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

	// テンプレートを選択
	htmlBody, textBody := signUpCodeTemplates(locale, job.Args.Code)

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "sign_up_code_email_subject")

	// メール送信
	err := w.mailSender.Send(ctx, mail.SendInput{
		To:       job.Args.Email,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
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
