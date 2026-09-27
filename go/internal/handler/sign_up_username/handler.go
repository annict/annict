// Package sign_up_usernameはサインアップユーザー名設定機能を提供します
package sign_up_username

import (
	"github.com/redis/go-redis/v9"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerユーザー名設定とユーザー登録のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	flashMgr         *session.FlashManager
	redisClient      *redis.Client
	completeSignUpUC *usecase.CompleteSignUpUsecase
}

// NewHandler新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	flashMgr *session.FlashManager,
	redisClient *redis.Client,
	completeSignUpUC *usecase.CompleteSignUpUsecase,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		flashMgr:         flashMgr,
		redisClient:      redisClient,
		completeSignUpUC: completeSignUpUC,
	}
}
