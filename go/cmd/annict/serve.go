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
	"github.com/annict/annict/go/internal/handler/db_episode"
	"github.com/annict/annict/go/internal/handler/db_episode_archive"
	"github.com/annict/annict/go/internal/handler/db_work"
	"github.com/annict/annict/go/internal/handler/db_work_archive"
	"github.com/annict/annict/go/internal/handler/db_work_deletion"
	"github.com/annict/annict/go/internal/handler/db_work_unarchive"
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
	"github.com/annict/annict/go/internal/httperror"
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

// dailyAt2AMScheduleは毎日深夜2時にジョブを実行するスケジュール。
type dailyAt2AMSchedule struct{}

// Nextは次の実行対象となる深夜2時を返し、1時59分以降は翌日に繰り越す。
func (s dailyAt2AMSchedule) Next(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), 2, 0, 0, 0, current.Location())

	// 現在時刻が今日の2時の1分前以降になっていれば、翌日の2時にする。
	if current.Hour() >= 2 || (current.Hour() == 1 && current.Minute() >= 59) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// hourlyScheduleは1時間ごと、毎時0分にジョブを実行する。
type hourlySchedule struct{}

// Nextはcurrentの次の毎時0分を返す。
func (s hourlySchedule) Next(current time.Time) time.Time {
	return time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), 0, 0, 0, current.Location()).
		Add(time.Hour)
}

