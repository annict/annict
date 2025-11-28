// Package sign_in_code はログイン確認コード検証機能を提供します
package sign_in_code

import (
	"database/sql"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
)

// Handler 6桁コード入力関連のHTTPハンドラーです
type Handler struct {
	cfg                *config.Config
	sessionMgr         *session.Manager
	userRepo           *repository.UserRepository
	db                 *sql.DB
	limiter            *ratelimit.Limiter
	sendSignInCodeUC   *usecase.SendSignInCodeUsecase
	verifySignInCodeUC *usecase.VerifySignInCodeUsecase
	createSessionUC    *usecase.CreateSessionUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	userRepo *repository.UserRepository,
	db *sql.DB,
	limiter *ratelimit.Limiter,
	sendSignInCodeUC *usecase.SendSignInCodeUsecase,
	verifySignInCodeUC *usecase.VerifySignInCodeUsecase,
	createSessionUC *usecase.CreateSessionUsecase,
) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionMgr:         sessionMgr,
		userRepo:           userRepo,
		db:                 db,
		limiter:            limiter,
		sendSignInCodeUC:   sendSignInCodeUC,
		verifySignInCodeUC: verifySignInCodeUC,
		createSessionUC:    createSessionUC,
	}
}
