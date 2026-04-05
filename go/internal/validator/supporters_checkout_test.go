package validator

import (
	"context"
	"testing"
)

func TestCreateSupportersCheckoutValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateSupportersCheckoutValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name:       "monthly plan",
			input:      CreateSupportersCheckoutValidatorInput{Plan: "monthly"},
			wantErrors: false,
		},
		{
			name:       "yearly plan",
			input:      CreateSupportersCheckoutValidatorInput{Plan: "yearly"},
			wantErrors: false,
		},
		{
			name:       "empty plan",
			input:      CreateSupportersCheckoutValidatorInput{Plan: ""},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "invalid plan",
			input:      CreateSupportersCheckoutValidatorInput{Plan: "weekly"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "uppercase plan",
			input:      CreateSupportersCheckoutValidatorInput{Plan: "MONTHLY"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
	}

	v := NewCreateSupportersCheckoutValidator()

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
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}
