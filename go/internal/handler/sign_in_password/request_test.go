package sign_in_password

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

// TestRequest_Validate バリデーションのテスト
func TestRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		password          string
		wantErrors        bool
		wantFieldErrors   []string
		wantErrorMessages map[string]string // フィールド名 -> 期待されるエラーメッセージ
	}{
		{
			name:       "正常系",
			password:   "password123",
			wantErrors: false,
		},
		{
			name:            "パスワードが空",
			password:        "",
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "パスワードがwhitespaceのみ",
			password:        "   ",
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "パスワードがタブとスペース",
			password:        " \t ",
			wantErrors:      true,
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &Request{
				Password: tt.password,
			}

			ctx := context.Background()
			errors := req.Validate(ctx)

			if tt.wantErrors {
				if errors == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
					return
				}

				// フィールドエラーの確認
				for _, field := range tt.wantFieldErrors {
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

	// パスワード必須エラーのメッセージ検証
	t.Run("password必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &Request{
			Password: "",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "sign_in_error_password_required")
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
}
