package model_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	if got := ve.Error(); got != "validation failed" {
		t.Errorf("Error() = %q、期待値 = %q", got, "validation failed")
	}
}

func TestValidationError_AddGlobal(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddGlobal("エラー1")
	ve.AddGlobal("エラー2")

	if len(ve.Global) != 2 {
		t.Fatalf("len(Global) = %d、期待値 = 2", len(ve.Global))
	}
	if ve.Global[0] != "エラー1" {
		t.Errorf("Global[0] = %q、期待値 = %q", ve.Global[0], "エラー1")
	}
	if ve.Global[1] != "エラー2" {
		t.Errorf("Global[1] = %q、期待値 = %q", ve.Global[1], "エラー2")
	}
}

func TestValidationError_AddField(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddField("email", "メールアドレスを入力してください")
	ve.AddField("email", "メールアドレスの形式が正しくありません")
	ve.AddField("password", "パスワードを入力してください")

	emailErrors := ve.Fields["email"]
	if len(emailErrors) != 2 {
		t.Fatalf("len(Fields[email]) = %d、期待値 = 2", len(emailErrors))
	}
	if emailErrors[0] != "メールアドレスを入力してください" {
		t.Errorf("Fields[email][0] = %q、期待値 = %q", emailErrors[0], "メールアドレスを入力してください")
	}

	passwordErrors := ve.Fields["password"]
	if len(passwordErrors) != 1 {
		t.Fatalf("len(Fields[password]) = %d、期待値 = 1", len(passwordErrors))
	}
}

func TestValidationError_AddField_NilFields(t *testing.T) {
	t.Parallel()

	ve := &model.ValidationError{}
	ve.AddField("email", "必須です")

	if !ve.HasFieldError("email") {
		t.Error("HasFieldError(email) = false、期待値 = true")
	}
}

