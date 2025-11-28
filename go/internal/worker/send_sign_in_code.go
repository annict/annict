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
	sign_in "github.com/annict/annict/internal/templates/emails/sign_in"
	"github.com/riverqueue/river"
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
	mailClient mail.MailSender
	cfg        *config.Config
}

// NewSendSignInCodeWorker は新しいSendSignInCodeWorkerを作成します
func NewSendSignInCodeWorker(queries *query.Queries, mailClient mail.MailSender, cfg *config.Config) *SendSignInCodeWorker {
	return &SendSignInCodeWorker{
		queries:    queries,
		mailClient: mailClient,
		cfg:        cfg,
	}
}

// renderSignInTemplate はtemplコンポーネントを使ってメールをレンダリングします
func renderSignInTemplate(ctx context.Context, locale, format, code string) (string, error) {
	var buf bytes.Buffer
	var err error

	// ロケールとフォーマットに応じてtemplコンポーネントを選択
	switch {
	case locale == "ja" && format == "text":
		err = sign_in.JaText(code).Render(ctx, &buf)
	case locale == "ja" && format == "html":
		err = sign_in.JaHTML(code).Render(ctx, &buf)
	case locale == "en" && format == "text":
		err = sign_in.EnText(code).Render(ctx, &buf)
	case locale == "en" && format == "html":
		err = sign_in.EnHTML(code).Render(ctx, &buf)
	default:
		// デフォルトは日本語
		if format == "html" {
			err = sign_in.JaHTML(code).Render(ctx, &buf)
		} else {
			err = sign_in.JaText(code).Render(ctx, &buf)
		}
	}

	if err != nil {
		return "", fmt.Errorf("テンプレートのレンダリングに失敗: %w", err)
	}

	return buf.String(), nil
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

	// テキストメール本文をレンダリング
	textBody, err := renderSignInTemplate(ctx, locale, "text", job.Args.Code)
	if err != nil {
		slog.ErrorContext(ctx, "テキストメールのレンダリングに失敗しました",
			"user_id", job.Args.UserID,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("テキストメールのレンダリングに失敗: %w", err)
	}

	// HTMLメール本文をレンダリング
	htmlBody, err := renderSignInTemplate(ctx, locale, "html", job.Args.Code)
	if err != nil {
		slog.ErrorContext(ctx, "HTMLメールのレンダリングに失敗しました",
			"user_id", job.Args.UserID,
			"locale", locale,
			"error", err,
		)
		return fmt.Errorf("HTMLメールのレンダリングに失敗: %w", err)
	}

	// 件名を取得（i18n使用）
	subject := i18n.T(ctx, "sign_in_code_email_subject")

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

	slog.InfoContext(ctx, "ログインコードを送信しました",
		"user_id", job.Args.UserID,
		"email", email,
		"locale", locale,
	)

	return nil
}
