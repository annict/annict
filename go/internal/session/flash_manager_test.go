package session

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestFlashManager() *FlashManager {
	return NewFlashManager("", false)
}

func TestFlashManager_SetSuccess(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()
	w := httptest.NewRecorder()

	fm.SetSuccess(w, "操作が成功しました")

	// Cookieが設定されていることを確認
	cookies := w.Result().Cookies()
	var flashCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == FlashCookieName {
			flashCookie = c
			break
		}
	}
	if flashCookie == nil {
		t.Fatal("フラッシュCookieが設定されていません")
	}

	// Base64デコードしてJSONを検証
	data, err := base64.StdEncoding.DecodeString(flashCookie.Value)
	if err != nil {
		t.Fatalf("Base64デコードに失敗: %v", err)
	}

	var flash Flash
	if err := json.Unmarshal(data, &flash); err != nil {
		t.Fatalf("JSONアンマーシャルに失敗: %v", err)
	}

	if flash.Type != FlashSuccess {
		t.Errorf("Type = %q, want %q", flash.Type, FlashSuccess)
	}
	if flash.Message != "操作が成功しました" {
		t.Errorf("Message = %q, want %q", flash.Message, "操作が成功しました")
	}
}

func TestFlashManager_SetError(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()
	w := httptest.NewRecorder()

	fm.SetError(w, "エラーが発生しました")

	cookies := w.Result().Cookies()
	var flashCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == FlashCookieName {
			flashCookie = c
			break
		}
	}
	if flashCookie == nil {
		t.Fatal("フラッシュCookieが設定されていません")
	}

	data, err := base64.StdEncoding.DecodeString(flashCookie.Value)
	if err != nil {
		t.Fatalf("Base64デコードに失敗: %v", err)
	}

	var flash Flash
	if err := json.Unmarshal(data, &flash); err != nil {
		t.Fatalf("JSONアンマーシャルに失敗: %v", err)
	}

	if flash.Type != FlashError {
		t.Errorf("Type = %q, want %q", flash.Type, FlashError)
	}
	if flash.Message != "エラーが発生しました" {
		t.Errorf("Message = %q, want %q", flash.Message, "エラーが発生しました")
	}
}

func TestFlashManager_GetFlash(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()

	// フラッシュメッセージを設定
	w := httptest.NewRecorder()
	fm.SetSuccess(w, "テストメッセージ")

	// 設定されたCookieを使ってリクエストを作成
	cookies := w.Result().Cookies()
	r := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}

	// フラッシュメッセージを取得
	w2 := httptest.NewRecorder()
	flash := fm.GetFlash(w2, r)

	if flash == nil {
		t.Fatal("フラッシュメッセージがnilです")
	}
	if flash.Type != FlashSuccess {
		t.Errorf("Type = %q, want %q", flash.Type, FlashSuccess)
	}
	if flash.Message != "テストメッセージ" {
		t.Errorf("Message = %q, want %q", flash.Message, "テストメッセージ")
	}

	// Cookieが削除されていることを確認（MaxAge=-1のCookieが設定される）
	clearCookies := w2.Result().Cookies()
	var clearCookie *http.Cookie
	for _, c := range clearCookies {
		if c.Name == FlashCookieName {
			clearCookie = c
			break
		}
	}
	if clearCookie == nil {
		t.Fatal("クリア用Cookieが設定されていません")
	}
	if clearCookie.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", clearCookie.MaxAge)
	}
}

func TestFlashManager_GetFlash_NoCookie(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	flash := fm.GetFlash(w, r)

	if flash != nil {
		t.Errorf("Cookieがない場合にnilを返すべき: got %+v", flash)
	}
}

func TestFlashManager_GetFlash_InvalidBase64(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  FlashCookieName,
		Value: "!!!invalid-base64!!!",
	})

	w := httptest.NewRecorder()
	flash := fm.GetFlash(w, r)

	if flash != nil {
		t.Errorf("不正なBase64でnilを返すべき: got %+v", flash)
	}

	// Cookieが削除されていることを確認
	clearCookies := w.Result().Cookies()
	var clearCookie *http.Cookie
	for _, c := range clearCookies {
		if c.Name == FlashCookieName {
			clearCookie = c
			break
		}
	}
	if clearCookie == nil {
		t.Fatal("クリア用Cookieが設定されていません")
	}
	if clearCookie.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", clearCookie.MaxAge)
	}
}

func TestFlashManager_Middleware_PutsFlashIntoContext(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()

	// 事前に Cookie を仕込む
	preW := httptest.NewRecorder()
	fm.SetSuccess(preW, "ok")
	cookies := preW.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("事前 Cookie が設定されていない")
	}

	var captured *Flash
	handler := fm.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = FlashFromContext(r.Context())
	}))

	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if captured == nil {
		t.Fatal("flash が context に積まれていない")
	}
	if captured.Type != FlashSuccess {
		t.Errorf("Type = %q, want %q", captured.Type, FlashSuccess)
	}
	if captured.Message != "ok" {
		t.Errorf("Message = %q, want %q", captured.Message, "ok")
	}

	// 直後に Cookie が削除されていること（MaxAge=-1 の Set-Cookie）
	foundClear := false
	for _, c := range w.Result().Cookies() {
		if c.Name == FlashCookieName && c.MaxAge < 0 {
			foundClear = true
		}
	}
	if !foundClear {
		t.Error("Middleware は Cookie を削除する Set-Cookie を発行すべき")
	}
}

func TestFlashManager_Middleware_NoCookie(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()

	var captured *Flash
	called := false
	handler := fm.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		captured = FlashFromContext(r.Context())
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("next.ServeHTTP が呼ばれていない")
	}
	if captured != nil {
		t.Errorf("Cookie が無い場合 FlashFromContext は nil を返すべき: got %+v", captured)
	}
}

func TestFlashManager_FlashFromContext_NoValue(t *testing.T) {
	t.Parallel()

	flash := FlashFromContext(context.Background())
	if flash != nil {
		t.Errorf("空 context では nil を返すべき: got %+v", flash)
	}
}

func TestFlashManager_GetFlash_InvalidJSON(t *testing.T) {
	t.Parallel()

	fm := newTestFlashManager()

	// 有効なBase64だが無効なJSON
	invalidJSON := base64.StdEncoding.EncodeToString([]byte("not-json"))
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  FlashCookieName,
		Value: invalidJSON,
	})

	w := httptest.NewRecorder()
	flash := fm.GetFlash(w, r)

	if flash != nil {
		t.Errorf("不正なJSONでnilを返すべき: got %+v", flash)
	}

	// Cookieが削除されていることを確認
	clearCookies := w.Result().Cookies()
	var clearCookie *http.Cookie
	for _, c := range clearCookies {
		if c.Name == FlashCookieName {
			clearCookie = c
			break
		}
	}
	if clearCookie == nil {
		t.Fatal("クリア用Cookieが設定されていません")
	}
	if clearCookie.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", clearCookie.MaxAge)
	}
}
