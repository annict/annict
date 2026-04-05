package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/validator"
)

// SendSignUpCodeUsecase は新規登録確認コードを生成・送信するユースケースです
type SendSignUpCodeUsecase struct {
	db             *sql.DB
	signUpCodeRepo *repository.SignUpCodeRepository
	userRepo       *repository.UserRepository
	dispatcher     *dispatcher.Dispatcher
	validator      *validator.CreateSignUpValidator
}

// NewSendSignUpCodeUsecase は新しいSendSignUpCodeUsecaseを作成します
func NewSendSignUpCodeUsecase(
	db *sql.DB,
	signUpCodeRepo *repository.SignUpCodeRepository,
	userRepo *repository.UserRepository,
	dispatcher *dispatcher.Dispatcher,
	validator *validator.CreateSignUpValidator,
) *SendSignUpCodeUsecase {
	return &SendSignUpCodeUsecase{
		db:             db,
		signUpCodeRepo: signUpCodeRepo,
		userRepo:       userRepo,
		dispatcher:     dispatcher,
		validator:      validator,
	}
}

// SendSignUpCodeInput はユースケースの入力パラメータです
type SendSignUpCodeInput struct {
	Email  string
	Locale string
}

// SendSignUpCodeResult はコード送信の結果を表します
type SendSignUpCodeResult struct {
	FormErrors *session.FormErrors // バリデーションエラー（nilなら成功）
	Code       string              // 平文コード（テスト用）
	Email      string              // メールアドレス
}

// Execute は新規登録確認コードを生成し、メール送信ジョブをエンキューします
func (uc *SendSignUpCodeUsecase) Execute(ctx context.Context, input SendSignUpCodeInput) (*SendSignUpCodeResult, error) {
	// 1. バリデーション
	valResult := uc.validator.Validate(ctx, validator.CreateSignUpValidatorInput{
		Email: input.Email,
	})
	if valResult.FormErrors != nil && valResult.FormErrors.HasErrors() {
		return &SendSignUpCodeResult{FormErrors: valResult.FormErrors}, nil
	}

	// 2. メールアドレスの重複チェック
	_, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		formErrors := &session.FormErrors{}
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_already_exists"))
		return &SendSignUpCodeResult{FormErrors: formErrors}, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("ユーザーの検索に失敗: %w", err)
	}

	// 3. トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	signUpCodeRepoTx := uc.signUpCodeRepo.WithTx(tx)

	// 既存の未使用コードを無効化
	if err := signUpCodeRepoTx.InvalidateByEmail(ctx, input.Email); err != nil {
		return nil, fmt.Errorf("古いコードの無効化に失敗: %w", err)
	}

	// 新しい6桁コードを生成
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("コード生成に失敗: %w", err)
	}

	// コードをbcryptでハッシュ化
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		return nil, fmt.Errorf("コードのハッシュ化に失敗: %w", err)
	}

	// コードをデータベースに保存（有効期限: 15分）
	if err := signUpCodeRepoTx.Create(ctx, repository.SignUpCodeCreateParams{
		Email:      input.Email,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}); err != nil {
		return nil, fmt.Errorf("新規登録確認コードの作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "新規登録確認コードを作成しました",
		"email", input.Email,
	)

	// メール送信ジョブをエンキュー
	if uc.dispatcher != nil {
		if err := uc.dispatcher.InsertSignUpCodeEmail(ctx, input.Email, code, input.Locale); err != nil {
			slog.ErrorContext(ctx, "新規登録確認コード送信ジョブのエンキューに失敗しました",
				"email", input.Email,
				"error", err,
			)
		} else {
			slog.InfoContext(ctx, "新規登録確認コード送信ジョブをエンキューしました",
				"email", input.Email,
			)
		}
	} else {
		slog.WarnContext(ctx, "Dispatcher が設定されていないため、メール送信ジョブをエンキューできませんでした",
			"email", input.Email,
		)
	}

	return &SendSignUpCodeResult{
		Code:  code,
		Email: input.Email,
	}, nil
}