func TestValidationError_HasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *model.ValidationError
		want  bool
	}{
		{
			name:  "エラーなし",
			setup: func() *model.ValidationError { return model.NewValidationError() },
			want:  false,
		},
		{
			name: "グローバルエラーあり",
			setup: func() *model.ValidationError {
				ve := model.NewValidationError()
				ve.AddGlobal("エラー")
				return ve
			},
			want: true,
		},
		{
			name: "フィールドエラーあり",
			setup: func() *model.ValidationError {
				ve := model.NewValidationError()
				ve.AddField("email", "必須です")
				return ve
			},
			want: true,
		},
		{
			name:  "nilレシーバー",
			setup: func() *model.ValidationError { return nil },
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ve := tt.setup()
			if got := ve.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_HasFieldError(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddField("email", "必須です")

	if !ve.HasFieldError("email") {
		t.Error("HasFieldError(email) = false、期待値 = true")
	}
	if ve.HasFieldError("password") {
		t.Error("HasFieldError(password) = true、期待値 = false")
	}

	// nilレシーバー
	var nilVE *model.ValidationError
	if nilVE.HasFieldError("email") {
		t.Error("nil.HasFieldError(email) = true、期待値 = false")
	}
}

func TestValidationError_GetFieldErrors(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddField("email", "必須です")
	ve.AddField("email", "形式が不正です")

	errs := ve.GetFieldErrors("email")
	if len(errs) != 2 {
		t.Fatalf("len(GetFieldErrors(email)) = %d、期待値 = 2", len(errs))
	}

	// 存在しないフィールド
	errs = ve.GetFieldErrors("password")
	if errs != nil {
		t.Errorf("GetFieldErrors(password) = %v、期待値 = nil", errs)
	}

	// nilレシーバー
	var nilVE *model.ValidationError
	errs = nilVE.GetFieldErrors("email")
	if errs != nil {
		t.Errorf("nil.GetFieldErrors(email) = %v、期待値 = nil", errs)
	}
}

func TestValidationError_FieldErrors(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddField("email", "必須です")
	ve.AddField("password", "短すぎます")

	errs := ve.FieldErrors()
	if len(errs) != 2 {
		t.Fatalf("len(FieldErrors()) = %d、期待値 = 2", len(errs))
	}

	// フィールド名とメッセージの組み合わせが含まれていることを確認
	found := map[string]bool{}
	for _, fe := range errs {
		found[fe.Field+":"+fe.Message] = true
	}
	if !found["email:必須です"] {
		t.Error("FieldErrors()にemail:必須ですが含まれていない")
	}
	if !found["password:短すぎます"] {
		t.Error("FieldErrors()にpassword:短すぎますが含まれていない")
	}

	// nilレシーバー
	var nilVE *model.ValidationError
	if errs := nilVE.FieldErrors(); errs != nil {
		t.Errorf("nil.FieldErrors() = %v、期待値 = nil", errs)
	}
}

// コンパイル時にerrorインターフェースの実装を保証する
var (
	_ error = (*model.ValidationError)(nil)
	_ error = (*model.AppError)(nil)
)

func TestAppError_Error(t *testing.T) {
	t.Parallel()

	ae := &model.AppError{
		Code:    model.AppErrCodeResourceNotFound,
		UserMsg: "リソースが見つかりません",
	}
	if got := ae.Error(); got != "リソースが見つかりません" {
		t.Errorf("Error() = %q、期待値 = %q", got, "リソースが見つかりません")
	}
}

func TestAppError_Unwrap(t *testing.T) {
	t.Parallel()

	inner := errors.New("DB接続エラー")
	ae := &model.AppError{
		Code:     model.AppErrCodeInternal,
		UserMsg:  "内部エラーが発生しました",
		Internal: inner,
	}

	if !errors.Is(ae, inner) {
		t.Error("errors.Is(ae, inner) = false、期待値 = true")
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	t.Parallel()

	ae := &model.AppError{
		Code:    model.AppErrCodeResourceNotFound,
		UserMsg: "見つかりません",
	}

	if ae.Unwrap() != nil {
		t.Errorf("Unwrap() = %v、期待値 = nil", ae.Unwrap())
	}
}

func TestAppError_LogString(t *testing.T) {
	t.Parallel()

	ae := &model.AppError{
		Code:     model.AppErrCodeForbidden,
		UserMsg:  "権限がありません",
		Internal: errors.New("認可エラー"),
		Metadata: map[string]string{"user_id": "123"},
	}

	got := ae.LogString()
	expected := fmt.Sprintf("Code: %d | Msg: %s | Cause: %v | Meta: %v",
		model.AppErrCodeForbidden, "権限がありません", ae.Internal, ae.Metadata)
	if got != expected {
		t.Errorf("LogString() = %q、期待値 = %q", got, expected)
	}
}

func TestNewAppError(t *testing.T) {
	t.Parallel()

	inner := errors.New("DB接続エラー")
	ae := model.NewAppError(model.AppErrCodeInternal, "内部エラーが発生しました", inner)

	if ae.Code != model.AppErrCodeInternal {
		t.Errorf("Code = %d、期待値 = %d", ae.Code, model.AppErrCodeInternal)
	}
	if ae.UserMsg != "内部エラーが発生しました" {
		t.Errorf("UserMsg = %q、期待値 = %q", ae.UserMsg, "内部エラーが発生しました")
	}
	if ae.Internal != inner {
		t.Errorf("Internal = %v、期待値 = %v", ae.Internal, inner)
	}
	if ae.Metadata != nil {
		t.Errorf("Metadata = %v、期待値 = nil", ae.Metadata)
	}
}

func TestNewAppError_NilInternal(t *testing.T) {
	t.Parallel()

	ae := model.NewAppError(model.AppErrCodeResourceNotFound, "見つかりません", nil)

	if ae.Code != model.AppErrCodeResourceNotFound {
		t.Errorf("Code = %d、期待値 = %d", ae.Code, model.AppErrCodeResourceNotFound)
	}
	if ae.Internal != nil {
		t.Errorf("Internal = %v、期待値 = nil", ae.Internal)
	}
}

func TestAppErrorCode_Constants(t *testing.T) {
	t.Parallel()

	// iota + 1で始まるため、ゼロ値と区別できることを確認
	if model.AppErrCodeResourceNotFound == 0 {
		t.Error("AppErrCodeResourceNotFoundが0だった")
	}

	// 各定数が異なる値であることを確認
	codes := []model.AppErrorCode{
		model.AppErrCodeResourceNotFound,
		model.AppErrCodeForbidden,
		model.AppErrCodeConflict,
		model.AppErrCodeInternal,
		model.AppErrCodeBusy,
	}
	seen := map[model.AppErrorCode]bool{}
	for _, code := range codes {
		if seen[code] {
			t.Errorf("重複したAppErrorCode = %d", code)
		}
		seen[code] = true
	}
}

func TestAsValidationError(t *testing.T) {
	t.Parallel()

	t.Run("ValidationErrorを取り出せる", func(t *testing.T) {
		t.Parallel()

		ve := model.NewValidationError()
		ve.AddField("email", "必須です")

		got := model.AsValidationError(ve)
		if got == nil {
			t.Fatal("AsValidationError() = nil、期待値 = 非nil")
		}
		if !got.HasFieldError("email") {
			t.Error("取り出したValidationErrorにemailエラーがない")
		}
	})

	t.Run("ラップされたValidationErrorを取り出せる", func(t *testing.T) {
		t.Parallel()

		ve := model.NewValidationError()
		ve.AddGlobal("エラー")
		wrapped := fmt.Errorf("バリデーションに失敗: %w", ve)

		got := model.AsValidationError(wrapped)
		if got == nil {
			t.Fatal("AsValidationError() = nil、期待値 = 非nil")
		}
		if len(got.Global) != 1 {
			t.Errorf("len(Global) = %d、期待値 = 1", len(got.Global))
		}
	})

	t.Run("異なるエラー型ではnilを返す", func(t *testing.T) {
		t.Parallel()

		err := errors.New("通常のエラー")
		got := model.AsValidationError(err)
		if got != nil {
			t.Errorf("AsValidationError() = %v、期待値 = nil", got)
		}
	})

	t.Run("nilではnilを返す", func(t *testing.T) {
		t.Parallel()

		got := model.AsValidationError(nil)
		if got != nil {
			t.Errorf("AsValidationError(nil) = %v、期待値 = nil", got)
		}
	})
}

func TestAsAppError(t *testing.T) {
	t.Parallel()

	t.Run("AppErrorを取り出せる", func(t *testing.T) {
		t.Parallel()

		ae := &model.AppError{
			Code:    model.AppErrCodeResourceNotFound,
			UserMsg: "見つかりません",
		}

		got := model.AsAppError(ae)
		if got == nil {
			t.Fatal("AsAppError() = nil、期待値 = 非nil")
		}
		if got.Code != model.AppErrCodeResourceNotFound {
			t.Errorf("Code = %d、期待値 = %d", got.Code, model.AppErrCodeResourceNotFound)
		}
	})

	t.Run("ラップされたAppErrorを取り出せる", func(t *testing.T) {
		t.Parallel()

		ae := &model.AppError{
			Code:    model.AppErrCodeForbidden,
			UserMsg: "権限がありません",
		}
		wrapped := fmt.Errorf("操作に失敗: %w", ae)

		got := model.AsAppError(wrapped)
		if got == nil {
			t.Fatal("AsAppError() = nil、期待値 = 非nil")
		}
		if got.Code != model.AppErrCodeForbidden {
			t.Errorf("Code = %d、期待値 = %d", got.Code, model.AppErrCodeForbidden)
		}
	})

	t.Run("異なるエラー型ではnilを返す", func(t *testing.T) {
		t.Parallel()

		err := errors.New("通常のエラー")
		got := model.AsAppError(err)
		if got != nil {
			t.Errorf("AsAppError() = %v、期待値 = nil", got)
		}
	})

	t.Run("nilではnilを返す", func(t *testing.T) {
		t.Parallel()

		got := model.AsAppError(nil)
		if got != nil {
			t.Errorf("AsAppError(nil) = %v、期待値 = nil", got)
		}
	})
}
