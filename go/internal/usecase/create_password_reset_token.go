package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/internal/password_reset"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/worker"
	"github.com/riverqueue/river"
)

// CreatePasswordResetTokenUsecase はパスワードリセットトークンを生成するユースケースです
type CreatePasswordResetTokenUsecase struct {
	db          *sql.DB
	queries     *query.Queries
	riverClient *worker.Client
}

// NewCreatePasswordResetTokenUsecase は新しいCreatePasswordResetTokenUsecaseを作成します
func NewCreatePasswordResetTokenUsecase(db *sql.DB, queries *query.Queries, riverClient *worker.Client) *CreatePasswordResetTokenUsecase {
	return &CreatePasswordResetTokenUsecase{
		db:          db,
		queries:     queries,
		riverClient: riverClient,
	}
}

// CreatePasswordResetTokenResult はトークン生成の結果を表します
type CreatePasswordResetTokenResult struct {
	Token  string // 平文トークン（メール送信用）
	UserID int64  // ユーザーID
}

// Execute はパスワードリセットトークンを生成します
func (uc *CreatePasswordResetTokenUsecase) Execute(ctx context.Context, userID int64) (*CreatePasswordResetTokenResult, error) {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// 既存の未使用トークンを無効化（削除）
	if err := queriesWithTx.DeleteUnusedPasswordResetTokensByUserID(ctx, userID); err != nil {
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
	params := query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	if _, err := queriesWithTx.CreatePasswordResetToken(ctx, params); err != nil {
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
	// 注: トランザクション外でエンキューするため、トークン保存とジョブエンキューの一貫性は完全ではありません
	// ただし、トークンは保存されているため、ジョブエンキューに失敗してもユーザーはリセットリンクを使用できます
	if uc.riverClient != nil {
		_, err := uc.riverClient.Client().Insert(ctx, worker.SendPasswordResetEmailArgs{
			UserID: userID,
			Token:  token,
		}, &river.InsertOpts{
			Queue: river.QueueDefault,
		})
		if err != nil {
			// ジョブエンキューに失敗してもトークンは有効なので、エラーログを出力して続行
			slog.ErrorContext(ctx, "パスワードリセットメール送信ジョブのエンキューに失敗しました",
				"user_id", userID,
				"error", err,
			)
		} else {
			slog.InfoContext(ctx, "パスワードリセットメール送信ジョブをエンキューしました",
				"user_id", userID,
			)
		}
	} else {
		slog.WarnContext(ctx, "River クライアントが設定されていないため、メール送信ジョブをエンキューできませんでした",
			"user_id", userID,
		)
	}

	return &CreatePasswordResetTokenResult{
		Token:  token,
		UserID: userID,
	}, nil
}
