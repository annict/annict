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

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	"github.com/annict/annict/go/internal/session"
)

// DeviceTokenCookieNameはデバイス (ブラウザ) を識別するCookieのキー名。
const DeviceTokenCookieName = "device_token"

// リバースプロキシが依存するフィーチャーフラグ判定の抽象化インターフェース。
type featureFlagChecker interface {
	IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID model.UserID, name model.FeatureFlagName) (bool, error)
}

// URLパターンと、その経路をゲートするフィーチャーフラグの対応を表す。
type featureFlaggedPattern struct {
	pattern *regexp.Regexp
	flag    model.FeatureFlagName
}

// 現在フラグでゲートされているパスは無い。次にRails版からGo版へ移す画面は
// この仕組みで段階的に公開するため、リストは削除せず空のまま残す (ADR 0002)。
var featureFlaggedPatterns []featureFlaggedPattern

// ReverseProxyMiddlewareはRails版へのリバースプロキシミドルウェア
type ReverseProxyMiddleware struct {
	railsURL *url.URL
	proxy    *httputil.ReverseProxy
	cfg      *config.Config
	// nil許容。フラグ機能不要時はnil
	featureFlagRepo featureFlagChecker
	// nil許容。テスト時やセッション不要時はnil
	sessionMgr *session.Manager
	// nil許容。SetRouterで設定する。Go版とRails版で分け合っているパスが
	// 登録済みのGoルートにマッチするかをミドルウェアが判定できるようにする。
	// nilのときはそれらのパスをGoチェーンに委ねる。
	router chi.Router
}

// Go版で処理するパス (ホワイトリスト)
// これらのパスはRails版にプロキシせず、Go版のハンドラーで処理する
var goHandledPaths = []string{
	"/static",           // 静的ファイル (CSS、JS、画像など)
	"/health",           // ヘルスチェックエンドポイント
	"/manifest.json",    // Web App Manifest
	"/sign_in/password", // パスワードログインページ・処理
	"/sign_in/code",     // 6桁コード入力・検証・再送信
	"/sign_in",          // メールアドレス入力・ログイン方法自動判定
	"/sign_out",         // ログアウト処理
	"/sign_up",          // 新規登録 (メールアドレス入力・確認コード送信)
	"/password/reset",   // パスワードリセット申請
	"/password/edit",    // パスワードリセット実行
	"/password",         // パスワード更新
	"/supporters",       // サポーターページ
	"/webhooks/stripe",  // Stripe Webhook受信
	"/ics",              // iCalendar配信 (Appleカレンダー互換パス)
}

// Goが配信する全画面エラーページ。isGoHandledPathが完全一致で判定する。
// cmd/annict/serve.goでルートとして登録し、拒否されたHTMXリクエストに遷移を指示したときに
// 到達する。
var goHandledErrorPaths = []string{
	httperror.NotFoundPath,
	httperror.ForbiddenPath,
	httperror.InvalidCSRFTokenPath,
	httperror.InternalServerErrorPath,
}

// Go版とRails版が分け合っているパスの接頭辞。Go版はルートを持つ画面だけを処理し、
// 残りはRails版が処理する。isGoSharedPathが接頭辞で判定したうえでmatchesGoRouteが
// ルーターに問い合わせるため、Go版にルートが無いパスはGoの404ではなくRails版に届く。
// 接頭辞の配下すべてをGo版に渡すgoHandledPathsとの違いはここにある。
//
// /db/ がこの接頭辞にあたる。作品とエピソードのCRUD画面はGo版にあるが、キャスト・
// スタッフ・放送枠・作品画像・シリーズなどAnnict DBの残りはRails版にしかない。
var goSharedPaths = []string{
	"/db/",
}

