package supporters_checkout

import (
	"context"
	"testing"
)

func TestCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name:       "monthly plan",
			input:      CreateValidatorInput{Plan: "monthly"},
			wantErrors: false,
		},
		{
			name:       "yearly plan",
			input:      CreateValidatorInput{Plan: "yearly"},
			wantErrors: false,
		},
		{
			name:       "empty plan",
			input:      CreateValidatorInput{Plan: ""},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "invalid plan",
			input:      CreateValidatorInput{Plan: "weekly"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "uppercase plan",
			input:      CreateValidatorInput{Plan: "MONTHLY"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
	}

	validator := NewCreateValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

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
