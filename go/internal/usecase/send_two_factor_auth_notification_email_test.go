package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// mockTwoFactorAuthNotificationEmailSenderはテスト用のモック
type mockTwoFactorAuthNotificationEmailSender struct {
	called bool
	to     string
	locale string
	event  model.TwoFactorAuthNotificationEvent
	err    error
}

func (m *mockTwoFactorAuthNotificationEmailSender) Send(_ context.Context, to, locale string, event model.TwoFactorAuthNotificationEvent) error {
	m.called = true
	m.to = to
	m.locale = locale
	m.event = event
	return m.err
}

// TestSendTwoFactorAuthNotificationEmailUsecase_Executeは、送信先と知っている種類があるときだけ送り、
// 送信の失敗をエラーとして返す (ジョブを再試行させる) ことを検証する。
func TestSendTwoFactorAuthNotificationEmailUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SendTwoFactorAuthNotificationEmailInput
		senderErr  error
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "有効化",
			input:      SendTwoFactorAuthNotificationEmailInput{Email: "test@example.com", Locale: "ja", Event: model.TwoFactorAuthNotificationEventEnabled},
			wantCalled: true,
		},
		{
			name:       "管理者による無効化",
			input:      SendTwoFactorAuthNotificationEmailInput{Email: "test@example.com", Locale: "en", Event: model.TwoFactorAuthNotificationEventDisabledByAdmin},
			wantCalled: true,
		},
		{
			name:    "メールアドレスが空",
			input:   SendTwoFactorAuthNotificationEmailInput{Email: "", Locale: "ja", Event: model.TwoFactorAuthNotificationEventEnabled},
			wantErr: true,
		},
		{
			name:    "未知の種類",
			input:   SendTwoFactorAuthNotificationEmailInput{Email: "test@example.com", Locale: "ja", Event: "unknown"},
			wantErr: true,
		},
		{
			name:       "送信の失敗",
			input:      SendTwoFactorAuthNotificationEmailInput{Email: "test@example.com", Locale: "ja", Event: model.TwoFactorAuthNotificationEventDisabled},
			senderErr:  errors.New("送信に失敗"),
			wantErr:    true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &mockTwoFactorAuthNotificationEmailSender{err: tt.senderErr}
			err := NewSendTwoFactorAuthNotificationEmailUsecase(sender).Execute(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute()のエラー = %v、エラーの有無の期待値 = %t", err, tt.wantErr)
			}
			if sender.called != tt.wantCalled {
				t.Errorf("Send()を呼んだか = %t、期待値 = %t", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && (sender.to != tt.input.Email || sender.locale != tt.input.Locale || sender.event != tt.input.Event) {
				t.Errorf("Send()の引数 = (%q, %q, %q)、期待値 = (%q, %q, %q)", sender.to, sender.locale, sender.event, tt.input.Email, tt.input.Locale, tt.input.Event)
			}
		})
	}
}
