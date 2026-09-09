package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
)

// wantSecurityHeaders spells out the headers Rails sends through `config.load_defaults 7.1`
// instead of reading them back from securityHeaders, so that the test states what a response
// has to look like rather than restating the implementation.
//
// [Ja] wantSecurityHeaders は Rails が `config.load_defaults 7.1` で送るヘッダーを、
// securityHeaders から読み出さずに書き下している。実装を言い換えるのではなく、レスポンスが
// どうあるべきかをテスト側で述べるため。
var wantSecurityHeaders = map[string]string{
	"X-Frame-Options":                   "SAMEORIGIN",
	"X-XSS-Protection":                  "0",
	"X-Content-Type-Options":            "nosniff",
	"X-Permitted-Cross-Domain-Policies": "none",
	"Referrer-Policy":                   "strict-origin-when-cross-origin",
}

// assertSecurityHeaders asserts that the response carries each header exactly once with the
// expected value. It checks the number of values too: a doubled header ("nosniff, nosniff")
// reads as present to Header.Get but is not what the reader receives.
//
// [Ja] assertSecurityHeaders は各ヘッダーが期待する値でちょうど 1 つずつ載っていることを
// 検証する。値の個数も見るのは、二重になったヘッダー ("nosniff, nosniff") が Header.Get では
// 付いているように見えるものの、読み手が受け取る値としては異なるため。
func assertSecurityHeaders(t *testing.T, header http.Header) {
	t.Helper()

	for name, want := range wantSecurityHeaders {
		values := header.Values(name)
		if len(values) != 1 {
			t.Errorf("%s の値の個数 = %d (%v), want 1", name, len(values), values)
			continue
		}
		if values[0] != want {
			t.Errorf("%s = %q, want %q", name, values[0], want)
		}
	}
}

func TestSecurityHeaders_SetsRailsDefaults(t *testing.T) {
	t.Parallel()

	handler := SecurityHeaders(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
	assertSecurityHeaders(t, rr.Header())
}

// TestSecurityHeaders_CoversErrorResponses covers the responses that do not come from a route:
// the status chi returns for an unregistered path, and what a handler writes when it fails.
//
// [Ja] TestSecurityHeaders_CoversErrorResponses はルート由来ではないレスポンス、すなわち
// 未登録のパスに対して chi が返す応答と、ハンドラーが失敗時に書き出す応答を対象にする。
func TestSecurityHeaders_CoversErrorResponses(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	r.Use(SecurityHeaders)
	r.Get("/boom", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "ハンドラーが返す 500", path: "/boom", wantStatus: http.StatusInternalServerError},
		{name: "chi が返す 404", path: "/unregistered", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d, want %d", rr.Code, tt.wantStatus)
			}
			assertSecurityHeaders(t, rr.Header())
		})
	}
}

// TestSecurityHeaders_StaticFilesKeepTypesNosniffAccepts checks the static file delivery, which
// nosniff constrains: with sniffing disabled, a stylesheet or a script whose declared type is
// not the one for its kind stops loading. http.FileServer derives the type from the extension,
// so the invariant is that the extensions this app ships resolve to those types.
//
// [Ja] TestSecurityHeaders_StaticFilesKeepTypesNosniffAccepts は nosniff が制約する静的
// ファイルの配信を検証する。sniffing を止めた状態では、種別に対応しない型を宣言した
// スタイルシートやスクリプトは読み込まれなくなる。http.FileServer は型を拡張子から決めるため、
// 本アプリケーションが配信する拡張子がその型に解決されることが不変条件になる。
func TestSecurityHeaders_StaticFilesKeepTypesNosniffAccepts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "style.css"), []byte(".a{color:red}"), 0o600); err != nil {
		t.Fatalf("テスト用 CSS の作成に失敗: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.js"), []byte("export const a = 1"), 0o600); err != nil {
		t.Fatalf("テスト用 JS の作成に失敗: %v", err)
	}

	r := chi.NewRouter()
	r.Use(SecurityHeaders)
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir(dir))))

	tests := []struct {
		name            string
		path            string
		wantContentType string
	}{
		{name: "スタイルシート", path: "/static/style.css", wantContentType: "text/css; charset=utf-8"},
		{name: "スクリプト", path: "/static/main.js", wantContentType: "text/javascript; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコード = %d, want %d", rr.Code, http.StatusOK)
			}
			if contentType := rr.Header().Get("Content-Type"); contentType != tt.wantContentType {
				t.Errorf("Content-Type = %q, want %q", contentType, tt.wantContentType)
			}
			assertSecurityHeaders(t, rr.Header())
		})
	}
}

