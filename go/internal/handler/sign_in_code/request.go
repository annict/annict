package sign_in_code

import (
	"context"
	"regexp"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// CreateRequest は6桁コード検証リクエストを表します
type CreateRequest struct {
	Code string // 6桁の数字コード
}

// Validate はリクエストをバリデーションします
func (r *CreateRequest) Validate(ctx context.Context) *session.FormErrors {
	formErrors := &session.FormErrors{}

	// コードが空の場合
	if r.Code == "" {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_required"))
		return formErrors
	}

	// コードが6桁の数字でない場合
	matched, err := regexp.MatchString(`^\d{6}$`, r.Code)
	if err != nil || !matched {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_invalid_format"))
		return formErrors
	}

	return nil
}
