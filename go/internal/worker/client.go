// Package worker はバックグラウンドワーカー機能を提供します
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/mail"
)

// Client は River クライアントのラッパー
type Client struct {
	riverClient *river.Client[pgx.Tx]
	pool        *pgxpool.Pool
}

// NewClientParams は NewClient に渡すパラメータです
type NewClientParams struct {
	CleanupExpiredTokens      ExpiredTokenCleaner
	CleanupExpiredSignInCodes ExpiredSignInCodeCleaner
}

// NewClient は新しい River クライアントを作成します
func NewClient(ctx context.Context, databaseURL string, params NewClientParams, cfg *config.Config) (*Client, error) {
	// pgxpool の作成
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// コネクションプール設定
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// メールクライアントの作成（メール送信用）
	var mailSender mail.Sender
	if cfg.ResendAPIKey != "" {
		mailSender = mail.NewResendClient(cfg.ResendAPIKey, cfg.ResendFromEmail, "Annict")
		slog.InfoContext(ctx, "Resend クライアントを初期化しました")
	} else {
		slog.WarnContext(ctx, "Resend API キーが設定されていません。メール送信機能は利用できません")
	}

	// River ワーカーの登録
	workers := river.NewWorkers()

	// メール送信ワーカーを登録
	if mailSender != nil {
		river.AddWorker(workers, NewSendEmailWorker(mailSender))
		slog.InfoContext(ctx, "SendEmailWorker を登録しました")
	}

	// トークンクリーンアップワーカーを登録
	river.AddWorker(workers, NewCleanupExpiredTokensWorker(params.CleanupExpiredTokens))
	slog.InfoContext(ctx, "CleanupExpiredTokensWorker を登録しました")

	// ログインコードクリーンアップワーカーを登録
	river.AddWorker(workers, NewCleanupExpiredSignInCodesWorker(params.CleanupExpiredSignInCodes))
	slog.InfoContext(ctx, "CleanupExpiredSignInCodesWorker を登録しました")

	// River クライアントの作成
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
		Logger:  slog.Default(),
	})
	if err != nil {
		pool.Close()
		return nil, err
	}

	return &Client{
		riverClient: riverClient,
		pool:        pool,
	}, nil
}

// Start は River クライアントを起動します
func (c *Client) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "River クライアントを起動します")
	return c.riverClient.Start(ctx)
}

// Stop は River クライアントを停止します
func (c *Client) Stop(ctx context.Context) error {
	slog.InfoContext(ctx, "River クライアントを停止します")
	if err := c.riverClient.Stop(ctx); err != nil {
		return err
	}
	c.pool.Close()
	return nil
}

// Client は River クライアントへのアクセスを提供します
func (c *Client) Client() *river.Client[pgx.Tx] {
	return c.riverClient
}
