// Package settings_account_deletionは退会の確認と受付のHTTPハンドラーを定義する。
package settings_account_deletion

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは退会の確認と受付のHTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                     *config.Config
	sessionManager          *session.Manager
	imageHelper             *image.Helper
	getAccountDeletionNewUC *usecase.GetAccountDeletionNewUsecase
	createAccountDeletionUC *usecase.CreateAccountDeletionUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	imageHelper *image.Helper,
	getAccountDeletionNewUC *usecase.GetAccountDeletionNewUsecase,
	createAccountDeletionUC *usecase.CreateAccountDeletionUsecase,
) *Handler {
	return &Handler{
		cfg:                     cfg,
		sessionManager:          sessionManager,
		imageHelper:             imageHelper,
		getAccountDeletionNewUC: getAccountDeletionNewUC,
		createAccountDeletionUC: createAccountDeletionUC,
	}
}
