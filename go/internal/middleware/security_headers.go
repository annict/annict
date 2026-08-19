package middleware

import "net/http"

// securityHeaders is the header set Rails sends on every response through its
// `config.load_defaults 7.1` defaults. Go serves the same domain, so a screen keeps the same
// protections after it moves over instead of losing them at the migration boundary.
// `X-XSS-Protection: 0` is one of those defaults: it turns the legacy XSS auditor off, which
// is what Rails settled on after the auditor itself proved to introduce holes.
//
// Headers Rails does not send (Content-Security-Policy, Strict-Transport-Security,
// Permissions-Policy) are deliberately absent. Enabling those changes what the pages
// themselves may do, or belongs to the layer that terminates TLS, and either way it is a
// decision about both applications rather than part of keeping a migrated response as it was.
//
// [Ja] securityHeaders は Rails が `config.load_defaults 7.1` の既定として全レスポンスに
// 付けているヘッダーの集合。Go は同じドメインを配信するため、移行した画面が移行の境目で
// 保護を失わないよう同じものを送る。`X-XSS-Protection: 0` もその既定の 1 つで、旧来の XSS
// auditor を無効にする指定。auditor 自体が穴を生むと分かった経緯から Rails が採った値。
//
// Rails が送っていないヘッダー (Content-Security-Policy / Strict-Transport-Security /
// Permissions-Policy) は意図的に含めない。これらの有効化はページ自身に許すことを変えるか、
// TLS を終端する層の責務であり、いずれにせよ移行したレスポンスの性質を保つこととは別に、
// 2 つのアプリケーションについて決める判断になる。
var securityHeaders = map[string]string{
	"X-Frame-Options":                   "SAMEORIGIN",
	"X-XSS-Protection":                  "0",
	"X-Content-Type-Options":            "nosniff",
	"X-Permitted-Cross-Domain-Policies": "none",
	"Referrer-Policy":                   "strict-origin-when-cross-origin",
}

// SecurityHeaders puts securityHeaders on every response the Go app renders, including the
// static files and the shared error pages.
//
// It has to be registered inside the reverse proxy middleware. httputil.ReverseProxy adds the
// upstream's headers to whatever the ResponseWriter already carries instead of replacing them,
// so setting these before a request is proxied would reach the reader as doubled values
// ("nosniff, nosniff") on every Rails-served page. Responses written before the proxy layer
// therefore never pass through here and call setSecurityHeaders themselves.
//
// [Ja] SecurityHeaders は Go が描画する全レスポンス (静的ファイルの配信と共通エラーページを
// 含む) に securityHeaders を付ける。
//
// 本ミドルウェアはリバースプロキシミドルウェアの内側に登録する必要がある。
// httputil.ReverseProxy は上流のヘッダーを ResponseWriter が既に持つ値へ上書きではなく追記
// するため、プロキシより前で設定すると Rails が配信する全ページで値が二重
// ("nosniff, nosniff") になって読み手に届く。そのためプロキシ層より前で書き出すレスポンスは
// ここを通らず、setSecurityHeaders を自分で呼ぶ。
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)

		next.ServeHTTP(w, r)
	})
}

// setSecurityHeaders writes securityHeaders onto w. Call it before the status line is written,
// from the paths that answer before the SecurityHeaders middleware runs.
//
// [Ja] setSecurityHeaders は w に securityHeaders を書き込む。SecurityHeaders ミドルウェアが
// 走るより前に応答を返す経路から、ステータス行を書き出す前に呼ぶ。
func setSecurityHeaders(w http.ResponseWriter) {
	header := w.Header()
	for name, value := range securityHeaders {
		header.Set(name, value)
	}
}
