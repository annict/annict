package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestSettingsAccountDeletionCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	v := NewSettingsAccountDeletionCreateValidator(repository.NewOAuthApplicationRepository(query.New(db).WithTx(tx)))

	userID := testutil.NewUserBuilder(t, tx).Build()
	// 論理削除済みのアプリケーションは退会を妨げない
	testutil.CreateTestOAuthApplication(t, tx, userID, true)
	ownerID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestOAuthApplication(t, tx, ownerID, false)

	tests := []struct {
		name           string
		input          SettingsAccountDeletionCreateValidatorInput
		wantValid      bool
		wantFieldError string
		wantGlobal     string
	}{
		{
			name:      "ユーザー名が一致する",
			input:     SettingsAccountDeletionCreateValidatorInput{UserID: userID, Username: "deletion_user", InputUsername: "deletion_user"},
			wantValid: true,
		},
		{
			name:      "前後の空白を除くと一致する",
			input:     SettingsAccountDeletionCreateValidatorInput{UserID: userID, Username: "deletion_user", InputUsername: " deletion_user "},
			wantValid: true,
		},
		{
			name:           "ユーザー名が一致しない",
			input:          SettingsAccountDeletionCreateValidatorInput{UserID: userID, Username: "deletion_user", InputUsername: "Deletion_User"},
			wantFieldError: "ユーザー名が一致しません",
		},
		{
			name:           "ユーザー名が空",
			input:          SettingsAccountDeletionCreateValidatorInput{UserID: userID, Username: "deletion_user", InputUsername: ""},
			wantFieldError: "ユーザー名が一致しません",
		},
		{
			name:       "論理削除されていないOAuthアプリケーションがある",
			input:      SettingsAccountDeletionCreateValidatorInput{UserID: ownerID, Username: "oauth_owner", InputUsername: "oauth_owner"},
			wantGlobal: "作成したアプリケーションが残っているため、退会できません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n.SetLocale(context.Background(), "ja")

			err := v.Validate(ctx, tt.input)

			if tt.wantValid {
				if err != nil {
					t.Fatalf("Validate()のエラー = %v", err)
				}
				return
			}

			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("ValidationErrorを期待したが、%vだった", err)
			}
			if tt.wantFieldError != "" {
				if got := ve.GetFieldErrors("username"); len(got) != 1 || got[0] != tt.wantFieldError {
					t.Errorf("usernameのエラー = %v、期待値 = [%s]", got, tt.wantFieldError)
				}
			}
			if tt.wantGlobal != "" {
				if len(ve.Global) != 1 || ve.Global[0] != tt.wantGlobal {
					t.Errorf("グローバルエラー = %v、期待値 = [%s]", ve.Global, tt.wantGlobal)
				}
				if ve.HasFieldError("username") {
					t.Error("OAuthアプリケーションが残っているときにユーザー名のエラーが出ている")
				}
			}
		})
	}
}