// TestSecurityHeaders_ProxiedResponseKeepsRailsHeaders builds the chain in the order serve.go
// registers it (the reverse proxy on the outside, SecurityHeaders inside) and checks both
// branches: a Rails-bound request receives Rails' own headers once, and a Go-bound request
// receives the ones this middleware sets.
//
// [Ja] TestSecurityHeaders_ProxiedResponseKeepsRailsHeaders は serve.go の登録順 (外側が
// リバースプロキシ、内側が SecurityHeaders) でチェーンを組み、両方の分岐を検証する。
// Rails 行きのリクエストは Rails 自身のヘッダーを 1 つずつ受け取り、Go 行きのリクエストは
// 本ミドルウェアが設定したヘッダーを受け取る。
func TestSecurityHeaders_ProxiedResponseKeepsRailsHeaders(t *testing.T) {
	t.Parallel()

	railsServer := newRailsServerWithSecurityHeaders(t)
	defer railsServer.Close()

	cfg := &config.Config{Domain: "annict-test.page"}

	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	handler := proxyMiddleware.Middleware(SecurityHeaders(testHandler()))

	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{name: "Rails 版へプロキシするパス", path: "/works", wantBody: "Rails response"},
		{name: "Go 版で処理するパス", path: "/health", wantBody: "OK"},
	}

	for _, tt := range tests {
		// The subtests stay sequential: the mock Rails server is closed when this function
		// returns, which is before parallel subtests would run.
		//
		// [Ja] サブテストは並行にしない。モックの Rails 版サーバーは本関数を抜けた時点で
		// 閉じられ、それは並行サブテストが走るより前になるため。
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコード = %d, want %d", rr.Code, http.StatusOK)
			}
			if rr.Body.String() != tt.wantBody {
				t.Errorf("レスポンスボディ = %q, want %q", rr.Body.String(), tt.wantBody)
			}
			assertSecurityHeaders(t, rr.Header())
		})
	}
}

// TestSecurityHeaders_RegisteredOutsideProxyDoublesRailsHeaders pins the behaviour the
// registration position rests on: with the middleware on the outside of the reverse proxy, the
// headers it sets and the ones Rails sends both reach the reader, because httputil.ReverseProxy
// appends the upstream's headers to the ones already on the ResponseWriter. It is not the
// configuration serve.go uses; it records why that configuration is not available.
//
// [Ja] TestSecurityHeaders_RegisteredOutsideProxyDoublesRailsHeaders は、登録位置の根拠に
// なっている挙動を固定する。本ミドルウェアをリバースプロキシの外側に置くと、本ミドルウェアが
// 設定したヘッダーと Rails が送ったヘッダーの両方が読み手に届く。httputil.ReverseProxy が
// 上流のヘッダーを ResponseWriter が既に持つ値へ追記するためである。serve.go が採る構成では
// なく、その構成を採れない理由を記録するテストである。
func TestSecurityHeaders_RegisteredOutsideProxyDoublesRailsHeaders(t *testing.T) {
	t.Parallel()

	railsServer := newRailsServerWithSecurityHeaders(t)
	defer railsServer.Close()

	cfg := &config.Config{Domain: "annict-test.page"}

	proxyMiddleware, err := NewReverseProxyMiddleware(railsServer.URL, cfg, nil, nil)
	if err != nil {
		t.Fatalf("ミドルウェアの作成に失敗: %v", err)
	}

	handler := SecurityHeaders(proxyMiddleware.Middleware(testHandler()))

	req := httptest.NewRequest(http.MethodGet, "/works", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Body.String() != "Rails response" {
		t.Fatalf("レスポンスボディ = %q, want %q", rr.Body.String(), "Rails response")
	}

	for name := range wantSecurityHeaders {
		if values := rr.Header().Values(name); len(values) != 2 {
			t.Errorf("%s の値の個数 = %d (%v), want 2 (二重付与)", name, len(values), values)
		}
	}
}

// newRailsServerWithSecurityHeaders returns a mock Rails app that sends the same header set the
// real one sends. Sending them is what makes a second layer of the same headers observable.
//
// [Ja] newRailsServerWithSecurityHeaders は本物と同じヘッダー集合を送るモックの Rails 版を
// 返す。同じヘッダーが二重になることを観測できるのは、これを送っているためである。
func newRailsServerWithSecurityHeaders(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, value := range wantSecurityHeaders {
			w.Header().Set(name, value)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Rails response"))
	}))
}
