package sign_up

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// Handler 新規登録関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	userRepo         *repository.UserRepository
	limiter          *ratelimit.Limiter
	sendSignUpCodeUC *usecase.SendSignUpCodeUsecase
	turnstileClient  *turnstile.Client
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	userRepo *repository.UserRepository,
	limiter *ratelimit.Limiter,
	sendSignUpCodeUC *usecase.SendSignUpCodeUsecase,
	turnstileClient *turnstile.Client,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		userRepo:         userRepo,
		limiter:          limiter,
		sendSignUpCodeUC: sendSignUpCodeUC,
		turnstileClient:  turnstileClient,
	}
}
