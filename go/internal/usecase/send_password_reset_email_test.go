package usecase

import (
	"context"
	"errors"
	"testing"
)

// mockPasswordResetEmailSender はテスト用のモック
type mockPasswordResetEmailSender struct {
	called   bool
	to       string
	resetURL string
	locale   string
	err      error
}

func (m *mockPasswordResetEmailSender) Send(_ context.Context, to, resetURL, locale string) error {
	m.called = true
	m.to = to
	m.resetURL = resetURL
	m.locale = locale
	return m.err
}

func TestSendPasswordResetEmailUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SendPasswordResetEmailInput
		senderErr  error
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "正常系: 日本語ロケール",
			input:      SendPasswordResetEmailInput{Email: "test@example.com", ResetURL: "https://example.com/password/edit?token=abc", Locale: "ja"},
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "正常系: 英語ロケール",
			input:      SendPasswordResetEmailInput{Email: "test@example.com", ResetURL: "https://example.com/password/edit?token=abc", Locale: "en"},
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "異常系: Email が空",
			input:      SendPasswordResetEmailInput{Email: "", ResetURL: "https://example.com/password/edit?token=abc", Locale: "ja"},
			wantErr:    true,
			wantCalled: false,
		},
		{
			name:       "異常系: Sender がエラー",
			input:      SendPasswordResetEmailInput{Email: "test@example.com", ResetURL: "https://example.com/password/edit?token=abc", Locale: "ja"},
			senderErr:  errors.New("メール送信エラー"),
			wantErr:    true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &mockPasswordResetEmailSender{err: tt.senderErr}
			uc := NewSendPasswordResetEmailUsecase(sender)

			err := uc.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if sender.called != tt.wantCalled {
				t.Errorf("Send called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && !tt.wantErr {
				if sender.to != tt.input.Email {
					t.Errorf("to = %s, want %s", sender.to, tt.input.Email)
				}
				if sender.resetURL != tt.input.ResetURL {
					t.Errorf("resetURL = %s, want %s", sender.resetURL, tt.input.ResetURL)
				}
				if sender.locale != tt.input.Locale {
					t.Errorf("locale = %s, want %s", sender.locale, tt.input.Locale)
				}
			}
		})
	}
}
