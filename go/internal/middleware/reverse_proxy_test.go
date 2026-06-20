package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/clientip"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
)

func TestReverseProxyMiddleware_GoHandledPaths(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：Go版で処理するパス
	testCases := []struct {
		name         string
		path         string
		expectedBody string
	}{
		{"静的ファイル", "/static/css/style.css", "Go response"},
		{"ヘルスチェック", "/health", "Go response"},
		{"Web App Manifest", "/manifest.json", "Go response"},
		{"パスワードログインページ", "/sign_in/password", "Go response"},
		{"パスワードリセット申請", "/password/reset", "Go response"},
		{"パスワードリセット実行", "/password/edit", "Go response"},
		{"パスワード更新", "/password", "Go response"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("レスポンスボディが期待と異なる: got %q want %q", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReverseProxyMiddleware_RailsProxiedPaths(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Forwarded-*ヘッダーが設定されていることを確認
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			t.Errorf("X-Forwarded-Protoが設定されていない: got %q", r.Header.Get("X-Forwarded-Proto"))
		}
		if r.Header.Get("X-Forwarded-Host") != "annict-test.page" {
			t.Errorf("X-Forwarded-Hostが設定されていない: got %q", r.Header.Get("X-Forwarded-Host"))
		}
		// X-Forwarded-ForとX-Real-IPが設定されていることを確認
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Errorf("X-Forwarded-Forが設定されていない")
		}
		if r.Header.Get("X-Real-IP") == "" {
			t.Errorf("X-Real-IPが設定されていない")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：Rails版にプロキシするパス
	testCases := []struct {
		name         string
		path         string
		expectedBody string
	}{
		{"トップページ", "/", "Rails response"},
		{"作品一覧", "/works", "Rails response"},
		{"作品詳細", "/works/1", "Rails response"},
		{"人気アニメページ", "/works/popular", "Rails response"},
		{"ユーザープロフィール", "/@username", "Rails response"},
		{"設定ページ", "/settings", "Rails response"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("レスポンスボディが期待と異なる: got %q want %q", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReverseProxyMiddleware_HeaderForwarding(t *testing.T) {
	// モックRailsサーバーを作成（ヘッダーチェック）
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 各種ヘッダーが転送されていることを確認
		headers := map[string]string{
			"CF-Connecting-IP": "1.2.3.4",
			"Origin":           "https://annict-test.page",
			"Referer":          "https://annict-test.page/previous",
			"Authorization":    "Basic dGVzdDp0ZXN0",
			"Cookie":           "_annict_session=test_session_id",
		}

		for name, expected := range headers {
			actual := r.Header.Get(name)
			if actual != expected {
				t.Errorf("ヘッダー %s が期待と異なる: got %q want %q", name, actual, expected)
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// リクエストを作成（ヘッダーを設定）
	req := httptest.NewRequest("GET", "/works", nil)
	req.Header.Set("CF-Connecting-IP", "1.2.3.4")
	req.Header.Set("Origin", "https://annict-test.page")
	req.Header.Set("Referer", "https://annict-test.page/previous")
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
	req.Header.Set("Cookie", "_annict_session=test_session_id")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestReverseProxyMiddleware_ErrorHandling(t *testing.T) {
	// モックRailsサーバーを作成（常にエラーを返す）
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 接続を即座に閉じる（エラーをシミュレート）
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("Hijackerをサポートしていない")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijackに失敗: %v", err)
		}
		_ = conn.Close()
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// リクエストを作成
	req := httptest.NewRequest("GET", "/works", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// エラーハンドリングにより502 Bad Gatewayが返ることを確認
	if rr.Code != http.StatusBadGateway {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusBadGateway)
	}
}

func TestIsGoHandledPath(t *testing.T) {
	cfg := &config.Config{Domain: "annict-test.page"}
	proxyMiddleware, _ := NewReverseProxyMiddleware("http://localhost:3000", cfg, nil, nil)

	testCases := []struct {
		path     string
		expected bool
	}{
		{"/static/css/style.css", true},
		{"/health", true},
		{"/manifest.json", true},
		{"/sign_in/password", true},
		{"/password/reset", true},
		{"/password/edit", true},
		{"/password", true},
		{"/sign_in", true},
		{"/works/popular", false},
		{"/works", false},
		{"/", false},
		{"/@username", false},
		{"/fragment/@username/tracking_heatmap", true},
		// "@user.name-with_dashes" exercises usernames that contain dots, hyphens, and underscores.
		//
		// [Ja] username にドット・ハイフン・アンダースコアが含まれるケースの検証。
		{"/fragment/@user.name-with_dashes/tracking_heatmap", true},
		{"/fragment/@username/records", false},
		{"/fragment/records", false},
		// Only an exact "/tracking_heatmap" suffix is allowed, so paths like "/tracking_heatmap/extra" must not match.
		//
		// [Ja] "/tracking_heatmap" の末尾完全一致のみを許可し、"/tracking_heatmap/extra" のような誤検知を避ける。
		{"/fragment/@username/tracking_heatmap/extra", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			actual := proxyMiddleware.isGoHandledPath(tc.path)
			if actual != tc.expected {
				t.Errorf("isGoHandledPath(%q) = %v, want %v", tc.path, actual, tc.expected)
			}
		})
	}
}

func TestIsAPISubdomain(t *testing.T) {
	cfg := &config.Config{Domain: "annict-test.page"}
	proxyMiddleware, _ := NewReverseProxyMiddleware("http://localhost:3000", cfg, nil, nil)

	testCases := []struct {
		host     string
		expected bool
	}{
		{"api.annict-test.page", true},
		{"api.annict-test.page:8080", true},
		{"annict-test.page", false},
		{"annict-test.page:8080", false},
		{"www.annict-test.page", false},
		{"API.annict-test.page", true}, // 大文字小文字を区別しない
	}

	for _, tc := range testCases {
		t.Run(tc.host, func(t *testing.T) {
			actual := proxyMiddleware.isAPISubdomain(tc.host)
			if actual != tc.expected {
				t.Errorf("isAPISubdomain(%q) = %v, want %v", tc.host, actual, tc.expected)
			}
		})
	}
}

func TestReverseProxyMiddleware_APISubdomain(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails API response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：APIサブドメインへのリクエストはすべてRails版にプロキシされる
	testCases := []struct {
		name         string
		host         string
		path         string
		expectedBody string
	}{
		{"GraphQL API", "api.annict-test.page", "/graphql", "Rails API response"},
		{"REST API", "api.annict-test.page", "/api/v1/works", "Rails API response"},
		{"OAuth", "api.annict-test.page", "/oauth/authorize", "Rails API response"},
		{"APIサブドメインの静的ファイル", "api.annict-test.page", "/static/css/style.css", "Rails API response"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Host = tc.host
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("レスポンスボディが期待と異なる: got %q want %q", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReverseProxyMiddleware_PreserveExistingHeaders(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Forwarded-Forヘッダーの確認
		// 注: httputil.ReverseProxyの標準動作により、RemoteAddr（192.0.2.1）が追加される
		// 実際の本番環境では、CloudflareがCF-Connecting-IPを設定するため問題ない
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		// 既存の値が含まれていることを確認（順序は保証されない）
		if !strings.Contains(xForwardedFor, "10.0.0.1") {
			t.Errorf("X-Forwarded-Forに10.0.0.1が含まれていない: got %q", xForwardedFor)
		}

		// 既存のX-Real-IPヘッダーがそのまま維持されていることを確認
		xRealIP := r.Header.Get("X-Real-IP")
		if xRealIP != "10.0.0.1" {
			t.Errorf("X-Real-IPが期待と異なる: got %q want %q", xRealIP, "10.0.0.1")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// リクエストを作成（既存のヘッダーを設定）
	req := httptest.NewRequest("GET", "/works", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	req.Header.Set("X-Real-IP", "10.0.0.1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestGetClientIP(t *testing.T) {
	testCases := []struct {
		name             string
		cfConnectingIP   string
		xForwardedFor    string
		remoteAddr       string
		expectedClientIP string
	}{
		{
			name:             "CF-Connecting-IPが優先される",
			cfConnectingIP:   "203.0.113.1",
			xForwardedFor:    "198.51.100.1",
			remoteAddr:       "192.0.2.1:1234",
			expectedClientIP: "203.0.113.1",
		},
		{
			name:             "CF-Connecting-IPがない場合、X-Forwarded-Forの最初のIP",
			cfConnectingIP:   "",
			xForwardedFor:    "198.51.100.1, 203.0.113.1",
			remoteAddr:       "192.0.2.1:1234",
			expectedClientIP: "198.51.100.1",
		},
		{
			name:             "X-Forwarded-Forが単一IPの場合",
			cfConnectingIP:   "",
			xForwardedFor:    "198.51.100.1",
			remoteAddr:       "192.0.2.1:1234",
			expectedClientIP: "198.51.100.1",
		},
		{
			name:             "両方ない場合、RemoteAddr",
			cfConnectingIP:   "",
			xForwardedFor:    "",
			remoteAddr:       "192.0.2.1:1234",
			expectedClientIP: "192.0.2.1",
		},
		{
			name:             "RemoteAddrにポート番号がない場合",
			cfConnectingIP:   "",
			xForwardedFor:    "",
			remoteAddr:       "192.0.2.1",
			expectedClientIP: "192.0.2.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.cfConnectingIP != "" {
				req.Header.Set("CF-Connecting-IP", tc.cfConnectingIP)
			}
			if tc.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}
			req.RemoteAddr = tc.remoteAddr

			actual := clientip.GetClientIP(req)
			if actual != tc.expectedClientIP {
				t.Errorf("clientip.GetClientIP() = %q, want %q", actual, tc.expectedClientIP)
			}
		})
	}
}

func TestReverseProxyMiddleware_CFConnectingIP(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CF-Connecting-IPヘッダーがそのまま転送されていることを確認
		cfIP := r.Header.Get("CF-Connecting-IP")
		if cfIP != "203.0.113.1" {
			t.Errorf("CF-Connecting-IPが期待と異なる: got %q want %q", cfIP, "203.0.113.1")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// リクエストを作成（CF-Connecting-IPヘッダーを設定）
	req := httptest.NewRequest("GET", "/works", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestReverseProxyMiddleware_ResponseHeaderTimeout(t *testing.T) {
	// レスポンスヘッダーの送信を遅延させるモックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 200ms遅延（テスト用のタイムアウトは100msに設定）
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Delayed response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// テスト用に短いタイムアウトを設定（100ms）
	// 注: 本番環境では30秒だが、テストを高速化するために短く設定
	if transport, ok := proxyMiddleware.proxy.Transport.(*http.Transport); ok {
		transport.ResponseHeaderTimeout = 100 * time.Millisecond
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// リクエストを作成
	req := httptest.NewRequest("GET", "/works", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// タイムアウトによりエラーハンドラーが502 Bad Gatewayを返すことを確認
	if rr.Code != http.StatusBadGateway {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusBadGateway)
	}
}

func TestReverseProxyMiddleware_HTTPMethods(t *testing.T) {
	// モックRailsサーバーを作成（HTTPメソッドを確認）
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HTTPメソッドをレスポンスボディに含める
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Method: " + r.Method))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：様々なHTTPメソッドがRails版にプロキシされることを確認
	testCases := []struct {
		method       string
		expectedBody string
	}{
		{"GET", "Method: GET"},
		{"POST", "Method: POST"},
		{"PUT", "Method: PUT"},
		{"PATCH", "Method: PATCH"},
		{"DELETE", "Method: DELETE"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/works", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("レスポンスボディが期待と異なる: got %q want %q", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReverseProxyMiddleware_RequestBodyForwarding(t *testing.T) {
	// モックRailsサーバーを作成（リクエストボディを確認）
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストボディを読み取り
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		defer func() { _ = r.Body.Close() }()

		// レスポンスにリクエストボディをエコーバック
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Received: " + string(body)))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：リクエストボディが正しく転送されることを確認
	testBody := `{"title":"テストアニメ","season_year":2024}`
	req := httptest.NewRequest("POST", "/works", strings.NewReader(testBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
	}

	expectedBody := "Received: " + testBody
	if rr.Body.String() != expectedBody {
		t.Errorf("レスポンスボディが期待と異なる: got %q want %q", rr.Body.String(), expectedBody)
	}
}

func TestReverseProxyMiddleware_MultipleHostnames(t *testing.T) {
	// モックRailsサーバーを作成
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：様々なホスト名でリクエストが処理されることを確認
	testCases := []struct {
		name         string
		host         string
		path         string
		expectedBody string
		description  string
	}{
		{
			name:         "メインドメイン",
			host:         "annict-test.page",
			path:         "/works",
			expectedBody: "Rails response",
			description:  "メインドメインはRails版にプロキシされる",
		},
		{
			name:         "メインドメイン（Go版で処理するパス）",
			host:         "annict-test.page",
			path:         "/sign_in/password",
			expectedBody: "Go response",
			description:  "メインドメインでもGo版で処理するパスはGo版で処理",
		},
		{
			name:         "APIサブドメイン",
			host:         "api.annict-test.page",
			path:         "/graphql",
			expectedBody: "Rails response",
			description:  "APIサブドメインはすべてRails版にプロキシされる",
		},
		{
			name:         "APIサブドメイン（Go版で処理するパスでも）",
			host:         "api.annict-test.page",
			path:         "/sign_in/password",
			expectedBody: "Rails response",
			description:  "APIサブドメインはGo版で処理するパスでもRails版にプロキシ",
		},
		{
			name:         "ポート番号付きメインドメイン",
			host:         "annict-test.page:8080",
			path:         "/works",
			expectedBody: "Rails response",
			description:  "ポート番号付きメインドメインはRails版にプロキシされる",
		},
		{
			name:         "ポート番号付きAPIサブドメイン",
			host:         "api.annict-test.page:8080",
			path:         "/graphql",
			expectedBody: "Rails response",
			description:  "ポート番号付きAPIサブドメインはRails版にプロキシされる",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Host = tc.host
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("%s: ステータスコードが期待と異なる: got %v want %v", tc.description, rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedBody {
				t.Errorf("%s: レスポンスボディが期待と異なる: got %q want %q", tc.description, rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReverseProxyMiddleware_LargeRequestBody(t *testing.T) {
	// モックRailsサーバーを作成（大きなリクエストボディを処理）
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストボディのサイズを確認
		body := make([]byte, r.ContentLength)
		n, _ := r.Body.Read(body)
		defer func() { _ = r.Body.Close() }()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Received bytes: " + string(rune(n))))
	}))
	defer railsServer.Close()

	// テスト用の設定
	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// リバースプロキシミドルウェアを作成
	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	// Go版で処理するハンドラー（ダミー）
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	// ミドルウェアを適用
	handler := proxyMiddleware.Middleware(goHandler)

	// テストケース：大きなリクエストボディが正しく転送されることを確認
	// 10KBのテストデータを作成
	largeBody := strings.Repeat("a", 10240)
	req := httptest.NewRequest("POST", "/works", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが期待と異なる: got %v want %v", rr.Code, http.StatusOK)
	}

	// レスポンスにサイズ情報が含まれていることを確認
	if !strings.Contains(rr.Body.String(), "Received bytes:") {
		t.Errorf("レスポンスが期待と異なる: got %q", rr.Body.String())
	}
}

// mockFeatureFlagChecker はテスト用のフィーチャーフラグチェッカー
type mockFeatureFlagChecker struct {
	enabled bool
	err     error
}

func (m *mockFeatureFlagChecker) IsEnabledByDeviceOrUser(_ context.Context, _ string, _ model.UserID, _ model.FeatureFlagName) (bool, error) {
	return m.enabled, m.err
}

func TestEnsureDeviceToken_GeneratesNewToken(t *testing.T) {
	cfg := &config.Config{
		Domain: "annict-test.page",
		Env:    "production",
	}

	mw := &ReverseProxyMiddleware{cfg: cfg}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	token := mw.ensureDeviceToken(rr, req)

	// トークンが生成されていること
	if token == "" {
		t.Fatal("トークンが空です")
	}

	// Cookieがセットされていること
	cookies := rr.Result().Cookies()
	var deviceCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == DeviceTokenCookieName {
			deviceCookie = c
			break
		}
	}

	if deviceCookie == nil {
		t.Fatal("device_token Cookieがセットされていません")
	}

	if deviceCookie.Value != token {
		t.Errorf("Cookieの値が一致しない: got %q want %q", deviceCookie.Value, token)
	}

	if !deviceCookie.HttpOnly {
		t.Error("HttpOnlyがtrueであるべき")
	}

	if !deviceCookie.Secure {
		t.Error("Secure（本番環境）がtrueであるべき")
	}

	if deviceCookie.SameSite != http.SameSiteLaxMode {
		t.Error("SameSiteがLaxであるべき")
	}

	if deviceCookie.MaxAge != 10*365*24*60*60 {
		t.Errorf("MaxAgeが10年分であるべき: got %d", deviceCookie.MaxAge)
	}
}

func TestEnsureDeviceToken_PreservesExistingToken(t *testing.T) {
	cfg := &config.Config{
		Domain: "annict-test.page",
		Env:    "production",
	}

	mw := &ReverseProxyMiddleware{cfg: cfg}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: DeviceTokenCookieName, Value: "existing-token"})
	rr := httptest.NewRecorder()

	token := mw.ensureDeviceToken(rr, req)

	if token != "existing-token" {
		t.Errorf("既存のトークンが返されるべき: got %q want %q", token, "existing-token")
	}

	// 新しいCookieがセットされていないこと
	cookies := rr.Result().Cookies()
	for _, c := range cookies {
		if c.Name == DeviceTokenCookieName {
			t.Error("既存のトークンがある場合、新しいCookieはセットされないべき")
		}
	}
}

func TestEnsureDeviceToken_DevelopmentNotSecure(t *testing.T) {
	cfg := &config.Config{
		Domain: "annict-test.page",
		Env:    "development",
	}

	mw := &ReverseProxyMiddleware{cfg: cfg}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	mw.ensureDeviceToken(rr, req)

	cookies := rr.Result().Cookies()
	for _, c := range cookies {
		if c.Name == DeviceTokenCookieName {
			if c.Secure {
				t.Error("開発環境ではSecureがfalseであるべき")
			}
			return
		}
	}
	t.Fatal("device_token Cookieがセットされていません")
}

func TestIsFeatureFlagEnabled_NilRepo(t *testing.T) {
	mw := &ReverseProxyMiddleware{
		featureFlagRepo: nil,
	}

	req := httptest.NewRequest("GET", "/some-path", nil)

	if mw.isFeatureFlagEnabled(req, "some-token") {
		t.Error("featureFlagRepoがnilの場合、falseを返すべき")
	}
}

func TestIsFeatureFlagEnabled_NoMatchingPattern(t *testing.T) {
	checker := &mockFeatureFlagChecker{enabled: true}
	mw := &ReverseProxyMiddleware{
		featureFlagRepo: checker,
	}

	// featureFlaggedPatternsが空なのでマッチしない
	req := httptest.NewRequest("GET", "/some-path", nil)

	if mw.isFeatureFlagEnabled(req, "some-token") {
		t.Error("マッチするパターンがない場合、falseを返すべき")
	}
}

func TestIsFeatureFlagEnabled_MatchingPatternEnabled(t *testing.T) {
	checker := &mockFeatureFlagChecker{enabled: true}
	mw := &ReverseProxyMiddleware{
		featureFlagRepo: checker,
	}

	// テスト用にパターンを一時的に追加
	original := featureFlaggedPatterns
	featureFlaggedPatterns = []featureFlaggedPattern{
		{pattern: regexp.MustCompile(`^/test-feature(/.*)?$`), flag: "test_feature"},
	}
	defer func() { featureFlaggedPatterns = original }()

	req := httptest.NewRequest("GET", "/test-feature/page", nil)

	if !mw.isFeatureFlagEnabled(req, "test-device-token") {
		t.Error("フラグが有効な場合、trueを返すべき")
	}
}

func TestIsFeatureFlagEnabled_MatchingPatternDisabled(t *testing.T) {
	checker := &mockFeatureFlagChecker{enabled: false}
	mw := &ReverseProxyMiddleware{
		featureFlagRepo: checker,
	}

	original := featureFlaggedPatterns
	featureFlaggedPatterns = []featureFlaggedPattern{
		{pattern: regexp.MustCompile(`^/test-feature(/.*)?$`), flag: "test_feature"},
	}
	defer func() { featureFlaggedPatterns = original }()

	req := httptest.NewRequest("GET", "/test-feature/page", nil)

	if mw.isFeatureFlagEnabled(req, "test-device-token") {
		t.Error("フラグが無効な場合、falseを返すべき")
	}
}

func TestIsFeatureFlagEnabled_ErrorFallsBackToFalse(t *testing.T) {
	checker := &mockFeatureFlagChecker{enabled: false, err: errors.New("db error")}
	mw := &ReverseProxyMiddleware{
		featureFlagRepo: checker,
	}

	original := featureFlaggedPatterns
	featureFlaggedPatterns = []featureFlaggedPattern{
		{pattern: regexp.MustCompile(`^/test-feature(/.*)?$`), flag: "test_feature"},
	}
	defer func() { featureFlaggedPatterns = original }()

	req := httptest.NewRequest("GET", "/test-feature/page", nil)

	if mw.isFeatureFlagEnabled(req, "test-device-token") {
		t.Error("エラー時はfalseを返すべき（Rails版にフォールバック）")
	}
}

func TestReverseProxyMiddleware_FeatureFlagRouting(t *testing.T) {
	// モックRailsサーバー
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	cfg := &config.Config{
		Domain: "annict-test.page",
		Env:    "production",
	}

	// テスト用にパターンを一時的に追加
	original := featureFlaggedPatterns
	featureFlaggedPatterns = []featureFlaggedPattern{
		{pattern: regexp.MustCompile(`^/test-feature(/.*)?$`), flag: "test_feature"},
	}
	defer func() { featureFlaggedPatterns = original }()

	t.Run("フラグ有効: Go版で処理される", func(t *testing.T) {
		checker := &mockFeatureFlagChecker{enabled: true}
		proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, checker, nil)
		if err != nil {
			t.Fatalf("ミドルウェアの作成に失敗: %v", err)
		}

		goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Go response"))
		})

		handler := proxyMiddleware.Middleware(goHandler)
		req := httptest.NewRequest("GET", "/test-feature/page", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Go response") {
			t.Errorf("Go版で処理されるべき: got %q", rr.Body.String())
		}
	})

	t.Run("フラグ無効: Rails版にプロキシされる", func(t *testing.T) {
		checker := &mockFeatureFlagChecker{enabled: false}
		proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, checker, nil)
		if err != nil {
			t.Fatalf("ミドルウェアの作成に失敗: %v", err)
		}

		goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Go response"))
		})

		handler := proxyMiddleware.Middleware(goHandler)
		req := httptest.NewRequest("GET", "/test-feature/page", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Rails response") {
			t.Errorf("Rails版にプロキシされるべき: got %q", rr.Body.String())
		}
	})

	t.Run("featureFlagRepoがnil: Rails版にプロキシされる", func(t *testing.T) {
		proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
		if err != nil {
			t.Fatalf("ミドルウェアの作成に失敗: %v", err)
		}

		goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Go response"))
		})

		handler := proxyMiddleware.Middleware(goHandler)
		req := httptest.NewRequest("GET", "/test-feature/page", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Rails response") {
			t.Errorf("featureFlagRepoがnilの場合、Rails版にプロキシされるべき: got %q", rr.Body.String())
		}
	})
}

func TestReverseProxyMiddleware_AnnictDBFeatureFlag(t *testing.T) {
	// モックRailsサーバー
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	cfg := &config.Config{
		Domain: "annict-test.page",
	}

	// Go版のハンドラー
	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go response"))
	})

	t.Run("フラグ有効: /db/配下のパスはGo版で処理される", func(t *testing.T) {
		checker := &mockFeatureFlagChecker{enabled: true}
		proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, checker, nil)
		if err != nil {
			t.Fatalf("ミドルウェアの作成に失敗: %v", err)
		}

		handler := proxyMiddleware.Middleware(goHandler)

		paths := []string{
			"/db/works",
			"/db/works/123/edit",
			"/db/works/123/episodes",
			"/db/works/new",
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				req := httptest.NewRequest("GET", path, nil)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if !strings.Contains(rr.Body.String(), "Go response") {
					t.Errorf("フラグ有効時、%s はGo版で処理されるべき: got %q", path, rr.Body.String())
				}
			})
		}
	})

	t.Run("フラグ無効: /db/配下のパスはRails版にプロキシされる", func(t *testing.T) {
		checker := &mockFeatureFlagChecker{enabled: false}
		proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, checker, nil)
		if err != nil {
			t.Fatalf("ミドルウェアの作成に失敗: %v", err)
		}

		handler := proxyMiddleware.Middleware(goHandler)

		paths := []string{
			"/db/works",
			"/db/works/123/edit",
			"/db/works/123/episodes",
			"/db/works/new",
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				req := httptest.NewRequest("GET", path, nil)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if !strings.Contains(rr.Body.String(), "Rails response") {
					t.Errorf("フラグ無効時、%s はRails版にプロキシされるべき: got %q", path, rr.Body.String())
				}
			})
		}
	})
}

func TestReverseProxyMiddleware_DeviceTokenCookieSetOnRequest(t *testing.T) {
	railsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
	defer railsServer.Close()

	cfg := &config.Config{
		Domain: "annict-test.page",
		Env:    "production",
	}

	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	goHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Go response"))
	})

	handler := proxyMiddleware.Middleware(goHandler)

	// device_tokenがない状態でリクエスト
	req := httptest.NewRequest("GET", "/works", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// device_token Cookieがセットされていること
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == DeviceTokenCookieName {
			found = true
			if c.Value == "" {
				t.Error("device_tokenの値が空であるべきではない")
			}
			break
		}
	}
	if !found {
		t.Error("device_token Cookieがセットされるべき")
	}
}
