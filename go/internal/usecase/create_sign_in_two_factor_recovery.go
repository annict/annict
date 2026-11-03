package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// CreateSignInTwoFactorRecoveryUsecaseは、第1の要素を確かめたユーザーのリカバリーコードを消費し、セッションを作る。
//
// リカバリーコードの消費とセッションの作成は1つのトランザクションで行う。
// コードだけ消費されてログインできない、またはコードを消費せずにログインできる状態を作らないため。
type CreateSignInTwoFactorRecoveryUsecase struct {
	db                            *sql.DB
	twoFactorKey                  *auth.TwoFactorKey
	createValidator               *validator.SignInTwoFactorRecoveryCreateValidator
	userRepo                      *repository.UserRepository
	userTwoFactorAuthRepo         *repository.UserTwoFactorAuthRepository
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository
	createSessionUC               *CreateSessionUsecase
}

// NewCreateSignInTwoFactorRecoveryUsecaseはCreateSignInTwoFactorRecoveryUsecaseを生成する。
func NewCreateSignInTwoFactorRecoveryUsecase(
	db *sql.DB,
	twoFactorKey *auth.TwoFactorKey,
	createValidator *validator.SignInTwoFactorRecoveryCreateValidator,
	userRepo *repository.UserRepository,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository,
	createSessionUC *CreateSessionUsecase,
) *CreateSignInTwoFactorRecoveryUsecase {
	return &CreateSignInTwoFactorRecoveryUsecase{
		db:                            db,
		twoFactorKey:                  twoFactorKey,
		createValidator:               createValidator,
		userRepo:                      userRepo,
		userTwoFactorAuthRepo:         userTwoFactorAuthRepo,
		userTwoFactorRecoveryCodeRepo: userTwoFactorRecoveryCodeRepo,
		createSessionUC:               createSessionUC,
	}
}

// CreateSignInTwoFactorRecoveryInputはCreateSignInTwoFactorRecoveryUsecase.Executeの入力。
type CreateSignInTwoFactorRecoveryInput struct {
	// UserIDは第1の要素を確かめたユーザーのID。ログイン途中の状態 (セッションの値) から読んだ値を渡す
	UserID model.UserID
	// Codeは入力されたままのリカバリーコード。大文字小文字・区切り・全角はここで正規化する
	Code string
}

// CreateSignInTwoFactorRecoveryOutputはCreateSignInTwoFactorRecoveryUsecase.Executeの結果。
type CreateSignInTwoFactorRecoveryOutput struct {
	// PublicIDは作ったセッションのPublicID (Cookieの値)
	PublicID string
	UserID   model.UserID
	Username string
}

// Executeはリカバリーコードを消費し、同じトランザクションでセッションを作る。
//
// 形式の誤り・一致しない・使用済みのコードは、コードの欄の*model.ValidationErrorで返す。
// 一致しないコードと使用済みのコードは同じ文言にし、どちらかを明かさない。
// ユーザーが退会した・二要素認証を無効にした・照合のあとに設定が入れ替わったときは、
// ログインをやり直させるためAppErrCodeConflictを返す。
func (uc *CreateSignInTwoFactorRecoveryUsecase) Execute(ctx context.Context, input CreateSignInTwoFactorRecoveryInput) (*CreateSignInTwoFactorRecoveryOutput, error) {
	code := auth.NormalizeRecoveryCode(input.Code)
	if err := uc.createValidator.Validate(ctx, validator.SignInTwoFactorRecoveryCreateValidatorInput{Code: code}); err != nil {
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

	digest := uc.twoFactorKey.RecoveryCodeDigest(code)
	publicID, err := uc.signIn(ctx, setting, digest, user.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	return &CreateSignInTwoFactorRecoveryOutput{
		PublicID: publicID,
		UserID:   input.UserID,
		Username: user.Username,
	}, nil
}

// signInは、照合した設定の行をロックしてからリカバリーコードを消費し、セッションを作る。
//
// 設定の行を先にロックするのは、無効化 (設定の行からリカバリーコードの行の順) とロックの順序をそろえ、
// 同時に送られた無効化とデッドロックしないようにするため。
// ロックできないとき (照合のあとに設定が無効になった・入れ替わった) は、ログインをやり直させる。
func (uc *CreateSignInTwoFactorRecoveryUsecase) signIn(ctx context.Context, setting *model.UserTwoFactorAuth, digest string, encryptedPassword string) (string, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	locked, err := uc.userTwoFactorAuthRepo.WithTx(tx).LockMatching(ctx, setting)
	if err != nil {
		return "", fmt.Errorf("二要素認証の設定のロックに失敗: %w", err)
	}
	if !locked {
		return "", signInTwoFactorUnavailableError(ctx, setting.UserID)
	}

	used, err := uc.userTwoFactorRecoveryCodeRepo.WithTx(tx).Use(ctx, setting.UserID, digest)
	if err != nil {
		return "", fmt.Errorf("リカバリーコードの消費に失敗: %w", err)
	}
	if !used {
		return "", incorrectRecoveryCodeError(ctx)
	}

	sessionResult, err := uc.createSessionUC.Execute(ctx, tx, setting.UserID, encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("セッションの作成に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return sessionResult.PublicID, nil
}

// incorrectRecoveryCodeErrorは、リカバリーコードが一致しない・使用済みであることをコードの欄のエラーで返す。
func incorrectRecoveryCodeError(ctx context.Context) *model.ValidationError {
	ve := model.NewValidationError()
	ve.AddField("code", i18n.T(ctx, "validation_recovery_code_incorrect"))
	return ve
}
