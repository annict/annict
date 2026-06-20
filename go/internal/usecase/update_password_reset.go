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

// ErrInvalidPasswordResetToken はパスワードリセットトークンが無効であることを示すエラーです。
// Handler はこのエラーを errors.Is で判定してトークン無効時の処理を行います。
var ErrInvalidPasswordResetToken = repository.ErrInvalidPasswordResetToken

// UpdatePasswordResetUsecase はパスワードリセットによるパスワード更新を行うユースケースです
type UpdatePasswordResetUsecase struct {
	db                     *sql.DB
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
	userRepo               *repository.UserRepository
	sessionRepo            *repository.SessionRepository
	validator              *validator.PasswordUpdateValidator
}

// NewUpdatePasswordResetUsecase は新しいUpdatePasswordResetUsecaseを作成します
func NewUpdatePasswordResetUsecase(db *sql.DB, passwordResetTokenRepo *repository.PasswordResetTokenRepository, userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, validator *validator.PasswordUpdateValidator) *UpdatePasswordResetUsecase {
	return &UpdatePasswordResetUsecase{
		db:                     db,
		passwordResetTokenRepo: passwordResetTokenRepo,
		userRepo:               userRepo,
		sessionRepo:            sessionRepo,
		validator:              validator,
	}
}

// UpdatePasswordResetInput はユースケースの入力パラメータです
type UpdatePasswordResetInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// UpdatePasswordResetOutput はパスワード更新の結果を表します
type UpdatePasswordResetOutput struct {
	UserID    model.UserID
	SessionID string // 新しいセッションID
}

// Execute はバリデーション・パスワード更新・セッション作成を行います
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

	// 3. トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	passwordResetTokenRepoTx := uc.passwordResetTokenRepo.WithTx(tx)
	userRepoTx := uc.userRepo.WithTx(tx)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	// パスワードを更新
	if err := userRepoTx.UpdatePassword(ctx, repository.UpdateUserPasswordParams{
		ID:                int64(resetToken.UserID),
		EncryptedPassword: hashedPassword,
	}); err != nil {
		return nil, fmt.Errorf("パスワードの更新に失敗: %w", err)
	}

	// トークンを使用済みにマーク
	if err := passwordResetTokenRepoTx.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return nil, fmt.Errorf("トークンの使用済みマークに失敗: %w", err)
	}

	// 新しいセッションを作成
	user, err := userRepoTx.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗: %w", err)
	}

	// NOT NULL制約があるフィールドの検証（fail-fast）
	if err := validator.ValidateNotNullTime(user.CreatedAt, "created_at", user.ID); err != nil {
		slog.ErrorContext(ctx, "データベース制約違反を検出しました（NOT NULL制約）",
			"table", "users",
			"field", "created_at",
			"user_id", user.ID,
			"error", err,
		)
		return nil, err
	}
	if err := validator.ValidateNotNullTime(user.UpdatedAt, "updated_at", user.ID); err != nil {
		slog.ErrorContext(ctx, "データベース制約違反を検出しました（NOT NULL制約）",
			"table", "users",
			"field", "updated_at",
			"user_id", user.ID,
			"error", err,
		)
		return nil, err
	}

	// CreateSessionUsecaseを使用してセッションを作成
	createSessionUC := NewCreateSessionUsecase(uc.sessionRepo)
	sessionResult, err := createSessionUC.Execute(ctx, tx, model.UserID(user.ID), user.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
	}
	sessionID := sessionResult.PublicID

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "パスワード変更が成功しました",
		"user_id", resetToken.UserID,
	)

	return &UpdatePasswordResetOutput{
		UserID:    resetToken.UserID,
		SessionID: sessionID,
	}, nil
}