// runServeはHTTPサーバーを起動する。設定の読み込み・依存の組み立て・ルートと
// 定期ジョブの登録を行い、シャットダウンまでListenAndServeでブロックする。
// `serve` サブコマンドの本体。
func runServe() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("サーバーを起動します", "env", cfg.Env)

	// Sentryの初期化。ReleaseにはAssetVersion (Gitコミットハッシュ短縮版)
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

	// slogのデフォルトロガーをSentry連携付きハンドラーに差し替える。
	// Errorレベル以上はSentryイベント化され、全レベルは引き続き標準エラー
	// 出力にも届く。slog.Errorを呼ぶ可能性のある処理 (DB接続、river起動など)
	// より前に呼ぶことで、起動時のエラーもSentryに届くようにする。
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

	// Redisクライアントの初期化 (Rate Limiting用とデータストア用)
	var limiter *ratelimit.Limiter
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			slog.Error("Redis URLのパースに失敗しました", "error", err)
			os.Exit(1)
		}

		// コネクションプール設定
		opt.PoolSize = 10
		opt.MinIdleConns = 2
		opt.ConnMaxIdleTime = 5 * time.Minute

		redisClient = redis.NewClient(opt)
		// Redis接続確認
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			slog.Warn("Redisへの接続に失敗しました (Rate Limitingは無効化されます)", "error", err)
			redisClient = nil
		} else {
			slog.Info("Redisに正常に接続しました")
			limiter = ratelimit.NewLimiter(redisClient)
		}
	} else {
		slog.Warn("Redis URLが設定されていません (Rate Limitingは無効化されます)")
	}

	// 下で登録する定期ジョブが使うクリーンアップUseCaseを組み立てる。いずれも対応する
	// `task cleanup-*` サブコマンドと同じヘルパー経由で組み立てるため、定期実行と手動実行で
	// 配線を共有できる。
	cleanupExpiredTokensUC := newCleanupExpiredTokensUsecase(queries)
	cleanupExpiredSignInCodesUC := newCleanupExpiredSignInCodesUsecase(queries)
	cleanupExpiredSessionsUC := newCleanupExpiredSessionsUsecase(queries)

	// フェーズ2のフル・リコンシリエーションバッチUseCaseを組み立てる (Worker用)。
	// works / episodesをanimes / anime_classificationsへ同期する下の定期ジョブから
	// 呼ばれる。配線はnewSyncAnimesUsecase経由で `task sync-animes` サブコマンドと共有する。
	syncAnimesUC := newSyncAnimesUsecase(db, queries)

	// Riverクライアントの初期化
	ctx := context.Background()
	riverClient, err := worker.NewClient(ctx, cfg.DatabaseDSN(), worker.NewClientParams{
		CleanupExpiredTokens:      cleanupExpiredTokensUC,
		CleanupExpiredSignInCodes: cleanupExpiredSignInCodesUC,
		CleanupExpiredSessions:    cleanupExpiredSessionsUC,
		SyncAnimes:                syncAnimesUC,
	}, cfg)
	if err != nil {
		slog.Error("Riverクライアントの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := riverClient.Stop(shutdownCtx); err != nil {
			slog.Warn("Riverクライアントの停止に失敗しました", "error", err)
		}
	}()

	// Riverクライアントを起動
	if err := riverClient.Start(ctx); err != nil {
		slog.Error("Riverクライアントの起動に失敗しました", "error", err)
		os.Exit(1)
	}

	// トークンクリーンアップを毎日深夜2時の定期実行ジョブとして登録する。
	periodicJobTokenCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredTokensArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobTokenCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "トークンクリーンアップ", "schedule", "毎日深夜2時")

	// ログインコードクリーンアップを毎日深夜2時の定期実行ジョブとして登録する。
	periodicJobSignInCodeCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredSignInCodesArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSignInCodeCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "ログインコードクリーンアップ", "schedule", "毎日深夜2時")

	// セッションクリーンアップを毎日深夜2時の定期実行ジョブとして登録する。
	periodicJobSessionCleanup := river.NewPeriodicJob(
		dailyAt2AMSchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.CleanupExpiredSessionsArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSessionCleanup)
	slog.Info("定期実行ジョブを登録しました", "job", "セッションクリーンアップ", "schedule", "毎日深夜2時")

	// animesリコンサイルを毎時登録する。フェーズ2ではまだ両書きが無く、本バッチが
	// animesを更新する唯一の経路のため、毎時実行で鮮度と差分メトリクスの取得頻度を優先する。
	// リコンサイルは冪等で、定常状態ではほぼ差分なしに収束する。
	periodicJobSyncAnimes := river.NewPeriodicJob(
		hourlySchedule{},
		func() (river.JobArgs, *river.InsertOpts) {
			return worker.SyncAnimesArgs{}, nil
		},
		nil,
	)

	riverClient.Client().PeriodicJobs().Add(periodicJobSyncAnimes)
	slog.Info("定期実行ジョブを登録しました", "job", "animes同期バッチ", "schedule", "毎時")

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
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

	// リクエストボディサイズ制限 (10MB)
	requestBodyLimitMW := authMiddleware.NewRequestBodyLimitMiddleware(10 * 1024 * 1024)
	r.Use(requestBodyLimitMW.Middleware)

	// メンテナンスミドルウェア。クライアントIPはclientip.GetClientIPで
	// 取得し、同関数がプロキシヘッダを自前で読むため、他のミドルウェアとの
	// 登録順はIP取得に影響しない。
	maintenanceMW := authMiddleware.NewMaintenanceMiddleware(cfg)
	r.Use(maintenanceMW.Middleware)

	// リバースプロキシミドルウェア
	// Go版で処理するかRails版にプロキシするかを判定
	// Rails版にプロキシする場合は後続のミドルウェアをスキップ
	//
	// あわせてルーターを注入し、Go版とRails版が分け合っているパス (例: /db/*) が
	// どのGoルートにもマッチしない場合に、ミドルウェアがこのレイヤー (下のSentry /
	// CSRFチェーンより前) でRailsへフォールバックできるようにする。
	if reverseProxyMW != nil {
		reverseProxyMW.SetRouter(r)
		r.Use(reverseProxyMW.Middleware)
	}

	// 以下はGo版で処理する場合のみ適用されるミドルウェア。Sentryの
	// チェーンをリバースプロキシの内側に置くことで、Rails版へプロキシされる
	// リクエストがGo版のSentryトランザクションに乗らないようにする。

	// SecurityHeadersは本セクションの先頭に置き、Goが描画する全レスポンス (静的ファイルの
	// 配信と、以下のミドルウェアが返すRecovererの500やCSRFの403を含む) を覆う。
	// リバースプロキシの外側へは動かせない。Railsは同じヘッダーを既に送っており、
	// httputil.ReverseProxyはそれを上書きではなく追記するため。
	r.Use(authMiddleware.SecurityHeaders)

	// Recovererはsentryhttpより前 (= 外側) に登録する。sentryhttpが
	// Repanic: trueで再panicしたものをここで握り潰して500を返すため。
	// Sentry SDK公式READMEも「recovery middlewareはsentryhttpより外側に
	// 置く」と指示している。
	r.Use(middleware.Recoverer)

	// sentryhttpはリクエスト単位のHubをcontextに積み、panicを捕捉して
	// Sentryに送信し、リクエストごとのSentryトランザクションを開始する。
	// Repanic: trueにより捕捉後に再panicすることで、外側のRecovererが
	// クライアントに500を返せる。
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r.Use(sentryHandler.Handle)

	// SentryTransactionはトランザクション名をchiのルートパターンに
	// 上書きするミドルウェア。sentryhttpの後 (= 内側) に登録することで、LIFOの
	// defer順序によりsentryhttpのtransaction.Finish() やrecoverWithSentry
	// より先に書き換えが走ることを保証する。
	r.Use(authMiddleware.SentryTransaction)

	r.Use(authMiddleware.MethodOverride) // Method Overrideミドルウェアを追加 (HTMLフォームからPUT/PATCH/DELETEを使用可能に)
	r.Use(authMW.Middleware)             // 認証ミドルウェアを追加 (ユーザー情報をコンテキストに設定)

	// Sentryユーザーコンテキストミドルウェアを追加 (認証済みユーザーのIDをSentryに設定)
	sentryUserContextMW := authMiddleware.NewSentryUserContextMiddleware()
	r.Use(sentryUserContextMW.Middleware)

	r.Use(authMiddleware.I18n) // I18nミドルウェアを追加 (ユーザーのlocaleを考慮)

	// 現在パスミドルウェアを追加 (サイドバーの現在ページハイライト用にaria-currentを付与)
	r.Use(templates.CurrentPathMiddleware)

	// Annict DBサイドバーのCookie設定をSSRの初期状態用に保存する。
	r.Use(templates.DBSidebarStateMiddleware)

	// CSRF保護ミドルウェアを追加
	csrfMiddleware := authMiddleware.NewCSRFMiddleware(sessionManager)
	r.Use(csrfMiddleware.Middleware)

	// フラッシュミドルウェアはCSRF検証後・全ハンドラー前に配置する。
	// CSRFが失敗するリクエストではflash Cookieを読み取らない (=クリアしない) ため、
	// 次回リクエストでもflashが失われない。
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

	// 視聴記録ヒートマップフラグメントハンドラーの初期化。
	recordRepo := repository.NewRecordRepository(queries)
	getTrackingHeatmapUC := usecase.NewGetTrackingHeatmapUsecase(userRepo, recordRepo)
	trackingHeatmapHandler := tracking_heatmap.NewHandler(getTrackingHeatmapUC)

	// サインインコードリポジトリの初期化
	signInCodeRepo := repository.NewSignInCodeRepository(queries)

	// Dispatcherの初期化
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
	// stripe-goクライアントを、UseCaseごとのinterfaceを実装するadapterで
	// ラップする。これにより各UseCaseは *stripe.Clientではなくseamに依存する。
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

	// 静的ファイルの配信 (Tailwind CLI + esbuildのビルド結果)
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// ルート設定
	r.Get("/", homeHandler.Show)
	r.Get(httperror.NotFoundPath, httperror.NotFound)
	r.Get(httperror.ForbiddenPath, httperror.Forbidden)
	r.Get(httperror.InvalidCSRFTokenPath, httperror.InvalidCSRFToken)
	r.Get(httperror.InternalServerErrorPath, httperror.InternalServerError)
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
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)
	getDBWorksUC := usecase.NewGetDBWorksUsecase(workRepo)
	getDBWorkFormOptionsUC := usecase.NewGetDBWorkFormOptionsUsecase(numberFormatRepo)
	getDBWorkEditUC := usecase.NewGetDBWorkEditUsecase(workRepo, numberFormatRepo)
	// 作品 作成 / 更新UseCaseはworkがsourceとする6つの別表を両書きするため、
	// フェーズ2の同期バッチと同じ別表リポジトリの束を受け取る。
	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	updateWorkUC := usecase.NewUpdateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	deleteWorkUC := usecase.NewDeleteWorkUsecase(db, workRepo, animeRepo)
	dbWorkHandler := db_work.NewHandler(cfg, sessionManager, flashMgr, imageHelper, getDBWorksUC, getDBWorkFormOptionsUC, getDBWorkEditUC, createWorkUC, updateWorkUC, deleteWorkUC)
	getDBWorkArchiveNewUC := usecase.NewGetDBWorkArchiveNewUsecase(workRepo)
	archiveWorkUC := usecase.NewArchiveWorkUsecase(db, workRepo, animeRepo)
	unarchiveWorkUC := usecase.NewUnarchiveWorkUsecase(db, workRepo, animeRepo)
	dbWorkArchiveHandler := db_work_archive.NewHandler(cfg, sessionManager, flashMgr, getDBWorkArchiveNewUC, archiveWorkUC, unarchiveWorkUC)
	getDBWorkUnarchiveNewUC := usecase.NewGetDBWorkUnarchiveNewUsecase(workRepo)
	dbWorkUnarchiveHandler := db_work_unarchive.NewHandler(cfg, sessionManager, getDBWorkUnarchiveNewUC)
	getDBWorkDeletionNewUC := usecase.NewGetDBWorkDeletionNewUsecase(workRepo)
	dbWorkDeletionHandler := db_work_deletion.NewHandler(cfg, sessionManager, getDBWorkDeletionNewUC)
	// 作品一覧は公開 (未ログインを含め誰でも閲覧可。committer限定の操作ボタンは
	// テンプレートで出し分ける)。作品の作成・編集・非公開・再公開はcommitterロールを要する
	// ため、New / Create / Edit / Updateと非公開の確認 / 実行 / 再公開エンドポイント、および
	// 公開の確認をRequireCommitterでまとめてゲートする。作品の削除はadmin専用のため
	// (ADR 0009でarchived=committer / deleted=adminに権限分離)、削除とその確認画面は
	// RequireAdminで別途ゲートする (確認画面には、それを送信できる者だけが到達する)。
	r.Get("/db/works", dbWorkHandler.Index)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireCommitter)
		r.Get("/db/works/new", dbWorkHandler.New)
		r.Post("/db/works", dbWorkHandler.Create)
		r.Get("/db/works/{id}/edit", dbWorkHandler.Edit)
		r.Patch("/db/works/{id}", dbWorkHandler.Update)
		r.Get("/db/works/{id}/archive/new", dbWorkArchiveHandler.New)
		r.Post("/db/works/{id}/archive", dbWorkArchiveHandler.Create)
		r.Get("/db/works/{id}/unarchive/new", dbWorkUnarchiveHandler.New)
		r.Delete("/db/works/{id}/archive", dbWorkArchiveHandler.Delete)
	})
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAdmin)
		r.Get("/db/works/{id}/deletion/new", dbWorkDeletionHandler.New)
		r.Delete("/db/works/{id}", dbWorkHandler.Delete)
	})

	// 作品のエピソード一覧は公開。RailsのDb::EpisodesController#indexが同コントローラ
	// で唯一authenticate_user! を持たないアクションであることに合わせている。エピソードの作成・
	// 編集・非公開・再公開はcommitterロールを要するため、一括作成フォームとその送信、編集
	// フォームとその送信、および非公開の確認とその送信・再公開はRequireCommitterでまとめて
	// ゲートする。エピソードの削除はadmin専用のため (ADR 0009でarchived=committer /
	// deleted=adminに権限分離。RailsのEpisodePolicyが行う分割とも同じ)、RequireAdminで
	// 別途ゲートする。
	episodeRepo := repository.NewEpisodeRepository(queries)
	getDBEpisodesUC := usecase.NewGetDBEpisodesUsecase(workRepo, episodeRepo)
	getDBEpisodeNewUC := usecase.NewGetDBEpisodeNewUsecase(workRepo)
	createEpisodesUC := usecase.NewCreateEpisodesUsecase(db, workRepo, episodeRepo, animeRepo, animeClassificationRepo, validator.NewDBEpisodeCreateValidator())
	getDBEpisodeEditUC := usecase.NewGetDBEpisodeEditUsecase(episodeRepo)
	updateEpisodeUC := usecase.NewUpdateEpisodeUsecase(db, episodeRepo, animeRepo, animeClassificationRepo, validator.NewDBEpisodeUpdateValidator())
	deleteEpisodeUC := usecase.NewDeleteEpisodeUsecase(db, episodeRepo, animeRepo)
	dbEpisodeHandler := db_episode.NewHandler(cfg, sessionManager, flashMgr, getDBEpisodesUC, getDBEpisodeNewUC, createEpisodesUC, getDBEpisodeEditUC, updateEpisodeUC, deleteEpisodeUC)
	getDBEpisodeArchiveNewUC := usecase.NewGetDBEpisodeArchiveNewUsecase(episodeRepo)
	archiveEpisodeUC := usecase.NewArchiveEpisodeUsecase(db, episodeRepo, animeRepo)
	unarchiveEpisodeUC := usecase.NewUnarchiveEpisodeUsecase(db, episodeRepo, animeRepo)
	dbEpisodeArchiveHandler := db_episode_archive.NewHandler(cfg, sessionManager, flashMgr, getDBEpisodeArchiveNewUC, archiveEpisodeUC, unarchiveEpisodeUC)
	r.Get("/db/works/{work_id}/episodes", dbEpisodeHandler.Index)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireCommitter)
		r.Get("/db/works/{work_id}/episodes/new", dbEpisodeHandler.New)
		r.Post("/db/works/{work_id}/episodes", dbEpisodeHandler.Create)
		r.Get("/db/episodes/{id}/edit", dbEpisodeHandler.Edit)
		r.Patch("/db/episodes/{id}", dbEpisodeHandler.Update)
		r.Get("/db/episodes/{id}/archive/new", dbEpisodeArchiveHandler.New)
		r.Post("/db/episodes/{id}/archive", dbEpisodeArchiveHandler.Create)
		r.Delete("/db/episodes/{id}/archive", dbEpisodeArchiveHandler.Delete)
	})
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAdmin)
		r.Delete("/db/episodes/{id}", dbEpisodeHandler.Delete)
	})

	// iCalendar配信
	r.Get("/@{username}/ics", icsHandler.Show) // メインのエンドポイント
	r.Get("/ics", icsHandler.Show)             // Appleカレンダー互換の代替パス (クエリパラメータでusernameを指定)

	// 視聴記録ヒートマップフラグメント。
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

	// Graceful shutdownのためのシグナルハンドリング
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
