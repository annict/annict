package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// TestSignInTwoFactorRecoveryCreateValidatorValidateは、発行するリカバリーコードの形 (区切りを除いた8文字) だけを受け付け、
// 空と形の違うコードをコードの欄のエラーにすることを検証する。
func TestSignInTwoFactorRecoveryCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	v := NewSignInTwoFactorRecoveryCreateValidator()

	tests := []struct {
		name           string
		code           string
		wantFieldError string
	}{
		{name: "リカバリーコードの形", code: "abcd2345"},
		{name: "空", code: "", wantFieldError: "入力してください"},
		{name: "7文字", code: "abcd234", wantFieldError: "リカバリーコード (xxxx-xxxxの形) を入力してください"},
		{name: "使わない文字を含む", code: "abcd2340", wantFieldError: "リカバリーコード (xxxx-xxxxの形) を入力してください"},
		{name: "認証アプリのコードの形", code: "123456", wantFieldError: "リカバリーコード (xxxx-xxxxの形) を入力してください"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			err := v.Validate(ctx, SignInTwoFactorRecoveryCreateValidatorInput{Code: tt.code})

			if tt.wantFieldError == "" {
				if err != nil {
					t.Fatalf("Validate()のエラー = %v", err)
				}
				return
			}
			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("Validate()のエラー = %v、ValidationErrorを期待", err)
			}
			if got := ve.GetFieldErrors("code"); len(got) != 1 || got[0] != tt.wantFieldError {
				t.Errorf("codeのフィールドエラー = %v、期待値 = [%q]", got, tt.wantFieldError)
			}
		})
	}
}
