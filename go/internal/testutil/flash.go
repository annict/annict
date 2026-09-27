package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/annict/annict/go/internal/session"
)

// NewTestFlashManagerはテスト用のFlashManagerを生成する。
// Cookieドメイン無し・非Secureで固定する。
func NewTestFlashManager() *session.FlashManager {
	return session.NewFlashManager("", false)
}

// ContextWithFlashはCookie経由でフラッシュメッセージをcontextに積んだcontextを返す。
// productionと同じFlashManager.Middlewareを経由するため、
// テストでも本番に近い経路でflashを扱える。
func ContextWithFlash(ctx context.Context, flashType, message string) context.Context {
	fm := NewTestFlashManager()

	// flashをCookieに書き込む
	preW := httptest.NewRecorder()
	switch flashType {
	case session.FlashSuccess:
		fm.SetSuccess(preW, message)
	case session.FlashError:
		fm.SetError(preW, message)
	case session.FlashWarning:
		fm.SetWarning(preW, message)
	case session.FlashInfo:
		fm.SetInfo(preW, message)
	}

	// Cookieを載せたリクエストをMiddlewareに通してcontextにflashを積む
	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	for _, c := range preW.Result().Cookies() {
		req.AddCookie(c)
	}

	resultCtx := ctx
	fm.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		resultCtx = r.Context()
	})).ServeHTTP(httptest.NewRecorder(), req)

	return resultCtx
}
