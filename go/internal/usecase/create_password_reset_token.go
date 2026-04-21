package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// CreatePasswordResetTokenUsecase はパスワードリセットトークンを生成するユースケースです
type CreatePasswordResetTokenUsecase struct {
	db                     *sql.DB
	userRepo               *repository.UserRepository
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
	cfg                    *config.Config
	dispatcher             *dispatcher.Dispatcher
	validator              *validator.PasswordResetCreateValidator
}

// NewCreatePasswordResetTokenUsecase は新しいCreatePasswordResetTokenUsecaseを作成します
func NewCreatePasswordResetTokenUsecase(db *sql.DB, userRepo *repository.UserRepository, passwordResetTokenRepo *repository.PasswordResetTokenRepository, cfg *config.Config, dispatcher *dispatcher.Dispatcher, validator *validator.PasswordResetCreateValidator) *CreatePasswordResetTokenUsecase {
	return &CreatePasswordResetTokenUsecase{
		db:                     db,
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		cfg:                    cfg,
		dispatcher:             dispatcher,
		validator:              validator,
	}
}

// CreatePasswordResetTokenInput はユースケースの入力パラメータです
type CreatePasswordResetTokenInput struct {
	Email string
}

// CreatePasswordResetTokenOutput はトークン生成の結果を表します
type CreatePasswordResetTokenOutput struct {
	Token  string // 平文トークン（メール送信用）
	UserID int64  // ユーザーID
}

// Execute はバリデーション・ユーザー検索・パスワードリセットトークン生成を行います。
// ユーザーが存在しない場合は nil を返す（セキュリティ対策: ユーザーの存在を明かさない）。
func (uc *CreatePasswordResetTokenUsecase) Execute(ctx context.Context, input CreatePasswordResetTokenInput) (*CreatePasswordResetTokenOutput, error) {
	// 1. バリデーション
	if err := uc.validator.Validate(ctx, validator.PasswordResetCreateValidatorInput{
		Email: input.Email,
	}); err != nil {
		return nil, err
	}

	// 2. ユーザー検索（存在しない場合もエラーを返さない - セキュリティ対策）
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ユーザーの検索に失敗: %w", err)
	}

	// 3. トークンを生成
	return uc.createToken(ctx, user.ID)
}

// createToken はパスワードリセットトークンを生成します
func (uc *CreatePasswordResetTokenUsecase) createToken(ctx context.Context, userID int64) (*CreatePasswordResetTokenOutput, error) {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	passwordResetTokenRepo := uc.passwordResetTokenRepo.WithTx(tx)

	// 既存の未使用トークンを無効化（削除）
	if err := passwordResetTokenRepo.DeleteUnusedByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("古いトークンの削除に失敗: %w", err)
	}

	// 新しいトークンを生成
	token, err := password_reset.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("トークン生成に失敗: %w", err)
	}

	// トークンをハッシュ化
	tokenDigest := password_reset.HashToken(token)

	// トークンをデータベースに保存（有効期限: 1時間）
	if _, err := passwordResetTokenRepo.Create(ctx, userID, tokenDigest, time.Now().Add(1*time.Hour)); err != nil {
		return nil, fmt.Errorf("パスワードリセットトークンの作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "パスワードリセットトークンを作成しました",
		"user_id", userID,
	)

	// メール送信ジョブをエンキュー
	if uc.dispatcher != nil {
		user, err := uc.userRepo.GetByID(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "ユーザー情報の取得に失敗しました",
				"user_id", userID,
				"error", err,
			)
		} else {
			resetURL := fmt.Sprintf("%s/password/edit?token=%s", uc.cfg.AppURL(), token)
			if err := uc.dispatcher.InsertPasswordResetEmail(ctx, user.Email, resetURL, user.Locale); err != nil {
				slog.ErrorContext(ctx, "パスワードリセットメール送信ジョブのエンキューに失敗しました",
					"user_id", userID,
					"error", err,
				)
			} else {
				slog.InfoContext(ctx, "パスワードリセットメール送信ジョブをエンキューしました",
					"user_id", userID,
				)
			}
		}
	} else {
		slog.WarnContext(ctx, "Dispatcher が設定されていないため、メール送信ジョブをエンキューできませんでした",
			"user_id", userID,
		)
	}

	return &CreatePasswordResetTokenOutput{
		Token:  token,
		UserID: userID,
	}, nil
}
