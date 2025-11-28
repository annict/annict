package password

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

func TestRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		token                string
		password             string
		passwordConfirmation string
		wantErrors           bool
		wantFields           []string
		wantErrorMessages    map[string]string
		description          string
	}{
		{
			name:                 "正常系",
			token:                "valid_token_123",
			password:             "newpassword123",
			passwordConfirmation: "newpassword123",
			wantErrors:           false,
			description:          "すべてのフィールドが有効な場合、エラーなし",
		},
		{
			name:                 "トークンが空文字列",
			token:                "",
			password:             "newpassword123",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"token"},
			wantErrorMessages: map[string]string{
				"token": "無効なリンクです",
			},
			description: "トークンが空の場合、エラーが返される",
		},
		{
			name:                 "トークンがwhitespaceのみ",
			token:                "   ",
			password:             "newpassword123",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"token"},
			wantErrorMessages: map[string]string{
				"token": "無効なリンクです",
			},
			description: "トークンがwhitespaceのみの場合、エラーが返される",
		},
		{
			name:                 "パスワードが空文字列",
			token:                "valid_token_123",
			password:             "",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "新しいパスワードを入力してください",
			},
			description: "パスワードが空の場合、エラーが返される",
		},
		{
			name:                 "パスワードがwhitespaceのみ",
			token:                "valid_token_123",
			password:             "   ",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "新しいパスワードを入力してください",
			},
			description: "パスワードがwhitespaceのみの場合、エラーが返される",
		},
		{
			name:                 "パスワード確認が空文字列",
			token:                "valid_token_123",
			password:             "newpassword123",
			passwordConfirmation: "",
			wantErrors:           true,
			wantFields:           []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "新しいパスワード（確認）を入力してください",
			},
			description: "パスワード確認が空の場合、エラーが返される",
		},
		{
			name:                 "パスワード確認がwhitespaceのみ",
			token:                "valid_token_123",
			password:             "newpassword123",
			passwordConfirmation: "   ",
			wantErrors:           true,
			wantFields:           []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "新しいパスワード（確認）を入力してください",
			},
			description: "パスワード確認がwhitespaceのみの場合、エラーが返される",
		},
		{
			name:                 "パスワードが一致しない",
			token:                "valid_token_123",
			password:             "newpassword123",
			passwordConfirmation: "differentpassword",
			wantErrors:           true,
			wantFields:           []string{"password_confirmation"},
			wantErrorMessages: map[string]string{
				"password_confirmation": "パスワードが一致しません",
			},
			description: "パスワードが一致しない場合、エラーが返される",
		},
		{
			name:                 "すべてのフィールドが空",
			token:                "",
			password:             "",
			passwordConfirmation: "",
			wantErrors:           true,
			wantFields:           []string{"token", "password", "password_confirmation"},
			description:          "すべてのフィールドが空の場合、複数のエラーが返される",
		},
		{
			name:                 "すべてのフィールドがwhitespace",
			token:                "   ",
			password:             "   ",
			passwordConfirmation: "   ",
			wantErrors:           true,
			wantFields:           []string{"token", "password", "password_confirmation"},
			description:          "すべてのフィールドがwhitespaceの場合、複数のエラーが返される",
		},
		{
			name:                 "トークンがタブのみ",
			token:                "\t\t",
			password:             "newpassword123",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"token"},
			description:          "トークンがタブのみの場合、エラーが返される",
		},
		{
			name:                 "パスワードが改行のみ",
			token:                "valid_token_123",
			password:             "\n\n",
			passwordConfirmation: "newpassword123",
			wantErrors:           true,
			wantFields:           []string{"password"},
			description:          "パスワードが改行のみの場合、エラーが返される",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &Request{
				Token:                tt.token,
				Password:             tt.password,
				PasswordConfirmation: tt.passwordConfirmation,
			}

			ctx := context.Background()
			errors := req.Validate(ctx)

			if tt.wantErrors {
				if errors == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
					return
				}

				// フィールドエラーの確認
				for _, field := range tt.wantFields {
					if _, exists := errors.Fields[field]; !exists {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}

				// I18nメッセージの内容検証
				if tt.wantErrorMessages != nil {
					for field, expectedMsg := range tt.wantErrorMessages {
						actualMsgs, exists := errors.Fields[field]
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
				if errors != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", errors)
				}
			}
		})
	}
}

// TestRequest_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestRequest_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// トークン必須エラーのメッセージ検証
	t.Run("token必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &Request{
			Token:                "",
			Password:             "newpassword123",
			PasswordConfirmation: "newpassword123",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "password_reset_token_invalid")
		actualMsgs, exists := errors.Fields["token"]
		if !exists {
			t.Fatal("tokenフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("tokenフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})

	// パスワード必須エラーのメッセージ検証
	t.Run("password必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &Request{
			Token:                "valid_token",
			Password:             "",
			PasswordConfirmation: "newpassword123",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_required")
		actualMsgs, exists := errors.Fields["password"]
		if !exists {
			t.Fatal("passwordフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("passwordフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})

	// パスワード確認必須エラーのメッセージ検証
	t.Run("password_confirmation必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &Request{
			Token:                "valid_token",
			Password:             "newpassword123",
			PasswordConfirmation: "",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_confirmation_required")
		actualMsgs, exists := errors.Fields["password_confirmation"]
		if !exists {
			t.Fatal("password_confirmationフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("password_confirmationフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})

	// パスワード不一致エラーのメッセージ検証
	t.Run("password不一致エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &Request{
			Token:                "valid_token",
			Password:             "newpassword123",
			PasswordConfirmation: "differentpassword",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "password_reset_password_mismatch")
		actualMsgs, exists := errors.Fields["password_confirmation"]
		if !exists {
			t.Fatal("password_confirmationフィールドのエラーが見つかりませんでした")
		}
		if len(actualMsgs) == 0 {
			t.Fatal("password_confirmationフィールドのエラーメッセージが空です")
		}
		actualMsg := actualMsgs[0]
		if actualMsg != expectedMsg {
			t.Errorf("エラーメッセージが一致しません\n期待: %q\n実際: %q", expectedMsg, actualMsg)
		}
	})
}
