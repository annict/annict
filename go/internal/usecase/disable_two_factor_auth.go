package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// errTwoFactorAuthNotEnabledは、無効にしようとした時点で二要素認証が有効でないことを表す内部エラー。
var errTwoFactorAuthNotEnabled = errors.New("二要素認証が有効ではありません")

// TwoFactorAuthDisableMethodは、無効化の本人確認で受け付けた入力の種類。ログに残して後から追えるようにする。
type TwoFactorAuthDisableMethod string

const (
	// TwoFactorAuthDisableMethodTOTPは認証アプリのコードで本人確認したことを表す
	TwoFactorAuthDisableMethodTOTP TwoFactorAuthDisableMethod = "totp"
	// TwoFactorAuthDisableMethodRecoveryCodeはリカバリーコードで本人確認したことを表す
	TwoFactorAuthDisableMethodRecoveryCode TwoFactorAuthDisableMethod = "recovery_code"
	// TwoFactorAuthDisableMethodPasswordはパスワードで本人確認したことを表す
	TwoFactorAuthDisableMethodPassword TwoFactorAuthDisableMethod = "password"
)

// DisableTwoFactorAuthUsecaseは、本人確認をしてから二要素認証を無効にし、設定とリカバリーコードを削除する。
//
// 本人確認は1つの入力欄で受け、認証アプリのコード・リカバリーコード・パスワードの順で照合する。
// どれとして照合して失敗したかは利用者に明かさない。
type DisableTwoFactorAuthUsecase struct {
	db                            *sql.DB
	twoFactorKey                  *auth.TwoFactorKey
	deleteValidator               *validator.SettingsTwoFactorAuthDeleteValidator
	userRepo                      *repository.UserRepository
	userTwoFactorAuthRepo         *repository.UserTwoFactorAuthRepository
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository
	dispatcher                    *dispatcher.Dispatcher
}

// NewDisableTwoFactorAuthUsecaseはDisableTwoFactorAuthUsecaseを生成する。
func NewDisableTwoFactorAuthUsecase(
	db *sql.DB,
	twoFactorKey *auth.TwoFactorKey,
	deleteValidator *validator.SettingsTwoFactorAuthDeleteValidator,
	userRepo *repository.UserRepository,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository,
	dispatcher *dispatcher.Dispatcher,
) *DisableTwoFactorAuthUsecase {
	return &DisableTwoFactorAuthUsecase{
		db:                            db,
		twoFactorKey:                  twoFactorKey,
		deleteValidator:               deleteValidator,
		userRepo:                      userRepo,
		userTwoFactorAuthRepo:         userTwoFactorAuthRepo,
		userTwoFactorRecoveryCodeRepo: userTwoFactorRecoveryCodeRepo,
		dispatcher:                    dispatcher,
	}
}

// DisableTwoFactorAuthInputはDisableTwoFactorAuthUsecase.Executeの入力。
type DisableTwoFactorAuthInput struct {
	UserID model.UserID
	// Credentialは入力されたままのパスワード・認証アプリのコード・リカバリーコードのいずれか
	Credential string
}

// DisableTwoFactorAuthOutputはDisableTwoFactorAuthUsecase.Executeの結果。
type DisableTwoFactorAuthOutput struct {
	// Methodは本人確認で受け付けた入力の種類
	Method TwoFactorAuthDisableMethod
}

// disableCredentialは、トランザクションの前に照合した本人確認の入力。
// コードの記録とリカバリーコードの消費は同時に送られた要求と競うため、トランザクションの中で確定させる。
type disableCredential struct {
	// totpStepは認証アプリのコードとして一致したタイムステップ。一致しなければ0
	totpStep int64
	// recoveryCodeDigestはリカバリーコードの形に正規化できた入力のダイジェスト。形が違えば空
	recoveryCodeDigest string
	// passwordMatchedはパスワードとして一致したか
	passwordMatched bool
}

// Executeは本人確認をして、二要素認証の設定とリカバリーコードを1つのトランザクションで削除する。
//
// 入力が空のとき・どれとしても一致しないときは、本人確認の欄の*model.ValidationErrorを返す。
// 二要素認証が有効でないときと、照合した設定を削除できなかったとき (同時に送られた別の無効化が先に削除した・
// 照合の後に別の画面で無効にして有効にし直した) は、AppErrCodeConflictの*model.AppErrorを返す。
func (uc *DisableTwoFactorAuthUsecase) Execute(ctx context.Context, input DisableTwoFactorAuthInput) (*DisableTwoFactorAuthOutput, error) {
	if err := uc.deleteValidator.Validate(ctx, validator.SettingsTwoFactorAuthDeleteValidatorInput{Credential: input.Credential}); err != nil {
		return nil, err
	}

	setting, err := uc.userTwoFactorAuthRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("二要素認証の設定の取得に失敗: %w", err)
	}
	if setting == nil || !setting.IsEnabled() {
		return nil, notEnabledError(ctx, input.UserID)
	}

	credential, err := uc.matchCredential(ctx, setting, input.Credential)
	if err != nil {
		return nil, err
	}

	method, err := uc.disable(ctx, setting, credential)
	if err != nil {
		return nil, err
	}

	enqueueTwoFactorAuthNotification(ctx, uc.dispatcher, uc.userRepo, input.UserID, model.TwoFactorAuthNotificationEventDisabled)

	return &DisableTwoFactorAuthOutput{Method: method}, nil
}

