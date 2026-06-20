package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	"github.com/annict/annict/go/internal/session"
)

// DeviceTokenCookieName is the cookie name used to identify a device (browser).
//
// [Ja] デバイス (ブラウザ) を識別する Cookie のキー名。
const DeviceTokenCookieName = "device_token"

// featureFlagChecker abstracts the feature flag evaluator that the reverse proxy depends on.
//
// [Ja] リバースプロキシが依存するフィーチャーフラグ判定の抽象化インターフェース。
type featureFlagChecker interface {
	IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID model.UserID, name model.FeatureFlagName) (bool, error)
}

// featureFlaggedPattern pairs a URL pattern with the feature flag that gates routing for it.
//
// [Ja] URL パターンと、その経路をゲートするフィーチャーフラグの対応を表す。
type featureFlaggedPattern struct {
	pattern *regexp.Regexp
	flag    model.FeatureFlagName
}

var featureFlaggedPatterns = []featureFlaggedPattern{
	{pattern: regexp.MustCompile(`^/db/`), flag: model.FeatureFlagGoAnnictDB},
}

// ReverseProxyMiddleware はRails版へのリバースプロキシミドルウェア
type ReverseProxyMiddleware struct {
	railsURL *url.URL
	proxy    *httputil.ReverseProxy
	cfg      *config.Config
	// optional; falls back to Rails when nil.
	//
	// [Ja] nil 許容。フラグ機能不要時は nil
	featureFlagRepo featureFlagChecker
	// optional; nil during tests or when session is not needed.
	//
	// [Ja] nil 許容。テスト時やセッション不要時は nil
	sessionMgr *session.Manager
}

// Go版で処理するパス（ホワイトリスト）
// これらのパスはRails版にプロキシせず、Go版のハンドラーで処理する
var goHandledPaths = []string{
	"/static",           // 静的ファイル（CSS、JS、画像など）
	"/health",           // ヘルスチェックエンドポイント
	"/manifest.json",    // Web App Manifest
	"/sign_in/password", // パスワードログインページ・処理
	"/sign_in/code",     // 6桁コード入力・検証・再送信
	"/sign_in",          // メールアドレス入力・ログイン方法自動判定
	"/sign_out",         // ログアウト処理
	"/sign_up",          // 新規登録（メールアドレス入力・確認コード送信）
	"/password/reset",   // パスワードリセット申請
	"/password/edit",    // パスワードリセット実行
	"/password",         // パスワード更新
	"/supporters",       // サポーターページ
	"/webhooks/stripe",  // Stripe Webhook受信
	"/ics",              // iCalendar配信（Apple カレンダー互換パス）
}

