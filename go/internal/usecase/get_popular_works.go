package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetPopularWorksUsecase は人気作品を取得するユースケースです
type GetPopularWorksUsecase struct {
	workRepo  *repository.WorkRepository
	castRepo  *repository.CastRepository
	staffRepo *repository.StaffRepository
}

// NewGetPopularWorksUsecase は新しいGetPopularWorksUsecaseを作成します
func NewGetPopularWorksUsecase(
	workRepo *repository.WorkRepository,
	castRepo *repository.CastRepository,
	staffRepo *repository.StaffRepository,
) *GetPopularWorksUsecase {
	return &GetPopularWorksUsecase{
		workRepo:  workRepo,
		castRepo:  castRepo,
		staffRepo: staffRepo,
	}
}

// GetPopularWorksOutput はユースケースの出力です
type GetPopularWorksOutput struct {
	Works []*model.Work
}

// Execute は人気作品をキャスト・スタッフ情報と共に取得します
func (uc *GetPopularWorksUsecase) Execute(ctx context.Context) (*GetPopularWorksOutput, error) {
	works, err := uc.workRepo.GetPopular(ctx)
	if err != nil {
		return nil, fmt.Errorf("人気作品の取得に失敗: %w", err)
	}

	if len(works) == 0 {
		return &GetPopularWorksOutput{Works: []*model.Work{}}, nil
	}

	workIDs := make([]model.WorkID, len(works))
	for i, w := range works {
		workIDs[i] = w.ID
	}

	casts, err := uc.castRepo.GetByWorkIDs(ctx, workIDs)
	if err != nil {
		return nil, fmt.Errorf("キャスト情報の取得に失敗: %w", err)
	}

	staffs, err := uc.staffRepo.GetByWorkIDs(ctx, workIDs)
	if err != nil {
		return nil, fmt.Errorf("スタッフ情報の取得に失敗: %w", err)
	}

	castsByWorkID := make(map[model.WorkID][]*model.Cast, len(works))
	for _, c := range casts {
		castsByWorkID[c.WorkID] = append(castsByWorkID[c.WorkID], c)
	}
	staffsByWorkID := make(map[model.WorkID][]*model.Staff, len(works))
	for _, s := range staffs {
		staffsByWorkID[s.WorkID] = append(staffsByWorkID[s.WorkID], s)
	}

	// workRepo.GetPopular が毎回新規 *model.Work を生成して返すことを前提に、
	// 取得した Work インスタンスへ Casts/Staffs を直接代入する。Repository 側で
	// キャッシュやプール再利用を導入する場合はこの前提を見直す必要がある。
	for _, w := range works {
		w.Casts = castsByWorkID[w.ID]
		w.Staffs = staffsByWorkID[w.ID]
	}

	return &GetPopularWorksOutput{Works: works}, nil
}
