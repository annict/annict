package worker

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/mail"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/testutil"
)

// TestPasswordResetTemplates はパスワードリセットメールテンプレートの選択とレンダリングをテストします
func TestPasswordResetTemplates(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		resetURL     string
		htmlContains []string
		textContains []string
	}{
		{
			name:     "日本語メール",
			locale:   "ja",
			resetURL: "https://example.com/reset?token=abc123",
			htmlContains: []string{
				"パスワードリセットのご案内",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
			textContains: []string{
				"パスワードリセットのリクエストを受け付けました",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
		},
		{
			name:     "英語メール",
			locale:   "en",
			resetURL: "https://example.com/reset?token=xyz789",
			htmlContains: []string{
				"Password Reset Request",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
			textContains: []string{
				"We have received a request to reset your password",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			htmlBody, textBody := passwordResetTemplates(tt.locale, tt.resetURL)

			// HTMLテンプレートをレンダリング
			var htmlBuf bytes.Buffer
			if err := htmlBody.Render(ctx, &htmlBuf); err != nil {
				t.Fatalf("HTMLテンプレートのレンダリングに失敗: %v", err)
			}
			for _, expected := range tt.htmlContains {
				if !strings.Contains(htmlBuf.String(), expected) {
					t.Errorf("HTMLに期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, htmlBuf.String())
				}
			}

			// テキストテンプレートをレンダリング
			var textBuf bytes.Buffer
			if err := textBody.Render(ctx, &textBuf); err != nil {
				t.Fatalf("テキストテンプレートのレンダリングに失敗: %v", err)
			}
			for _, expected := range tt.textContains {
				if !strings.Contains(textBuf.String(), expected) {
					t.Errorf("テキストに期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, textBuf.String())
				}
			}
		})
	}
}

// TestSendPasswordResetEmailWorker_Work はメール送信ワーカーの動作をテストします
func TestSendPasswordResetEmailWorker_Work(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	ctx := context.Background()

	cfg := &config.Config{
		Domain: "example.com",
	}

	tests := []struct {
		name         string
		locale       string
		username     string
		email        string
		htmlContains string
	}{
		{
			name:         "日本語ロケールのユーザー",
			locale:       "ja",
			username:     "ja_user_pw_reset_work",
			email:        "ja_pw_reset_work@example.com",
			htmlContains: "パスワードリセットのご案内",
		},
		{
			name:         "英語ロケールのユーザー",
			locale:       "en",
			username:     "en_user_pw_reset_work",
			email:        "en_pw_reset_work@example.com",
			htmlContains: "Password Reset Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailSender := mail.NewNoopSender()

			userID := testutil.NewUserBuilder(t, tx).
				WithUsername(tt.username).
				WithEmail(tt.email).
				WithLocale(tt.locale).
				Build()

			worker := NewSendPasswordResetEmailWorker(queries, mailSender, cfg)

			job := &river.Job[SendPasswordResetEmailArgs]{
				Args: SendPasswordResetEmailArgs{
					UserID: userID,
					Token:  "test-token-123",
				},
			}

			err := worker.Work(ctx, job)
			if err != nil {
				t.Fatalf("Work() error = %v", err)
			}

			if len(mailSender.SentEmails) != 1 {
				t.Fatalf("SentEmails length = %d, want 1", len(mailSender.SentEmails))
			}

			sentEmail := mailSender.SentEmails[0]
			if sentEmail.To != tt.email {
				t.Errorf("To = %q, want %q", sentEmail.To, tt.email)
			}

			// HTMLテンプレートの内容を検証
			var htmlBuf bytes.Buffer
			if err := sentEmail.HTMLBody.Render(ctx, &htmlBuf); err != nil {
				t.Fatalf("HTMLBody.Render() error = %v", err)
			}
			if !strings.Contains(htmlBuf.String(), tt.htmlContains) {
				t.Errorf("HTMLにロケールに応じた文字列が見つかりません: %q", tt.htmlContains)
			}
			if !strings.Contains(htmlBuf.String(), "test-token-123") {
				t.Error("HTMLにリセットトークンが含まれていません")
			}
		})
	}
}
