package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/repository"
)

// CheckHealthUsecaseはヘルスチェックを行うユースケースです
type CheckHealthUsecase struct {
	workRepo *repository.WorkRepository
}

// NewCheckHealthUsecaseは新しいCheckHealthUsecaseを作成します
func NewCheckHealthUsecase(workRepo *repository.WorkRepository) *CheckHealthUsecase {
	return &CheckHealthUsecase{
		workRepo: workRepo,
	}
}

// CheckHealthOutputはユースケースの出力です
type CheckHealthOutput struct {
	DBHealthy bool
	DBError   string
}

// Executeはデータベースの接続状態を確認します
func (uc *CheckHealthUsecase) Execute(ctx context.Context) (*CheckHealthOutput, error) {
	_, err := uc.workRepo.GetByID(ctx, 1)
	if err != nil && err != sql.ErrNoRows {
		return &CheckHealthOutput{
			DBHealthy: false,
			DBError:   fmt.Sprintf("unhealthy: %v", err),
		}, nil
	}

	return &CheckHealthOutput{
		DBHealthy: true,
	}, nil
}
