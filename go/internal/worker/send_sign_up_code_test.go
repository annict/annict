package worker

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/mail"
)

// TestSignUpCodeTemplates は新規登録確認コードテンプレートの選択とレンダリングをテストします
func TestSignUpCodeTemplates(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		code         string
		htmlContains []string
		textContains []string
	}{
		{
			name:   "日本語メール",
			locale: "ja",
			code:   "123456",
			htmlContains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
			textContains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
		},
		{
			name:   "英語メール",
			locale: "en",
			code:   "654321",
			htmlContains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
			textContains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			htmlBody, textBody := signUpCodeTemplates(tt.locale, tt.code)

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

// TestSendSignUpCodeWorker_Work は新規登録確認コード送信ワーカーの動作をテストします
func TestSendSignUpCodeWorker_Work(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{
		Domain: "example.com",
	}

	tests := []struct {
		name         string
		locale       string
		email        string
		htmlContains string
	}{
		{
			name:         "日本語ロケール",
			locale:       "ja",
			email:        "ja_sign_up_work@example.com",
			htmlContains: "Annictへようこそ！",
		},
		{
			name:         "英語ロケール",
			locale:       "en",
			email:        "en_sign_up_work@example.com",
			htmlContains: "Welcome to Annict!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailSender := mail.NewNoopSender()

			worker := NewSendSignUpCodeWorker(mailSender, cfg)

			job := &river.Job[SendSignUpCodeArgs]{
				Args: SendSignUpCodeArgs{
					Email:  tt.email,
					Code:   "654321",
					Locale: tt.locale,
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
			if !strings.Contains(htmlBuf.String(), "654321") {
				t.Error("HTMLに確認コードが含まれていません")
			}
		})
	}
}
