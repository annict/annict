// Package configはconfig機能を提供します
package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// 環境
	Env string

	// データベース
	DatabaseURL string

	// Redis
	RedisURL string

	// Resend (メール送信)
	ResendAPIKey    string
	ResendFromEmail string
	ResendFromName  string

	// サーバー
	Port   string
	Domain string

	// Cookie設定
	CookieDomain string

	// セッション (将来的に使用)
	SessionSecure   string
	SessionHTTPOnly string

	// Storage (S3互換オブジェクトストレージ - Cloudflare R2)
	S3BucketName      string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Region          string

	// imgproxy設定
	ImgproxyEndpoint string
	ImgproxyKey      string
	ImgproxySalt     string

	// アセットバージョン (CDNキャッシュ対策用)
	AssetVersion string

	// Rate Limiting設定
	DisableRateLimit bool

	// Rails版アプリのURL (リバースプロキシ用)
	RailsAppURL string

	// Cloudflare Turnstile (Bot対策)
	TurnstileSiteKey   string
	TurnstileSecretKey string

	// メンテナンスモード
	MaintenanceMode bool
	AdminIPs        []string

	// Sentry (エラー追跡)
	SentryDSN              string
	SentryEnvironment      string
	SentryTracesSampleRate float64
	SentryDebug            bool

	// Stripe (決済)
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePriceMonthlyID string
	StripePriceYearlyID  string

	// 季節設定 (サイドバー表示用)
	SeasonPrevious string
	SeasonCurrent  string
	SeasonNext     string
}

