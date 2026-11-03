package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// ErrInvalidPasswordResetTokenはパスワードリセットトークンが無効であることを示すエラーです。
// Handlerはこのエラーをerrors.Isで判定してトークン無効時の処理を行います。
var ErrInvalidPasswordResetToken = repository.ErrInvalidPasswordResetToken

// UpdatePasswordResetUsecaseはパスワードリセットによるパスワード更新を行うユースケースです。
//
// パスワードを更新してもセッションは作らず、利用者には新しいパスワードでログインし直してもらう。
// ログインの経路を通さないと、二要素認証を有効にしたユーザーが認証アプリのコードを求められずにログインできてしまうため。
type UpdatePasswordResetUsecase struct {
	db                     *sql.DB
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
	userRepo               *repository.UserRepository
	validator              *validator.PasswordUpdateValidator
}

// NewUpdatePasswordResetUsecaseは新しいUpdatePasswordResetUsecaseを作成します
func NewUpdatePasswordResetUsecase(db *sql.DB, passwordResetTokenRepo *repository.PasswordResetTokenRepository, userRepo *repository.UserRepository, validator *validator.PasswordUpdateValidator) *UpdatePasswordResetUsecase {
	return &UpdatePasswordResetUsecase{
		db:                     db,
		passwordResetTokenRepo: passwordResetTokenRepo,
		userRepo:               userRepo,
		validator:              validator,
	}
}

// UpdatePasswordResetInputはユースケースの入力パラメータです
type UpdatePasswordResetInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// UpdatePasswordResetOutputはパスワード更新の結果を表します
type UpdatePasswordResetOutput struct {
	UserID model.UserID
}

// Executeはバリデーションとパスワードの更新を行います。セッションは作りません。
func (uc *UpdatePasswordResetUsecase) Execute(ctx context.Context, input UpdatePasswordResetInput) (*UpdatePasswordResetOutput, error) {
	// 1. バリデーション
	if err := uc.validator.Validate(ctx, validator.PasswordUpdateValidatorInput{
		Token:                input.Token,
		Password:             input.Password,
		PasswordConfirmation: input.PasswordConfirmation,
	}); err != nil {
		return nil, err
	}

	// 2. トークンの有効性を検証
	resetToken, err := uc.passwordResetTokenRepo.GetValidByToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}

	// 3. パスワードをハッシュ化する。bcryptは重い計算のため、トランザクションを開く前に行う
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	// 4. パスワードの更新とトークンの使用済み化を1つのトランザクションで行う
	if err := uc.updatePassword(ctx, resetToken, hashedPassword); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "パスワード変更が成功しました",
		"user_id", resetToken.UserID,
	)

	return &UpdatePasswordResetOutput{
		UserID: resetToken.UserID,
	}, nil
}

// updatePasswordは、パスワードの更新とトークンの使用済み化を1つのトランザクションで行う。
// トークンだけが使用済みになって、パスワードが変わらない状態を作らないため。
func (uc *UpdatePasswordResetUsecase) updatePassword(ctx context.Context, resetToken *model.PasswordResetToken, hashedPassword string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	updated, err := uc.userRepo.WithTx(tx).UpdatePassword(ctx, repository.UpdateUserPasswordParams{
		ID:                int64(resetToken.UserID),
		EncryptedPassword: hashedPassword,
	})
	if err != nil {
		return fmt.Errorf("パスワードの更新に失敗: %w", err)
	}
	if !updated {
		// トークンの検証後に退会・利用停止が確定した場合も、トークンを消費せずに拒否する。
		return ErrInvalidPasswordResetToken
	}

	marked, err := uc.passwordResetTokenRepo.WithTx(tx).MarkAsUsed(ctx, resetToken.ID)
	if err != nil {
		return fmt.Errorf("トークンの使用済みマークに失敗: %w", err)
	}
	if !marked {
		// トークンの検証後に、同じトークンを使った別のリセットが先に確定した場合は、パスワードの更新も戻して拒否する。
		return ErrInvalidPasswordResetToken
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}
	return nil
}
