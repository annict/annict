package sign_in_code

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             CreateValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
	}{
		{
			name:       "正常: 6桁の数字",
			input:      CreateValidatorInput{Code: "123456"},
			wantErrors: false,
		},
		{
			name:       "エラー: コードが空",
			input:      CreateValidatorInput{Code: ""},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードを入力してください",
			},
		},
		{
			name:       "エラー: 5桁の数字",
			input:      CreateValidatorInput{Code: "12345"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 7桁の数字",
			input:      CreateValidatorInput{Code: "1234567"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 6桁の英数字",
			input:      CreateValidatorInput{Code: "12345a"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 6桁のアルファベット",
			input:      CreateValidatorInput{Code: "abcdef"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: スペースを含む",
			input:      CreateValidatorInput{Code: "123 456"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
	}

	validator := NewCreateValidator()

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

// TestCreateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestCreateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewCreateValidator()

	t.Run("code必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateValidatorInput{Code: ""}
		result := validator.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_in_code_error_code_required")
		actualMsgs, exists := result.FormErrors.Fields["code"]
		if !exists {
			t.Fatal("codeフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("codeフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})

	t.Run("codeフォーマットエラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateValidatorInput{Code: "12345"}
		result := validator.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_in_code_error_code_invalid_format")
		actualMsgs, exists := result.FormErrors.Fields["code"]
		if !exists {
			t.Fatal("codeフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("codeフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})
}
