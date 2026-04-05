package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestCreateSignInPasswordValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             CreateSignInPasswordValidatorInput
		wantErrors        bool
		wantFieldErrors   []string
		wantErrorMessages map[string]string
	}{
		{
			name:       "正常系",
			input:      CreateSignInPasswordValidatorInput{Password: "password123"},
			wantErrors: false,
		},
		{
			name:            "パスワードが空",
			input:           CreateSignInPasswordValidatorInput{Password: ""},
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "パスワードがwhitespaceのみ",
			input:           CreateSignInPasswordValidatorInput{Password: "   "},
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "パスワードがタブとスペース",
			input:           CreateSignInPasswordValidatorInput{Password: " \t "},
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
	}

	v := NewCreateSignInPasswordValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := v.Validate(ctx, tt.input)

			if tt.wantErrors {
				if result.FormErrors == nil || !result.FormErrors.HasErrors() {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFieldErrors {
					if _, exists := result.FormErrors.Fields[field]; !exists {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}

				if tt.wantErrorMessages != nil {
					for field, expectedMsg := range tt.wantErrorMessages {
						actualMsgs, exists := result.FormErrors.Fields[field]
						if !exists {
							t.Errorf("フィールド %s のエラーメッセージが見つかりませんでした", field)
							continue
						}
						if len(actualMsgs) == 0 {
							t.Errorf("フィールド %s のエラーメッセージが空です", field)
							continue
						}
						if actualMsgs[0] != expectedMsg {
							t.Errorf("フィールド %s のエラーメッセージが一致しません\n期待: %q\n実際: %q", field, expectedMsg, actualMsgs[0])
						}
					}
				}
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}

// TestCreateSignInPasswordValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestCreateSignInPasswordValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewCreateSignInPasswordValidator()

	t.Run("password必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateSignInPasswordValidatorInput{Password: ""}
		result := v.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_in_error_password_required")
		actualMsgs, exists := result.FormErrors.Fields["password"]
		if !exists {
			t.Fatal("passwordフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("passwordフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})
}
