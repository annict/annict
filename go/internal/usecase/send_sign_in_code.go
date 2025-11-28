package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/worker"
	"github.com/riverqueue/river"
)

// SendSignInCodeUsecase は6桁のログインコードを生成・送信するユースケースです
type SendSignInCodeUsecase struct {
	db          *sql.DB
	queries     *query.Queries
	riverClient *worker.Client
}

// NewSendSignInCodeUsecase は新しいSendSignInCodeUsecaseを作成します
func NewSendSignInCodeUsecase(db *sql.DB, queries *query.Queries, riverClient *worker.Client) *SendSignInCodeUsecase {
	return &SendSignInCodeUsecase{
		db:          db,
		queries:     queries,
		riverClient: riverClient,
	}
}

// SendSignInCodeResult はコード送信の結果を表します
type SendSignInCodeResult struct {
	Code   string // 平文コード（テスト用、本番では使用しない）
	UserID int64  // ユーザーID
}

// Execute は6桁のログインコードを生成し、メール送信ジョブをエンキューします
func (uc *SendSignInCodeUsecase) Execute(ctx context.Context, userID int64) (*SendSignInCodeResult, error) {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queriesWithTx := uc.queries.WithTx(tx)

	// 既存の未使用コードを無効化
	if err := queriesWithTx.InvalidateUserSignInCodes(ctx, userID); err != nil {
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
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	if _, err := queriesWithTx.CreateSignInCode(ctx, params); err != nil {
		return nil, fmt.Errorf("ログインコードの作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログインコードを作成しました",
		"user_id", userID,
	)

	// メール送信ジョブをエンキュー
	if uc.riverClient != nil {
		_, err := uc.riverClient.Client().Insert(ctx, worker.SendSignInCodeArgs{
			UserID: userID,
			Code:   code,
		}, &river.InsertOpts{
			Queue: river.QueueDefault,
		})
		if err != nil {
			// ジョブエンキューに失敗してもコードは有効なので、エラーログを出力して続行
			slog.ErrorContext(ctx, "ログインコード送信ジョブのエンキューに失敗しました",
				"user_id", userID,
				"error", err,
			)
		} else {
			slog.InfoContext(ctx, "ログインコード送信ジョブをエンキューしました",
				"user_id", userID,
			)
		}
	} else {
		slog.WarnContext(ctx, "River クライアントが設定されていないため、メール送信ジョブをエンキューできませんでした",
			"user_id", userID,
		)
	}

	return &SendSignInCodeResult{
		Code:   code,
		UserID: userID,
	}, nil
}
