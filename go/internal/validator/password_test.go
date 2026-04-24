package validator

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

func TestPasswordUpdateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             PasswordUpdateValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
	}{
		{
			name: "正常系",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "newpassword123",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: false,
		},
		{
			name: "トークンが空文字列",
			input: PasswordUpdateValidatorInput{
				Token:                "",
				Password:             "newpassword123",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"token"},
			wantErrorMessages: map[string]string{
				"token": "無効なリンクです",
			},
		},
		{
			name: "トークンがwhitespaceのみ",
			input: PasswordUpdateValidatorInput{
				Token:                "   ",
				Password:             "newpassword123",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"token"},
			wantErrorMessages: map[string]string{
				"token": "無効なリンクです",
			},
		},
		{
			name: "パスワードが空文字列",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "新しいパスワードを入力してください",
			},
		},
		{
			name: "パスワードがwhitespaceのみ",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "   ",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "新しいパスワードを入力してください",
			},
		},
		{
			name: "パスワード確認が空文字列",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "newpassword123",
				PasswordConfirmation: "",
			},
			wantErrors: true,
			wantFields: []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "新しいパスワード（確認）を入力してください",
			},
		},
		{
			name: "パスワード確認がwhitespaceのみ",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "newpassword123",
				PasswordConfirmation: "   ",
			},
			wantErrors: true,
			wantFields: []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "新しいパスワード（確認）を入力してください",
			},
		},
		{
			name: "パスワードが一致しない",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "newpassword123",
				PasswordConfirmation: "differentpassword",
			},
			wantErrors: true,
			wantFields: []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "パスワードが一致しません",
			},
		},
		{
			name: "すべてのフィールドが空",
			input: PasswordUpdateValidatorInput{
				Token:                "",
				Password:             "",
				PasswordConfirmation: "",
			},
			wantErrors: true,
			wantFields: []string{"token", "password", "password_confirmation"},
		},
		{
			name: "すべてのフィールドがwhitespace",
			input: PasswordUpdateValidatorInput{
				Token:                "   ",
				Password:             "   ",
				PasswordConfirmation: "   ",
			},
			wantErrors: true,
			wantFields: []string{"token", "password", "password_confirmation"},
		},
		{
			name: "トークンがタブのみ",
			input: PasswordUpdateValidatorInput{
				Token:                "\t\t",
				Password:             "newpassword123",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"token"},
		},
		{
			name: "パスワードが改行のみ",
			input: PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             "\n\n",
				PasswordConfirmation: "newpassword123",
			},
			wantErrors: true,
			wantFields: []string{"password"},
		},
	}

	v := NewPasswordUpdateValidator()

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

// TestPasswordUpdateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestPasswordUpdateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewPasswordUpdateValidator()

	t.Run("token必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := PasswordUpdateValidatorInput{
			Token:                "",
			Password:             "newpassword123",
			PasswordConfirmation: "newpassword123",
		}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "password_reset_token_invalid")
		actualMsgs, exists := ve.Fields["token"]
		if !exists {
			t.Fatal("tokenフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("tokenフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})

	t.Run("password必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := PasswordUpdateValidatorInput{
			Token:                "valid_token",
			Password:             "",
			PasswordConfirmation: "newpassword123",
		}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_required")
		actualMsgs, exists := ve.Fields["password"]
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

	t.Run("password_confirmation必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := PasswordUpdateValidatorInput{
			Token:                "valid_token",
			Password:             "newpassword123",
			PasswordConfirmation: "",
		}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_confirmation_required")
		actualMsgs, exists := ve.Fields["password_confirmation"]
		if !exists {
			t.Fatal("password_confirmationフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("password_confirmationフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})

	t.Run("password不一致エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := PasswordUpdateValidatorInput{
			Token:                "valid_token",
			Password:             "newpassword123",
			PasswordConfirmation: "differentpassword",
		}
		err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_mismatch")
		actualMsgs, exists := ve.Fields["password_confirmation"]
		if !exists {
			t.Fatal("password_confirmationフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("password_confirmationフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})
}

