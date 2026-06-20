// Package password はパスワード変更機能を提供します
package password

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はパスワード編集・更新のHTTPハンドラーです
type Handler struct {
	cfg                     *config.Config
	sessionMgr              *session.Manager
	flashMgr                *session.FlashManager
	limiter                 *ratelimit.Limiter
	getPasswordResetTokenUC *usecase.GetPasswordResetTokenUsecase
	updatePasswordResetUC   *usecase.UpdatePasswordResetUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, sessionMgr *session.Manager, flashMgr *session.FlashManager, limiter *ratelimit.Limiter, getPasswordResetTokenUC *usecase.GetPasswordResetTokenUsecase, updatePasswordResetUC *usecase.UpdatePasswordResetUsecase) *Handler {
	return &Handler{
		cfg:                     cfg,
		sessionMgr:              sessionMgr,
		flashMgr:                flashMgr,
		limiter:                 limiter,
		getPasswordResetTokenUC: getPasswordResetTokenUC,
		updatePasswordResetUC:   updatePasswordResetUC,
	}
}
