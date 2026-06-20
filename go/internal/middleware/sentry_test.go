package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
)

// errorEventType is the zero value of sentry.Event.Type, which marks error
// events (as opposed to "transaction" events).
//
// [Ja] sentry.Event.Type のゼロ値。error イベント ("transaction" 以外) を表す。
const errorEventType = ""

// captureTransport collects every event the Sentry client would otherwise ship
// over the network, so tests can assert against them in-process. It is safe for
// concurrent use because the Sentry hub may call SendEvent from multiple
// goroutines during transaction.Finish + recoverWithSentry interleaving.
//
// [Ja] Sentry クライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用 Transport。transaction.Finish と recoverWithSentry が別ゴルーチンから
// SendEvent を呼ぶ可能性があるため排他制御で守る。
type captureTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (t *captureTransport) Configure(_ sentry.ClientOptions)        {}
func (t *captureTransport) Flush(_ time.Duration) bool              { return true }
func (t *captureTransport) FlushWithContext(_ context.Context) bool { return true }
func (t *captureTransport) Close()                                  {}
func (t *captureTransport) SendEventWithContext(_ context.Context, e *sentry.Event) {
	t.SendEvent(e)
}

func (t *captureTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

func (t *captureTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]*sentry.Event, len(t.events))
	copy(out, t.events)
	return out
}

