package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
)

// VerifySignInCodeUsecase は6桁のログインコードを検証するユースケースです
type VerifySignInCodeUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewVerifySignInCodeUsecase は新しいVerifySignInCodeUsecaseを作成します
func NewVerifySignInCodeUsecase(db *sql.DB, queries *query.Queries) *VerifySignInCodeUsecase {
	return &VerifySignInCodeUsecase{
		db:      db,
		queries: queries,
	}
}

var (
	// ErrCodeNotFound はコードが見つからない場合のエラーです
	ErrCodeNotFound = errors.New("コードが見つからないか、有効期限が切れています")
	// ErrCodeInvalid はコードが間違っている場合のエラーです
	ErrCodeInvalid = errors.New("コードが正しくありません")
	// ErrCodeAttemptsExceeded は試行回数が上限に達した場合のエラーです
	ErrCodeAttemptsExceeded = errors.New("試行回数が上限に達しました。新しいコードを送信してください")
)

// Execute は6桁のログインコードを検証します
// コードが正しい場合はnilを返し、コードを使用済みにします
// コードが間違っている場合は試行回数をインクリメントし、エラーを返します
func (uc *VerifySignInCodeUsecase) Execute(ctx context.Context, userID int64, code string) error {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// 有効なコードを取得（未使用 AND 有効期限内）
	signInCode, err := queriesWithTx.GetValidSignInCode(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCodeNotFound
		}
		return fmt.Errorf("コードの取得に失敗: %w", err)
	}

	// 試行回数チェック（5回まで）
	if signInCode.Attempts >= 5 {
		// 試行回数が上限に達している場合、コードを無効化
		if err := queriesWithTx.MarkSignInCodeAsUsed(ctx, signInCode.ID); err != nil {
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
		if err := queriesWithTx.IncrementSignInCodeAttempts(ctx, signInCode.ID); err != nil {
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
	if err := queriesWithTx.MarkSignInCodeAsUsed(ctx, signInCode.ID); err != nil {
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
