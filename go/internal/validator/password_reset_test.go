package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestCreatePasswordResetValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             CreatePasswordResetValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
	}{
		{
			name: "正常系",
			input: CreatePasswordResetValidatorInput{
				Email: "user@example.com",
			},
			wantErrors: false,
		},
		{
			name: "メールアドレスが空文字列",
			input: CreatePasswordResetValidatorInput{
				Email: "",
			},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
		},
		{
			name: "メールアドレスがwhitespaceのみ",
			input: CreatePasswordResetValidatorInput{
				Email: "   ",
			},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
		},
		{
			name: "メールアドレスがタブのみ",
			input: CreatePasswordResetValidatorInput{
				Email: "\t\t",
			},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
		},
		{
			name: "メールアドレスが改行のみ",
			input: CreatePasswordResetValidatorInput{
				Email: "\n\n",
			},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
		},
		{
			name: "メールアドレスが混合whitespace",
			input: CreatePasswordResetValidatorInput{
				Email: " \t\n ",
			},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
		},
	}

	v := NewCreatePasswordResetValidator()

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

				for _, field := range tt.wantFields {
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

// TestCreatePasswordResetValidator_ValidateI18nMessages はI18nメッセージの内容を検証するテスト
func TestCreatePasswordResetValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewCreatePasswordResetValidator()

	t.Run("email必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreatePasswordResetValidatorInput{
			Email: "",
		}
		result := v.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "password_reset_email_required")
		actualMsgs, exists := result.FormErrors.Fields["email"]
		if !exists {
			t.Fatal("emailフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("emailフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})
}
