package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestCreateSignInValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             CreateSignInValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
		description       string
	}{
		{
			name:        "正常系",
			input:       CreateSignInValidatorInput{Email: "user@example.com"},
			wantErrors:  false,
			description: "有効なメールアドレスの場合、エラーなし",
		},
		{
			name:       "メールアドレスが空文字列",
			input:      CreateSignInValidatorInput{Email: ""},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "空文字列の場合、エラーが返される",
		},
		{
			name:       "メールアドレスがwhitespaceのみ",
			input:      CreateSignInValidatorInput{Email: "   "},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "whitespaceのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスがタブのみ",
			input:      CreateSignInValidatorInput{Email: "\t\t"},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "タブのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが改行のみ",
			input:      CreateSignInValidatorInput{Email: "\n\n"},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "改行のみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが混合whitespace",
			input:      CreateSignInValidatorInput{Email: " \t\n "},
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "混合whitespaceの場合、エラーが返される",
		},
	}

	validator := NewCreateSignInValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

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
						actualMsg := actualMsgs[0]
						if actualMsg != expectedMsg {
							t.Errorf("フィールド %s のエラーメッセージが一致しません\n期待: %q\n実際: %q", field, expectedMsg, actualMsg)
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

// TestCreateSignInValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestCreateSignInValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewCreateSignInValidator()

	t.Run("email必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateSignInValidatorInput{Email: ""}
		result := validator.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_in_email_required")
		actualMsgs, exists := result.FormErrors.Fields["email"]
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
