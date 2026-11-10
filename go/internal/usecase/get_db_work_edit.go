package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkEditUsecaseはAnnict DB管理画面の作品編集フォームに必要なデータ
// (対象のworkと、その選択肢となるnumber format) を取得するユースケース。
type GetDBWorkEditUsecase struct {
	workRepo         *repository.WorkRepository
	numberFormatRepo *repository.NumberFormatRepository
	featureFlagRepo  *repository.FeatureFlagRepository
}

func NewGetDBWorkEditUsecase(
	workRepo *repository.WorkRepository,
	numberFormatRepo *repository.NumberFormatRepository,
	featureFlagRepo *repository.FeatureFlagRepository,
) *GetDBWorkEditUsecase {
	return &GetDBWorkEditUsecase{
		workRepo:         workRepo,
		numberFormatRepo: numberFormatRepo,
		featureFlagRepo:  featureFlagRepo,
	}
}

type GetDBWorkEditInput struct {
	User   *model.User
	WorkID model.WorkID
	// DeviceTokenは閲覧中のデバイスのトークン (Cookieが無いときは空文字列)。履歴へのリンクを出すかを、
	// 履歴の画面をゲートするリバースプロキシと同じく、デバイスとユーザーのどちらかで判定するのに使う。
	DeviceToken string
}

type GetDBWorkEditOutput struct {
	Work          *model.Work
	NumberFormats []model.NumberFormat
	// CanShowRevisionsは作品の編集画面の見出しに、作品の履歴へのリンクを出すかを表す。履歴の画面はフィーチャー
	// フラグdb_edit_historyが有効なデバイスかユーザーにだけ公開するため、リンクも同じ判定で出す。
	CanShowRevisions bool
}

// Executeはworkを取得する前にコミッターを認可し、編集対象のworkとフォームの選択肢を
// 返す。workが存在しない場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、
// Handler側で404に変換する。
func (uc *GetDBWorkEditUsecase) Execute(ctx context.Context, input GetDBWorkEditInput) (*GetDBWorkEditOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	work, err := uc.workRepo.GetForEditByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if work == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("NumberFormatの取得に失敗: %w", err)
	}

	canShowRevisions, err := canShowDBRevisions(ctx, uc.featureFlagRepo, input.DeviceToken, input.User)
	if err != nil {
		return nil, err
	}

	return &GetDBWorkEditOutput{
		Work:             work,
		NumberFormats:    numberFormats,
		CanShowRevisions: canShowRevisions,
	}, nil
}