// NewReverseProxyMiddlewareは新しいReverseProxyMiddlewareを作成
func NewReverseProxyMiddleware(railsURL string, cfg *config.Config, featureFlagRepo featureFlagChecker, sessionMgr *session.Manager) (*ReverseProxyMiddleware, error) {
	parsedURL, err := url.Parse(railsURL)
	if err != nil {
		return nil, err
	}

	// httputil.ReverseProxyを作成
	proxy := &httputil.ReverseProxy{}

	// カスタムのHTTP Transportを設定 (タイムアウトと接続プーリング)
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

	// プロキシのRewrite関数でヘッダー設定を行う。httputil.ReverseProxyはRewrite呼び出し前に
	// Forwarded / X-Forwarded-For / X-Forwarded-Host / X-Forwarded-ProtoをOut.Headerから削除するため、
	// 元の値を参照したい場合はpr.In.Headerから取得する必要がある。
	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		// URLをRails版のホストに書き換える。SetURLはOut.Host = "" をセットしてしまうため、
		// 続けてOut.Host = In.Hostを設定し、クライアントが送ってきたHostヘッダをそのままRails版に
		// 転送する挙動を維持する。
		pr.SetURL(parsedURL)
		pr.Out.Host = pr.In.Host

		// クライアントIPアドレスを取得 (優先順位: CF-Connecting-IP > X-Forwarded-Forの最初のIP > RemoteAddr)
		clientIP := clientip.GetClientIP(pr.In)

		// X-Forwarded-Forの設定
		if originalXForwardedFor := pr.In.Header.Get("X-Forwarded-For"); originalXForwardedFor != "" {
			// 既存の値を維持 (Cloudflareなどが設定した値を保持)
			pr.Out.Header.Set("X-Forwarded-For", originalXForwardedFor)
		} else {
			// 既存の値がない場合、clientIPを設定
			pr.Out.Header.Set("X-Forwarded-For", clientIP)
		}

		// X-Real-IPの設定 (既存の値がない場合のみclientIPを設定)
		if originalXRealIP := pr.In.Header.Get("X-Real-IP"); originalXRealIP != "" {
			pr.Out.Header.Set("X-Real-IP", originalXRealIP)
		} else {
			pr.Out.Header.Set("X-Real-IP", clientIP)
		}

		// X-Forwarded-Protoの設定
		pr.Out.Header.Set("X-Forwarded-Proto", "https")

		// X-Forwarded-Hostの設定
		pr.Out.Header.Set("X-Forwarded-Host", cfg.Domain)

		// ログ出力 (開発者向け)
		slog.Info("リバースプロキシでRails版にリクエストを転送",
			"path", pr.In.URL.Path,
			"method", pr.In.Method,
			"target", parsedURL.String()+pr.In.URL.Path,
			"client_ip", clientIP,
		)
	}

	// レスポンス処理後のログ出力 (成功時)
	proxy.ModifyResponse = func(resp *http.Response) error {
		// プロキシが成功した場合のレスポンスログを出力 (開発者向け)
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

		// 開発者向けの詳細エラーログ。source属性を載せておくと
		// SentryのbeforeSendで本イベントを破棄できる (Rails側のプロキシ
		// 失敗はRailsのSentryプロジェクトで扱うべきため)。
		slog.ErrorContext(ctx, "Rails版へのプロキシでエラーが発生",
			annictSentry.SourceAttrKey, annictSentry.ReverseProxySource,
			"error", err,
			"path", r.URL.Path,
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
		)

		// 以下のページはRailsから中継したものではなくGoが描画するため、他のGoの
		// レスポンスと同じヘッダーを付ける。SecurityHeadersミドルウェアは本ミドルウェアの
		// 内側にあり、リクエストはそこまで到達していないため、付与を任せられない。
		setSecurityHeaders(w)

		// 本ハンドラーはGoのミドルウェアチェーンの外側であるプロキシ層で動くため、
		// コンテキストにはまだロケールが載っていない。ここで解決し、共通エラーページが
		// チェーンの内側で配信されるページと同じく読み手の言語で表示されるようにする。
		httperror.BadGateway(w, r.WithContext(i18n.SetLocale(ctx, resolveLocale(r))))
	}

	return &ReverseProxyMiddleware{
		railsURL:        parsedURL,
		proxy:           proxy,
		cfg:             cfg,
		featureFlagRepo: featureFlagRepo,
		sessionMgr:      sessionMgr,
	}, nil
}

// MiddlewareはHTTPミドルウェアを返す
func (m *ReverseProxyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// api.annict.com (または開発環境の相当するホスト) の場合、すべてRails版にプロキシ
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
			// パスは有効な移行フラグでゲートされている。Goルートにマッチするときだけ
			// Goチェーンに渡し、マッチしない場合はその画面がGo未実装のため、ここで
			// Railsへプロキシする (下のフラグ無効時と同じレイヤー)。chiのNotFound
			// ハンドラー (Rails行きのリクエストがスキップすべきSentry / CSRF
			// ミドルウェアチェーンの内側で走る) へ流さないためにこうする。
			if m.matchesGoRoute(r) {
				next.ServeHTTP(w, r)
				return
			}
			m.proxy.ServeHTTP(w, r)
			return
		}

		// パスはGo版とRails版が分け合っているもの。Goルートにマッチするときだけ
		// Goチェーンに渡し、マッチしない場合はその画面がRails版にしかないため、
		// 下のプロキシに流す。
		if m.isGoSharedPath(r.URL.Path) && m.matchesGoRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Rails版にプロキシ
		m.proxy.ServeHTTP(w, r)
	})
}

