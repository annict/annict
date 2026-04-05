package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestCreateSignUpUsernameValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		input             CreateSignUpUsernameValidatorInput
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
	}{
		{
			name:       "正常: 有効なトークンとユーザー名",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "testuser"},
			wantErrors: false,
		},
		{
			name:       "正常: アンダースコアを含むユーザー名",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "test_user_1"},
			wantErrors: false,
		},
		{
			name:       "正常: 20文字のユーザー名",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "12345678901234567890"},
			wantErrors: false,
		},
		{
			name:       "エラー: トークンが空",
			input:      CreateSignUpUsernameValidatorInput{Token: "", Username: "testuser"},
			wantErrors: true,
			wantFields: []string{"token"},
		},
		{
			name:       "エラー: ユーザー名が空",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: ""},
			wantErrors: true,
			wantFields: []string{"username"},
		},
		{
			name:       "エラー: ユーザー名が21文字以上",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "123456789012345678901"},
			wantErrors: true,
			wantFields: []string{"username"},
		},
		{
			name:       "エラー: ユーザー名にハイフンを含む",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "test-user"},
			wantErrors: true,
			wantFields: []string{"username"},
		},
		{
			name:       "エラー: ユーザー名に日本語を含む",
			input:      CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "テストユーザー"},
			wantErrors: true,
			wantFields: []string{"username"},
		},
		{
			name:       "エラー: トークンとユーザー名の両方が空",
			input:      CreateSignUpUsernameValidatorInput{Token: "", Username: ""},
			wantErrors: true,
			wantFields: []string{"token", "username"},
		},
	}

	v := NewCreateSignUpUsernameValidator()

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

// TestCreateSignUpUsernameValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestCreateSignUpUsernameValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := NewCreateSignUpUsernameValidator()

	t.Run("token必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateSignUpUsernameValidatorInput{Token: "", Username: "testuser"}
		result := v.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_username_error_token_missing")
		actualMsgs, exists := result.FormErrors.Fields["token"]
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

	t.Run("username必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: ""}
		result := v.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_username_error_username_required")
		actualMsgs, exists := result.FormErrors.Fields["username"]
		if !exists {
			t.Fatal("usernameフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("usernameフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})

	t.Run("usernameフォーマットエラーメッセージ", func(t *testing.T) {
		t.Parallel()

		input := CreateSignUpUsernameValidatorInput{Token: "valid_token", Username: "invalid-name!"}
		result := v.Validate(ctx, input)

		if result.FormErrors == nil || !result.FormErrors.HasErrors() {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}

		expectedMsg := i18n.T(ctx, "sign_up_username_error_username_format")
		actualMsgs, exists := result.FormErrors.Fields["username"]
		if !exists {
			t.Fatal("usernameフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("usernameフィールドのエラーメッセージが空です")
		}
		if actualMsgs[0] != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsgs[0])
		}
	})
}
