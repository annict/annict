package usecase

import (
	"context"
	"errors"
	"testing"
)

// mockSignInCodeEmailSender はテスト用のモック
type mockSignInCodeEmailSender struct {
	called bool
	to     string
	code   string
	locale string
	err    error
}

func (m *mockSignInCodeEmailSender) Send(_ context.Context, to, code, locale string) error {
	m.called = true
	m.to = to
	m.code = code
	m.locale = locale
	return m.err
}

func TestSendSignInCodeEmailUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SendSignInCodeEmailInput
		senderErr  error
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "正常系: 日本語ロケール",
			input:      SendSignInCodeEmailInput{Email: "test@example.com", Code: "123456", Locale: "ja"},
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "正常系: 英語ロケール",
			input:      SendSignInCodeEmailInput{Email: "test@example.com", Code: "123456", Locale: "en"},
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "異常系: Email が空",
			input:      SendSignInCodeEmailInput{Email: "", Code: "123456", Locale: "ja"},
			wantErr:    true,
			wantCalled: false,
		},
		{
			name:       "異常系: Sender がエラー",
			input:      SendSignInCodeEmailInput{Email: "test@example.com", Code: "123456", Locale: "ja"},
			senderErr:  errors.New("メール送信エラー"),
			wantErr:    true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &mockSignInCodeEmailSender{err: tt.senderErr}
			uc := NewSendSignInCodeEmailUsecase(sender)

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
				if sender.code != tt.input.Code {
					t.Errorf("code = %s, want %s", sender.code, tt.input.Code)
				}
				if sender.locale != tt.input.Locale {
					t.Errorf("locale = %s, want %s", sender.locale, tt.input.Locale)
				}
			}
		})
	}
}
