package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// TestSignInTwoFactorCreateValidatorValidateは、6桁の数字だけを受け付け、
// 空と形の違うコードをコードの欄のエラーにすることを検証する。
func TestSignInTwoFactorCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	v := NewSignInTwoFactorCreateValidator()

	tests := []struct {
		name           string
		code           string
		wantFieldError string
	}{
		{name: "6桁の数字", code: "012345"},
		{name: "空", code: "", wantFieldError: "入力してください"},
		{name: "7桁の数字", code: "0123456", wantFieldError: "認証アプリに表示された6桁の数字を入力してください"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			err := v.Validate(ctx, SignInTwoFactorCreateValidatorInput{Code: tt.code})

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
