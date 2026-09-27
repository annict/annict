// Package workerはバックグラウンドワーカー機能を提供します
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/email"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	"github.com/annict/annict/go/internal/usecase"
)

// ClientはRiverクライアントのラッパー
type Client struct {
	riverClient *river.Client[pgx.Tx]
	pool        *pgxpool.Pool
}

// NewClientParamsはNewClientに渡すパラメータです
type NewClientParams struct {
	CleanupExpiredTokens      ExpiredTokenCleaner
	CleanupExpiredSignInCodes ExpiredSignInCodeCleaner
	CleanupExpiredSessions    ExpiredSessionCleaner
	SyncAnimes                AnimesSyncer
}

// NewClientは新しいRiverクライアントを作成します
func NewClient(ctx context.Context, databaseURL string, params NewClientParams, cfg *config.Config) (*Client, error) {
	// pgxpoolの作成
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

	// メール送信クライアントの作成
	var emailSender email.Sender
	if cfg.ResendAPIKey != "" {
		emailSender = email.NewResendSender(cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.ResendFromName)
		slog.InfoContext(ctx, "Resendクライアントを初期化しました")
	} else {
		slog.WarnContext(ctx, "Resend APIキーが設定されていません。メール送信機能は利用できません")
	}

	// Riverワーカーの登録
	workers := river.NewWorkers()

	// メール送信ワーカーを登録
	if emailSender != nil {
		signInCodeSender := email.NewSignInCodeSender(emailSender)
		sendSignInCodeEmailUC := usecase.NewSendSignInCodeEmailUsecase(signInCodeSender)
		river.AddWorker(workers, NewSendSignInCodeEmailWorker(sendSignInCodeEmailUC))
		slog.InfoContext(ctx, "SendSignInCodeEmailWorkerを登録しました")

		signUpCodeSender := email.NewSignUpCodeSender(emailSender)
		sendSignUpCodeEmailUC := usecase.NewSendSignUpCodeEmailUsecase(signUpCodeSender)
		river.AddWorker(workers, NewSendSignUpCodeEmailWorker(sendSignUpCodeEmailUC))
		slog.InfoContext(ctx, "SendSignUpCodeEmailWorkerを登録しました")

		passwordResetSender := email.NewPasswordResetSender(emailSender)
		sendPasswordResetEmailUC := usecase.NewSendPasswordResetEmailUsecase(passwordResetSender)
		river.AddWorker(workers, NewSendPasswordResetEmailWorker(sendPasswordResetEmailUC))
		slog.InfoContext(ctx, "SendPasswordResetEmailWorkerを登録しました")
	}

	// トークンクリーンアップワーカーを登録
	river.AddWorker(workers, NewCleanupExpiredTokensWorker(params.CleanupExpiredTokens))
	slog.InfoContext(ctx, "CleanupExpiredTokensWorkerを登録しました")

	// ログインコードクリーンアップワーカーを登録
	river.AddWorker(workers, NewCleanupExpiredSignInCodesWorker(params.CleanupExpiredSignInCodes))
	slog.InfoContext(ctx, "CleanupExpiredSignInCodesWorkerを登録しました")

	// セッションクリーンアップワーカーを登録する。
	river.AddWorker(workers, NewCleanupExpiredSessionsWorker(params.CleanupExpiredSessions))
	slog.InfoContext(ctx, "CleanupExpiredSessionsWorkerを登録しました")

	// animesリコンサイルバッチワーカーを登録する。
	river.AddWorker(workers, NewSyncAnimesWorker(params.SyncAnimes))
	slog.InfoContext(ctx, "SyncAnimesWorkerを登録しました")

	// Riverにはslog.Default() ではなく素の標準エラー出力ハンドラーで
	// ログを取らせる。slog.Default() のハンドラーはErrorレベルをSentryへ
	// ファンアウトする。Riverのバックグラウンドサービスは一時的な接続失敗を
	// Errorレベルで出力し自力で回復する。producerは次回ポーリングでジョブ取得を
	// 再試行し、notifier (LISTEN/NOTIFY) やelectorはバックオフ後に再接続する
	// (例: DB接続の瞬断時)。これらをdefaultロガー経由にすると自己回復する
	// ノイズがSentryイベント化されてしまう。本当のジョブ失敗は
	// RiverWorkerMiddlewareが意図的に捕捉するので、標準エラー出力のみのロガーに
	// しても対処可能なシグナルは失われない。デフォルトロガーの基底でもある
	// sentry.NewBaseHandlerから生成するため、形式と詳細度は一致し、Sentryへの
	// ファンアウトだけを外している。
	riverLogger := slog.New(annictSentry.NewBaseHandler())

	// Riverクライアントの作成
	// SentryミドルウェアはConfig.Middlewareに登録する。
	// 将来riverのアップデートで削除される可能性のあるWorkerMiddleware
	// フィールドは使わないことで、削除時の再対応を不要にする。
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
		Middleware: []rivertype.Middleware{
			annictSentry.RiverWorkerMiddleware(),
		},
		Logger: riverLogger,
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

// StartはRiverクライアントを起動します
func (c *Client) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "Riverクライアントを起動します")
	return c.riverClient.Start(ctx)
}

// StopはRiverクライアントを停止します
func (c *Client) Stop(ctx context.Context) error {
	slog.InfoContext(ctx, "Riverクライアントを停止します")
	if err := c.riverClient.Stop(ctx); err != nil {
		return err
	}
	c.pool.Close()
	return nil
}

// ClientはRiverクライアントへのアクセスを提供します
func (c *Client) Client() *river.Client[pgx.Tx] {
	return c.riverClient
}
