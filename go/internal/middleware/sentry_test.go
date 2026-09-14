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

// sentry.Event.Typeのゼロ値。errorイベント ("transaction" 以外) を表す。
const errorEventType = ""

// Sentryクライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用Transport。transaction.FinishとrecoverWithSentryが別ゴルーチンから
// SendEventを呼ぶ可能性があるため排他制御で守る。
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

// テストごとに独立したHub + captureTransportを作る。グローバルHubには
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
		t.Fatalf("sentry.NewClient()のエラー = %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// sentryhttpがグローバルHubをcloneするのを避けるため、テスト用Hubを
// リクエストcontextに積むミドルウェアを返す。
func attachHub(hub *sentry.Hub) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := sentry.SetHubOnContext(r.Context(), hub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// 本番のcmd/annict/serve.goと同じSentry周りのチェーン
// (Recoverer → sentryhttp → SentryTransaction) を組んだルーターを作る。
// テスト用のHub差し込み (attachHub) のみ追加で噛ませる。
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

// イベントを種別で絞り込む ("" はerrorイベントにマッチ)。
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

	// chiのRecovererが再panicを握り潰し、500を返す経路を確認する。
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusInternalServerError)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()

	errEvents := findEvents(events, errorEventType)
	if len(errEvents) != 1 {
		t.Fatalf("エラーイベントの件数 = %d、期待値 = 1 (events=%+v)", len(errEvents), events)
	}
	if got, want := errEvents[0].Transaction, "GET /works/{work_id}"; got != want {
		t.Errorf("エラーイベントのTransaction = %q、期待値 = %q", got, want)
	}

	txEvents := findEvents(events, "transaction")
	if len(txEvents) != 1 {
		t.Fatalf("トランザクションイベントの件数 = %d、期待値 = 1", len(txEvents))
	}
	if got, want := txEvents[0].Transaction, "GET /works/{work_id}"; got != want {
		t.Errorf("トランザクションイベントのTransaction = %q、期待値 = %q", got, want)
	}
	if got := txEvents[0].TransactionInfo; got == nil || got.Source != sentry.SourceRoute {
		t.Errorf("トランザクションイベントのTransactionInfo.Source = %+v、期待値 = %q", got, sentry.SourceRoute)
	}
}

func TestSentryTransaction_CapturedErrorCarriesRoutePattern(t *testing.T) {
	t.Parallel()

	hub, transport := newTestHub(t)

	router := buildRouter(hub, func(r chi.Router) {
		r.Get("/@{username}/ics", func(w http.ResponseWriter, req *http.Request) {
			ctxHub := sentry.GetHubFromContext(req.Context())
			if ctxHub == nil {
				t.Error("requestのcontextにhubが無い")
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
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()

	errEvents := findEvents(events, errorEventType)
	if len(errEvents) != 1 {
		t.Fatalf("エラーイベントの件数 = %d、期待値 = 1", len(errEvents))
	}

	// ハンドラー実行中のCaptureExceptionは本ミドルウェアのdeferより
	// 先に走るが、SentryTransactionが仕込んだEventProcessorが
	// chi.RouteContext().RoutePattern() をキャプチャ時に読むためTransaction
	// が乗る。chiはハンドラー実行時点で既にルートパターンを確定させている。
	if got, want := errEvents[0].Transaction, "GET /@{username}/ics"; got != want {
		t.Errorf("エラーイベントのTransaction = %q、期待値 = %q", got, want)
	}
}

func TestSentryTransaction_NoChiContext_NoOp(t *testing.T) {
	t.Parallel()

	// chiを介さず直接呼び出すとRouteContextが無い状態になる。本
	// ミドルウェアはそのままno-opで通すこと (静的ファイル等で安全に動く)。
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
		t.Fatal("後続のハンドラーが呼ばれなかった")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}

	hub.Flush(100 * time.Millisecond)
	if len(transport.Events()) != 0 {
		t.Errorf("イベントの件数 = %d、期待値 = 0", len(transport.Events()))
	}
}

func TestSentryTransaction_UnmatchedRoute_NoOp(t *testing.T) {
	t.Parallel()

	// chiがマッチできなかった場合RoutePatternは "" になる。本ミドル
	// ウェアは何も上書きせず、404ハンドラーをそのまま走らせる。
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
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
	}
}

func TestSentryUserContextMiddleware_WithAuthenticatedUser(t *testing.T) {
	// Sentryを初期化 (テスト用にDSNは空にする)
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

	// SentryのHubをコンテキストに注入 (sentryhttp.Handlerと同様の動作をシミュレート)
	hub := sentry.CurrentHub().Clone()
	ctx = sentry.SetHubOnContext(ctx, hub)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	// Hubのスコープからユーザー情報を確認
	// sentry-goのAPIではスコープから直接ユーザー情報を取得する方法がないため、
	// BeforeSendフックを使って検証する別のアプローチを使用
	_ = capturedUserID
	_ = capturedUsername
}

func TestSentryUserContextMiddleware_WithoutAuthenticatedUser(t *testing.T) {
	// Sentryを初期化 (テスト用にDSNは空にする)
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

	// リクエストを作成 (認証なし)
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
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	// ハンドラーが呼び出されたことを確認
	if !handlerCalled {
		t.Error("ハンドラーが呼ばれなかった")
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

	// リクエストを作成 (SentryのHubはコンテキストに注入しない)
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ミドルウェアを適用 (Hubがなくてもエラーにならないことを確認)
	sentryMW.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードが200であることを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	// ハンドラーが呼び出されたことを確認
	if !handlerCalled {
		t.Error("ハンドラーが呼ばれなかった")
	}
}

func TestSentryUserContextMiddleware_SetsCorrectUserInfo(t *testing.T) {
	// Sentryを初期化 (BeforeSendフックでユーザー情報を検証)
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

	// テスト用のハンドラー (エラーをキャプチャしてユーザー情報を検証)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hubからエラーをキャプチャ (ユーザー情報が設定されていることを検証するため)
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
		t.Errorf("ユーザーID = %v、期待値 = %v", capturedUser.ID, "789")
	}
	if capturedUser.Username != "verifyuser" {
		t.Errorf("ユーザー名 = %v、期待値 = %v", capturedUser.Username, "verifyuser")
	}
}
