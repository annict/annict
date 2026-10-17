package middleware

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
)

// webRateLimitKeyPrefixはWebサイトのレート制限のキーの名前空間。
// Rails版がAPIホストで使う `api:` と分け、同じ `rate_limits` テーブルで別々に数える。
const webRateLimitKeyPrefix = "web:ip:"

// webRateLimitWindowはWebサイトのレート制限の窓の長さ (毎分0秒から59秒の固定窓)
const webRateLimitWindow = time.Minute

// レート制限で数えない静的なファイルのパスの接頭辞。Go版の静的ファイル (/static) と、
// Go版が中継するRails版のアセット (Propshaftの `/assets` とpublic配下のディレクトリ) を含む。
var rateLimitExemptPathPrefixes = []string{
	"/static/",
	"/assets/",
	"/images/",
	"/js/",
}

// レート制限で数えないパス (完全一致)。Rails版のpublic直下の静的なファイルと、
// 人のリクエストではないヘルスチェック・StripeのWebhookを含む。
var rateLimitExemptPaths = []string{
	"/health",
	"/webhooks/stripe",
	"/favicon.ico",
	"/icon.png",
	"/icon.svg",
	"/robots.txt",
	"/ads.txt",
}

// RateLimitMiddlewareはWebサイトに接続元IP単位のレート制限を掛けるミドルウェア。
// リバースプロキシより前に置き、Go版が描画するページとRails版へ中継するページの両方を数える。
type RateLimitMiddleware struct {
	cfg      *config.Config
	limiter  *ratelimit.Limiter
	sessions *session.Manager
	now      func() time.Time
}

// NewRateLimitMiddlewareは新しいRateLimitMiddlewareを作成
func NewRateLimitMiddleware(cfg *config.Config, limiter *ratelimit.Limiter, sessions *session.Manager) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cfg:      cfg,
		limiter:  limiter,
		sessions: sessions,
		now:      time.Now,
	}
}

// MiddlewareはHTTPミドルウェアを返す
// 上限を超えた場合は429に `Retry-After` を付けて共通のエラーページを返す
func (m *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cfg.DisableRateLimit || m.isExempt(r) {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		result, err := m.limiter.Check(ctx, ratelimit.CheckInput{
			Key:    webRateLimitKeyPrefix + clientip.GetClientIP(r),
			Limit:  m.cfg.WebRateLimitPerMinute,
			Window: webRateLimitWindow,
		})
		if err != nil {
			// カウンターの読み書きに失敗してもサイトは止めず、制限を掛けずに通す。
			// 全リクエストで通る経路のため、Sentryへ送るErrorではなくWarnで残す
			// (原因のDB障害は他の経路のエラーで通知される)。
			slog.WarnContext(ctx, "レート制限の判定に失敗したため制限せずに通します", "error", err, "path", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		if result.Allowed {
			next.ServeHTTP(w, r)
			return
		}

		// 本ミドルウェアはリバースプロキシより前で応答するため、その内側に登録した
		// SecurityHeadersミドルウェアは本レスポンスに対して走らない。
		setSecurityHeaders(w)
		w.Header().Set("Retry-After", strconv.Itoa(m.retryAfterSeconds(result.ResetAt)))

		// 認証ミドルウェアより前で応答するため、429のときだけセッションから
		// ユーザーを取得し、通常ページと同じ優先順位でロケールを解決する。
		user, err := m.sessions.GetCurrentUser(ctx, r)
		if err != nil {
			slog.WarnContext(ctx, "429ページの言語設定を取得できなかったためブラウザの言語を使います", "error", err)
		}
		if user != nil {
			ctx = context.WithValue(ctx, UserContextKey, user)
		}
		// resolveLocaleはリクエストのコンテキストからユーザーを読むため、先にユーザーを載せる
		r = r.WithContext(ctx)
		r = r.WithContext(i18n.SetLocale(ctx, resolveLocale(r)))
		httperror.TooManyRequests(w, r)
	})
}

// isExemptはレート制限で数えないリクエストかどうかを返す
func (m *RateLimitMiddleware) isExempt(r *http.Request) bool {
	// APIホストの方針はRails版 (rack-attack) で掛けるため、Go版を経由する構成でも数えない
	if isAPIHost(r.Host, m.cfg.Domain) {
		return true
	}

	path := r.URL.Path
	for _, exemptPath := range rateLimitExemptPaths {
		if path == exemptPath {
			return true
		}
	}
	for _, prefix := range rateLimitExemptPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// retryAfterSecondsは窓が終わるまでの秒数を返す。端数は切り上げ、最小で1秒とする
func (m *RateLimitMiddleware) retryAfterSeconds(resetAt time.Time) int {
	seconds := int(math.Ceil(resetAt.Sub(m.now()).Seconds()))
	return max(seconds, 1)
}
