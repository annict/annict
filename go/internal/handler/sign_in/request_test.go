package sign_in

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

func TestCreateRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		email             string
		wantErrors        bool
		wantFields        []string
		wantErrorMessages map[string]string
		description       string
	}{
		{
			name:        "正常系",
			email:       "user@example.com",
			wantErrors:  false,
			description: "有効なメールアドレスの場合、エラーなし",
		},
		{
			name:       "メールアドレスが空文字列",
			email:      "",
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "空文字列の場合、エラーが返される",
		},
		{
			name:       "メールアドレスがwhitespaceのみ",
			email:      "   ",
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "whitespaceのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスがタブのみ",
			email:      "\t\t",
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "タブのみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが改行のみ",
			email:      "\n\n",
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "改行のみの場合、エラーが返される",
		},
		{
			name:       "メールアドレスが混合whitespace",
			email:      " \t\n ",
			wantErrors: true,
			wantFields: []string{"email"},
			wantErrorMessages: map[string]string{
				"email": "メールアドレスを入力してください",
			},
			description: "混合whitespaceの場合、エラーが返される",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &CreateRequest{
				Email: tt.email,
			}

			ctx := context.Background()
			errors := req.Validate(ctx)

			if tt.wantErrors {
				if errors == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
					return
				}

				for _, field := range tt.wantFields {
					if _, exists := errors.Fields[field]; !exists {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}

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

// TestCreateRequest_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestCreateRequest_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("email必須エラーメッセージ", func(t *testing.T) {
		t.Parallel()
		req := &CreateRequest{
			Email: "",
		}
		errors := req.Validate(ctx)
		if errors == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}

		expectedMsg := i18n.T(ctx, "sign_in_email_required")
		actualMsgs, exists := errors.Fields["email"]
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