// Loadは環境変数から設定を読み込みます
func Load() (*Config, error) {
	// APP_ENVの値を取得 (デフォルト: dev)
	// dev: 開発環境、test: テスト環境、prod: 本番環境
	//
	// すべての環境でGoプロセス起動時には既に環境変数がセット済みです：
	// - ローカル開発/テスト: op run --env-file=".env" が処理済み
	// - CI環境: GitHub Actionsが設定済み
	// - 本番環境: Dokkuが設定済み
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	cfg := &Config{
		Env: env,
	}

	// 必須の環境変数をチェック
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("必須の環境変数DATABASE_URLが設定されていません")
	}

	cfg.Port = os.Getenv("ANNICT_PORT")
	if cfg.Port == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_PORTが設定されていません")
	}

	cfg.Domain = os.Getenv("ANNICT_DOMAIN")
	if cfg.Domain == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_DOMAINが設定されていません")
	}

	cfg.CookieDomain = os.Getenv("ANNICT_COOKIE_DOMAIN")
	if cfg.CookieDomain == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_COOKIE_DOMAINが設定されていません")
	}

	cfg.SessionSecure = os.Getenv("ANNICT_SESSION_SECURE")
	if cfg.SessionSecure == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_SESSION_SECUREが設定されていません")
	}

	cfg.SessionHTTPOnly = os.Getenv("ANNICT_SESSION_HTTPONLY")
	if cfg.SessionHTTPOnly == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_SESSION_HTTPONLYが設定されていません")
	}

	// Storage関連の設定 (Cloudflare R2)
	// これらはオプショナル (seed実行時など、画像アップロードが必要な場合のみ必須)
	cfg.S3BucketName = os.Getenv("ANNICT_S3_BUCKET_NAME")
	cfg.S3Endpoint = os.Getenv("ANNICT_S3_ENDPOINT")
	cfg.S3AccessKeyID = os.Getenv("ANNICT_S3_ACCESS_KEY_ID")
	cfg.S3SecretAccessKey = os.Getenv("ANNICT_S3_SECRET_ACCESS_KEY")

	// R2のリージョンはデフォルトで "auto" を使用
	cfg.S3Region = os.Getenv("ANNICT_S3_REGION")
	if cfg.S3Region == "" {
		cfg.S3Region = "auto"
	}

	// imgproxy関連の設定
	cfg.ImgproxyEndpoint = os.Getenv("ANNICT_IMGPROXY_ENDPOINT")
	if cfg.ImgproxyEndpoint == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_IMGPROXY_ENDPOINTが設定されていません")
	}

	cfg.ImgproxyKey = os.Getenv("ANNICT_IMGPROXY_KEY")
	if cfg.ImgproxyKey == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_IMGPROXY_KEYが設定されていません")
	}

	cfg.ImgproxySalt = os.Getenv("ANNICT_IMGPROXY_SALT")
	if cfg.ImgproxySalt == "" {
		return nil, fmt.Errorf("必須の環境変数ANNICT_IMGPROXY_SALTが設定されていません")
	}

	// Redis (オプショナル - パスワードリセット機能で使用)
	cfg.RedisURL = os.Getenv("ANNICT_REDIS_URL")

	// Resend (オプショナル - パスワードリセット機能で使用)
	cfg.ResendAPIKey = os.Getenv("ANNICT_RESEND_API_KEY")
	cfg.ResendFromEmail = os.Getenv("ANNICT_RESEND_FROM_EMAIL")
	cfg.ResendFromName = os.Getenv("ANNICT_RESEND_FROM_NAME")
	if cfg.ResendFromName == "" {
		cfg.ResendFromName = "Annict"
	}

	// Rate Limiting設定 (オプショナル - 開発環境でRate Limitingを無効化)
	cfg.DisableRateLimit = os.Getenv("ANNICT_DISABLE_RATE_LIMIT") == "true"

	// Rails版アプリのURL (オプショナル - リバースプロキシ機能で使用)
	cfg.RailsAppURL = os.Getenv("ANNICT_RAILS_APP_URL")

	// Cloudflare Turnstile (オプショナル - ログイン・サインアップフォームで使用)
	// テスト環境では空文字列でも動作する (モック設定として使用)
	cfg.TurnstileSiteKey = os.Getenv("ANNICT_TURNSTILE_SITE_KEY")
	cfg.TurnstileSecretKey = os.Getenv("ANNICT_TURNSTILE_SECRET_KEY")

	// ANNICT_TURNSTILE_DISABLEは、非本番環境でsite/secretの2キーを
	// 未設定にする代わりに1フラグでTurnstileを無効化するためのもの。有効時は
	// 両キーを空に落とし、既存の「キー空」経路 (turnstile.Verifyは常に成功し、
	// ウィジェットは描画されない) に委ねる。
	//
	// 本番ではフラグを意図的に無視し (fail-closed)、warnログを出すだけにする。
	// TurnstileはBot対策なので、ANNICT_TURNSTILE_DISABLE=trueが誤って本番に
	// 漏れても、黙って無効化されてはならない。
	if os.Getenv("ANNICT_TURNSTILE_DISABLE") == "true" {
		if cfg.IsProduction() {
			slog.Warn("ANNICT_TURNSTILE_DISABLEは本番環境では無視されます (Bot対策を維持するためfail-closed)")
		} else {
			cfg.TurnstileSiteKey = ""
			cfg.TurnstileSecretKey = ""
		}
	}

	// メンテナンスモード (オプショナル - "on"のときメンテナンスモードを有効化)
	cfg.MaintenanceMode = os.Getenv("ANNICT_MAINTENANCE_MODE") == "on"

	// 管理者IP (オプショナル - カンマ区切りで複数指定可能)
	adminIPStr := os.Getenv("ANNICT_ADMIN_IP")
	if adminIPStr != "" {
		cfg.AdminIPs = parseAdminIPs(adminIPStr)
	}

	// Sentry (オプショナル - エラー追跡サービス)
	cfg.SentryDSN = os.Getenv("ANNICT_SENTRY_DSN")
	cfg.SentryEnvironment = os.Getenv("ANNICT_SENTRY_ENVIRONMENT")
	if cfg.SentryEnvironment == "" {
		cfg.SentryEnvironment = env
	}
	cfg.SentryTracesSampleRate = parseSentryTracesSampleRate(os.Getenv("ANNICT_SENTRY_TRACES_SAMPLE_RATE"))
	cfg.SentryDebug = os.Getenv("ANNICT_SENTRY_DEBUG") == "true"

	// Stripe (オプショナル - サポーター決済)
	cfg.StripeSecretKey = os.Getenv("ANNICT_STRIPE_SECRET_KEY")
	cfg.StripeWebhookSecret = os.Getenv("ANNICT_STRIPE_WEBHOOK_SECRET")
	cfg.StripePriceMonthlyID = os.Getenv("ANNICT_STRIPE_PRICE_MONTHLY_ID")
	cfg.StripePriceYearlyID = os.Getenv("ANNICT_STRIPE_PRICE_YEARLY_ID")

	// アセットバージョン (Gitコミットハッシュ) を設定
	cfg.AssetVersion = getGitCommitHash()

	// 季節設定 (オプショナル - サイドバー表示用)
	cfg.SeasonPrevious = os.Getenv("ANNICT_SEASON_PREVIOUS")
	cfg.SeasonCurrent = os.Getenv("ANNICT_SEASON_CURRENT")
	cfg.SeasonNext = os.Getenv("ANNICT_SEASON_NEXT")

	return cfg, nil
}

