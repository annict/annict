package sign_up_username

import (
	"context"
	"regexp"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

// usernameRegex ユーザー名の形式（1～20文字の半角英数字とアンダースコア）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// CreateRequest ユーザー登録リクエスト
type CreateRequest struct {
	Token    string
	Username string
}

// Validate リクエストのバリデーションを行います
func (req *CreateRequest) Validate(ctx context.Context) *session.FormErrors {
	formErrors := session.FormErrors{}

	// トークン必須チェック
	if req.Token == "" {
		formErrors.AddFieldError("token", i18n.T(ctx, "sign_up_username_error_token_missing"))
	}

	// ユーザー名必須チェック
	if req.Username == "" {
		formErrors.AddFieldError("username", i18n.T(ctx, "sign_up_username_error_username_required"))
	} else if !usernameRegex.MatchString(req.Username) {
		// ユーザー名形式チェック（20文字以内、半角英数字とアンダースコアのみ）
		formErrors.AddFieldError("username", i18n.T(ctx, "sign_up_username_error_username_format"))
	}

	if formErrors.HasErrors() {
		return &formErrors
	}
	return nil
}
