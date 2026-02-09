package worker

import (
	"context"
	"strings"
	"testing"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/mail"
)

// TestBuildSignUpCodeEmail は新規登録確認コードメールの構築をテストします
func TestBuildSignUpCodeEmail(t *testing.T) {
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
				"123456",
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
				"654321",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			args, err := BuildSignUpCodeEmail(ctx, "test@example.com", tt.code, tt.locale)
			if err != nil {
				t.Fatalf("BuildSignUpCodeEmail() error = %v", err)
			}

			if args.To != "test@example.com" {
				t.Errorf("To = %q, want %q", args.To, "test@example.com")
			}

			if args.Subject == "" {
				t.Error("Subject is empty")
			}

			for _, expected := range tt.htmlContains {
				if !strings.Contains(args.HTMLBody, expected) {
					t.Errorf("HTMLBodyに期待される文字列が見つかりません: %q", expected)
				}
			}

			for _, expected := range tt.textContains {
				if !strings.Contains(args.TextBody, expected) {
					t.Errorf("TextBodyに期待される文字列が見つかりません: %q", expected)
				}
			}
		})
	}
}

// TestBuildSignInCodeEmail はログインコードメールの構築をテストします
func TestBuildSignInCodeEmail(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		code         string
		htmlContains []string
	}{
		{
			name:   "日本語メール",
			locale: "ja",
			code:   "111111",
			htmlContains: []string{
				"111111",
			},
		},
		{
			name:   "英語メール",
			locale: "en",
			code:   "222222",
			htmlContains: []string{
				"222222",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			args, err := BuildSignInCodeEmail(ctx, "user@example.com", tt.code, tt.locale)
			if err != nil {
				t.Fatalf("BuildSignInCodeEmail() error = %v", err)
			}

			if args.To != "user@example.com" {
				t.Errorf("To = %q, want %q", args.To, "user@example.com")
			}

			if args.Subject == "" {
				t.Error("Subject is empty")
			}

			for _, expected := range tt.htmlContains {
				if !strings.Contains(args.HTMLBody, expected) {
					t.Errorf("HTMLBodyに期待される文字列が見つかりません: %q", expected)
				}
			}
		})
	}
}

// TestBuildPasswordResetEmail はパスワードリセットメールの構築をテストします
func TestBuildPasswordResetEmail(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		resetURL     string
		htmlContains []string
	}{
		{
			name:     "日本語メール",
			locale:   "ja",
			resetURL: "https://example.com/password/edit?token=abc123",
			htmlContains: []string{
				"https://example.com/password/edit?token=abc123",
			},
		},
		{
			name:     "英語メール",
			locale:   "en",
			resetURL: "https://example.com/password/edit?token=def456",
			htmlContains: []string{
				"https://example.com/password/edit?token=def456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			args, err := BuildPasswordResetEmail(ctx, "reset@example.com", tt.resetURL, tt.locale)
			if err != nil {
				t.Fatalf("BuildPasswordResetEmail() error = %v", err)
			}

			if args.To != "reset@example.com" {
				t.Errorf("To = %q, want %q", args.To, "reset@example.com")
			}

			if args.Subject == "" {
				t.Error("Subject is empty")
			}

			for _, expected := range tt.htmlContains {
				if !strings.Contains(args.HTMLBody, expected) {
					t.Errorf("HTMLBodyに期待される文字列が見つかりません: %q", expected)
				}
			}
		})
	}
}

// TestNormalizeLocale はロケール正規化のテストです
func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"en", "en"},
		{"ja", "ja"},
		{"", "ja"},
		{"fr", "ja"},
		{"zh", "ja"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeLocale(tt.input)
			if got != tt.want {
				t.Errorf("normalizeLocale(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestSendEmailWorker_Work はメール送信ワーカーの動作をテストします
func TestSendEmailWorker_Work(t *testing.T) {
	ctx := context.Background()
	mailSender := mail.NewNoopSender()
	w := NewSendEmailWorker(mailSender)

	job := &river.Job[SendEmailArgs]{
		Args: SendEmailArgs{
			To:       "test@example.com",
			Subject:  "テスト件名",
			HTMLBody: "<h1>テスト</h1>",
			TextBody: "テスト",
		},
	}

	err := w.Work(ctx, job)
	if err != nil {
		t.Fatalf("Work() error = %v", err)
	}

	if len(mailSender.SentRawEmails) != 1 {
		t.Fatalf("SentRawEmails length = %d, want 1", len(mailSender.SentRawEmails))
	}

	sent := mailSender.SentRawEmails[0]
	if sent.To != "test@example.com" {
		t.Errorf("To = %q, want %q", sent.To, "test@example.com")
	}
	if sent.Subject != "テスト件名" {
		t.Errorf("Subject = %q, want %q", sent.Subject, "テスト件名")
	}
	if sent.HTMLBody != "<h1>テスト</h1>" {
		t.Errorf("HTMLBody = %q, want %q", sent.HTMLBody, "<h1>テスト</h1>")
	}
	if sent.TextBody != "テスト" {
		t.Errorf("TextBody = %q, want %q", sent.TextBody, "テスト")
	}
}

// TestSendEmailWorker_Work_EmptyTo はメールアドレスが空の場合にエラーを返すことをテストします
func TestSendEmailWorker_Work_EmptyTo(t *testing.T) {
	ctx := context.Background()
	mailSender := mail.NewNoopSender()
	w := NewSendEmailWorker(mailSender)

	job := &river.Job[SendEmailArgs]{
		Args: SendEmailArgs{
			To:       "",
			Subject:  "テスト",
			HTMLBody: "<h1>テスト</h1>",
		},
	}

	err := w.Work(ctx, job)
	if err == nil {
		t.Error("expected error for empty To, got nil")
	}
}
