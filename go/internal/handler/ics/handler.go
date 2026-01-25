// Package ics はiCalendar形式のカレンダー配信機能を提供します
package ics

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
)

// Handler はiCalendar配信関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	userCalendarRepo *repository.UserCalendarRepository
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, userCalendarRepo *repository.UserCalendarRepository) *Handler {
	return &Handler{
		cfg:              cfg,
		userCalendarRepo: userCalendarRepo,
	}
}