// newTestHub builds a per-test Sentry Hub backed by a captureTransport so each
// test sees its own isolated event stream and we never touch the global hub.
//
// [Ja] テストごとに独立した Hub + captureTransport を作る。グローバル Hub には
// 一切触らない。
func newTestHub(t *testing.T) (*sentry.Hub, *captureTransport) {
	t.Helper()
	transport := &captureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              "https://public@example.com/1",
		Transport:        transport,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient: %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// attachHub returns chi middleware that pins the given Hub on the request
// context so the downstream sentryhttp middleware adopts it instead of cloning
// the global hub.
//
// [Ja] sentryhttp がグローバル Hub を clone するのを避けるため、テスト用 Hub を
// リクエスト context に積むミドルウェアを返す。
func attachHub(hub *sentry.Hub) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := sentry.SetHubOnContext(r.Context(), hub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// buildRouter constructs a router whose middleware chain mirrors the production
// setup in cmd/server/main.go around Sentry (Recoverer → sentryhttp →
// SentryTransaction) plus an attachHub shim used only by tests.
//
// [Ja] 本番の cmd/server/main.go と同じ Sentry 周りのチェーン
// (Recoverer → sentryhttp → SentryTransaction) を組んだルーターを作る。
// テスト用の Hub 差し込み (attachHub) のみ追加で噛ませる。
func buildRouter(hub *sentry.Hub, register func(chi.Router)) *chi.Mux {
	sentryHTTP := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(attachHub(hub))
	r.Use(sentryHTTP.Handle)
	r.Use(middleware.SentryTransaction)
	register(r)
	return r
}

// findEvents filters events by type ("" matches error events).
//
// [Ja] イベントを種別で絞り込む ("" は error イベントにマッチ)。
func findEvents(events []*sentry.Event, eventType string) []*sentry.Event {
	var out []*sentry.Event
	for _, e := range events {
		if eventType == errorEventType && e.Type != "transaction" {
			out = append(out, e)
			continue
		}
		if e.Type == eventType {
			out = append(out, e)
		}
	}
	return out
}

func TestSentryTransaction_PanicEventCarriesRoutePattern(t *testing.T) {
	t.Parallel()

	hub, transport := newTestHub(t)

	router := buildRouter(hub, func(r chi.Router) {
		r.Get("/works/{work_id}", func(_ http.ResponseWriter, _ *http.Request) {
			panic(errors.New("boom"))
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/works/123", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// chi's Recoverer must absorb the re-panicked error and return 500.
	//
	// [Ja] chi の Recoverer が再 panic を握り潰し、500 を返す経路を確認する。
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()

	errEvents := findEvents(events, errorEventType)
	if len(errEvents) != 1 {
		t.Fatalf("expected 1 error event, got %d (events=%+v)", len(errEvents), events)
	}
	if got, want := errEvents[0].Transaction, "GET /works/{work_id}"; got != want {
		t.Errorf("error event Transaction = %q, want %q", got, want)
	}

	txEvents := findEvents(events, "transaction")
	if len(txEvents) != 1 {
		t.Fatalf("expected 1 transaction event, got %d", len(txEvents))
	}
	if got, want := txEvents[0].Transaction, "GET /works/{work_id}"; got != want {
		t.Errorf("transaction event Transaction = %q, want %q", got, want)
	}
	if got := txEvents[0].TransactionInfo; got == nil || got.Source != sentry.SourceRoute {
		t.Errorf("transaction event TransactionInfo.Source = %+v, want %q", got, sentry.SourceRoute)
	}
}

func TestSentryTransaction_CapturedErrorCarriesRoutePattern(t *testing.T) {
	t.Parallel()

	hub, transport := newTestHub(t)

	router := buildRouter(hub, func(r chi.Router) {
		r.Get("/@{username}/ics", func(w http.ResponseWriter, req *http.Request) {
			ctxHub := sentry.GetHubFromContext(req.Context())
			if ctxHub == nil {
				t.Error("hub missing from request context")
				return
			}
			ctxHub.CaptureException(errors.New("calendar lookup failed"))
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/@alice/ics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()

	errEvents := findEvents(events, errorEventType)
	if len(errEvents) != 1 {
		t.Fatalf("expected 1 error event, got %d", len(errEvents))
	}

	// The handler captures the exception while it is still running, i.e.
	// before the defer in SentryTransaction has updated the span. The route
	// pattern still ends up on the event because SentryTransaction installs
	// an EventProcessor that reads chi.RouteContext().RoutePattern() at
	// capture time -- and chi has already populated that pattern by the time
	// the handler runs.
	//
	// [Ja] ハンドラー実行中の CaptureException は本ミドルウェアの defer より
	// 先に走るが、SentryTransaction が仕込んだ EventProcessor が
	// chi.RouteContext().RoutePattern() をキャプチャ時に読むため Transaction
	// が乗る。chi はハンドラー実行時点で既にルートパターンを確定させている。
	if got, want := errEvents[0].Transaction, "GET /@{username}/ics"; got != want {
		t.Errorf("error event Transaction = %q, want %q", got, want)
	}
}

func TestSentryTransaction_NoChiContext_NoOp(t *testing.T) {
	t.Parallel()

	// Direct invocation without chi guarantees there is no route context to
	// read. The middleware must silently no-op rather than panic, so calls
	// from static-file or non-chi paths remain safe.
	//
	// [Ja] chi を介さず直接呼び出すと RouteContext が無い状態になる。本
	// ミドルウェアはそのまま no-op で通すこと (静的ファイル等で安全に動く)。
	hub, transport := newTestHub(t)

	called := false
	handler := middleware.SentryTransaction(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	req = req.WithContext(sentry.SetHubOnContext(req.Context(), hub))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("downstream handler was not invoked")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	hub.Flush(100 * time.Millisecond)
	if len(transport.Events()) != 0 {
		t.Errorf("did not expect any events, got %d", len(transport.Events()))
	}
}

func TestSentryTransaction_UnmatchedRoute_NoOp(t *testing.T) {
	t.Parallel()

	// When chi cannot match the URL, RoutePattern returns "" and there is no
	// pattern to record. The middleware must still let the downstream 404
	// handler run instead of overwriting Transaction with a misleading value.
	//
	// [Ja] chi がマッチできなかった場合 RoutePattern は "" になる。本ミドル
	// ウェアは何も上書きせず、404 ハンドラーをそのまま走らせる。
	hub, _ := newTestHub(t)

	router := buildRouter(hub, func(r chi.Router) {
		r.Get("/known", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/does/not/exist", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

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
	user := &model.User{
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
	user := &model.User{
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
	user := &model.User{
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
