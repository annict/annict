package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates"
)

// DBEditHistoryLinkMiddlewareは、Annict DBの画面のサイドバーに編集履歴の項目を出すかを判定し、コンテキストへ
// 保存する。サイドバーはAnnict DBのすべての画面に出るため、画面ごとの読み取りユースケースではなくここで判定する。
// 判定は、編集履歴の画面をゲートするReverseProxyMiddlewareと同じく、フィーチャーフラグdb_edit_historyが
// 閲覧中のデバイスかユーザーに有効かで行う。Annict DBの画面 (/db配下) 以外のリクエストでは判定しない。
type DBEditHistoryLinkMiddleware struct {
	featureFlagRepo featureFlagChecker
}

// NewDBEditHistoryLinkMiddlewareは新しいDBEditHistoryLinkMiddlewareを作成する。
func NewDBEditHistoryLinkMiddleware(featureFlagRepo featureFlagChecker) *DBEditHistoryLinkMiddleware {
	return &DBEditHistoryLinkMiddleware{featureFlagRepo: featureFlagRepo}
}

// Middlewareは認証ミドルウェアがユーザーをコンテキストに載せた後に置く。
func (m *DBEditHistoryLinkMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/db" && !strings.HasPrefix(r.URL.Path, "/db/") {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		var userID model.UserID
		if user := GetUserFromContext(ctx); user != nil {
			userID = user.ID
		}
		deviceToken := GetDeviceToken(r)
		if deviceToken == "" && userID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		visible, err := m.featureFlagRepo.IsEnabledByDeviceOrUser(ctx, deviceToken, userID, model.FeatureFlagDBEditHistory)
		if err != nil {
			// 項目を出せなくても画面は使えるため、エラーにせず出さない。
			slog.WarnContext(ctx, "編集履歴のフィーチャーフラグの判定に失敗", "error", err)
			visible = false
		}
		next.ServeHTTP(w, r.WithContext(templates.SetDBEditHistoryLinkVisible(ctx, visible)))
	})
}
