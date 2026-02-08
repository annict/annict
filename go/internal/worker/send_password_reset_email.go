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
	password_reset "github.com/annict/annict/go/internal/templates/emails/password_reset"
)

// SendPasswordResetEmailArgs はパスワードリセットメール送信ジョブの引数です
type SendPasswordResetEmailArgs struct {
	UserID int64  `json:"user_id"` // ユーザーID
	Token  string `json:"token"`   // 平文トークン（メール本文に含めるため）
}

// Kind はジョブの種類を返します
func (SendPasswordResetEmailArgs) Kind() string {
	return "send_password_reset_email"
}

// SendPasswordResetEmailWorker はパスワードリセットメール送信ワーカーです
type SendPasswordResetEmailWorker struct {
	river.WorkerDefaults[SendPasswordResetEmailArgs]
	queries    *query.Queries
	mailSender mail.Sender
	cfg        *config.Config
}

// NewSendPasswordResetEmailWorker は新しいSendPasswordResetEmailWorkerを作成します
func NewSendPasswordResetEmailWorker(queries *query.Queries, mailSender mail.Sender, cfg *config.Config) *SendPasswordResetEmailWorker {
	return &SendPasswordResetEmailWorker{
		queries:    queries,
		mailSender: mailSender,
		cfg:        cfg,
	}
}

// passwordResetTemplates はロケールに応じたメールテンプレートを返します
func passwordResetTemplates(locale, resetURL string) (htmlBody, textBody templ.Component) {
	switch locale {
	case "en":
		return password_reset.EnHTML(resetURL), password_reset.EnText(resetURL)
	default:
		return password_reset.JaHTML(resetURL), password_reset.JaText(resetURL)
	}
}

// Work はパスワードリセットメールを送信します
func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[SendPasswordResetEmailArgs]) error {
	slog.InfoContext(ctx, "パスワードリセットメール送信ジョブを開始します",
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

	// リセットURLを生成
	resetURL := fmt.Sprintf("%s/password/edit?token=%s", w.cfg.AppURL(), job.Args.Token)

	// テンプレートを選択
	htmlBody, textBody := passwordResetTemplates(locale, resetURL)

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "password_reset_email_subject")

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

	slog.InfoContext(ctx, "パスワードリセットメールを送信しました",
		"user_id", job.Args.UserID,
		"email", email,
		"locale", locale,
	)

	return nil
}
