package usecase

import (
	"context"
	"errors"
	"testing"
)

// mockAccountDeletionEmailSenderはテスト用のモック
type mockAccountDeletionEmailSender struct {
	called            bool
	to                string
	locale            string
	isStripeSupporter bool
	err               error
}

func (m *mockAccountDeletionEmailSender) Send(_ context.Context, to, locale string, isStripeSupporter bool) error {
	m.called = true
	m.to = to
	m.locale = locale
	m.isStripeSupporter = isStripeSupporter
	return m.err
}

func TestSendAccountDeletionEmailUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SendAccountDeletionEmailInput
		senderErr  error
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "正常系: 日本語ロケール",
			input:      SendAccountDeletionEmailInput{Email: "test@example.com", Locale: "ja"},
			wantCalled: true,
		},
		{
			name:       "正常系: 英語ロケールのStripeのサポーター",
			input:      SendAccountDeletionEmailInput{Email: "test@example.com", Locale: "en", IsStripeSupporter: true},
			wantCalled: true,
		},
		{
			name:       "異常系: Emailが空",
			input:      SendAccountDeletionEmailInput{Email: "", Locale: "ja"},
			wantErr:    true,
			wantCalled: false,
		},
		{
			name:       "異常系: Senderがエラー",
			input:      SendAccountDeletionEmailInput{Email: "test@example.com", Locale: "ja"},
			senderErr:  errors.New("メール送信エラー"),
			wantErr:    true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &mockAccountDeletionEmailSender{err: tt.senderErr}
			uc := NewSendAccountDeletionEmailUsecase(sender)

			err := uc.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute()のエラー = %v、期待値 = %v", err, tt.wantErr)
			}
			if sender.called != tt.wantCalled {
				t.Errorf("Send()の呼び出し = %v、期待値 = %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && !tt.wantErr {
				if sender.to != tt.input.Email {
					t.Errorf("to = %s、期待値 = %s", sender.to, tt.input.Email)
				}
				if sender.locale != tt.input.Locale {
					t.Errorf("locale = %s、期待値 = %s", sender.locale, tt.input.Locale)
				}
				if sender.isStripeSupporter != tt.input.IsStripeSupporter {
					t.Errorf("isStripeSupporter = %v、期待値 = %v", sender.isStripeSupporter, tt.input.IsStripeSupporter)
				}
			}
		})
	}
}
