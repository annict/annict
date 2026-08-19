package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/templates/pages/maintenance"
)

// MaintenanceMiddleware はメンテナンスモード時にアクセスを制限するミドルウェア
type MaintenanceMiddleware struct {
	cfg *config.Config
}

// NewMaintenanceMiddleware は新しいMaintenanceMiddlewareを作成
func NewMaintenanceMiddleware(cfg *config.Config) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		cfg: cfg,
	}
}

// Middleware はHTTPミドルウェアを返す
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

		// htmx swaps every response except 204 and 304, and an hx-delete without hx-target
		// swaps into the element that issued it, so without this the maintenance document
		// would be placed inside the button that was clicked. Every path answers with this
		// page while maintenance is on, so reloading shows it full screen and there is
		// nowhere else to send the reader (unlike the httperror pages, each of which
		// navigates to a route of its own).
		//
		// [Ja] htmx は 204 と 304 以外のレスポンスをスワップし、hx-target を指定していない
		// hx-delete のスワップ先はリクエスト元自身になるため、指示しなければメンテナンスの
		// 文書が押したボタンの中へ挿入される。メンテナンス中はどのパスもこのページを返すため、
		// リロードすれば全画面で表示され、他に送り先は要らない (それぞれ専用のルートへ遷移する
		// httperror のページとはこの点が違う)。
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Refresh", "true")
		}

		// メンテナンスページを返す（503 Service Unavailable）
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Retry-After", "3600") // 1時間後にリトライを推奨
		w.WriteHeader(http.StatusServiceUnavailable)

		// メンテナンスページテンプレートをレンダリング
		component := maintenance.Page()
		_ = component.Render(r.Context(), w)
	})
}

// isAdminIP はリクエスト元IPが管理者IPかどうかをチェック
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
