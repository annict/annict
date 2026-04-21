package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// SendSignInCodeUsecase はサインインの入力検証・ユーザー検索・ログインコード送信を担当するユースケースです
type SendSignInCodeUsecase struct {
	db             *sql.DB
	signInCodeRepo *repository.SignInCodeRepository
	userRepo       *repository.UserRepository
	dispatcher     *dispatcher.Dispatcher
	validator      *validator.SignInCreateValidator
}

// NewSendSignInCodeUsecase は新しいSendSignInCodeUsecaseを作成します
func NewSendSignInCodeUsecase(
	db *sql.DB,
	signInCodeRepo *repository.SignInCodeRepository,
	userRepo *repository.UserRepository,
	dispatcher *dispatcher.Dispatcher,
	validator *validator.SignInCreateValidator,
) *SendSignInCodeUsecase {
	return &SendSignInCodeUsecase{
		db:             db,
		signInCodeRepo: signInCodeRepo,
		userRepo:       userRepo,
		dispatcher:     dispatcher,
		validator:      validator,
	}
}

// SendSignInCodeInput はユースケースの入力パラメータです
type SendSignInCodeInput struct {
	Email string
}

// SendSignInCodeOutput はユースケースの結果を表します
type SendSignInCodeOutput struct {
	UserID      int64  // ユーザーID
	Email       string // メールアドレス
	HasPassword bool   // パスワードログインを使用するかどうか
	Code        string // 平文コード（テスト用、コードログインの場合のみ）
}

// Execute はサインインの入力検証・ユーザー検索を行い、パスワードなしユーザーの場合はコードを生成・送信します
func (uc *SendSignInCodeUsecase) Execute(ctx context.Context, input SendSignInCodeInput) (*SendSignInCodeOutput, error) {
	// 1. バリデーション
	if err := uc.validator.Validate(ctx, validator.SignInCreateValidatorInput{
		Email: input.Email,
	}); err != nil {
		return nil, err
	}

	// 2. ユーザー検索
	user, err := uc.userRepo.GetByEmailForSignIn(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ve := model.NewValidationError()
			ve.AddField("email", i18n.T(ctx, "sign_in_user_not_found"))
			return nil, ve
		}
		return nil, fmt.Errorf("ユーザーの検索に失敗: %w", err)
	}

	// 3. パスワードの有無を確認
	if user.EncryptedPassword != "" {
		slog.InfoContext(ctx, "ユーザーはパスワードログインを使用します",
			"user_id", user.ID,
			"email", user.Email,
		)
		return &SendSignInCodeOutput{
			UserID:      user.ID,
			Email:       user.Email,
			HasPassword: true,
		}, nil
	}

	// 4. パスワードなしの場合: コードを生成・送信
	slog.InfoContext(ctx, "ユーザーはメールログインを使用します",
		"user_id", user.ID,
		"email", user.Email,
	)

	code, err := uc.generateAndSendCode(ctx, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &SendSignInCodeOutput{
		UserID:      user.ID,
		Email:       user.Email,
		HasPassword: false,
		Code:        code,
	}, nil
}

// generateAndSendCode は6桁のログインコードを生成し、メール送信ジョブをエンキューします
func (uc *SendSignInCodeUsecase) generateAndSendCode(ctx context.Context, userID int64, email string) (string, error) {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	signInCodeRepoTx := uc.signInCodeRepo.WithTx(tx)

	// 既存の未使用コードを無効化
	if err := signInCodeRepoTx.InvalidateByUserID(ctx, userID); err != nil {
		return "", fmt.Errorf("古いコードの無効化に失敗: %w", err)
	}

	// 新しい6桁コードを生成
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("コード生成に失敗: %w", err)
	}

	// コードをbcryptでハッシュ化
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		return "", fmt.Errorf("コードのハッシュ化に失敗: %w", err)
	}

	// コードをデータベースに保存（有効期限: 15分）
	if err := signInCodeRepoTx.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}); err != nil {
		return "", fmt.Errorf("ログインコードの作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログインコードを作成しました",
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
			if err := uc.dispatcher.InsertSignInCodeEmail(ctx, user.Email, code, user.Locale); err != nil {
				slog.ErrorContext(ctx, "ログインコード送信ジョブのエンキューに失敗しました",
					"user_id", userID,
					"error", err,
				)
			} else {
				slog.InfoContext(ctx, "ログインコード送信ジョブをエンキューしました",
					"user_id", userID,
				)
			}
		}
	} else {
		slog.WarnContext(ctx, "Dispatcher が設定されていないため、メール送信ジョブをエンキューできませんでした",
			"user_id", userID,
		)
	}

	return code, nil
}
