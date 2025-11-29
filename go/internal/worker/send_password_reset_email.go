package worker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/mail"
	"github.com/annict/annict/internal/query"
	password_reset "github.com/annict/annict/internal/templates/emails/password_reset"
	"github.com/riverqueue/river"
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
	mailClient mail.MailSender
	cfg        *config.Config
}

// NewSendPasswordResetEmailWorker は新しいSendPasswordResetEmailWorkerを作成します
func NewSendPasswordResetEmailWorker(queries *query.Queries, mailClient mail.MailSender, cfg *config.Config) *SendPasswordResetEmailWorker {
	return &SendPasswordResetEmailWorker{
		queries:    queries,
		mailClient: mailClient,
		cfg:        cfg,
	}
}

// renderEmailTemplate はtemplコンポーネントを使ってメールをレンダリングします
func renderEmailTemplate(ctx context.Context, locale, format, resetURL string) (string, error) {
	var buf bytes.Buffer
	var err error

	// ロケールとフォーマットに応じてtemplコンポーネントを選択
	switch {
	case locale == "ja" && format == "text":
		err = password_reset.JaText(resetURL).Render(ctx, &buf)
	case locale == "ja" && format == "html":
		err = password_reset.JaHTML(resetURL).Render(ctx, &buf)
	case locale == "en" && format == "text":
		err = password_reset.EnText(resetURL).Render(ctx, &buf)
	case locale == "en" && format == "html":
		err = password_reset.EnHTML(resetURL).Render(ctx, &buf)
	default:
		// デフォルトは日本語
		if format == "html" {
			err = password_reset.JaHTML(resetURL).Render(ctx, &buf)
		} else {
			err = password_reset.JaText(resetURL).Render(ctx, &buf)
		}
	}

	if err != nil {
		return "", fmt.Errorf("テンプレートのレンダリングに失敗: %w", err)
	}

	return buf.String(), nil
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

	// テキストメール本文をレンダリング
	textBody, err := renderEmailTemplate(ctx, locale, "text", resetURL)
	if err != nil {
		slog.ErrorContext(ctx, "テキストメールのレンダリングに失敗しました",
			"user_id", job.Args.UserID,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("テキストメールのレンダリングに失敗: %w", err)
	}

	// HTMLメール本文をレンダリング
	htmlBody, err := renderEmailTemplate(ctx, locale, "html", resetURL)
	if err != nil {
		slog.ErrorContext(ctx, "HTMLメールのレンダリングに失敗しました",
			"user_id", job.Args.UserID,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("HTMLメールのレンダリングに失敗: %w", err)
	}

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "password_reset_email_subject")

	// メール送信
	err = w.mailClient.SendMultipartEmail(ctx, email, subject, textBody, htmlBody)
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