// SetRouterはchiルーターを注入し、Go版とRails版が分け合っているパス
// (例: /db/*) が登録済みのGoルートにマッチするかをミドルウェアが判定できるようにする。
// ルーター生成後・配信開始前にセットアップで1回呼ぶ。そうしたパスがどのGoルートにも
// マッチしない場合、その画面はRails版のものなので、MiddlewareはchiのNotFound
// ハンドラー (Rails行きのリクエストがスキップすべきSentry / CSRFミドルウェアチェーンの
// 内側で走る) に委ねず、自身でRailsへプロキシする。
func (m *ReverseProxyMiddleware) SetRouter(router chi.Router) {
	m.router = router
}

// matchesGoRouteはリクエストをGoのルートが処理するかどうかを返す。ルーターが無い場合は
// すべてGoチェーンへ渡す (SetRouter呼び出し前の挙動)。
//
// HTMLフォームはGETとPOSTしか送れないため、PUT / PATCH / DELETEで登録したルートには
// MethodOverrideが適用する _methodパラメータ経由で到達する。MethodOverrideは本ミドルウェアの
// 内側で動くため、ここで見えるメソッドはブラウザが送ったPOSTのままであり、それだけで判定すると
// 該当ルートを取りこぼして実装済みの画面をRailsへプロキシしてしまう。ここで _methodを読むことは
// できない。_methodはボディにあり、プロキシするリクエストはそのボディを必要とするため。そこで、
// どのルートにもマッチしないPOSTは、オーバーライドが生みうるメソッドで再判定する。_methodを
// 持たないPOSTはこの結果Goチェーンへ渡りRailsではなく405で終わるが、Goが所有していて
// POSTを受け付けないパスに対する答えとしてはそのほうが正しい。
func (m *ReverseProxyMiddleware) matchesGoRoute(r *http.Request) bool {
	if m.router == nil {
		return true
	}

	if m.router.Match(chi.NewRouteContext(), r.Method, r.URL.Path) {
		return true
	}

	if r.Method != http.MethodPost {
		return false
	}

	for _, method := range overridableMethods {
		if m.router.Match(chi.NewRouteContext(), method, r.URL.Path) {
			return true
		}
	}

	return false
}

// リクエストからdevice_token Cookieを取得し、未設定の場合は新規生成してレスポンスにセットしたうえで返す。
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
		// 10年
		MaxAge:   10 * 365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	return token
}

// リクエストパスがフィーチャーフラグで有効化されているかを判定する。device tokenはensureDeviceTokenで
// 確保済みのトークンを受け取る。featureFlagRepoがnilの場合や判定に失敗した場合はfalseを返し、呼び出し側に
// Rails版へのフォールバックを促す。
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

// isGoHandledPathはGo版で処理するパスかどうかを判定
func (m *ReverseProxyMiddleware) isGoHandledPath(path string) bool {
	// goHandledPathsは前方一致で判定するため、一覧に入れると /errors配下すべてを
	// Go側が引き取ることになる。Goが持つエラーページは下記の一覧だけなので完全一致で判定する。
	for _, p := range goHandledErrorPaths {
		if path == p {
			return true
		}
	}

	for _, p := range goHandledPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	// /@{username}/icsパターンの判定
	if strings.HasPrefix(path, "/@") && strings.HasSuffix(path, "/ics") {
		return true
	}

	// /fragment/@{username}/tracking_heatmapパターンの判定。
	// /fragment/ 配下のうち、tracking_heatmapだけがGo版に移行している段階で、
	// 他の /fragment/... はGo版実装が揃うまでRails版が処理する。
	if strings.HasPrefix(path, "/fragment/@") && strings.HasSuffix(path, "/tracking_heatmap") {
		return true
	}

	return false
}

// isGoSharedPathはパスがGo版とRails版で分け合っている接頭辞の配下かどうかを返す。
// 呼び出し側はmatchesGoRouteと組み合わせ、どちらが処理するかを判定する。
func (m *ReverseProxyMiddleware) isGoSharedPath(path string) bool {
	for _, p := range goSharedPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// isAPISubdomainはAPIサブドメイン (api.annict.comまたはapi.annict-dev.page) かどうかを判定
func (m *ReverseProxyMiddleware) isAPISubdomain(host string) bool {
	// ポート番号を除去 (開発環境では :8080などのポートが含まれる場合がある)
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
