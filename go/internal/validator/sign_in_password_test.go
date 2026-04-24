package validator

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestSignInPasswordCreateValidatorValidate_FormatErrors(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(query.New(db)).WithTx(tx)

	tests := []struct {
		name              string
		input             SignInPasswordCreateValidatorInput
		wantFieldErrors   []string
		wantErrorMessages map[string]string
	}{
		{
			name:            "パスワードが空",
			input:           SignInPasswordCreateValidatorInput{EmailOrUsername: "user@example.com", Password: ""},
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "パスワードがwhitespaceのみ",
			input:           SignInPasswordCreateValidatorInput{EmailOrUsername: "user@example.com", Password: "   "},
			wantFieldErrors: []string{"password"},
			wantErrorMessages: map[string]string{
				"password": "パスワードを入力してください",
			},
		},
		{
			name:            "メールアドレスが空",
			input:           SignInPasswordCreateValidatorInput{EmailOrUsername: "", Password: "password123"},
			wantFieldErrors: []string{"email_or_username"},
		},
		{
			name:            "両方空",
			input:           SignInPasswordCreateValidatorInput{EmailOrUsername: "", Password: ""},
			wantFieldErrors: []string{"email_or_username", "password"},
		},
	}

	v := NewSignInPasswordCreateValidator(userRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			output, err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if ve == nil {
				t.Fatal("エラーが期待されましたが、エラーがありませんでした")
			}
			if output != nil {
				t.Error("エラー時は output が nil になるべきです")
			}

			for _, field := range tt.wantFieldErrors {
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
		})
	}
}

func TestSignInPasswordCreateValidatorValidate_StateErrors(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(query.New(db)).WithTx(tx)

	// テストユーザーを作成
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("validator_password_test").
		WithEmail("validator_password_test@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	v := NewSignInPasswordCreateValidator(userRepo)

	t.Run("正常系（メールアドレス）", func(t *testing.T) {
		ctx := context.Background()
		output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
			EmailOrUsername: "validator_password_test@example.com",
			Password:        "correctpassword",
		})
		if err != nil {
			t.Errorf("エラーは期待されていませんでしたが、返されました: %v", err)
		}
		if output == nil {
			t.Fatal("成功時は output が返されるべきです")
		}
		if output.User.Email != "validator_password_test@example.com" {
			t.Errorf("User.Email = %q, want %q", output.User.Email, "validator_password_test@example.com")
		}
	})

	t.Run("正常系（ユーザー名）", func(t *testing.T) {
		ctx := context.Background()
		output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
			EmailOrUsername: "validator_password_test",
			Password:        "correctpassword",
		})
		if err != nil {
			t.Errorf("エラーは期待されていませんでしたが、返されました: %v", err)
		}
		if output == nil {
			t.Fatal("成功時は output が返されるべきです")
		}
		if output.User.Username != "validator_password_test" {
			t.Errorf("User.Username = %q, want %q", output.User.Username, "validator_password_test")
		}
	})

	t.Run("ユーザーが見つからない", func(t *testing.T) {
		ctx := context.Background()
		output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
			EmailOrUsername: "nonexistent@example.com",
			Password:        "password123",
		})
		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatal("バリデーションエラーが期待されましたが、返されませんでした")
		}
		if output != nil {
			t.Error("エラー時は output が nil になるべきです")
		}
		if len(ve.Global) == 0 {
			t.Error("グローバルエラーが期待されましたが、ありませんでした")
		}
	})

	t.Run("パスワードが一致しない", func(t *testing.T) {
		ctx := context.Background()
		output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
			EmailOrUsername: "validator_password_test@example.com",
			Password:        "wrongpassword",
		})
		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatal("バリデーションエラーが期待されましたが、返されませんでした")
		}
		if output != nil {
			t.Error("エラー時は output が nil になるべきです")
		}
		if len(ve.Global) == 0 {
			t.Error("グローバルエラーが期待されましたが、ありませんでした")
		}
	})
}

// TestSignInPasswordCreateValidator_ValidateI18nMessages I18nメッセージの内容を検証するテスト
func TestSignInPasswordCreateValidator_ValidateI18nMessages(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(query.New(db)).WithTx(tx)

	ctx := context.Background()
	v := NewSignInPasswordCreateValidator(userRepo)

	t.Run("password必須エラーメッセージ", func(t *testing.T) {
		input := SignInPasswordCreateValidatorInput{EmailOrUsername: "user@example.com", Password: ""}
		output, err := v.Validate(ctx, input)
		ve := model.AsValidationError(err)

		if ve == nil {
			t.Fatal("エラーが期待されましたが、エラーがありませんでした")
		}
		if output != nil {
			t.Error("エラー時は output が nil になるべきです")
		}

		expectedMsg := i18n.T(ctx, "sign_in_error_password_required")
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
}
