package sign_up

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler 新規登録関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	limiter          *ratelimit.Limiter
	sendSignUpCodeUC *usecase.SendSignUpCodeUsecase
	turnstileClient  *turnstile.Client
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	limiter *ratelimit.Limiter,
	sendSignUpCodeUC *usecase.SendSignUpCodeUsecase,
	turnstileClient *turnstile.Client,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		limiter:          limiter,
		sendSignUpCodeUC: sendSignUpCodeUC,
		turnstileClient:  turnstileClient,
	}
}
