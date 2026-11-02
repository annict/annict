package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// errSignInTwoFactorUnavailableは、第1の要素を確かめたあとにユーザーが退会した・二要素認証の設定が変わったことを表す内部エラー。
var errSignInTwoFactorUnavailable = errors.New("ログイン途中のユーザーが二要素認証を使えなくなりました")

// CreateSignInTwoFactorUsecaseは、第1の要素を確かめたユーザーの認証アプリのコードを照合し、
// ログインを続けてよいユーザーを返す。
//
// セッションの発行は、パスワードやメールのコードでのログインと同じくCreateSessionUsecaseの別の段階とする。
// 書き込みは使ったタイムステップの記録 (条件付きのUPDATE) 1回のため、トランザクションは開かない。
type CreateSignInTwoFactorUsecase struct {
	twoFactorKey          *auth.TwoFactorKey
	createValidator       *validator.SignInTwoFactorCreateValidator
	userRepo              *repository.UserRepository
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository
}

// NewCreateSignInTwoFactorUsecaseはCreateSignInTwoFactorUsecaseを生成する。
func NewCreateSignInTwoFactorUsecase(
	twoFactorKey *auth.TwoFactorKey,
	createValidator *validator.SignInTwoFactorCreateValidator,
	userRepo *repository.UserRepository,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
) *CreateSignInTwoFactorUsecase {
	return &CreateSignInTwoFactorUsecase{
		twoFactorKey:          twoFactorKey,
		createValidator:       createValidator,
		userRepo:              userRepo,
		userTwoFactorAuthRepo: userTwoFactorAuthRepo,
	}
}

// CreateSignInTwoFactorInputはCreateSignInTwoFactorUsecase.Executeの入力。
type CreateSignInTwoFactorInput struct {
	// UserIDは第1の要素を確かめたユーザーのID。ログイン途中の状態 (セッションの値) から読んだ値を渡す
	UserID model.UserID
	// Codeは入力されたままのコード。空白・ハイフン・全角の数字はここで正規化する
	Code string
}

// CreateSignInTwoFactorOutputはCreateSignInTwoFactorUsecase.Executeの結果。
type CreateSignInTwoFactorOutput struct {
	UserID model.UserID
	// EncryptedPasswordはセッションの作成に渡すパスワードのハッシュ
	EncryptedPassword string
	Username          string
}

// Executeはコードを照合し、合っていればそのタイムステップを使用済みにしてユーザーを返す。
//
// 形式の誤り・照合の失敗・使用済みのコードは、コードの欄の*model.ValidationErrorで返す。
// 使用済みのコードも照合の失敗と同じ文言にし、認証アプリの最新のコードを入れ直してもらう。
// ユーザーが退会した・二要素認証の設定が変わったときは、ログインをやり直させるためAppErrCodeConflictを返す。
func (uc *CreateSignInTwoFactorUsecase) Execute(ctx context.Context, input CreateSignInTwoFactorInput) (*CreateSignInTwoFactorOutput, error) {
	code := auth.NormalizeTOTPCode(input.Code)
	if err := uc.createValidator.Validate(ctx, validator.SignInTwoFactorCreateValidatorInput{Code: code}); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.GetActiveByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, signInTwoFactorUnavailableError(ctx, input.UserID)
		}
		return nil, fmt.Errorf("ユーザーの取得に失敗: %w", err)
	}

	setting, err := uc.userTwoFactorAuthRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("二要素認証の設定の取得に失敗: %w", err)
	}
	if setting == nil || !setting.IsEnabled() {
		return nil, signInTwoFactorUnavailableError(ctx, input.UserID)
	}

	secret, err := decryptTwoFactorSecret(uc.twoFactorKey, setting)
	if err != nil {
		return nil, err
	}
	step, ok := auth.MatchTOTPCode(secret, code, time.Now())
	if !ok {
		return nil, incorrectSignInTOTPCodeError(ctx)
	}

	if err := uc.recordMatchedStep(ctx, setting, step); err != nil {
		return nil, err
	}

	return &CreateSignInTwoFactorOutput{
		UserID:            input.UserID,
		EncryptedPassword: user.EncryptedPassword,
		Username:          user.Username,
	}, nil
}

// recordMatchedStepは、照合した設定と同じ行に限ってステップを記録する。
// 更新できないときは設定の変更とコードの再使用を区別し、設定が変わった場合はログインをやり直させる。
func (uc *CreateSignInTwoFactorUsecase) recordMatchedStep(ctx context.Context, setting *model.UserTwoFactorAuth, step int64) error {
	used, err := uc.userTwoFactorAuthRepo.UseMatchingStep(ctx, setting, step)
	if err != nil {
		return fmt.Errorf("使ったタイムステップの記録に失敗: %w", err)
	}
	if used {
		return nil
	}

	current, err := uc.userTwoFactorAuthRepo.FindByUserID(ctx, setting.UserID)
	if err != nil {
		return fmt.Errorf("更新できなかった二要素認証の設定の取得に失敗: %w", err)
	}
	if current == nil || !current.IsEnabled() || current.ID != setting.ID || !bytes.Equal(current.SecretCiphertext, setting.SecretCiphertext) {
		return signInTwoFactorUnavailableError(ctx, setting.UserID)
	}
	return incorrectSignInTOTPCodeError(ctx)
}

// incorrectSignInTOTPCodeErrorは、コードが合わなかったことをコードの欄のエラーで返す。
func incorrectSignInTOTPCodeError(ctx context.Context) *model.ValidationError {
	ve := model.NewValidationError()
	ve.AddField("code", i18n.T(ctx, "validation_totp_code_incorrect"))
	return ve
}

// signInTwoFactorUnavailableErrorは、ログイン途中のユーザーが二要素認証を使えなくなったことを表す
// AppErrCodeConflictの*model.AppErrorを返す。ハンドラーはログイン途中の状態を消してログインをやり直させる。
func signInTwoFactorUnavailableError(ctx context.Context, userID model.UserID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeConflict,
		UserMsg:  i18n.T(ctx, "flash_sign_in_two_factor_restart"),
		Internal: errSignInTwoFactorUnavailable,
		Metadata: map[string]string{"user_id": userID.String()},
	}
}
