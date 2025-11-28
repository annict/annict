// Package sign_up_username はサインアップユーザー名設定機能を提供します
package sign_up_username

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
	"github.com/redis/go-redis/v9"
)

// Handler ユーザー名設定とユーザー登録のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	redisClient      *redis.Client
	completeSignUpUC *usecase.CompleteSignUpUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	redisClient *redis.Client,
	completeSignUpUC *usecase.CompleteSignUpUsecase,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		redisClient:      redisClient,
		completeSignUpUC: completeSignUpUC,
	}
}
