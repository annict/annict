package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// UpdatePersonUsecaseはAnnict DBの人物編集フォームの1回の送信を適用する。
type UpdatePersonUsecase struct {
	personRepo *repository.PersonRepository
	validator  *validator.DBPersonUpdateValidator
}

// NewUpdatePersonUsecaseはUpdatePersonUsecaseを生成する。
func NewUpdatePersonUsecase(
	personRepo *repository.PersonRepository,
	validator *validator.DBPersonUpdateValidator,
) *UpdatePersonUsecase {
	return &UpdatePersonUsecase{
		personRepo: personRepo,
		validator:  validator,
	}
}

// UpdatePersonInputは送信された編集フォームを保持する。値が文字列なのはフォームが送信
// するものが文字列であるため。UpdatedAtはフォームを開いた時点の版。
type UpdatePersonInput struct {
	PersonID model.PersonID
	// Userは変更の帰属先となる編集者であり、送信を受け付けるかを決めるロールのsource
	// でもある。
	User              *model.User
	Name              string
	NameKana          string
	NameEn            string
	Nickname          string
	NicknameEn        string
	Gender            string
	BloodType         string
	PrefectureID      string
	Birthday          string
	Height            string
	URL               string
	URLEn             string
	WikipediaURL      string
	WikipediaURLEn    string
	TwitterUsername   string
	TwitterUsernameEn string
	UpdatedAt         string
}

// UpdatePersonOutputは更新された人物を報告する。
type UpdatePersonOutput struct {
	PersonID model.PersonID
}

// Executeは送信された値を人物に適用する。
//
// 認可と形式検証の後、保存済みの版を確認する。DBを要する入力検証はその後に行う。
// 更新は1文 (編集履歴の記録を含む) で完結するため、トランザクションは開かない。
func (uc *UpdatePersonUsecase) Execute(ctx context.Context, input UpdatePersonInput) (*UpdatePersonOutput, error) {
	// 送信はcommitterに限られ、変更者に帰属するため、ユーザーを伴わない呼び出しと
	// 一般ユーザーの呼び出しは何かを読む前に拒否する。
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	fields, err := uc.validator.ValidateFormat(ctx, validator.DBPersonUpdateValidatorInput{
		Name:              input.Name,
		NameKana:          input.NameKana,
		NameEn:            input.NameEn,
		Nickname:          input.Nickname,
		NicknameEn:        input.NicknameEn,
		Gender:            input.Gender,
		BloodType:         input.BloodType,
		PrefectureID:      input.PrefectureID,
		Birthday:          input.Birthday,
		Height:            input.Height,
		URL:               input.URL,
		URLEn:             input.URLEn,
		WikipediaURL:      input.WikipediaURL,
		WikipediaURLEn:    input.WikipediaURLEn,
		TwitterUsername:   input.TwitterUsername,
		TwitterUsernameEn: input.TwitterUsernameEn,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	current, err := uc.personRepo.GetForEditByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗しました: %w", err)
	}
	if current == nil {
		return nil, personNotFoundError(ctx, input.PersonID)
	}
	if !current.UpdatedAt.Equal(fields.UpdatedAt) {
		return nil, personConflictError(ctx, input.PersonID)
	}
	fields, err = uc.validator.ValidateDatabase(ctx, input.PersonID, fields)
	if err != nil {
		return nil, err
	}

	updated, err := uc.personRepo.Update(ctx, repository.UpdatePersonParams{
		ID:                input.PersonID,
		Name:              fields.Name,
		NameKana:          fields.NameKana,
		NameEn:            fields.NameEn,
		Nickname:          fields.Nickname,
		NicknameEn:        fields.NicknameEn,
		Gender:            fields.Gender,
		BloodType:         fields.BloodType,
		PrefectureID:      fields.PrefectureID,
		Birthday:          fields.Birthday,
		Height:            fields.Height,
		URL:               fields.URL,
		URLEn:             fields.URLEn,
		WikipediaURL:      fields.WikipediaURL,
		WikipediaURLEn:    fields.WikipediaURLEn,
		TwitterUsername:   fields.TwitterUsername,
		TwitterUsernameEn: fields.TwitterUsernameEn,
		Version:           fields.UpdatedAt,
		UserID:            input.User.ID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrPersonNameTaken) {
			ve := model.NewValidationError()
			ve.AddField("name", i18n.T(ctx, "validation_person_name_already_taken"))
			return nil, ve
		}
		return nil, fmt.Errorf("人物の更新に失敗しました: %w", err)
	}
	if updated {
		return &UpdatePersonOutput{PersonID: input.PersonID}, nil
	}

	// どの行も更新されなかった理由は、人物が無い (存在しないか、削除・掲載停止された) か、版が進んだかの
	// どちらか。前者は404、後者は競合として、保存済みの値を並べたフォームで返す。
	current, err = uc.personRepo.GetForEditByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗しました: %w", err)
	}
	if current == nil {
		return nil, personNotFoundError(ctx, input.PersonID)
	}

	return nil, personConflictError(ctx, input.PersonID)
}

func personNotFoundError(ctx context.Context, personID model.PersonID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_person_not_found"),
		Metadata: map[string]string{"person_id": personID.String()},
	}
}

func personConflictError(ctx context.Context, personID model.PersonID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeConflict,
		UserMsg:  i18n.T(ctx, "validation_version_conflict"),
		Metadata: map[string]string{"person_id": personID.String()},
	}
}
