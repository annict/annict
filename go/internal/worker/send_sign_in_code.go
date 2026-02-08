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
	"github.com/annict/annict/go/internal/query"
	sign_in "github.com/annict/annict/go/internal/templates/emails/sign_in"
)

// SendSignInCodeArgs はログインコード送信ジョブの引数です
type SendSignInCodeArgs struct {
	UserID int64  `json:"user_id"` // ユーザーID
	Code   string `json:"code"`    // 平文コード（メール本文に含めるため）
}

// Kind はジョブの種類を返します
func (SendSignInCodeArgs) Kind() string {
	return "send_sign_in_code"
}

// SendSignInCodeWorker はログインコード送信ワーカーです
type SendSignInCodeWorker struct {
	river.WorkerDefaults[SendSignInCodeArgs]
	queries    *query.Queries
	mailSender mail.Sender
	cfg        *config.Config
}

// NewSendSignInCodeWorker は新しいSendSignInCodeWorkerを作成します
func NewSendSignInCodeWorker(queries *query.Queries, mailSender mail.Sender, cfg *config.Config) *SendSignInCodeWorker {
	return &SendSignInCodeWorker{
		queries:    queries,
		mailSender: mailSender,
		cfg:        cfg,
	}
}

// signInCodeTemplates はロケールに応じたメールテンプレートを返します
func signInCodeTemplates(locale, code string) (htmlBody, textBody templ.Component) {
	switch locale {
	case "en":
		return sign_in.EnHTML(code), sign_in.EnText(code)
	default:
		return sign_in.JaHTML(code), sign_in.JaText(code)
	}
}

// Work はログインコードを送信します
func (w *SendSignInCodeWorker) Work(ctx context.Context, job *river.Job[SendSignInCodeArgs]) error {
	slog.InfoContext(ctx, "ログインコード送信ジョブを開始します",
		"user_id", job.Args.UserID,
	)

	// ユーザー情報を取得
	user, err := w.queries.GetUserByID(ctx, job.Args.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "ユーザー情報の取得に失敗しました",
			"user_id", job.Args.UserID,
			"error", err,
		)
		return fmt.Errorf("ユーザー情報の取得に失敗: %w", err)
	}

	// メールアドレスの取得
	email := user.Email
	if email == "" {
		slog.ErrorContext(ctx, "ユーザーのメールアドレスが空です",
			"user_id", job.Args.UserID,
		)
		return fmt.Errorf("ユーザーのメールアドレスが空です")
	}

	// ユーザーのロケールを取得（デフォルトは日本語）
	locale := user.Locale
	if locale == "" {
		locale = "ja"
	}
	// "en"以外は日本語として扱う
	if locale != "en" {
		locale = "ja"
	}

	// テンプレートを選択
	htmlBody, textBody := signInCodeTemplates(locale, job.Args.Code)

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "sign_in_code_email_subject")

	// メール送信
	err = w.mailSender.Send(ctx, mail.SendInput{
		To:       email,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	if err != nil {
		slog.ErrorContext(ctx, "メール送信に失敗しました",
			"user_id", job.Args.UserID,
			"email", email,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログインコードを送信しました",
		"user_id", job.Args.UserID,
		"email", email,
		"locale", locale,
	)

	return nil
}
