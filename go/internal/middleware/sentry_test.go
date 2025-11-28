package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/query"
	"github.com/getsentry/sentry-go"
)

func TestSentryUserContextMiddleware_WithAuthenticatedUser(t *testing.T) {
	// Sentryを初期化（テスト用にDSNは空にする）
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "",
	})
	if err != nil {
		t.Fatalf("Sentryの初期化に失敗しました: %v", err)
	}
	defer sentry.Flush(0)

	// ミドルウェアを作成
	sentryMW := middleware.NewSentryUserContextMiddleware()

	// Sentryのユーザー情報を検証するためのハンドラー
	var capturedUserID string
	var capturedUsername string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hubからユーザー情報を取得して検証
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			scope := hub.Scope()
			// 直接scopeからユーザー情報を取得することはできないが、
			// ミドルウェアが正しく実行されていることは確認できる
			_ = scope // スコープは存在する
		}
		w.WriteHeader(http.StatusOK)
	})

	// 認証済みユーザー情報をコンテキストに設定
	user := &query.GetUserByIDRow{
		ID:       123,
		Username: "testuser",
	}

	// リクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)

	// SentryのHubをコンテキストに注入（sentryhttp.Handlerと同様の動作をシミュレート）
	hub := sentry.CurrentHub().Clone()
	ctx = sentry.SetHubOnContext(ctx, hub)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Hubのスコープからユーザー情報を確認
	// sentry-goのAPIではスコープから直接ユーザー情報を取得する方法がないため、
	// BeforeSendフックを使って検証する別のアプローチを使用
	_ = capturedUserID
	_ = capturedUsername
}

func TestSentryUserContextMiddleware_WithoutAuthenticatedUser(t *testing.T) {
	// Sentryを初期化（テスト用にDSNは空にする）
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "",
	})
	if err != nil {
		t.Fatalf("Sentryの初期化に失敗しました: %v", err)
	}
	defer sentry.Flush(0)

	// ミドルウェアを作成
	sentryMW := middleware.NewSentryUserContextMiddleware()

	// テスト用のハンドラー
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// リクエストを作成（認証なし）
	req := httptest.NewRequest("GET", "/test", nil)

	// SentryのHubをコンテキストに注入
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(req.Context(), hub)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// ハンドラーが呼び出されたことを確認
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestSentryUserContextMiddleware_WithoutSentryHub(t *testing.T) {
	// ミドルウェアを作成
	sentryMW := middleware.NewSentryUserContextMiddleware()

	// テスト用のハンドラー
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// 認証済みユーザー情報をコンテキストに設定
	user := &query.GetUserByIDRow{
		ID:       456,
		Username: "anotheruser",
	}

	// リクエストを作成（SentryのHubはコンテキストに注入しない）
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用（Hubがなくてもエラーにならないことを確認）
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// ハンドラーが呼び出されたことを確認
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestSentryUserContextMiddleware_SetsCorrectUserInfo(t *testing.T) {
	// Sentryを初期化（BeforeSendフックでユーザー情報を検証）
	var capturedUser sentry.User
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			capturedUser = event.User
			return event
		},
	})
	if err != nil {
		t.Fatalf("Sentryの初期化に失敗しました: %v", err)
	}
	defer sentry.Flush(0)

	// ミドルウェアを作成
	sentryMW := middleware.NewSentryUserContextMiddleware()

	// テスト用のハンドラー（エラーをキャプチャしてユーザー情報を検証）
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hubからエラーをキャプチャ（ユーザー情報が設定されていることを検証するため）
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.CaptureMessage("test message")
		}
		w.WriteHeader(http.StatusOK)
	})

	// 認証済みユーザー情報をコンテキストに設定
	user := &query.GetUserByIDRow{
		ID:       789,
		Username: "verifyuser",
	}

	// リクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)

	// SentryのHubをコンテキストに注入
	hub := sentry.CurrentHub().Clone()
	ctx = sentry.SetHubOnContext(ctx, hub)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ユーザー情報が正しく設定されていることを確認
	if capturedUser.ID != "789" {
		t.Errorf("wrong user ID: got %v want %v", capturedUser.ID, "789")
	}
	if capturedUser.Username != "verifyuser" {
		t.Errorf("wrong username: got %v want %v", capturedUser.Username, "verifyuser")
	}
}
