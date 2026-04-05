// Package ics はiCalendar形式のカレンダー配信機能を提供します
package ics

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はiCalendar配信関連のHTTPハンドラーです
type Handler struct {
	cfg               *config.Config
	getUserCalendarUC *usecase.GetUserCalendarUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, getUserCalendarUC *usecase.GetUserCalendarUsecase) *Handler {
	return &Handler{
		cfg:               cfg,
		getUserCalendarUC: getUserCalendarUC,
	}
}
