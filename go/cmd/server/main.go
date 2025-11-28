package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/handler/health"
	"github.com/annict/annict/internal/handler/home"
	"github.com/annict/annict/internal/handler/manifest"
	"github.com/annict/annict/internal/handler/password"
	"github.com/annict/annict/internal/handler/password_reset"
	"github.com/annict/annict/internal/handler/popular_work"
	"github.com/annict/annict/internal/handler/sign_in"
	"github.com/annict/annict/internal/handler/sign_in_code"
	"github.com/annict/annict/internal/handler/sign_in_password"
	"github.com/annict/annict/internal/handler/sign_up"
	"github.com/annict/annict/internal/handler/sign_up_code"
	"github.com/annict/annict/internal/handler/sign_up_username"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/image"
	authMiddleware "github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	annictSentry "github.com/annict/annict/internal/sentry"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
	"github.com/annict/annict/internal/worker"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
)

// dailyAt2AMSchedule は毎日深夜2時に実行するスケジュール
type dailyAt2AMSchedule struct{}

// Next は次の実行時刻を返します
func (s dailyAt2AMSchedule) Next(current time.Time) time.Time {
	// 毎日深夜2時に実行
	next := time.Date(current.Year(), current.Month(), current.Day(), 2, 0, 0, 0, current.Location())

	// 今日の2時を過ぎている場合は明日の2時に設定
	if current.Hour() >= 2 || (current.Hour() == 1 && current.Minute() >= 59) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

func main() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("サーバーを起動します", "env", cfg.Env)

	// Sentryの初期化
	err = annictSentry.Init(annictSentry.Config{
		DSN:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		Debug:            cfg.SentryDebug,
	})
	if err != nil {
		slog.Error("Sentryの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer annictSentry.Flush(2 * time.Second)

	// データベース接続
	db, err := sql.Open("postgres", cfg.DatabaseDSN())
	if err != nil {
		slog.Error("データベースへの接続に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("データベース接続のクローズに失敗しました", "error", err)
		}
	}()

	// コネクションプール設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// データベース接続確認
	if err := db.Ping(); err != nil {
		slog.Error("データベースへの疎通確認に失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("データベースに正常に接続しました")

	// 簡単なクエリで接続テスト
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		slog.Error("データベースクエリに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("データベースに接続しました", "database", dbName)

	// sqlcのクエリーインスタンスを作成
	queries := query.New(db)

	// Redisクライアントの初期化（Rate Limiting用とデータストア用）
	var limiter *ratelimit.Limiter
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			slog.Error("Redis URL のパースに失敗しました", "error", err)
			os.Exit(1)
		}

		// コネクションプール設定
		opt.PoolSize = 10
		opt.MinIdleConns = 2
		opt.ConnMaxIdleTime = 5 * time.Minute

		redisClient = redis.NewClient(opt)
		// Redis接続確認
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			slog.Warn("Redis への接続に失敗しました（Rate Limiting は無効化されます）", "error", err)
			redisClient = nil
		} else {
			slog.Info("Redis に正常に接続しました")
			limiter = ratelimit.NewLimiter(redisClient)
		}
	} else {
		slog.Warn("Redis URL が設定されていません（Rate Limiting は無効化されます）")
	}

	// River クライアントの初期化
	ctx := context.Background()
	riverClient, err := worker.NewClient(ctx, cfg.DatabaseDSN(), queries, cfg)
	if err != nil {
		slog.Error("River クライアントの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := riverClient.Stop(shutdownCtx); err != nil {
			slog.Warn("River クライアントの停止に失敗しました", "error", err)
		}
	}()

	// River クライアントを起動
	if err := riverClient.Start(ctx); err != nil {
		slog.Error("River クライアントの起動に失敗しました", "error", err)
		os.Exit(1)
	}

	// 定期実行ジョブの登録: トークンクリーンアップ（毎日深夜2時）
	periodicJobTokenCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredTokensArgs{}, nil
		},
		nil, // オプションなし
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobTokenCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "トークンクリーンアップ", "schedule", "毎日深夜2時")

	// 定期実行ジョブの登録: ログインコードクリーンアップ（毎日深夜2時）
	periodicJobSignInCodeCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredSignInCodesArgs{}, nil
		},
		nil, // オプションなし
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSignInCodeCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "ログインコードクリーンアップ", "schedule", "毎日深夜2時")

	// セッションリポジトリの初期化
	sessionRepo := repository.NewSessionRepository(queries)

	// セッションマネージャーの初期化
	sessionManager := session.NewManager(sessionRepo, cfg)

	// 認証ミドルウェアの初期化
	authMW := authMiddleware.NewAuthMiddleware(sessionManager, sessionRepo)

	// リバースプロキシミドルウェアの初期化
	var reverseProxyMW *authMiddleware.ReverseProxyMiddleware
	if cfg.RailsAppURL != "" {
		var err error
		reverseProxyMW, err = authMiddleware.NewReverseProxyMiddleware(cfg.RailsAppURL, cfg)
		if err != nil {
			slog.Error("リバースプロキシミドルウェアの初期化に失敗しました", "error", err)
			os.Exit(1)
		}
		slog.Info("リバースプロキシミドルウェアを有効化しました", "rails_app_url", cfg.RailsAppURL)
	}

	// Chiルーターの設定
	r := chi.NewRouter()

	// Sentryミドルウェアの作成
	// パニック発生時にSentryにエラーを送信し、リクエスト情報をキャプチャする
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true, // パニックを再発生させて middleware.Logger でログ出力可能にする
	})

	// ミドルウェア
	r.Use(middleware.Logger)
	r.Use(func(next http.Handler) http.Handler {
		return sentryHandler.Handle(next)
	})
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// リクエストボディサイズ制限（10MB）
	requestBodyLimitMW := authMiddleware.NewRequestBodyLimitMiddleware(10 * 1024 * 1024)
	r.Use(requestBodyLimitMW.Middleware)

	// メンテナンスミドルウェア（RealIPの後に配置し、クライアントIPを正しく取得）
	maintenanceMW := authMiddleware.NewMaintenanceMiddleware(cfg)
	r.Use(maintenanceMW.Middleware)

	// リバースプロキシミドルウェア
	// Go版で処理するかRails版にプロキシするかを判定
	// Rails版にプロキシする場合は後続のミドルウェアをスキップ
	if reverseProxyMW != nil {
		r.Use(reverseProxyMW.Middleware)
	}

	// 以下はGo版で処理する場合のみ適用されるミドルウェア
	r.Use(authMiddleware.MethodOverride) // Method Overrideミドルウェアを追加（HTMLフォームからPUT/PATCH/DELETEを使用可能に）
	r.Use(authMW.Middleware)             // 認証ミドルウェアを追加（ユーザー情報をコンテキストに設定）

	// Sentryユーザーコンテキストミドルウェアを追加（認証済みユーザーのIDをSentryに設定）
	sentryUserContextMW := authMiddleware.NewSentryUserContextMiddleware()
	r.Use(sentryUserContextMW.Middleware)

	r.Use(i18n.Middleware) // I18nミドルウェアを追加（ユーザーのlocaleを考慮）

	// CSRF保護ミドルウェアを追加
	csrfMiddleware := authMiddleware.NewCSRFMiddleware(sessionManager)
	r.Use(csrfMiddleware.Middleware)

	// リポジトリの初期化
	workRepo := repository.NewWorkRepository(queries)

	// ヘルスチェックハンドラーの初期化
	healthHandler := health.NewHandler(cfg, workRepo)

	// ホームページハンドラーの初期化
	homeHandler := home.NewHandler(cfg)

	// 人気作品ハンドラーの初期化
	imageHelper := image.NewHelper(cfg)
	popularWorkHandler := popular_work.NewHandler(cfg, workRepo, imageHelper, sessionManager)

	// ユーザーリポジトリの初期化
	userRepo := repository.NewUserRepository(queries)

	// 6桁コード送信ユースケースの初期化
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, riverClient)

	// 新規登録確認コード送信ユースケースの初期化
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, queries, riverClient)

	// Turnstileクライアントの初期化
	turnstileClient := turnstile.NewClient(cfg.TurnstileSiteKey, cfg.TurnstileSecretKey)

	// サインインハンドラーの初期化
	signInHandler := sign_in.NewHandler(cfg, sessionManager, userRepo, sendSignInCodeUC, turnstileClient)

	// 新規登録ハンドラーの初期化
	signUpHandler := sign_up.NewHandler(cfg, sessionManager, userRepo, limiter, sendSignUpCodeUC, turnstileClient)

	// 新規登録確認コード検証ユースケースの初期化
	verifySignUpCodeUC := usecase.NewVerifySignUpCodeUsecase(db, queries)

	// 新規登録確認コード入力ハンドラーの初期化
	signUpCodeHandler := sign_up_code.NewHandler(cfg, sessionManager, db, limiter, redisClient, sendSignUpCodeUC, verifySignUpCodeUC)

	// ユーザー名設定とユーザー登録ハンドラーの初期化
	completeSignUpUC := usecase.NewCompleteSignUpUsecase(db, queries, redisClient)
	signUpUsernameHandler := sign_up_username.NewHandler(cfg, sessionManager, redisClient, completeSignUpUC)

	// 6桁コード入力ハンドラーの初期化
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)
	signInCodeHandler := sign_in_code.NewHandler(cfg, sessionManager, userRepo, db, limiter, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// パスワードログインハンドラーの初期化
	signInPasswordHandler := sign_in_password.NewHandler(cfg, userRepo, sessionManager, createSessionUC)

	// パスワードリセット申請ハンドラーの初期化
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, queries, riverClient)
	passwordResetHandler := password_reset.NewHandler(cfg, userRepo, sessionManager, limiter, turnstileClient, createPasswordResetTokenUC)

	// パスワード編集・更新ハンドラーの初期化
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	updatePasswordUC := usecase.NewUpdatePasswordResetUsecase(db, queries)
	passwordHandler := password.NewHandler(cfg, db, passwordResetTokenRepo, sessionManager, limiter, updatePasswordUC)

	// Web App Manifestハンドラーの初期化
	manifestHandler := manifest.NewHandler(cfg)

	// 静的ファイルの配信 (Tailwind CLI + esbuild のビルド結果)
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// ルート設定
	r.Get("/", homeHandler.Show)
	r.Get("/health", healthHandler.Show)
	r.Get("/manifest.json", manifestHandler.Show)
	r.Get("/works/popular", popularWorkHandler.Index)
	r.Get("/sign_in", signInHandler.New)
	r.Post("/sign_in", signInHandler.Create)
	r.Get("/sign_in/code", signInCodeHandler.Show)
	r.Post("/sign_in/code", signInCodeHandler.Create)
	r.Patch("/sign_in/code", signInCodeHandler.Update)
	r.Get("/sign_in/password", signInPasswordHandler.New)
	r.Post("/sign_in/password", signInPasswordHandler.Create)
	r.Get("/sign_up", signUpHandler.New)
	r.Post("/sign_up", signUpHandler.Create)
	r.Get("/sign_up/code", signUpCodeHandler.New)
	r.Post("/sign_up/code", signUpCodeHandler.Create)
	r.Patch("/sign_up/code", signUpCodeHandler.Update)
	r.Get("/sign_up/username", signUpUsernameHandler.New)
	r.Post("/sign_up/username", signUpUsernameHandler.Create)

	// パスワードリセット申請
	r.Get("/password/reset", passwordResetHandler.New)
	r.Post("/password/reset", passwordResetHandler.Create)

	// パスワード編集・更新
	r.Get("/password/edit", passwordHandler.Edit)
	r.Patch("/password", passwordHandler.Update) // HTMLフォームからは_methodパラメータでPATCHを送信

	// サーバー起動
	// Dockerコンテナ内で動かす場合、0.0.0.0でリッスンする必要がある
	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	slog.Info("HTTPサーバーを起動します", "addr", addr)
	slog.Info("アクセス先", "local", fmt.Sprintf("http://localhost:%s", cfg.Port), "domain", fmt.Sprintf("https://%s", cfg.Domain))

	// HTTPサーバーの作成
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Graceful shutdown のためのシグナルハンドリング
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		slog.Info("シャットダウンシグナルを受信しました。サーバーを停止します...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("サーバーのシャットダウンに失敗しました", "error", err)
		}
	}()

	// サーバー起動
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("サーバーの起動に失敗しました", "error", err)
		os.Exit(1)
	}

	slog.Info("サーバーが正常に停止しました")
}
