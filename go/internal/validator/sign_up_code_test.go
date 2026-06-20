package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

func TestSignUpCodeCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             SignUpCodeCreateValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
	}{
		{
			name:       "正常: 6桁の数字",
			input:      SignUpCodeCreateValidatorInput{Code: "123456"},
			wantErrors: false,
		},
		{
			name:       "エラー: コードが空",
			input:      SignUpCodeCreateValidatorInput{Code: ""},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードを入力してください",
			},
		},
		{
			name:       "エラー: 5桁の数字",
			input:      SignUpCodeCreateValidatorInput{Code: "12345"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 7桁の数字",
			input:      SignUpCodeCreateValidatorInput{Code: "1234567"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 6桁の英数字",
			input:      SignUpCodeCreateValidatorInput{Code: "12345a"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: 6桁のアルファベット",
			input:      SignUpCodeCreateValidatorInput{Code: "abcdef"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
		{
			name:       "エラー: スペースを含む",
			input:      SignUpCodeCreateValidatorInput{Code: "123 456"},
			wantErrors: true,
			wantFields: []string{"code"},
			wantErrorMessages: map[string]string{
				"code": "コードは6桁の数字で入力してください",
			},
		},
	}

	v := NewSignUpCodeCreateValidator()

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
						if actualMsgs[0] != expectedMsg {
							t.Errorf("フィールド %s のエラーメッセージが一致しません\n期待: %q\n実際: %q", field, expectedMsg, actualMsgs[0])
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

// TestSignUpCodeCreateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestSignUpCodeCreateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewSignUpCodeCreateValidator()

	t.Run("code必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := SignUpCodeCreateValidatorInput{Code: ""}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_code_error_code_required")
		actualMsgs, exists := ve.Fields["code"]
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

		input := SignUpCodeCreateValidatorInput{Code: "12345"}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_code_error_code_invalid_format")
		actualMsgs, exists := ve.Fields["code"]
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
