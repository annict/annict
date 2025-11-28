package sign_in_code

import (
	"context"
	"testing"
)

func TestCreateRequest_Validate(t *testing.T) {
	// 日本語ロケールでテスト
	ctx := context.Background()

	tests := []struct {
		name      string
		code      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "正常: 6桁の数字",
			code:      "123456",
			wantError: false,
		},
		{
			name:      "エラー: コードが空",
			code:      "",
			wantError: true,
			errorMsg:  "コードを入力してください",
		},
		{
			name:      "エラー: 5桁の数字",
			code:      "12345",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
		{
			name:      "エラー: 7桁の数字",
			code:      "1234567",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
		{
			name:      "エラー: 6桁の英数字",
			code:      "12345a",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
		{
			name:      "エラー: 6桁のアルファベット",
			code:      "abcdef",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
		{
			name:      "エラー: スペースを含む",
			code:      "123 456",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateRequest{
				Code: tt.code,
			}

			formErrors := req.Validate(ctx)

			if tt.wantError {
				if formErrors == nil {
					t.Errorf("エラーを期待していましたが、エラーが返されませんでした")
					return
				}

				errors, exists := formErrors.Fields["code"]
				if !exists {
					t.Errorf("codeフィールドのエラーを期待していましたが、エラーがありませんでした")
					return
				}

				if len(errors) == 0 {
					t.Errorf("codeフィールドのエラーメッセージが空です")
					return
				}

				if errors[0] != tt.errorMsg {
					t.Errorf("期待したエラーメッセージと異なります: got %v, want %v", errors[0], tt.errorMsg)
				}
			} else {
				if formErrors != nil {
					t.Errorf("エラーが返されました: %v", formErrors)
				}
			}
		})
	}
}

func TestCreateRequest_Validate_English(t *testing.T) {
	// 英語ロケールでテスト
	ctx := context.Background()

	tests := []struct {
		name      string
		code      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid: 6-digit number",
			code:      "123456",
			wantError: false,
		},
		{
			name:      "Error: Empty code",
			code:      "",
			wantError: true,
			errorMsg:  "コードを入力してください",
		},
		{
			name:      "Error: 5-digit number",
			code:      "12345",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
		{
			name:      "Error: 7-digit number",
			code:      "1234567",
			wantError: true,
			errorMsg:  "コードは6桁の数字で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateRequest{
				Code: tt.code,
			}

			formErrors := req.Validate(ctx)

			if tt.wantError {
				if formErrors == nil {
					t.Errorf("Expected error but got none")
					return
				}

				errors, exists := formErrors.Fields["code"]
				if !exists {
					t.Errorf("Expected error for 'code' field but got none")
					return
				}

				if len(errors) == 0 {
					t.Errorf("Error messages for 'code' field are empty")
					return
				}

				if errors[0] != tt.errorMsg {
					t.Errorf("Unexpected error message: got %v, want %v", errors[0], tt.errorMsg)
				}
			} else {
				if formErrors != nil {
					t.Errorf("Unexpected error: %v", formErrors)
				}
			}
		})
	}
}
