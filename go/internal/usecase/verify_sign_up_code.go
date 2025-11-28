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

// VerifySignUpCodeUsecase は6桁の新規登録確認コードを検証するユースケースです
type VerifySignUpCodeUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewVerifySignUpCodeUsecase は新しいVerifySignUpCodeUsecaseを作成します
func NewVerifySignUpCodeUsecase(db *sql.DB, queries *query.Queries) *VerifySignUpCodeUsecase {
	return &VerifySignUpCodeUsecase{
		db:      db,
		queries: queries,
	}
}

// Execute は6桁の新規登録確認コードを検証します
// コードが正しい場合はnilを返し、コードを使用済みにします
// コードが間違っている場合は試行回数をインクリメントし、エラーを返します
func (uc *VerifySignUpCodeUsecase) Execute(ctx context.Context, email string, code string) error {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// 有効なコードを取得（未使用 AND 有効期限内）
	signUpCode, err := queriesWithTx.GetValidSignUpCode(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCodeNotFound
		}
		return fmt.Errorf("コードの取得に失敗: %w", err)
	}

	// 試行回数チェック（5回まで）
	if signUpCode.Attempts >= 5 {
		// 試行回数が上限に達している場合、コードを無効化
		if err := queriesWithTx.MarkSignUpCodeAsUsed(ctx, signUpCode.ID); err != nil {
			return fmt.Errorf("コードの無効化に失敗: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
		}

		slog.WarnContext(ctx, "試行回数が上限に達したため、コードを無効化しました",
			"email", email,
			"sign_up_code_id", signUpCode.ID,
		)

		return ErrCodeAttemptsExceeded
	}

	// コード検証（bcryptで比較）
	if !auth.VerifyCode(code, signUpCode.CodeDigest) {
		// コードが間違っている場合、試行回数をインクリメント
		if err := queriesWithTx.IncrementSignUpCodeAttempts(ctx, signUpCode.ID); err != nil {
			return fmt.Errorf("試行回数のインクリメントに失敗: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
		}

		slog.WarnContext(ctx, "コードが正しくありません",
			"email", email,
			"sign_up_code_id", signUpCode.ID,
			"attempts", signUpCode.Attempts+1,
		)

		return ErrCodeInvalid
	}

	// コードが正しい場合、使用済みにする
	if err := queriesWithTx.MarkSignUpCodeAsUsed(ctx, signUpCode.ID); err != nil {
		return fmt.Errorf("コードの使用済み設定に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "コード検証に成功しました",
		"email", email,
		"sign_up_code_id", signUpCode.ID,
	)

	return nil
}
