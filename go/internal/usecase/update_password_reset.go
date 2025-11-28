package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/password_reset"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/validator"
)

// UpdatePasswordResetUsecase はパスワードリセットによるパスワード更新を行うユースケースです
type UpdatePasswordResetUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewUpdatePasswordResetUsecase は新しいUpdatePasswordResetUsecaseを作成します
func NewUpdatePasswordResetUsecase(db *sql.DB, queries *query.Queries) *UpdatePasswordResetUsecase {
	return &UpdatePasswordResetUsecase{
		db:      db,
		queries: queries,
	}
}

// UpdatePasswordResetResult はパスワード更新の結果を表します
type UpdatePasswordResetResult struct {
	UserID    int64
	SessionID string // 新しいセッションID
}

// Execute はパスワードを更新し、新しいセッションを作成します
func (uc *UpdatePasswordResetUsecase) Execute(ctx context.Context, token, newPassword string) (*UpdatePasswordResetResult, error) {
	// トークンをハッシュ化してデータベースから検索
	tokenDigest := password_reset.HashToken(token)
	resetToken, err := uc.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
	if err == sql.ErrNoRows {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "not_found",
		)
		return nil, fmt.Errorf("invalid token")
	} else if err != nil {
		return nil, fmt.Errorf("パスワードリセットトークンの取得に失敗: %w", err)
	}

	// トークンの有効性をチェック
	if resetToken.UsedAt.Valid {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "used",
		)
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "expired",
		)
		return nil, fmt.Errorf("invalid token")
	}

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	// パスワードを更新
	params := query.UpdateUserPasswordParams{
		ID:                resetToken.UserID,
		EncryptedPassword: hashedPassword,
	}

	if err := queriesWithTx.UpdateUserPassword(ctx, params); err != nil {
		return nil, fmt.Errorf("パスワードの更新に失敗: %w", err)
	}

	// トークンを使用済みにマーク
	if err := queriesWithTx.MarkPasswordResetTokenAsUsed(ctx, resetToken.ID); err != nil {
		return nil, fmt.Errorf("トークンの使用済みマークに失敗: %w", err)
	}

	// 新しいセッションを作成
	// まずユーザー情報を取得（encrypted_passwordが必要）
	user, err := queriesWithTx.GetUserByID(ctx, resetToken.UserID)
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
	createSessionUC := NewCreateSessionUsecase(uc.queries)
	sessionResult, err := createSessionUC.Execute(ctx, tx, user.ID, user.EncryptedPassword, "")
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

	return &UpdatePasswordResetResult{
		UserID:    resetToken.UserID,
		SessionID: sessionID,
	}, nil
}
