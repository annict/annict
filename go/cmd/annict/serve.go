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

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/handler/db_work"
	"github.com/annict/annict/go/internal/handler/health"
	"github.com/annict/annict/go/internal/handler/home"
	"github.com/annict/annict/go/internal/handler/ics"
	"github.com/annict/annict/go/internal/handler/manifest"
	"github.com/annict/annict/go/internal/handler/password"
	"github.com/annict/annict/go/internal/handler/password_reset"
	"github.com/annict/annict/go/internal/handler/popular_work"
	"github.com/annict/annict/go/internal/handler/sign_in"
	"github.com/annict/annict/go/internal/handler/sign_in_code"
	"github.com/annict/annict/go/internal/handler/sign_in_password"
	"github.com/annict/annict/go/internal/handler/sign_out"
	"github.com/annict/annict/go/internal/handler/sign_up"
	"github.com/annict/annict/go/internal/handler/sign_up_code"
	"github.com/annict/annict/go/internal/handler/sign_up_username"
	"github.com/annict/annict/go/internal/handler/supporters"
	"github.com/annict/annict/go/internal/handler/supporters_checkout"
	"github.com/annict/annict/go/internal/handler/supporters_portal"
	"github.com/annict/annict/go/internal/handler/tracking_heatmap"
	stripewebhook "github.com/annict/annict/go/internal/handler/webhooks/stripe"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
	"github.com/annict/annict/go/internal/worker"
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

// hourlySchedule runs a job once an hour, on the hour.
//
// [Ja] hourlySchedule は 1 時間ごと、毎時 0 分にジョブを実行する。
type hourlySchedule struct{}

// Next returns the next top-of-the-hour after current.
//
// [Ja] Next は current の次の毎時 0 分を返す。
func (s hourlySchedule) Next(current time.Time) time.Time {
	return time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), 0, 0, 0, current.Location()).
		Add(time.Hour)
}

