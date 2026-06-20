// Package sign_in_code はログイン確認コード検証機能を提供します
package sign_in_code

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler 6桁コード入力関連のHTTPハンドラーです
type Handler struct {
	cfg                *config.Config
	sessionMgr         *session.Manager
	flashMgr           *session.FlashManager
	limiter            *ratelimit.Limiter
	sendSignInCodeUC   *usecase.SendSignInCodeUsecase
	verifySignInCodeUC *usecase.VerifySignInCodeUsecase
	createSessionUC    *usecase.CreateSessionUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	flashMgr *session.FlashManager,
	limiter *ratelimit.Limiter,
	sendSignInCodeUC *usecase.SendSignInCodeUsecase,
	verifySignInCodeUC *usecase.VerifySignInCodeUsecase,
	createSessionUC *usecase.CreateSessionUsecase,
) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionMgr:         sessionMgr,
		flashMgr:           flashMgr,
		limiter:            limiter,
		sendSignInCodeUC:   sendSignInCodeUC,
		verifySignInCodeUC: verifySignInCodeUC,
		createSessionUC:    createSessionUC,
	}
}