// matchCredentialは、トランザクションの前にできる照合 (TOTPのコードとパスワード) を済ませる。
//
// パスワードのハッシュの照合は重いCPUの計算のため、トランザクションを開く前に行う。
// 認証アプリのコードとして一致したときは、パスワードを照合しない。
func (uc *DisableTwoFactorAuthUsecase) matchCredential(ctx context.Context, setting *model.UserTwoFactorAuth, input string) (*disableCredential, error) {
	totpCode := auth.NormalizeTOTPCode(input)
	if auth.IsTOTPCodeFormat(totpCode) {
		secret, err := decryptTwoFactorSecret(uc.twoFactorKey, setting)
		if err != nil {
			return nil, err
		}
		if step, ok := auth.MatchTOTPCode(secret, totpCode, time.Now()); ok {
			return &disableCredential{totpStep: step}, nil
		}
	}

	credential := &disableCredential{}
	if recoveryCode := auth.NormalizeRecoveryCode(input); auth.IsRecoveryCodeFormat(recoveryCode) {
		credential.recoveryCodeDigest = uc.twoFactorKey.RecoveryCodeDigest(recoveryCode)
	}

	user, err := uc.userRepo.GetActiveByID(ctx, setting.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザーの取得に失敗: %w", err)
	}
	// パスワードを設定していない (メールのコードでログインする) ユーザーは、ハッシュが空のため一致しない。
	credential.passwordMatched = user.EncryptedPassword != "" && auth.CheckPassword(user.EncryptedPassword, input) == nil

	return credential, nil
}

// disableは本人確認を確定させ、二要素認証の設定とリカバリーコードを削除する。
//
// 認証アプリのコードは照合した設定にステップを記録してから、その設定を削除する。
// リカバリーコードも先に設定を削除し、設定行からコード行へロックを取る順序をそろえる。
// コードを消費できなければロールバックして、設定の削除も戻す。
// リカバリーコードとして使えなかった (一致しない・使用済み) ときは、パスワードとして一致していれば受け付ける。
func (uc *DisableTwoFactorAuthUsecase) disable(ctx context.Context, setting *model.UserTwoFactorAuth, credential *disableCredential) (TwoFactorAuthDisableMethod, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	authRepo := uc.userTwoFactorAuthRepo.WithTx(tx)
	recoveryCodeRepo := uc.userTwoFactorRecoveryCodeRepo.WithTx(tx)

	totpUsed := false
	if credential.totpStep > 0 {
		used, err := authRepo.UseMatchingStep(ctx, setting, credential.totpStep)
		if err != nil {
			return "", fmt.Errorf("タイムステップの記録に失敗: %w", err)
		}
		totpUsed = used
	}

	deleted, err := authRepo.DeleteMatching(ctx, setting)
	if err != nil {
		return "", fmt.Errorf("二要素認証の設定の削除に失敗: %w", err)
	}
	if !deleted {
		return "", twoFactorAuthChangedError(ctx, setting.UserID)
	}

	var method TwoFactorAuthDisableMethod
	switch {
	case totpUsed:
		method = TwoFactorAuthDisableMethodTOTP
	case credential.recoveryCodeDigest != "":
		used, err := recoveryCodeRepo.Use(ctx, setting.UserID, credential.recoveryCodeDigest)
		if err != nil {
			return "", fmt.Errorf("リカバリーコードの消費に失敗: %w", err)
		}
		if used {
			method = TwoFactorAuthDisableMethodRecoveryCode
		}
	}
	if method == "" && credential.passwordMatched {
		method = TwoFactorAuthDisableMethodPassword
	}
	if method == "" {
		return "", credentialIncorrectError(ctx)
	}

	if err := recoveryCodeRepo.DeleteByUserID(ctx, setting.UserID); err != nil {
		return "", fmt.Errorf("リカバリーコードの削除に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return method, nil
}

// twoFactorAuthChangedErrorは、本人確認のあとに設定が削除・再登録されたことを表す競合のエラーを返す。
func twoFactorAuthChangedError(ctx context.Context, userID model.UserID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeConflict,
		UserMsg:  i18n.T(ctx, "flash_two_factor_auth_changed"),
		Metadata: map[string]string{"user_id": userID.String()},
	}
}

// credentialIncorrectErrorは、本人確認の入力がどれとしても一致しなかったことを表す*model.ValidationErrorを返す。
// どれとして照合して失敗したかを明かさないよう、1つの文面にする。
func credentialIncorrectError(ctx context.Context) *model.ValidationError {
	ve := model.NewValidationError()
	ve.AddField("credential", i18n.T(ctx, "validation_two_factor_auth_credential_incorrect"))
	return ve
}

// notEnabledErrorは、二要素認証が有効でないことを表すAppErrCodeConflictの*model.AppErrorを返す。
// ハンドラーはこのエラーで設定画面へ送り、今の状態を見せる。
func notEnabledError(ctx context.Context, userID model.UserID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeConflict,
		UserMsg:  i18n.T(ctx, "flash_two_factor_auth_not_enabled"),
		Internal: errTwoFactorAuthNotEnabled,
		Metadata: map[string]string{"user_id": userID.String()},
	}
}
