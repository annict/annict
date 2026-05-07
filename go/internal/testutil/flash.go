package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/annict/annict/go/internal/session"
)

// NewTestFlashManager はテスト用の FlashManager を生成する。
// Cookie ドメイン無し・非 Secure で固定する。
func NewTestFlashManager() *session.FlashManager {
	return session.NewFlashManager("", false)
}

// ContextWithFlash は Cookie 経由でフラッシュメッセージを context に積んだ context を返す。
// production と同じ FlashManager.Middleware を経由するため、
// テストでも本番に近い経路で flash を扱える。
func ContextWithFlash(ctx context.Context, flashType, message string) context.Context {
	fm := NewTestFlashManager()

	// flash を Cookie に書き込む
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

	// Cookie を載せたリクエストを Middleware に通して context に flash を積む
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
