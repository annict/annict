package sentry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/riverqueue/river/rivertype"
)

// riverCaptureTransport collects events emitted by the Sentry client so tests
// can assert against them without hitting the network.
//
// [Ja] Sentry クライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用 Transport。
type riverCaptureTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (t *riverCaptureTransport) Configure(_ sentry.ClientOptions)        {}
func (t *riverCaptureTransport) Flush(_ time.Duration) bool              { return true }
func (t *riverCaptureTransport) FlushWithContext(_ context.Context) bool { return true }
func (t *riverCaptureTransport) Close()                                  {}
func (t *riverCaptureTransport) SendEventWithContext(_ context.Context, e *sentry.Event) {
	t.SendEvent(e)
}

func (t *riverCaptureTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

func (t *riverCaptureTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]*sentry.Event, len(t.events))
	copy(out, t.events)
	return out
}

// newRiverTestHub builds an isolated Sentry Hub backed by riverCaptureTransport
// so each test sees its own event stream and the global hub stays untouched.
//
// [Ja] テストごとに独立した Hub + riverCaptureTransport を作る。グローバル Hub
// には一切触らない。
func newRiverTestHub(t *testing.T) (*sentry.Hub, *riverCaptureTransport) {
	t.Helper()
	transport := &riverCaptureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient: %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// newJobRow builds a minimal JobRow with the fields the middleware reads.
//
// [Ja] ミドルウェアが参照するフィールドのみを埋めた最小限の JobRow を生成する。
func newJobRow(kind string, attempt int) *rivertype.JobRow {
	return &rivertype.JobRow{Kind: kind, Attempt: attempt}
}

func TestRiverWorkerMiddleware_NoErrorDoesNotSendEvent(t *testing.T) {
	t.Parallel()

	hub, transport := newRiverTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), hub)

	mw := RiverWorkerMiddleware()
	err := mw.Work(ctx, newJobRow("send_sign_in_code_email", 1), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Work() unexpected error: %v", err)
	}

	hub.Flush(2 * time.Second)
	if got := len(transport.Events()); got != 0 {
		t.Errorf("成功ジョブで Sentry イベントが送られてはいけない: got %d events", got)
	}
}

func TestRiverWorkerMiddleware_ErrorCapturesEvent(t *testing.T) {
	t.Parallel()

	hub, transport := newRiverTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), hub)

	jobErr := errors.New("ジョブ失敗")
	mw := RiverWorkerMiddleware()
	err := mw.Work(ctx, newJobRow("send_password_reset_email", 3), func(ctx context.Context) error {
		return jobErr
	})
	if !errors.Is(err, jobErr) {
		t.Fatalf("Work() = %v, want %v (river のリトライ判断のためそのまま伝搬すること)", err, jobErr)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("失敗ジョブは Sentry に 1 件送るべき: got %d events", len(events))
	}
	got := events[0]
	if len(got.Exception) != 1 || got.Exception[0].Value != "ジョブ失敗" {
		t.Errorf("event.Exception = %+v, want exception with value %q", got.Exception, "ジョブ失敗")
	}
	if got.Tags["job.kind"] != "send_password_reset_email" {
		t.Errorf("event.Tags[job.kind] = %q, want %q", got.Tags["job.kind"], "send_password_reset_email")
	}
	if got.Tags["job.attempt"] != "3" {
		t.Errorf("event.Tags[job.attempt] = %q, want %q", got.Tags["job.attempt"], "3")
	}
}

func TestRiverWorkerMiddleware_DropsIgnorableErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "context.Canceledは送らない",
			err:  context.Canceled,
		},
		{
			name: "context.Canceledのラップは送らない",
			err:  fmt.Errorf("ジョブ中断: %w", context.Canceled),
		},
		{
			name: "http.ErrAbortHandlerは送らない",
			err:  http.ErrAbortHandler,
		},
		{
			name: "http.ErrAbortHandlerのラップは送らない",
			err:  fmt.Errorf("ハンドラー中断: %w", http.ErrAbortHandler),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hub, transport := newRiverTestHub(t)
			ctx := sentry.SetHubOnContext(context.Background(), hub)

			mw := RiverWorkerMiddleware()
			err := mw.Work(ctx, newJobRow("cleanup_expired_tokens", 1), func(ctx context.Context) error {
				return tt.err
			})
			if !errors.Is(err, tt.err) {
				t.Fatalf("Work() = %v, want %v (river のリトライ判断のためそのまま伝搬すること)", err, tt.err)
			}

			hub.Flush(2 * time.Second)
			if got := len(transport.Events()); got != 0 {
				t.Errorf("無視対象のエラーは Sentry に送らないこと: got %d events", got)
			}
		})
	}
}

func TestRiverWorkerMiddleware_BindsHubToContext(t *testing.T) {
	t.Parallel()

	// Workers usually surface intermediate errors via slog.ErrorContext which
	// routes through sentryslog → the Hub on ctx. Verify the middleware binds a
	// (cloned) Hub to ctx so an explicit CaptureException inside the inner
	// function reaches the same transport, and that the job tags set by the
	// middleware are present on that capture too.
	//
	// [Ja] ワーカーは多くの場合 slog.ErrorContext で中間エラーを Sentry に流す
	// (sentryslog 経由で ctx の Hub を使う)。本テストでは inner で
	// hub.CaptureException を直接呼び、ミドルウェアが ctx に bind した Hub と
	// 同じ transport にイベントが届くこと・ジョブタグが乗っていることを確認する。
	parentHub, transport := newRiverTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), parentHub)

	mw := RiverWorkerMiddleware()
	err := mw.Work(ctx, newJobRow("send_sign_up_code_email", 2), func(ctx context.Context) error {
		jobHub := sentry.GetHubFromContext(ctx)
		if jobHub == nil {
			t.Fatal("inner ctx に Hub が bind されていない")
		}
		if jobHub == parentHub {
			t.Error("ジョブ単位の Hub は Clone である必要がある (parent と同一インスタンスは NG)")
		}
		jobHub.CaptureException(errors.New("中間エラー"))
		return nil
	})
	if err != nil {
		t.Fatalf("Work() unexpected error: %v", err)
	}

	parentHub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("inner からの CaptureException は 1 件届くべき: got %d events", len(events))
	}
	got := events[0]
	if len(got.Exception) != 1 || got.Exception[0].Value != "中間エラー" {
		t.Errorf("event.Exception = %+v, want exception with value %q", got.Exception, "中間エラー")
	}
	if got.Tags["job.kind"] != "send_sign_up_code_email" {
		t.Errorf("event.Tags[job.kind] = %q, want %q", got.Tags["job.kind"], "send_sign_up_code_email")
	}
	if got.Tags["job.attempt"] != "2" {
		t.Errorf("event.Tags[job.attempt] = %q, want %q", got.Tags["job.attempt"], "2")
	}
}