// NewReverseProxyMiddleware は新しいReverseProxyMiddlewareを作成
func NewReverseProxyMiddleware(railsURL string, cfg *config.Config, featureFlagRepo featureFlagChecker, sessionMgr *session.Manager) (*ReverseProxyMiddleware, error) {
	parsedURL, err := url.Parse(railsURL)
	if err != nil {
		return nil, err
	}

	// httputil.ReverseProxyを作成
	proxy := &httputil.ReverseProxy{}

	// カスタムのHTTP Transportを設定（タイムアウトと接続プーリング）
	proxy.Transport = &http.Transport{
		// 接続タイムアウト: 10秒
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// レスポンスヘッダー読み取りタイムアウト: 30秒
		ResponseHeaderTimeout: 30 * time.Second,
		// 接続プーリングの設定
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Customize header rewriting via the proxy's Rewrite function. ReverseProxy strips Forwarded /
	// X-Forwarded-For / X-Forwarded-Host / X-Forwarded-Proto from Out.Header before calling Rewrite,
	// so read the original values from pr.In.Header when they need to be preserved.
	//
	// [Ja] プロキシの Rewrite 関数でヘッダー設定を行う。httputil.ReverseProxy は Rewrite 呼び出し前に
	// Forwarded / X-Forwarded-For / X-Forwarded-Host / X-Forwarded-Proto を Out.Header から削除するため、
	// 元の値を参照したい場合は pr.In.Header から取得する必要がある。
	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		// Rewrite the URL to the Rails host. SetURL sets Out.Host = "", so follow it with
		// Out.Host = In.Host to keep forwarding the client's Host header to Rails unchanged.
		//
		// [Ja] URL を Rails 版のホストに書き換える。SetURL は Out.Host = "" をセットしてしまうため、
		// 続けて Out.Host = In.Host を設定し、クライアントが送ってきた Host ヘッダをそのまま Rails 版に
		// 転送する挙動を維持する。
		pr.SetURL(parsedURL)
		pr.Out.Host = pr.In.Host

		// Get the client IP (priority: CF-Connecting-IP > first IP of X-Forwarded-For > RemoteAddr).
		//
		// [Ja] クライアント IP アドレスを取得 (優先順位: CF-Connecting-IP > X-Forwarded-For の最初の IP > RemoteAddr)
		clientIP := clientip.GetClientIP(pr.In)

		// Set X-Forwarded-For.
		//
		// [Ja] X-Forwarded-For の設定
		if originalXForwardedFor := pr.In.Header.Get("X-Forwarded-For"); originalXForwardedFor != "" {
			// Keep the existing value (preserve what Cloudflare etc. set).
			//
			// [Ja] 既存の値を維持 (Cloudflare などが設定した値を保持)
			pr.Out.Header.Set("X-Forwarded-For", originalXForwardedFor)
		} else {
			// Set clientIP when there is no existing value.
			//
			// [Ja] 既存の値がない場合、clientIPを設定
			pr.Out.Header.Set("X-Forwarded-For", clientIP)
		}

		// Set X-Real-IP (set clientIP only when there is no existing value).
		//
		// [Ja] X-Real-IP の設定 (既存の値がない場合のみ clientIP を設定)
		if originalXRealIP := pr.In.Header.Get("X-Real-IP"); originalXRealIP != "" {
			pr.Out.Header.Set("X-Real-IP", originalXRealIP)
		} else {
			pr.Out.Header.Set("X-Real-IP", clientIP)
		}

		// Set X-Forwarded-Proto.
		//
		// [Ja] X-Forwarded-Proto の設定
		pr.Out.Header.Set("X-Forwarded-Proto", "https")

		// Set X-Forwarded-Host.
		//
		// [Ja] X-Forwarded-Host の設定
		pr.Out.Header.Set("X-Forwarded-Host", cfg.Domain)

		// Log output (for developers).
		//
		// [Ja] ログ出力 (開発者向け)
		slog.Info("リバースプロキシでRails版にリクエストを転送",
			"path", pr.In.URL.Path,
			"method", pr.In.Method,
			"target", parsedURL.String()+pr.In.URL.Path,
			"client_ip", clientIP,
		)
	}

	// レスポンス処理後のログ出力（成功時）
	proxy.ModifyResponse = func(resp *http.Response) error {
		// プロキシが成功した場合のレスポンスログを出力（開発者向け）
		slog.Info("Rails版からレスポンスを受信",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"path", resp.Request.URL.Path,
			"method", resp.Request.Method,
		)
		return nil
	}

	// エラーハンドラーをカスタマイズ
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		ctx := r.Context()

		// Detailed error log for developers. The source attribute lets the
		// Sentry beforeSend hook drop this event: Rails-side proxy failures
		// belong to the Rails Sentry project, not the Go one.
		//
		// [Ja] 開発者向けの詳細エラーログ。source 属性を載せておくと
		// Sentry の beforeSend で本イベントを破棄できる (Rails 側のプロキシ
		// 失敗は Rails の Sentry プロジェクトで扱うべきため)。
		slog.ErrorContext(ctx, "Rails版へのプロキシでエラーが発生",
			annictSentry.SourceAttrKey, annictSentry.ReverseProxySource,
			"error", err,
			"path", r.URL.Path,
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
		)

		// 502エラーテンプレートをレンダリング
		if err := render502Error(w, r, cfg); err != nil {
			// テンプレートのレンダリングに失敗した場合は、シンプルなテキストを返す
			slog.ErrorContext(ctx, "502エラーテンプレートのレンダリングに失敗",
				"error", err,
			)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			// フォールバックエラーレスポンスなので、書き込みエラーは無視
			_, _ = w.Write([]byte("<html><body><h1>502 Bad Gateway</h1><p>Service Unavailable</p></body></html>"))
		}
	}

	return &ReverseProxyMiddleware{
		railsURL:        parsedURL,
		proxy:           proxy,
		cfg:             cfg,
		featureFlagRepo: featureFlagRepo,
		sessionMgr:      sessionMgr,
	}, nil
}

// Middleware はHTTPミドルウェアを返す
func (m *ReverseProxyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// api.annict.com（または開発環境の相当するホスト）の場合、すべてRails版にプロキシ
		if m.isAPISubdomain(r.Host) {
			m.proxy.ServeHTTP(w, r)
			return
		}

		deviceToken := m.ensureDeviceToken(w, r)

		// Go版で処理するパスかどうかをチェック
		if m.isGoHandledPath(r.URL.Path) {
			// Go版で処理する
			next.ServeHTTP(w, r)
			return
		}

		if m.isFeatureFlagEnabled(r, deviceToken) {
			next.ServeHTTP(w, r)
			return
		}

		// Rails版にプロキシ
		m.proxy.ServeHTTP(w, r)
	})
}

