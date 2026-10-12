package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// CreateAccountDeletionUsecaseは退会を受け付けるユースケース。ユーザーを匿名化して論理削除し、
// データの削除ジョブと退会完了メールのジョブを投入する。データの削除そのものは
// DeleteUserUsecaseがジョブで行う。
type CreateAccountDeletionUsecase struct {
	db                   *sql.DB
	userRepo             *repository.UserRepository
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	dispatcher           *dispatcher.Dispatcher
	validator            *validator.SettingsAccountDeletionCreateValidator
}

func NewCreateAccountDeletionUsecase(
	db *sql.DB,
	userRepo *repository.UserRepository,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	dispatcher *dispatcher.Dispatcher,
	validator *validator.SettingsAccountDeletionCreateValidator,
) *CreateAccountDeletionUsecase {
	return &CreateAccountDeletionUsecase{
		db:                   db,
		userRepo:             userRepo,
		stripeSubscriberRepo: stripeSubscriberRepo,
		dispatcher:           dispatcher,
		validator:            validator,
	}
}

type CreateAccountDeletionInput struct {
	User *model.User
	// Usernameは同意の確認として利用者が入力したユーザー名
	Username string
}

func (uc *CreateAccountDeletionUsecase) Execute(ctx context.Context, input CreateAccountDeletionInput) error {
	user := input.User
	if user == nil {
		return fmt.Errorf("ユーザーがnilです")
	}

	anonymizedUsername := newAnonymizedUsername()
	return uc.createAccountDeletion(ctx, user.ID, input.Username, anonymizedUsername, anonymizedUsername+"@example.com")
}

// newAnonymizedUsernameは退会したユーザーのユーザー名を置き換える値を返す。
// Rails版の退会処理と同じく、ハイフンをアンダースコアにしたUUIDにする。
func newAnonymizedUsername() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "_")
}

// createAccountDeletionはユーザーの匿名化・論理削除と、削除ジョブ・退会完了メールのジョブの投入を
// 1つのトランザクションで行う。論理削除だけがコミットされると、ログインできないままデータが残り続けるため。
func (uc *CreateAccountDeletionUsecase) createAccountDeletion(ctx context.Context, userID model.UserID, inputUsername, username, email string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	userRepo := uc.userRepo.WithTx(tx)
	lockedUser, err := userRepo.LockForAccountDeletion(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("退会対象のユーザーのロックに失敗: %w", err)
	}
	if lockedUser.Deleted {
		return nil
	}

	if err := uc.validator.WithTx(tx).Validate(ctx, validator.SettingsAccountDeletionCreateValidatorInput{
		UserID:        userID,
		Username:      lockedUser.Username,
		InputUsername: inputUsername,
	}); err != nil {
		return err
	}

	// 確認ページと同じく、有効なサブスクリプションがあるときだけ解約の案内をメールに書く。
	isStripeSupporter, err := uc.stripeSubscriberRepo.WithTx(tx).HasActiveSubscription(ctx, lockedUser.StripeSubscriberID)
	if err != nil {
		return fmt.Errorf("StripeSubscriberの取得に失敗: %w", err)
	}

	updated, err := userRepo.AnonymizeAndSoftDelete(ctx, userID, lockedUser.Username, username, email)
	if err != nil {
		return fmt.Errorf("ユーザーの匿名化に失敗: %w", err)
	}
	// ユーザー行をロックして最新の状態を確認済みのため、0件なら前提が崩れている。
	if updated == 0 {
		return fmt.Errorf("ロックしたユーザーを更新できませんでした: user_id=%d", userID)
	}

	if err := uc.dispatcher.EnqueueDeleteUserTx(ctx, tx, int64(userID)); err != nil {
		return fmt.Errorf("削除ジョブの投入に失敗: %w", err)
	}

	if err := uc.dispatcher.EnqueueAccountDeletionEmailTx(ctx, tx, lockedUser.Email, lockedUser.Locale, isStripeSupporter); err != nil {
		return fmt.Errorf("退会完了メールのジョブの投入に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "退会を受け付けました", "user_id", userID)
	return nil
}
