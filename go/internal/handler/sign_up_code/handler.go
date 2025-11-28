// Package sign_up_code はサインアップ確認コード検証機能を提供します
package sign_up_code

import (
	"database/sql"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
	"github.com/redis/go-redis/v9"
)

// Handler 新規登録確認コード関連のHTTPハンドラーです
type Handler struct {
	cfg                *config.Config
	sessionMgr         *session.Manager
	db                 *sql.DB
	limiter            *ratelimit.Limiter
	redisClient        *redis.Client
	sendSignUpCodeUC   *usecase.SendSignUpCodeUsecase
	verifySignUpCodeUC *usecase.VerifySignUpCodeUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	db *sql.DB,
	limiter *ratelimit.Limiter,
	redisClient *redis.Client,
	sendSignUpCodeUC *usecase.SendSignUpCodeUsecase,
	verifySignUpCodeUC *usecase.VerifySignUpCodeUsecase,
) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionMgr:         sessionMgr,
		db:                 db,
		limiter:            limiter,
		redisClient:        redisClient,
		sendSignUpCodeUC:   sendSignUpCodeUC,
		verifySignUpCodeUC: verifySignUpCodeUC,
	}
}
