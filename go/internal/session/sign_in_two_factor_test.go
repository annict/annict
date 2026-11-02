package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
)

// newRequestWithCookiesは、wが設定したCookieを載せた新しいリクエストを作る。
func newRequestWithCookies(w *httptest.ResponseRecorder) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}
	return req
}

// TestSignInTwoFactorは、ログイン途中の状態を書くとメールアドレスの入力からの一時的な値が消え、
// 期限までだけユーザーIDを読めて、消したあとは読めないことを検証する。
func TestSignInTwoFactor(t *testing.T) {
	_, _, sessionRepo := setupTestDB(t)
	manager := NewManager(sessionRepo, &config.Config{CookieDomain: ".test.example.com", SessionSecure: "false", SessionHTTPOnly: "true"})
	ctx := context.Background()
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	// メールアドレスの入力で書かれる一時的な値を用意する。
	w := httptest.NewRecorder()
	if err := manager.SetValue(ctx, w, httptest.NewRequest(http.MethodGet, "/test", nil), "sign_in_email", "user@example.com"); err != nil {
		t.Fatalf("SetValue()のエラー = %v", err)
	}
	req := newRequestWithCookies(w)
	if err := manager.SetValue(ctx, httptest.NewRecorder(), req, "sign_in_user_id", "42"); err != nil {
		t.Fatalf("SetValue()のエラー = %v", err)
	}

	if _, ok, err := manager.SignInTwoFactorUserID(ctx, req, now); err != nil || ok {
		t.Errorf("始める前のSignInTwoFactorUserID() = (ok:%t, %v)、期待値 = (false, nil)", ok, err)
	}

	if err := manager.StartSignInTwoFactor(ctx, httptest.NewRecorder(), req, model.UserID(42), now); err != nil {
		t.Fatalf("StartSignInTwoFactor()のエラー = %v", err)
	}
	for _, key := range []string{"sign_in_email", "sign_in_user_id"} {
		if got, err := manager.GetValue(ctx, req, key); err != nil || got != "" {
			t.Errorf("%sのGetValue() = (%q, %v)、期待値 = (\"\", nil)", key, got, err)
		}
	}

	tests := []struct {
		name   string
		at     time.Time
		wantOK bool
	}{
		{name: "直後", at: now, wantOK: true},
		{name: "期限の1秒前", at: now.Add(SignInTwoFactorTTL - time.Second), wantOK: true},
		{name: "期限ちょうど", at: now.Add(SignInTwoFactorTTL), wantOK: false},
	}
	for _, tt := range tests {
		userID, ok, err := manager.SignInTwoFactorUserID(ctx, req, tt.at)
		if err != nil {
			t.Fatalf("%s: SignInTwoFactorUserID()のエラー = %v", tt.name, err)
		}
		if ok != tt.wantOK || (ok && userID != 42) {
			t.Errorf("%s: SignInTwoFactorUserID() = (%d, %t)、期待値 = (42, %t)", tt.name, userID, ok, tt.wantOK)
		}
	}

	if err := manager.ClearSignInTwoFactor(ctx, req); err != nil {
		t.Fatalf("ClearSignInTwoFactor()のエラー = %v", err)
	}
	if _, ok, err := manager.SignInTwoFactorUserID(ctx, req, now); err != nil || ok {
		t.Errorf("消したあとのSignInTwoFactorUserID() = (ok:%t, %v)、期待値 = (false, nil)", ok, err)
	}
}

// TestSignInTwoFactorUserID_Brokenは、形の崩れた値をログイン途中の状態として扱わないことを検証する。
func TestSignInTwoFactorUserID_Broken(t *testing.T) {
	_, _, sessionRepo := setupTestDB(t)
	manager := NewManager(sessionRepo, &config.Config{CookieDomain: ".test.example.com", SessionSecure: "false", SessionHTTPOnly: "true"})
	ctx := context.Background()

	for _, value := range []string{"42", "abc:1", "0:9999999999", "42:abc"} {
		w := httptest.NewRecorder()
		if err := manager.SetValue(ctx, w, httptest.NewRequest(http.MethodGet, "/test", nil), signInTwoFactorKey, value); err != nil {
			t.Fatalf("SetValue()のエラー = %v", err)
		}
		if _, ok, err := manager.SignInTwoFactorUserID(ctx, newRequestWithCookies(w), time.Unix(0, 0)); err != nil || ok {
			t.Errorf("値 %qのSignInTwoFactorUserID() = (ok:%t, %v)、期待値 = (false, nil)", value, ok, err)
		}
	}
}
