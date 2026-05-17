package session

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// FlashCookieName はフラッシュメッセージを格納するCookieのキー名
const FlashCookieName = "annict_flash"

// FlashManager はCookieベースでフラッシュメッセージを管理する構造体
type FlashManager struct {
	cookieDomain  string
	sessionSecure bool
}

// NewFlashManager は FlashManager を生成する
func NewFlashManager(cookieDomain string, sessionSecure bool) *FlashManager {
	return &FlashManager{
		cookieDomain:  cookieDomain,
		sessionSecure: sessionSecure,
	}
}

// SetSuccess は成功メッセージを設定する
func (fm *FlashManager) SetSuccess(w http.ResponseWriter, message string) {
	fm.setFlash(w, FlashSuccess, message)
}

// SetError はエラーメッセージを設定する
func (fm *FlashManager) SetError(w http.ResponseWriter, message string) {
	fm.setFlash(w, FlashError, message)
}

// SetWarning は警告メッセージを設定する
func (fm *FlashManager) SetWarning(w http.ResponseWriter, message string) {
	fm.setFlash(w, FlashWarning, message)
}

// SetInfo は情報メッセージを設定する
func (fm *FlashManager) SetInfo(w http.ResponseWriter, message string) {
	fm.setFlash(w, FlashInfo, message)
}

// setFlash はフラッシュメッセージをCookieに設定する
func (fm *FlashManager) setFlash(w http.ResponseWriter, flashType, message string) {
	flash := Flash{
		Type:    flashType,
		Message: message,
	}
	data, err := json.Marshal(flash)
	if err != nil {
		return
	}

	// CookieのValueにはBase64エンコードして保存
	// JSONの特殊文字（ダブルクォートなど）がCookieで無効な文字として扱われるため
	encoded := base64.StdEncoding.EncodeToString(data)

	cookie := &http.Cookie{
		Name:     FlashCookieName,
		Value:    encoded,
		Path:     "/",
		Domain:   fm.cookieDomain,
		Secure:   fm.sessionSecure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetFlash はフラッシュメッセージを取得し、Cookieから削除する
func (fm *FlashManager) GetFlash(w http.ResponseWriter, r *http.Request) *Flash {
	cookie, err := r.Cookie(FlashCookieName)
	if err != nil {
		return nil
	}

	// Base64デコード
	data, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		fm.clearFlash(w)
		return nil
	}

	var flash Flash
	if err := json.Unmarshal(data, &flash); err != nil {
		fm.clearFlash(w)
		return nil
	}

	// Cookieを削除
	fm.clearFlash(w)

	return &flash
}

type flashContextKey struct{}

// Middleware はリクエストからフラッシュメッセージを読み取り、context に格納するミドルウェアです。
// Cookie からフラッシュを読み取り後、自動的に Cookie を削除します。
func (fm *FlashManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flash := fm.GetFlash(w, r)
		if flash != nil {
			ctx := context.WithValue(r.Context(), flashContextKey{}, flash)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// FlashFromContext は context からフラッシュメッセージを取得します
func FlashFromContext(ctx context.Context) *Flash {
	flash, _ := ctx.Value(flashContextKey{}).(*Flash)
	return flash
}

// clearFlash はフラッシュCookieを削除する
func (fm *FlashManager) clearFlash(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     FlashCookieName,
		Value:    "",
		Path:     "/",
		Domain:   fm.cookieDomain,
		Secure:   fm.sessionSecure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
