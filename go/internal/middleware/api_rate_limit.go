package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
)

// REST API V2のレート制限は、Rails版がREST API V1・GraphQL APIに掛けるrack-attackと
// 同じキー・上限・窓で数え、3つのAPIで1つのカウンターを共有する。
// キーはRails版の `rate_limits` の `key` (throttleの名前と識別子をつないだもの) と同じ形にする
const (
	// apiRateLimitContextKeyは、1回の計数で得た判定を後続のミドルウェアへ渡すためのキー
	apiRateLimitContextKey contextKey = "api_rate_limit"
	// apiTokenRateLimitKeyPrefixはトークンを送ったリクエストを、トークンと接続元IPの組で数えるキーの接頭辞
	apiTokenRateLimitKeyPrefix = "api:token:"
	// apiIPRateLimitKeyPrefixはトークンを送らないリクエストを、接続元IPで数えるキーの接頭辞
	apiIPRateLimitKeyPrefix = "api:ip:"
	// apiTokenRateLimitは、トークンと接続元IPの組で1つの窓に受け付けるリクエスト数
	apiTokenRateLimit = 600
	// apiIPRateLimitは、トークンを送らないリクエストを接続元IPで1つの窓に受け付けるリクエスト数
	apiIPRateLimit = 60
	// apiRateLimitWindowは窓の長さ (毎分0秒から59秒の固定窓)
	apiRateLimitWindow = time.Minute
)

// apiRateLimiterはREST API V2のリクエスト数を数えるインターフェース。
// ratelimit.Limiterがこのインターフェースを満たす
type apiRateLimiter interface {
	Check(ctx context.Context, input ratelimit.CheckInput) (*ratelimit.CheckResult, error)
}

// APIRateLimitはREST API V2にレート制限を掛けるミドルウェア。
// トークンを照合する前に、受け取ったトークンの文字列で数える (Rails版と同じく、受け付けられないトークンも数える)。
// 数えたリクエストの応答に `RateLimit-Limit`・`RateLimit-Remaining`・`RateLimit-Reset` を付け、上限を超えたら429を返す。
// 本文サイズ超過の413とメンテナンス中の503を数えたうえで429より優先するため、
// 計数 (CountMiddleware) と429の判定 (EnforceMiddleware) を、それらの早期応答の前後に分けて登録する
type APIRateLimit struct {
	cfg      *config.Config
	limiter  apiRateLimiter
	problems *apierror.Writer
	now      func() time.Time
}

// apiRateLimitResultは、計数の時点の判定と、429に付ける再試行までの時間
type apiRateLimitResult struct {
	allowed    bool
	resetAfter time.Duration
}

// NewAPIRateLimitは新しいAPIRateLimitを作成する
func NewAPIRateLimit(cfg *config.Config, limiter apiRateLimiter, problems *apierror.Writer) *APIRateLimit {
	return &APIRateLimit{
		cfg:      cfg,
		limiter:  limiter,
		problems: problems,
		now:      time.Now,
	}
}

// CountMiddlewareはAPIのリクエストを1回数え、残量ヘッダーと判定を残す。
// 413・503も数えるため、それらの早期応答より外側に置く。
// CORSのプリフライトはAPICORSが手前で応えるため数えない。
// Web画面と、Rails版が数えるAPIホストへの要求も数えない
func (l *APIRateLimit) CountMiddleware(next http.Handler) http.Handler {
	count := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l.cfg.DisableRateLimit {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		input := apiRateLimitInput(r)
		result, err := l.limiter.Check(ctx, input)
		if err != nil {
			// Webサイトのレート制限と同じく、カウンターの読み書きに失敗してもAPIは止めない
			slog.WarnContext(ctx, "APIのレート制限の判定に失敗したため制限せずに通します", "error", err)
			next.ServeHTTP(w, r)
			return
		}

		resetAfter := apiRateLimitResetAfter(result.ResetAt, l.now())
		h := w.Header()
		h.Set("RateLimit-Limit", strconv.Itoa(input.Limit))
		h.Set("RateLimit-Remaining", strconv.Itoa(result.Remaining))
		h.Set("RateLimit-Reset", strconv.FormatInt(int64(resetAfter/time.Second), 10))

		ctx = context.WithValue(ctx, apiRateLimitContextKey, apiRateLimitResult{
			allowed:    result.Allowed,
			resetAfter: resetAfter,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
	// APIの計数はルーターの内側のRecovererより先に動くため、ここでもパニックを500に変換する
	recovered := NewRecoverer(l.problems).Middleware(count)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAPIPath(r.URL.Path) || isAPIHost(r.Host, l.cfg.Domain) {
			next.ServeHTTP(w, r)
			return
		}
		recovered.ServeHTTP(w, r)
	})
}

// EnforceMiddlewareはCountMiddlewareの判定を使い、上限超過なら429を返す。
// 本文サイズ超過の413とメンテナンスの503を優先するため、本文制限・メンテナンスの内側に置く
func (l *APIRateLimit) EnforceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, ok := r.Context().Value(apiRateLimitContextKey).(apiRateLimitResult)
		if ok && !result.allowed {
			l.problems.TooManyRequests(w, r, result.resetAfter)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// apiRateLimitInputは、リクエストを数えるカウンターのキーと上限を返す。
// Bearerトークンを送ったリクエストはトークンのSHA-256と接続元IPの組で、送らないリクエストは接続元IPで数える。
// トークンはカウンターのキーに残るため、ハッシュにしてから使う
func apiRateLimitInput(r *http.Request) ratelimit.CheckInput {
	ip := clientip.GetClientIP(r)
	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		digest := sha256.Sum256([]byte(token))
		return ratelimit.CheckInput{
			Key:    apiTokenRateLimitKeyPrefix + hex.EncodeToString(digest[:]) + ":" + ip,
			Limit:  apiTokenRateLimit,
			Window: apiRateLimitWindow,
		}
	}
	return ratelimit.CheckInput{
		Key:    apiIPRateLimitKeyPrefix + ip,
		Limit:  apiIPRateLimit,
		Window: apiRateLimitWindow,
	}
}

// apiRateLimitResetAfterは、窓が終わって残りの数が戻るまでの時間を秒単位に切り上げて返す。
// Rails版と同じく最小で1秒にし、`RateLimit-Reset` と429の `Retry-After` に同じ値を使う
func apiRateLimitResetAfter(resetAt, now time.Time) time.Duration {
	seconds := int64((resetAt.Sub(now) + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return time.Duration(seconds) * time.Second
}
