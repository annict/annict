package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// TestSettingsTwoFactorAuthCreateValidatorValidateは、6桁の数字だけを受け付け、
// 空と形の違うコードをコードの欄のエラーにすることを検証する。
func TestSettingsTwoFactorAuthCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	v := NewSettingsTwoFactorAuthCreateValidator()

	tests := []struct {
		name           string
		code           string
		wantFieldError string
	}{
		{name: "6桁の数字", code: "012345"},
		{name: "空", code: "", wantFieldError: "入力してください"},
		{name: "5桁の数字", code: "01234", wantFieldError: "認証アプリに表示された6桁の数字を入力してください"},
		{name: "英字を含む", code: "01234a", wantFieldError: "認証アプリに表示された6桁の数字を入力してください"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			err := v.Validate(ctx, SettingsTwoFactorAuthCreateValidatorInput{Code: tt.code})

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

// TestSettingsTwoFactorAuthDeleteValidatorValidateは、本人確認の入力が空のときだけエラーにすることを検証する。
// どの形でもない入力はパスワードとして照合するため、形式は確かめない。
func TestSettingsTwoFactorAuthDeleteValidatorValidate(t *testing.T) {
	t.Parallel()

	v := NewSettingsTwoFactorAuthDeleteValidator()
	ctx := i18n.SetLocale(context.Background(), "ja")

	for _, credential := range []string{"123456", "abcd-2345", "a"} {
		if err := v.Validate(ctx, SettingsTwoFactorAuthDeleteValidatorInput{Credential: credential}); err != nil {
			t.Errorf("Validate(%q)のエラー = %v", credential, err)
		}
	}

	ve := model.AsValidationError(v.Validate(ctx, SettingsTwoFactorAuthDeleteValidatorInput{Credential: ""}))
	if ve == nil {
		t.Fatal("空のValidate()のエラー = nil、ValidationErrorを期待")
	}
	if got := ve.GetFieldErrors("credential"); len(got) != 1 || got[0] != "入力してください" {
		t.Errorf("credentialのフィールドエラー = %v、期待値 = [\"入力してください\"]", got)
	}
}
