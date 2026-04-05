package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/validator"
)

// VerifySignInCodeUsecase は6桁のログインコードを検証するユースケースです
type VerifySignInCodeUsecase struct {
	db             *sql.DB
	signInCodeRepo *repository.SignInCodeRepository
	userRepo       *repository.UserRepository
	v              *validator.CreateSignInCodeValidator
}

// NewVerifySignInCodeUsecase は新しいVerifySignInCodeUsecaseを作成します
func NewVerifySignInCodeUsecase(
	db *sql.DB,
	signInCodeRepo *repository.SignInCodeRepository,
	userRepo *repository.UserRepository,
	v *validator.CreateSignInCodeValidator,
) *VerifySignInCodeUsecase {
	return &VerifySignInCodeUsecase{
		db:             db,
		signInCodeRepo: signInCodeRepo,
		userRepo:       userRepo,
		v:              v,
	}
}

// VerifySignInCodeInput はユースケースの入力パラメータです
type VerifySignInCodeInput struct {
	UserID int64
	Code   string
}

// VerifySignInCodeResult はユースケースの結果を表します
type VerifySignInCodeResult struct {
	FormErrors        *session.FormErrors // バリデーションエラー（nilなら成功）
	EncryptedPassword string              // ユーザーのパスワードハッシュ（セッション作成用）
	Username          string              // ユーザー名（ログ用）
}

var (
	// ErrCodeNotFound はコードが見つからない場合のエラーです
	ErrCodeNotFound = errors.New("コードが見つからないか、有効期限が切れています")
	// ErrCodeInvalid はコードが間違っている場合のエラーです
	ErrCodeInvalid = errors.New("コードが正しくありません")
	// ErrCodeAttemptsExceeded は試行回数が上限に達した場合のエラーです
	ErrCodeAttemptsExceeded = errors.New("試行回数が上限に達しました。新しいコードを送信してください")
)

// Execute は6桁のログインコードを検証し、成功時にユーザー情報を返します
func (uc *VerifySignInCodeUsecase) Execute(ctx context.Context, input VerifySignInCodeInput) (*VerifySignInCodeResult, error) {
	// 1. バリデーション
	valResult := uc.v.Validate(ctx, validator.CreateSignInCodeValidatorInput{
		Code: input.Code,
	})
	if valResult.FormErrors != nil && valResult.FormErrors.HasErrors() {
		return &VerifySignInCodeResult{FormErrors: valResult.FormErrors}, nil
	}

	// 2. コード検証
	if err := uc.verifyCode(ctx, input.UserID, input.Code); err != nil {
		return nil, err
	}

	// 3. ユーザー情報を取得
	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ユーザーが見つかりません (user_id=%d): %w", input.UserID, err)
		}
		return nil, fmt.Errorf("ユーザー情報の取得に失敗: %w", err)
	}

	return &VerifySignInCodeResult{
		EncryptedPassword: user.EncryptedPassword,
		Username:          user.Username,
	}, nil
}

// verifyCode は6桁のログインコードを検証します
func (uc *VerifySignInCodeUsecase) verifyCode(ctx context.Context, userID int64, code string) error {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	signInCodeRepoTx := uc.signInCodeRepo.WithTx(tx)

	// 有効なコードを取得（未使用 AND 有効期限内）
	signInCode, err := signInCodeRepoTx.GetValidByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCodeNotFound
		}
		return fmt.Errorf("コードの取得に失敗: %w", err)
	}

	// 試行回数チェック（5回まで）
	if signInCode.Attempts >= 5 {
		// 試行回数が上限に達している場合、コードを無効化
		if err := signInCodeRepoTx.MarkAsUsed(ctx, signInCode.ID); err != nil {
			return fmt.Errorf("コードの無効化に失敗: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
		}

		slog.WarnContext(ctx, "試行回数が上限に達したため、コードを無効化しました",
			"user_id", userID,
			"sign_in_code_id", signInCode.ID,
		)

		return ErrCodeAttemptsExceeded
	}

	// コード検証（bcryptで比較）
	if !auth.VerifyCode(code, signInCode.CodeDigest) {
		// コードが間違っている場合、試行回数をインクリメント
		if err := signInCodeRepoTx.IncrementAttempts(ctx, signInCode.ID); err != nil {
			return fmt.Errorf("試行回数のインクリメントに失敗: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
		}

		slog.WarnContext(ctx, "コードが正しくありません",
			"user_id", userID,
			"sign_in_code_id", signInCode.ID,
			"attempts", signInCode.Attempts+1,
		)

		return ErrCodeInvalid
	}

	// コードが正しい場合、使用済みにする
	if err := signInCodeRepoTx.MarkAsUsed(ctx, signInCode.ID); err != nil {
		return fmt.Errorf("コードの使用済み設定に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "コード検証に成功しました",
		"user_id", userID,
		"sign_in_code_id", signInCode.ID,
	)

	return nil
}
