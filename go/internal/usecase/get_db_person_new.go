package usecase

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// GetDBPersonNewUsecaseはDB管理画面の人物の一括登録フォームのユースケース。フォームは
// 1つのtextareaに全行を入力する形で、読み込む保存済みの状態を持たないため、閲覧者の
// 認可だけを行う。
type GetDBPersonNewUsecase struct{}

// NewGetDBPersonNewUsecaseは新しいGetDBPersonNewUsecaseを作成する。
func NewGetDBPersonNewUsecase() *GetDBPersonNewUsecase {
	return &GetDBPersonNewUsecase{}
}

// GetDBPersonNewInputはユースケースの入力。
type GetDBPersonNewInput struct {
	User *model.User
}

// Executeは閲覧者がcommitterであることを確かめる。committerでなければ
// AppErrCodeForbiddenの *model.AppErrorを返し、Handler側で403に変換する。webのルートは
// committerのミドルウェアで塞がれているが、この確認によりルールがUseCase側に残る。
func (uc *GetDBPersonNewUsecase) Execute(ctx context.Context, input GetDBPersonNewInput) error {
	if input.User == nil || !input.User.IsCommitter() {
		return &model.AppError{
			Code:    model.AppErrCodeForbidden,
			UserMsg: i18n.T(ctx, "error_forbidden"),
		}
	}

	return nil
}
