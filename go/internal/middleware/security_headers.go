package middleware

import "net/http"

// securityHeadersはRailsが `config.load_defaults 7.1` の既定として全レスポンスに
// 付けているヘッダーの集合。Goは同じドメインを配信するため、移行した画面が移行の境目で
// 保護を失わないよう同じものを送る。`X-XSS-Protection: 0` もその既定の1つで、旧来のXSS
// auditorを無効にする指定。auditor自体が穴を生むと分かった経緯からRailsが採った値。
//
// Railsが送っていないヘッダー (Content-Security-Policy / Strict-Transport-Security /
// Permissions-Policy) は意図的に含めない。これらの有効化はページ自身に許すことを変えるか、
// TLSを終端する層の責務であり、いずれにせよ移行したレスポンスの性質を保つこととは別に、
// 2つのアプリケーションについて決める判断になる。
var securityHeaders = map[string]string{
	"X-Frame-Options":                   "SAMEORIGIN",
	"X-XSS-Protection":                  "0",
	"X-Content-Type-Options":            "nosniff",
	"X-Permitted-Cross-Domain-Policies": "none",
	"Referrer-Policy":                   "strict-origin-when-cross-origin",
}

// SecurityHeadersはGoが描画する全レスポンス (静的ファイルの配信と共通エラーページを
// 含む) にsecurityHeadersを付ける。
//
// 本ミドルウェアはリバースプロキシミドルウェアの内側に登録する必要がある。
// httputil.ReverseProxyは上流のヘッダーをResponseWriterが既に持つ値へ上書きではなく追記
// するため、プロキシより前で設定するとRailsが配信する全ページで値が二重
// ("nosniff, nosniff") になって読み手に届く。そのためプロキシ層より前で書き出すレスポンスは
// ここを通らず、setSecurityHeadersを自分で呼ぶ。
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)

		next.ServeHTTP(w, r)
	})
}

// setSecurityHeadersはwにsecurityHeadersを書き込む。SecurityHeadersミドルウェアが
// 走るより前に応答を返す経路から、ステータス行を書き出す前に呼ぶ。
func setSecurityHeaders(w http.ResponseWriter) {
	header := w.Header()
	for name, value := range securityHeaders {
		header.Set(name, value)
	}
}
