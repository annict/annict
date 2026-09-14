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

// Sentryクライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用Transport。
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

// テストごとに独立したHub + riverCaptureTransportを作る。グローバルHub
// には一切触らない。
func newRiverTestHub(t *testing.T) (*sentry.Hub, *riverCaptureTransport) {
	t.Helper()
	transport := &riverCaptureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient()のエラー = %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// ミドルウェアが参照するフィールドのみを埋めた最小限のJobRowを生成する。
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
		t.Fatalf("Work()の想定外のエラー = %v", err)
	}

	hub.Flush(2 * time.Second)
	if got := len(transport.Events()); got != 0 {
		t.Errorf("成功したジョブのSentryイベント件数 = %d、期待値 = 0", got)
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
		t.Fatalf("Work() = %v、期待値 = %v (riverのリトライ判断のためそのまま伝搬すること)", err, jobErr)
	}

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("失敗したジョブのSentryイベント件数 = %d、期待値 = 1", len(events))
	}
	got := events[0]
	if len(got.Exception) != 1 || got.Exception[0].Value != "ジョブ失敗" {
		t.Errorf("event.Exception = %+v、期待値 = valueが%qの例外", got.Exception, "ジョブ失敗")
	}
	if got.Tags["job.kind"] != "send_password_reset_email" {
		t.Errorf("event.Tags[job.kind] = %q、期待値 = %q", got.Tags["job.kind"], "send_password_reset_email")
	}
	if got.Tags["job.attempt"] != "3" {
		t.Errorf("event.Tags[job.attempt] = %q、期待値 = %q", got.Tags["job.attempt"], "3")
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
				t.Fatalf("Work() = %v、期待値 = %v (riverのリトライ判断のためそのまま伝搬すること)", err, tt.err)
			}

			hub.Flush(2 * time.Second)
			if got := len(transport.Events()); got != 0 {
				t.Errorf("無視対象のエラーのSentryイベント件数 = %d、期待値 = 0", got)
			}
		})
	}
}

func TestRiverWorkerMiddleware_BindsHubToContext(t *testing.T) {
	t.Parallel()

	// ワーカーは多くの場合slog.ErrorContextで中間エラーをSentryに流す
	// (NewSlogHandler経由でctxのHubを使う)。本テストではinnerで
	// hub.CaptureExceptionを直接呼び、ミドルウェアがctxにbindしたHubと
	// 同じtransportにイベントが届くこと・ジョブタグが乗っていることを確認する。
	parentHub, transport := newRiverTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), parentHub)

	mw := RiverWorkerMiddleware()
	err := mw.Work(ctx, newJobRow("send_sign_up_code_email", 2), func(ctx context.Context) error {
		jobHub := sentry.GetHubFromContext(ctx)
		if jobHub == nil {
			t.Fatal("内側のコールバックのctxにHubが関連付けられていない")
		}
		if jobHub == parentHub {
			t.Error("ジョブ単位のHubはCloneである必要がある (parentHubと同一インスタンスはNG)")
		}
		jobHub.CaptureException(errors.New("中間エラー"))
		return nil
	})
	if err != nil {
		t.Fatalf("Work()の想定外のエラー = %v", err)
	}

	parentHub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("内側のコールバックからのCaptureExceptionのイベント件数 = %d、期待値 = 1", len(events))
	}
	got := events[0]
	if len(got.Exception) != 1 || got.Exception[0].Value != "中間エラー" {
		t.Errorf("event.Exception = %+v、期待値 = valueが%qの例外", got.Exception, "中間エラー")
	}
	if got.Tags["job.kind"] != "send_sign_up_code_email" {
		t.Errorf("event.Tags[job.kind] = %q、期待値 = %q", got.Tags["job.kind"], "send_sign_up_code_email")
	}
	if got.Tags["job.attempt"] != "2" {
		t.Errorf("event.Tags[job.attempt] = %q、期待値 = %q", got.Tags["job.attempt"], "2")
	}
}