// TestPasswordUpdateValidator_PasswordStrength はパスワード強度エラーが i18n 翻訳に解決されることを検証する
func TestPasswordUpdateValidator_PasswordStrength(t *testing.T) {
	t.Parallel()

	v := NewPasswordUpdateValidator()

	tests := []struct {
		name           string
		password       string
		translationKey string
		templateData   map[string]any
	}{
		{
			name:           "最小文字数未満",
			password:       "short1!",
			translationKey: "password_strength_min_length",
			templateData:   map[string]any{"MinLength": auth.MinPasswordLength},
		},
		{
			name:           "最大文字数超過",
			password:       strings.Repeat("a", auth.MaxPasswordLength+1),
			translationKey: "password_strength_max_length",
			templateData:   map[string]any{"MaxLength": auth.MaxPasswordLength},
		},
		{
			name:           "印字可能ASCII以外（スペース）",
			password:       "password with space",
			translationKey: "password_strength_invalid_chars",
		},
		{
			name:           "印字可能ASCII以外（日本語）",
			password:       "パスワード12345",
			translationKey: "password_strength_invalid_chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			input := PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             tt.password,
				PasswordConfirmation: tt.password,
			}
			err := v.Validate(ctx, input)
			ve := model.AsValidationError(err)

			if ve == nil {
				t.Fatal("ValidationError が期待されましたが返されませんでした")
			}

			var expectedMsg string
			if tt.templateData != nil {
				expectedMsg = i18n.T(ctx, tt.translationKey, tt.templateData)
			} else {
				expectedMsg = i18n.T(ctx, tt.translationKey)
			}
			actualMsgs, exists := ve.Fields["password"]
			if !exists {
				t.Fatal("password フィールドのエラーが見つかりませんでした")
			}
			if len(actualMsgs) == 0 {
				t.Fatal("password フィールドのエラーメッセージが空です")
			}
			if actualMsgs[0] != expectedMsg {
				t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
			}
		})
	}
}

// TestPasswordUpdateValidator_PasswordStrengthLocales は日本語・英語ロケールで翻訳が切り替わることを検証する
func TestPasswordUpdateValidator_PasswordStrengthLocales(t *testing.T) {
	t.Parallel()

	v := NewPasswordUpdateValidator()

	tests := []struct {
		name        string
		locale      string
		password    string
		expectedMsg string
	}{
		{
			name:        "最小長エラー（日本語）",
			locale:      "ja",
			password:    "short1!",
			expectedMsg: fmt.Sprintf("パスワードは%d文字以上である必要があります", auth.MinPasswordLength),
		},
		{
			name:        "最大長エラー（日本語）",
			locale:      "ja",
			password:    strings.Repeat("a", auth.MaxPasswordLength+1),
			expectedMsg: fmt.Sprintf("パスワードは%d文字以内である必要があります", auth.MaxPasswordLength),
		},
		{
			name:        "無効な文字エラー（日本語）",
			locale:      "ja",
			password:    "password with space",
			expectedMsg: "パスワードは半角英数記号のみ使用できます",
		},
		{
			name:        "最小長エラー（英語）",
			locale:      "en",
			password:    "short1!",
			expectedMsg: fmt.Sprintf("Password must be at least %d characters long", auth.MinPasswordLength),
		},
		{
			name:        "最大長エラー（英語）",
			locale:      "en",
			password:    strings.Repeat("a", auth.MaxPasswordLength+1),
			expectedMsg: fmt.Sprintf("Password must be no more than %d characters long", auth.MaxPasswordLength),
		},
		{
			name:        "無効な文字エラー（英語）",
			locale:      "en",
			password:    "password with space",
			expectedMsg: "Password can only use alphanumeric characters and symbols",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			input := PasswordUpdateValidatorInput{
				Token:                "valid_token_123",
				Password:             tt.password,
				PasswordConfirmation: tt.password,
			}
			err := v.Validate(ctx, input)
			ve := model.AsValidationError(err)

			if ve == nil {
				t.Fatal("ValidationError が期待されましたが返されませんでした")
			}

			actualMsgs, exists := ve.Fields["password"]
			if !exists {
				t.Fatal("password フィールドのエラーが見つかりませんでした")
			}
			if len(actualMsgs) == 0 {
				t.Fatal("password フィールドのエラーメッセージが空です")
			}
			if actualMsgs[0] != tt.expectedMsg {
				t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", tt.expectedMsg, actualMsgs[0])
			}
		})
	}
}