// runServe starts the HTTP server: it loads config, wires dependencies, registers
// routes and periodic jobs, and blocks on ListenAndServe until shutdown. It is the
// body of the `serve` subcommand.
//
// [Ja] runServe は HTTP サーバーを起動する。設定の読み込み・依存の組み立て・ルートと
// 定期ジョブの登録を行い、シャットダウンまで ListenAndServe でブロックする。
// `serve` サブコマンドの本体。
func runServe() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("サーバーを起動します", "env", cfg.Env)

	// Initialize Sentry. Release reuses AssetVersion (short Git commit hash) so
	// that events can be distinguished per deploy.
	//
	// [Ja] Sentry の初期化。Release には AssetVersion (Git コミットハッシュ短縮版)
	// を流用し、デプロイごとにイベントを区別できるようにする。
	err = annictSentry.Init(annictSentry.Config{
		DSN:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		Release:          cfg.AssetVersion,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		Debug:            cfg.SentryDebug,
	})
	if err != nil {
		slog.Error("Sentryの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer annictSentry.Flush(2 * time.Second)

	// Route slog through Sentry: Error-level records become Sentry events while
	// every level still reaches stderr through the underlying text handler.
	// SetDefault must happen before any code path that can call slog.Error
	// (DB connection, river client, etc.) so the Sentry handler covers
	// startup failures too.
	//
	// [Ja] slog のデフォルトロガーを Sentry 連携付きハンドラーに差し替える。
	// Error レベル以上は Sentry イベント化され、全レベルは引き続き標準エラー
	// 出力にも届く。slog.Error を呼ぶ可能性のある処理 (DB 接続、river 起動など)
	// より前に呼ぶことで、起動時のエラーも Sentry に届くようにする。
	slog.SetDefault(slog.New(annictSentry.NewSlogHandler(annictSentry.NewBaseHandler())))

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

	// クリーンアップ UseCase の作成（Worker 用）
	cleanupTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	cleanupCodeRepo := repository.NewSignInCodeRepository(queries)
	cleanupExpiredTokensUC := usecase.NewCleanupExpiredTokensUsecase(cleanupTokenRepo)
	cleanupExpiredSignInCodesUC := usecase.NewCleanupExpiredSignInCodesUsecase(cleanupCodeRepo)

	// Wire the phase 2 full-reconciliation batch usecase (for the Worker). It is
	// invoked by the periodic job below to sync works / episodes into animes /
	// anime_classifications. The wiring is shared with the `sync-animes` subcommand
	// via newSyncAnimesUsecase.
	//
	// [Ja] フェーズ 2 のフル・リコンシリエーションバッチ UseCase を組み立てる (Worker 用)。
	// works / episodes を animes / anime_classifications へ同期する下の定期ジョブから
	// 呼ばれる。配線は newSyncAnimesUsecase 経由で `sync-animes` サブコマンドと共有する。
	syncAnimesUC := newSyncAnimesUsecase(db, queries)

	// River クライアントの初期化
	ctx := context.Background()
	riverClient, err := worker.NewClient(ctx, cfg.DatabaseDSN(), worker.NewClientParams{
		CleanupExpiredTokens:      cleanupExpiredTokensUC,
		CleanupExpiredSignInCodes: cleanupExpiredSignInCodesUC,
		SyncAnimes:                syncAnimesUC,
	}, cfg)
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
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobTokenCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "トークンクリーンアップ", "schedule", "毎日深夜2時")

	// 定期実行ジョブの登録: ログインコードクリーンアップ（毎日深夜2時）
	periodicJobSignInCodeCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredSignInCodesArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSignInCodeCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "ログインコードクリーンアップ", "schedule", "毎日深夜2時")

	// Register the hourly animes reconciliation. During phase 2 there is no
	// dual-write yet, so this batch is the only path that updates animes; running
	// it hourly favors freshness and the cadence of the diff metric. The
	// reconciliation is idempotent and converges to near-zero diffs in steady state.
	//
	// [Ja] animes リコンサイルを毎時登録する。フェーズ 2 ではまだ両書きが無く、本バッチが
	// animes を更新する唯一の経路のため、毎時実行で鮮度と差分メトリクスの取得頻度を優先する。
	// リコンサイルは冪等で、定常状態ではほぼ差分なしに収束する。
	periodicJobSyncAnimes := river.NewPeriodicJob(
		hourlySchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.SyncAnimesArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSyncAnimes)
	slog.Info("定期実行ジョブを登録しました", "job", "animes 同期バッチ", "schedule", "毎時")

	// セッションリポジトリの初期化
	sessionRepo := repository.NewSessionRepository(queries)

	// セッションマネージャーの初期化
	sessionManager := session.NewManager(sessionRepo, cfg)

	// フラッシュマネージャーの初期化
	flashMgr := session.NewFlashManager(cfg.CookieDomain, cfg.SessionSecure == "true")

	// 認証ミドルウェアの初期化
	authMW := authMiddleware.NewAuthMiddleware(sessionManager)

	// フィーチャーフラグリポジトリの初期化
	featureFlagRepo := repository.NewFeatureFlagRepository(queries)

	// リバースプロキシミドルウェアの初期化
	var reverseProxyMW *authMiddleware.ReverseProxyMiddleware
	if cfg.RailsAppURL != "" {
		var err error
		reverseProxyMW, err = authMiddleware.NewReverseProxyMiddleware(cfg.RailsAppURL, cfg, featureFlagRepo, sessionManager)
		if err != nil {
			slog.Error("リバースプロキシミドルウェアの初期化に失敗しました", "error", err)
			os.Exit(1)
		}
		slog.Info("リバースプロキシミドルウェアを有効化しました", "rails_app_url", cfg.RailsAppURL)
	}

	// Chiルーターの設定
	r := chi.NewRouter()

	// ミドルウェア
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

	// リクエストボディサイズ制限（10MB）
	requestBodyLimitMW := authMiddleware.NewRequestBodyLimitMiddleware(10 * 1024 * 1024)
	r.Use(requestBodyLimitMW.Middleware)

	// Maintenance middleware. It resolves the client IP via clientip.GetClientIP,
	// which reads the proxy headers itself, so the order relative to other
	// middleware does not affect IP resolution.
	//
	// [Ja] メンテナンスミドルウェア。クライアント IP は clientip.GetClientIP で
	// 取得し、同関数がプロキシヘッダを自前で読むため、他のミドルウェアとの
	// 登録順は IP 取得に影響しない。
	maintenanceMW := authMiddleware.NewMaintenanceMiddleware(cfg)
	r.Use(maintenanceMW.Middleware)

	// リバースプロキシミドルウェア
	// Go版で処理するかRails版にプロキシするかを判定
	// Rails版にプロキシする場合は後続のミドルウェアをスキップ
	if reverseProxyMW != nil {
		r.Use(reverseProxyMW.Middleware)
	}

	// The middleware below applies only to requests handled by the Go app.
	// Registering the Sentry chain inside the reverse proxy keeps requests
	// proxied to the Rails app out of the Go app's Sentry transactions.
	//
	// [Ja] 以下は Go 版で処理する場合のみ適用されるミドルウェア。Sentry の
	// チェーンをリバースプロキシの内側に置くことで、Rails 版へプロキシされる
	// リクエストが Go 版の Sentry トランザクションに乗らないようにする。

	// Recoverer must be registered before sentryhttp (= outer in the chain) so
	// that the re-panic from sentryhttp (Repanic: true) is caught here and a
	// 500 response is returned. The official sentryhttp README also says to
	// place the recovery middleware on the outside of sentryhttp.
	//
	// [Ja] Recoverer は sentryhttp より前 (= 外側) に登録する。sentryhttp が
	// Repanic: true で再 panic したものをここで握り潰して 500 を返すため。
	// Sentry SDK 公式 README も「recovery middleware は sentryhttp より外側に
	// 置く」と指示している。
	r.Use(middleware.Recoverer)

	// sentryhttp captures panics, sets a per-request Hub on the context, and
	// starts a Sentry transaction per request. Repanic: true re-throws after
	// capture so that Recoverer (outer) returns 500 to the client.
	//
	// [Ja] sentryhttp はリクエスト単位の Hub を context に積み、panic を捕捉して
	// Sentry に送信し、リクエストごとの Sentry トランザクションを開始する。
	// Repanic: true により捕捉後に再 panic することで、外側の Recoverer が
	// クライアントに 500 を返せる。
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r.Use(sentryHandler.Handle)

	// SentryTransaction rewrites the transaction name with chi's route pattern.
	// It must be registered after sentryhttp (= inside the wrapper) so that the
	// LIFO defer order guarantees the rewrite happens before sentryhttp's
	// transaction.Finish() / recoverWithSentry run.
	//
	// [Ja] SentryTransaction はトランザクション名を chi のルートパターンに
	// 上書きするミドルウェア。sentryhttp の後 (= 内側) に登録することで、LIFO の
	// defer 順序により sentryhttp の transaction.Finish() や recoverWithSentry
	// より先に書き換えが走ることを保証する。
	r.Use(authMiddleware.SentryTransaction)

	r.Use(authMiddleware.MethodOverride) // Method Overrideミドルウェアを追加（HTMLフォームからPUT/PATCH/DELETEを使用可能に）
	r.Use(authMW.Middleware)             // 認証ミドルウェアを追加（ユーザー情報をコンテキストに設定）

	// Sentryユーザーコンテキストミドルウェアを追加（認証済みユーザーのIDをSentryに設定）
	sentryUserContextMW := authMiddleware.NewSentryUserContextMiddleware()
	r.Use(sentryUserContextMW.Middleware)

	r.Use(i18n.Middleware) // I18nミドルウェアを追加（ユーザーのlocaleを考慮）

	// 現在パスミドルウェアを追加（サイドバーの現在ページハイライト用に aria-current を付与）
	r.Use(templates.CurrentPathMiddleware)

	// CSRF保護ミドルウェアを追加
	csrfMiddleware := authMiddleware.NewCSRFMiddleware(sessionManager)
	r.Use(csrfMiddleware.Middleware)

	// フラッシュミドルウェアは CSRF 検証後・全ハンドラー前に配置する。
	// CSRF が失敗するリクエストでは flash Cookie を読み取らない（=クリアしない）ため、
	// 次回リクエストでも flash が失われない。
	r.Use(flashMgr.Middleware)

	// リポジトリの初期化
	workRepo := repository.NewWorkRepository(queries)
	castRepo := repository.NewCastRepository(queries)
	staffRepo := repository.NewStaffRepository(queries)
	userCalendarRepo := repository.NewUserCalendarRepository(queries)

	// ヘルスチェックハンドラーの初期化
	checkHealthUC := usecase.NewCheckHealthUsecase(workRepo)
	healthHandler := health.NewHandler(cfg, checkHealthUC)

	// ホームページハンドラーの初期化
	homeHandler := home.NewHandler(cfg)

	// 人気作品ハンドラーの初期化
	imageHelper := image.NewHelper(cfg)
	getPopularWorksUC := usecase.NewGetPopularWorksUsecase(workRepo, castRepo, staffRepo)
	popularWorkHandler := popular_work.NewHandler(cfg, getPopularWorksUC, imageHelper)

	// iCalendar配信ハンドラーの初期化
	getUserCalendarUC := usecase.NewGetUserCalendarUsecase(userCalendarRepo)
	icsHandler := ics.NewHandler(cfg, getUserCalendarUC)

	// ユーザーリポジトリの初期化
	userRepo := repository.NewUserRepository(queries)

	// Initialize the tracking heatmap fragment handler.
	//
	// [Ja] 視聴記録ヒートマップフラグメントハンドラーの初期化。
	recordRepo := repository.NewRecordRepository(queries)
	getTrackingHeatmapUC := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)
	trackingHeatmapHandler := tracking_heatmap.NewHandler(getTrackingHeatmapUC)

	// サインインコードリポジトリの初期化
	signInCodeRepo := repository.NewSignInCodeRepository(queries)

	// Dispatcher の初期化
	d := dispatcher.NewDispatcher(riverClient.Client())

	// 6桁コード送信ユースケースの初期化
	signInValidator := validator.NewSignInCreateValidator()
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, signInCodeRepo, userRepo, d, signInValidator)

	// サインアップコードリポジトリの初期化
	signUpCodeRepo := repository.NewSignUpCodeRepository(queries)

	// 新規登録確認コード送信ユースケースの初期化
	signUpValidator := validator.NewSignUpCreateValidator()
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, signUpCodeRepo, userRepo, d, signUpValidator)

	// Turnstileクライアントの初期化
	turnstileClient := turnstile.NewClient(cfg.TurnstileSiteKey, cfg.TurnstileSecretKey)

	// サインインハンドラーの初期化
	signInHandler := sign_in.NewHandler(cfg, sessionManager, flashMgr, sendSignInCodeUC, turnstileClient)

	// 新規登録ハンドラーの初期化
	signUpHandler := sign_up.NewHandler(cfg, sessionManager, flashMgr, limiter, sendSignUpCodeUC, turnstileClient)

	// 新規登録確認コード検証ユースケースの初期化
	signUpCodeValidator := validator.NewSignUpCodeCreateValidator()
	verifySignUpCodeUC := usecase.NewVerifySignUpCodeUsecase(db, signUpCodeRepo, signUpCodeValidator)

	// 新規登録確認コード入力ハンドラーの初期化
	signUpCodeHandler := sign_up_code.NewHandler(cfg, sessionManager, flashMgr, db, limiter, redisClient, sendSignUpCodeUC, verifySignUpCodeUC)

	// ユーザー名設定とユーザー登録ハンドラーの初期化
	profileRepo := repository.NewProfileRepository(queries)
	settingRepo := repository.NewSettingRepository(queries)
	emailNotificationRepo := repository.NewEmailNotificationRepository(queries)
	signUpUsernameValidator := validator.NewSignUpUsernameCreateValidator()
	completeSignUpUC := usecase.NewCompleteSignUpUsecase(db, userRepo, profileRepo, settingRepo, emailNotificationRepo, sessionRepo, redisClient, signUpUsernameValidator)
	signUpUsernameHandler := sign_up_username.NewHandler(cfg, sessionManager, flashMgr, redisClient, completeSignUpUC)

	// 6桁コード入力ハンドラーの初期化
	signInCodeValidator := validator.NewSignInCodeCreateValidator()
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, signInCodeValidator)
	createSessionUC := usecase.NewCreateSessionUsecase(sessionRepo)
	signInCodeHandler := sign_in_code.NewHandler(cfg, sessionManager, flashMgr, limiter, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// パスワードログインハンドラーの初期化
	signInPasswordValidator := validator.NewSignInPasswordCreateValidator(userRepo)
	authenticateByPasswordUC := usecase.NewAuthenticateByPasswordUsecase(createSessionUC, signInPasswordValidator)
	signInPasswordHandler := sign_in_password.NewHandler(cfg, sessionManager, flashMgr, authenticateByPasswordUC)

	// ログアウトハンドラーの初期化
	signOutHandler := sign_out.NewHandler(sessionManager)

	// パスワードリセット申請ハンドラーの初期化
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	passwordResetValidator := validator.NewPasswordResetCreateValidator()
	createPasswordResetTokenUC := usecase.NewCreatePasswordResetTokenUsecase(db, userRepo, passwordResetTokenRepo, cfg, d, passwordResetValidator)
	passwordResetHandler := password_reset.NewHandler(cfg, sessionManager, limiter, turnstileClient, createPasswordResetTokenUC)

	// パスワード編集・更新ハンドラーの初期化
	updatePasswordValidator := validator.NewPasswordUpdateValidator()
	getPasswordResetTokenUC := usecase.NewGetPasswordResetTokenUsecase(passwordResetTokenRepo)
	updatePasswordUC := usecase.NewUpdatePasswordResetUsecase(db, passwordResetTokenRepo, userRepo, sessionRepo, updatePasswordValidator)
	passwordHandler := password.NewHandler(cfg, sessionManager, flashMgr, limiter, getPasswordResetTokenUC, updatePasswordUC)

	// Web App Manifestハンドラーの初期化
	manifestHandler := manifest.NewHandler(cfg)

	// サポーターページハンドラーの初期化
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)
	annictStripeCfg := &annictStripe.Config{
		SecretKey:      cfg.StripeSecretKey,
		WebhookSecret:  cfg.StripeWebhookSecret,
		PriceMonthlyID: cfg.StripePriceMonthlyID,
		PriceYearlyID:  cfg.StripePriceYearlyID,
	}
	stripeClient := annictStripe.NewClient(cfg.StripeSecretKey)
	// Wrap the stripe-go client in an adapter that implements the per-UseCase
	// interfaces, so the UseCases depend on the seam rather than *stripe.Client.
	//
	// [Ja] stripe-go クライアントを、UseCase ごとの interface を実装する adapter で
	// ラップする。これにより各 UseCase は *stripe.Client ではなく seam に依存する。
	stripeAdapter := annictStripe.NewAdapter(stripeClient)
	getSupporterStatusUC := usecase.NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)
	supportersHandler := supporters.NewHandler(cfg, sessionManager, imageHelper, getSupporterStatusUC, annictStripeCfg, stripeClient)

	// Stripe Checkoutハンドラーの初期化
	createSupportersCheckoutValidator := validator.NewSupportersCheckoutCreateValidator()
	createCheckoutSessionUC := usecase.NewCreateCheckoutSessionUsecase(cfg, stripeSubscriberRepo, annictStripeCfg, stripeAdapter, createSupportersCheckoutValidator)
	supportersCheckoutHandler := supporters_checkout.NewHandler(flashMgr, createCheckoutSessionUC)

	// Stripe Customer Portalハンドラーの初期化
	createPortalSessionUC := usecase.NewCreatePortalSessionUsecase(cfg, stripeSubscriberRepo, stripeAdapter)
	supportersPortalHandler := supporters_portal.NewHandler(flashMgr, createPortalSessionUC)

	// Stripe Webhookハンドラーの初期化
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)
	createStripeSubscriberUC := usecase.NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, stripeAdapter)
	updateStripeSubscriberUC := usecase.NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
	deleteStripeSubscriberUC := usecase.NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
	processStripeWebhookUC := usecase.NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createStripeSubscriberUC, updateStripeSubscriberUC, deleteStripeSubscriberUC)
	stripeWebhookHandler := stripewebhook.NewHandler(cfg, processStripeWebhookUC)

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
	r.Delete("/sign_out", signOutHandler.Delete) // Rails UJSからのDELETEリクエスト
	r.Post("/sign_out", signOutHandler.Delete)   // Go版HTMLフォームからのPOST + _method=DELETE
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

	// サポーターページ
	r.Get("/supporters", supportersHandler.Show)
	r.Post("/supporters/checkout", supportersCheckoutHandler.Create)
	r.Post("/supporters/portal", supportersPortalHandler.Create)

	// Stripe Webhook
	r.Post("/webhooks/stripe", stripeWebhookHandler.Create)

	// DB管理画面
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	listDbWorksUC := usecase.NewListDbWorksUsecase(workRepo)
	getDbWorkFormOptionsUC := usecase.NewGetDbWorkFormOptionsUsecase(numberFormatRepo)
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, validator.NewDbWorkCreateValidator())
	dbWorkHandler := db_work.NewHandler(cfg, sessionManager, flashMgr, listDbWorksUC, getDbWorkFormOptionsUC, createWorkUC)
	r.Get("/db/works", dbWorkHandler.Index)
	r.Get("/db/works/new", dbWorkHandler.New)
	r.Post("/db/works", dbWorkHandler.Create)

	// iCalendar配信
	r.Get("/@{username}/ics", icsHandler.Show) // メインのエンドポイント
	r.Get("/ics", icsHandler.Show)             // Apple カレンダー互換の代替パス（クエリパラメータで username を指定）

	// Tracking heatmap fragment.
	//
	// [Ja] 視聴記録ヒートマップフラグメント。
	r.Get("/fragment/@{username}/tracking_heatmap", trackingHeatmapHandler.Show)

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
