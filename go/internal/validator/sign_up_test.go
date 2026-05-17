package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

func TestSignUpCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             SignUpCreateValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
		description       string
	}{
		{
			name:        "正常系",
			input:       SignUpCreateValidatorInput{Email: "user@example.com"},
			wantErrors:  false,
			description: "有効なメールアドレスの場合、エラーなし",
		},
		{
			name:       "メールアドレスが空文字列",
			input:      SignUpCreateValidatorInput{Email: ""},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "空文字列の場合、エラーが返される",
		},
		{
			name:       "メールアドレスがwhitespaceのみ",
			input:      SignUpCreateValidatorInput{Email: "   "},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "whitespaceのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスの形式が不正",
			input:      SignUpCreateValidatorInput{Email: "invalid-email"},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスの形式が正しくありません",
			},
			description: "不正な形式の場合、エラーが返される",
		},
	}

	v := NewSignUpCreateValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if _, exists := ve.Fields[field]; !exists {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}

				if tt.wantErrorMessages != nil {
					for field, expectedMsg := range tt.wantErrorMessages {
						actualMsgs, exists := ve.Fields[field]
						if !exists {
							t.Errorf("フィールド %s のエラーメッセージが見つかりませんでした", field)
							continue
						}
						if len(actualMsgs) == 0 {
							t.Errorf("フィールド %s のエラーメッセージが空です", field)
							continue
						}
						actualMsg := actualMsgs[0]
						if actualMsg != expectedMsg {
							t.Errorf("フィールド %s のエラーメッセージが一致しません\n期待: %q\n実際: %q", field, expectedMsg, actualMsg)
						}
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}

// TestSignUpCreateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestSignUpCreateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewSignUpCreateValidator()

	t.Run("email必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := SignUpCreateValidatorInput{Email: ""}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_error_email_required")
		actualMsgs, exists := ve.Fields["email"]
		if !exists {
			t.Fatal("emailフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("emailフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})

	t.Run("email形式エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := SignUpCreateValidatorInput{Email: "invalid-email"}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_error_email_invalid")
		actualMsgs, exists := ve.Fields["email"]
		if !exists {
			t.Fatal("emailフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("emailフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})
}