// DatabaseDSNはPostgreSQL接続文字列を返します
func (c *Config) DatabaseDSN() string {
	return c.DatabaseURL
}

// IsDevは開発環境かどうかを返します
func (c *Config) IsDev() bool {
	return c.Env == "dev"
}

// IsProductionは本番環境かどうかを返します
func (c *Config) IsProduction() bool {
	return c.Env == "prod"
}

// AppURLはアプリケーションのベースURLを返します
func (c *Config) AppURL() string {
	if c.IsProduction() {
		return "https://" + c.Domain
	}
	return "https://" + c.Domain
}

// 実行中ビルドのGitコミットハッシュ (短縮版) を返す。Sentryのreleaseと、
// CDNキャッシュ対策用のCSS/JSクエリパラメータに使う。
//
// GIT_REVを最優先する。Dokkuのデプロイ先コンテナには .gitディレクトリが無いため
// `git rev-parse` は失敗し、そのままだと "dev" にフォールバックしてしまう。Dokkuは
// 代わりにデプロイ時のコミットハッシュをGIT_REV環境変数で渡す (プラットフォームが
// 提供する変数なのでANNICT_ プレフィックスは付けない)。ローカルのgitコマンドは
// 開発用のフォールバックで、最後の手段が "dev"。
func getGitCommitHash() string {
	// Dokkuはここに完全なデプロイSHAを渡すので、release / キャッシュ
	// バスティング用の識別子を短く安定させるため7文字に短縮する。ローカルの
	// `git rev-parse --short` の通常の出力長に概ね合わせている。
	if rev := strings.TrimSpace(os.Getenv("GIT_REV")); rev != "" {
		const shortHashLen = 7
		if len(rev) > shortHashLen {
			return rev[:shortHashLen]
		}
		return rev
	}

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// Gitが利用できない場合は "dev" を返す (開発環境用のフォールバック)。
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

// GetAssetVersionはアセットのバージョン文字列を返します
// 開発環境: 現在時刻のUnixタイムスタンプ (ミリ秒) を返す (キャッシュを無効化)
// 本番/テスト環境: Gitコミットハッシュを返す (起動時に設定された値)
func (c *Config) GetAssetVersion() string {
	if c.IsDev() {
		// 開発環境では毎回異なる値を返す (現在時刻のUnixタイムスタンプ、ミリ秒)
		return strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	// 本番/テスト環境では起動時に設定されたGitコミットハッシュを返す
	return c.AssetVersion
}

// parseAdminIPsはカンマ区切りのIP文字列をスライスに変換します
// 各IPアドレスの前後の空白は除去されます
func parseAdminIPs(s string) []string {
	parts := strings.Split(s, ",")
	ips := make([]string, 0, len(parts))
	for _, p := range parts {
		ip := strings.TrimSpace(p)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// parseSentryTracesSampleRateは文字列からトレースサンプリングレートをパースします
// 空文字列またはパースに失敗した場合はデフォルト値0.5を返します
func parseSentryTracesSampleRate(s string) float64 {
	if s == "" {
		return 0.5
	}
	rate, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.5
	}
	if rate < 0.0 || rate > 1.0 {
		return 0.5
	}
	return rate
}
