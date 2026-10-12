package validator

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// SettingsAccountDeletionCreateValidatorは退会の受付のバリデーションを行う
type SettingsAccountDeletionCreateValidator struct {
	oauthApplicationRepo *repository.OAuthApplicationRepository
}

// NewSettingsAccountDeletionCreateValidatorはSettingsAccountDeletionCreateValidatorを生成する
func NewSettingsAccountDeletionCreateValidator(oauthApplicationRepo *repository.OAuthApplicationRepository) *SettingsAccountDeletionCreateValidator {
	return &SettingsAccountDeletionCreateValidator{oauthApplicationRepo: oauthApplicationRepo}
}

// WithTxは退会の受付と同じトランザクションでOAuthアプリケーションを確認するValidatorを返す。
func (v *SettingsAccountDeletionCreateValidator) WithTx(tx *sql.Tx) *SettingsAccountDeletionCreateValidator {
	return &SettingsAccountDeletionCreateValidator{oauthApplicationRepo: v.oauthApplicationRepo.WithTx(tx)}
}

// SettingsAccountDeletionCreateValidatorInputはバリデーションの入力パラメータ
type SettingsAccountDeletionCreateValidatorInput struct {
	UserID model.UserID
	// Usernameはログイン中のユーザーのユーザー名
	Username string
	// InputUsernameは同意の確認として利用者が入力したユーザー名
	InputUsername string
}

// Validateはバリデーションを行い、失敗時は *model.ValidationErrorを返す
func (v *SettingsAccountDeletionCreateValidator) Validate(ctx context.Context, input SettingsAccountDeletionCreateValidatorInput) error {
	ve := model.NewValidationError()

	// OAuthアプリケーションを残したまま退会すると、そのアプリケーションの利用者が
	// 管理者のいないアプリケーションに認可を与え続けることになるため、先に削除してもらう。
	hasKeptOAuthApplications, err := v.oauthApplicationRepo.ExistsKeptByOwnerID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("OAuthアプリケーションの有無の取得に失敗: %w", err)
	}
	if hasKeptOAuthApplications {
		ve.AddGlobal(i18n.T(ctx, "validation_account_deletion_oauth_applications_kept"))
		return ve
	}

	if strings.TrimSpace(input.InputUsername) != input.Username {
		ve.AddField("username", i18n.T(ctx, "validation_account_deletion_username_mismatch"))
		return ve
	}

	return nil
}
