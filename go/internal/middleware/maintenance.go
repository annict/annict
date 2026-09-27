package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/templates/pages/maintenance"
)

// MaintenanceMiddlewareはメンテナンスモード時にアクセスを制限するミドルウェア
type MaintenanceMiddleware struct {
	cfg *config.Config
}

// NewMaintenanceMiddlewareは新しいMaintenanceMiddlewareを作成
func NewMaintenanceMiddleware(cfg *config.Config) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		cfg: cfg,
	}
}

// MiddlewareはHTTPミドルウェアを返す
// メンテナンスモードが有効で、管理者IP以外からのアクセスの場合は503を返す
func (m *MaintenanceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// メンテナンスモードが無効の場合は通常処理
		if !m.cfg.MaintenanceMode {
			next.ServeHTTP(w, r)
			return
		}

		// 管理者IPからのアクセスは通常処理
		if m.isAdminIP(r) {
			next.ServeHTTP(w, r)
			return
		}

		// 本ミドルウェアはリバースプロキシより前で応答するため、その内側に登録した
		// SecurityHeadersミドルウェアは本レスポンスに対して走らない。
		setSecurityHeaders(w)

		// htmxは204と304以外のレスポンスをスワップし、hx-targetを指定していない
		// hx-deleteのスワップ先はリクエスト元自身になるため、指示しなければメンテナンスの
		// 文書が押したボタンの中へ挿入される。メンテナンス中はどのパスもこのページを返すため、
		// リロードすれば全画面で表示され、他に送り先は要らない (それぞれ専用のルートへ遷移する
		// httperrorのページとはこの点が違う)。
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Refresh", "true")
		}

		// メンテナンスページを返す (503 Service Unavailable)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Retry-After", "3600") // 1時間後にリトライを推奨
		w.WriteHeader(http.StatusServiceUnavailable)

		// メンテナンスページテンプレートをレンダリング
		component := maintenance.Page()
		_ = component.Render(r.Context(), w)
	})
}

// isAdminIPはリクエスト元IPが管理者IPかどうかをチェック
func (m *MaintenanceMiddleware) isAdminIP(r *http.Request) bool {
	// 管理者IPが設定されていない場合は常にfalse
	if len(m.cfg.AdminIPs) == 0 {
		return false
	}

	clientIP := clientip.GetClientIP(r)

	// 管理者IPリストに含まれているかチェック
	for _, adminIP := range m.cfg.AdminIPs {
		if clientIP == adminIP {
			return true
		}
	}

	return false
}
