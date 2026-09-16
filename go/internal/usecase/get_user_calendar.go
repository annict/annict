package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetUserCalendarUsecaseはユーザーのカレンダーデータを取得するユースケースです
type GetUserCalendarUsecase struct {
	userCalendarRepo *repository.UserCalendarRepository
}

// NewGetUserCalendarUsecaseは新しいGetUserCalendarUsecaseを作成します
func NewGetUserCalendarUsecase(userCalendarRepo *repository.UserCalendarRepository) *GetUserCalendarUsecase {
	return &GetUserCalendarUsecase{
		userCalendarRepo: userCalendarRepo,
	}
}

// GetUserCalendarInputはユースケースの入力です
type GetUserCalendarInput struct {
	Username string
	Now      time.Time
}

// GetUserCalendarOutputはユースケースの出力です
type GetUserCalendarOutput struct {
	UserCalendar *model.UserCalendar
}

// Executeはユーザーのカレンダーデータを取得します
func (uc *GetUserCalendarUsecase) Execute(ctx context.Context, input GetUserCalendarInput) (*GetUserCalendarOutput, error) {
	userCalendar, err := uc.userCalendarRepo.GetByUsername(ctx, input.Username, input.Now)
	if err != nil {
		return nil, fmt.Errorf("カレンダーデータの取得に失敗: %w", err)
	}
	return &GetUserCalendarOutput{UserCalendar: userCalendar}, nil
}
