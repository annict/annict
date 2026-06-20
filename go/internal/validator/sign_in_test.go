package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

func TestSignInCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             SignInCreateValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
		description       string
	}{
		{
			name:        "正常系",
			input:       SignInCreateValidatorInput{Email: "user@example.com"},
			wantErrors:  false,
			description: "有効なメールアドレスの場合、エラーなし",
		},
		{
			name:       "メールアドレスが空文字列",
			input:      SignInCreateValidatorInput{Email: ""},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "空文字列の場合、エラーが返される",
		},
		{
			name:       "メールアドレスがwhitespaceのみ",
			input:      SignInCreateValidatorInput{Email: "   "},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "whitespaceのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスがタブのみ",
			input:      SignInCreateValidatorInput{Email: "\t\t"},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "タブのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが改行のみ",
			input:      SignInCreateValidatorInput{Email: "\n\n"},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "改行のみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが混合whitespace",
			input:      SignInCreateValidatorInput{Email: " \t\n "},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "混合whitespaceの場合、エラーが返される",
		},
	}

	v := NewSignInCreateValidator()

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

// TestSignInCreateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestSignInCreateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewSignInCreateValidator()

	t.Run("email必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := SignInCreateValidatorInput{Email: ""}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_in_error_email_required")
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
