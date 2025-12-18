// Package sign_up_code はサインアップ確認コード検証機能を提供します
package sign_up_code

import (
	"database/sql"

	"github.com/redis/go-redis/v9"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
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
