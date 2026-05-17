// Package sign_up_username はサインアップユーザー名設定機能を提供します
package sign_up_username

import (
	"github.com/redis/go-redis/v9"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler ユーザー名設定とユーザー登録のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	flashMgr         *session.FlashManager
	redisClient      *redis.Client
	completeSignUpUC *usecase.CompleteSignUpUsecase
}

// NewHandler 新しいHandlerを作成します
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