// ensureDeviceToken returns the device_token cookie value from the request, generating and setting a new one on the
// response when missing. An empty string is returned only when token generation itself fails.
//
// [Ja] リクエストから device_token Cookie を取得し、未設定の場合は新規生成してレスポンスにセットしたうえで返す。
// トークン生成自体に失敗した場合のみ空文字列を返す。
func (m *ReverseProxyMiddleware) ensureDeviceToken(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(DeviceTokenCookieName); err == nil && c.Value != "" {
		return c.Value
	}

	token, err := auth.GenerateSecureToken()
	if err != nil {
		slog.ErrorContext(r.Context(), "デバイストークンの生成に失敗", "error", err)
		return ""
	}

	secure := m.cfg.Env != "development"
	http.SetCookie(w, &http.Cookie{
		Name:  DeviceTokenCookieName,
		Value: token,
		Path:  "/",
		// 10 years.
		//
		// [Ja] 10 年
		MaxAge:   10 * 365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	return token
}

// isFeatureFlagEnabled reports whether the request path is enabled by a feature flag. It receives the device token
// returned by ensureDeviceToken. When featureFlagRepo is nil or evaluation fails, it returns false so the caller falls
// back to the Rails proxy.
//
// [Ja] リクエストパスがフィーチャーフラグで有効化されているかを判定する。device token は ensureDeviceToken で
// 確保済みのトークンを受け取る。featureFlagRepo が nil の場合や判定に失敗した場合は false を返し、呼び出し側に
// Rails 版へのフォールバックを促す。
func (m *ReverseProxyMiddleware) isFeatureFlagEnabled(r *http.Request, deviceToken string) bool {
	if m.featureFlagRepo == nil {
		return false
	}

	var flagName model.FeatureFlagName
	matched := false
	for _, fp := range featureFlaggedPatterns {
		if fp.pattern.MatchString(r.URL.Path) {
			flagName = fp.flag
			matched = true
			break
		}
	}
	if !matched {
		return false
	}

	ctx := r.Context()

	var userID model.UserID
	if m.sessionMgr != nil {
		sessionID, err := m.sessionMgr.GetSessionID(r)
		if err == nil && sessionID != "" {
			sessionData, err := m.sessionMgr.GetSession(ctx, sessionID)
			if err == nil && sessionData != nil && sessionData.UserID != nil {
				userID = *sessionData.UserID
			}
		}
	}

	enabled, err := m.featureFlagRepo.IsEnabledByDeviceOrUser(ctx, deviceToken, userID, flagName)
	if err != nil {
		slog.ErrorContext(ctx, "フィーチャーフラグの判定に失敗", "error", err, "flag", flagName, "path", r.URL.Path)
		return false
	}

	return enabled
}

// isGoHandledPath はGo版で処理するパスかどうかを判定
func (m *ReverseProxyMiddleware) isGoHandledPath(path string) bool {
	for _, p := range goHandledPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	// /@{username}/ics パターンの判定
	if strings.HasPrefix(path, "/@") && strings.HasSuffix(path, "/ics") {
		return true
	}

	// /fragment/@{username}/tracking_heatmap pattern. Only the tracking
	// heatmap fragment endpoint moves to Go; other /fragment/... paths are
	// still served by Rails until their Go versions land.
	//
	// [Ja] /fragment/@{username}/tracking_heatmap パターンの判定。
	// /fragment/ 配下のうち、tracking_heatmap だけが Go 版に移行している段階で、
	// 他の /fragment/... は Go 版実装が揃うまで Rails 版が処理する。
	if strings.HasPrefix(path, "/fragment/@") && strings.HasSuffix(path, "/tracking_heatmap") {
		return true
	}

	return false
}

// isAPISubdomain はAPIサブドメイン（api.annict.com または api.annict-dev.page）かどうかを判定
func (m *ReverseProxyMiddleware) isAPISubdomain(host string) bool {
	// ポート番号を除去（開発環境では :8080 などのポートが含まれる場合がある）
	hostWithoutPort := host
	if idx := strings.Index(host, ":"); idx != -1 {
		hostWithoutPort = host[:idx]
	}

	// APIサブドメインのパターン
	apiSubdomains := []string{
		"api." + m.cfg.Domain,           // 例: api.annict.com, api.annict-dev.page
		"api." + m.cfg.Domain + ":8080", // 開発環境でポート付きの場合
	}

	for _, apiSubdomain := range apiSubdomains {
		if strings.EqualFold(hostWithoutPort, apiSubdomain) || strings.EqualFold(host, apiSubdomain) {
			return true
		}
	}

	return false
}

// render502Error は502エラーページをレンダリング
// 注: リバースプロキシのエラーハンドラーはi18nミドルウェアより前に実行されるため、
// シンプルなHTMLを返す。
func render502Error(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	// シンプルな502エラーページ（日本語）
	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>サービス接続エラー - Annict</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 600px;
            padding: 2rem;
            text-align: center;
        }
        h1 {
            font-size: 2rem;
            color: #333;
            margin-bottom: 1rem;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 2rem;
        }
        a {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background-color: #3b82f6;
            color: white;
            text-decoration: none;
            border-radius: 0.375rem;
            transition: background-color 0.2s;
        }
        a:hover {
            background-color: #2563eb;
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⚠️</div>
        <h1>サービス接続エラー</h1>
        <p>申し訳ございません。現在サービスに接続できません。<br>しばらくしてから再度お試しください。</p>
        <a href="/">トップページに戻る</a>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	_, err := w.Write([]byte(html))
	return err
}
