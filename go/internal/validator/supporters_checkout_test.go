package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestSupportersCheckoutCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SupportersCheckoutCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name:       "月額プラン",
			input:      SupportersCheckoutCreateValidatorInput{Plan: "monthly"},
			wantErrors: false,
		},
		{
			name:       "年額プラン",
			input:      SupportersCheckoutCreateValidatorInput{Plan: "yearly"},
			wantErrors: false,
		},
		{
			name:       "プランが空",
			input:      SupportersCheckoutCreateValidatorInput{Plan: ""},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "不正なプラン",
			input:      SupportersCheckoutCreateValidatorInput{Plan: "weekly"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
		{
			name:       "プランが大文字",
			input:      SupportersCheckoutCreateValidatorInput{Plan: "MONTHLY"},
			wantErrors: true,
			wantFields: []string{"plan"},
		},
	}

	v := NewSupportersCheckoutCreateValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if _, exists := ve.Fields[field]; !exists {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}
