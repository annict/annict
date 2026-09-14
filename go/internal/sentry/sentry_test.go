package sentry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestInit_EmptyDSN(t *testing.T) {
	// DSNが空の場合は初期化をスキップしてnilを返す
	cfg := Config{
		DSN:              "",
		Environment:      "test",
		TracesSampleRate: 0.5,
		Debug:            false,
	}

	err := Init(cfg)
	if err != nil {
		t.Errorf("DSNが空のときのInit() = %v、期待値 = nil", err)
	}
}

func TestInit_InvalidDSN(t *testing.T) {
	// 無効なDSNの場合はエラーを返す
	cfg := Config{
		DSN:              "invalid-dsn",
		Environment:      "test",
		TracesSampleRate: 0.5,
		Debug:            false,
	}

	err := Init(cfg)
	if err == nil {
		t.Error("不正なDSNでInit()がエラーを返さなかった")
	}
}

func TestInit_SetsReleaseAndTracingOptions(t *testing.T) {
	// InitがReleaseをクライアントオプションに伝搬し、トレースを有効化する
	// ことを検証する (EnableTracing無しではTracesSampleRateを渡してもトレースが
	// 送られない)。グローバルHubのクライアントを差し替えるため並行実行しない。
	cfg := Config{
		DSN:              "https://public@o0.ingest.sentry.io/1",
		Environment:      "test",
		Release:          "abc1234",
		TracesSampleRate: 0.5,
		Debug:            false,
	}

	// テスト後にグローバルHubをクライアント無しの状態へ戻し、後続テストが
	// liveクライアントにイベントを送らないようにする。
	t.Cleanup(func() {
		sentry.CurrentHub().BindClient(nil)
	})

	if err := Init(cfg); err != nil {
		t.Fatalf("Init()のエラー = %v", err)
	}

	client := sentry.CurrentHub().Client()
	if client == nil {
		t.Fatal("Init()後はクライアントが設定されているべき")
	}

	opts := client.Options()
	if opts.Release != "abc1234" {
		t.Errorf("Release = %q、期待値 = %q", opts.Release, "abc1234")
	}
	if !opts.EnableTracing {
		t.Error("EnableTracingはtrueであるべき")
	}
	if len(opts.IgnoreErrors) == 0 {
		t.Error("IgnoreErrorsが設定されているべき")
	}
}

func TestCaptureError_NilContext(t *testing.T) {
	// コンテキストにHubがない場合でもパニックしない
	ctx := context.Background()
	err := errors.New("test error")

	// パニックしなければOK
	CaptureError(ctx, err)
}

func TestCaptureMessage_NilContext(t *testing.T) {
	// コンテキストにHubがない場合でもパニックしない
	ctx := context.Background()

	// パニックしなければOK
	CaptureMessage(ctx, "test message")
}

func TestBeforeSend_FiltersRequestHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name: "Authorizationヘッダーをマスク",
			headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"Content-Type":  "application/json",
			},
			expected: map[string]string{
				"Authorization": "[FILTERED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "Cookieヘッダーをマスク",
			headers: map[string]string{
				"Cookie":       "session_id=abc123",
				"Content-Type": "text/html",
			},
			expected: map[string]string{
				"Cookie":       "[FILTERED]",
				"Content-Type": "text/html",
			},
		},
		{
			name: "X-CSRF-Tokenヘッダーをマスク",
			headers: map[string]string{
				"X-CSRF-Token": "csrf-token-value",
				"Accept":       "text/html",
			},
			expected: map[string]string{
				"X-CSRF-Token": "[FILTERED]",
				"Accept":       "text/html",
			},
		},
		{
			name: "大文字小文字を区別しない",
			headers: map[string]string{
				"authorization": "Bearer secret-token",
				"COOKIE":        "session_id=abc123",
				"x-csrf-token":  "csrf-value",
			},
			expected: map[string]string{
				"authorization": "[FILTERED]",
				"COOKIE":        "[FILTERED]",
				"x-csrf-token":  "[FILTERED]",
			},
		},
		{
			name: "センシティブでないヘッダーは変更しない",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0",
			},
			expected: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &sentry.Event{
				Request: &sentry.Request{
					Headers: tt.headers,
				},
			}

			result := beforeSend(event, nil)

			for key, expectedValue := range tt.expected {
				if result.Request.Headers[key] != expectedValue {
					t.Errorf("ヘッダー%s = %q、期待値 = %q", key, result.Request.Headers[key], expectedValue)
				}
			}
		})
	}
}

func TestBeforeSend_FiltersRequestData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "passwordフィールドをマスク",
			data:     "username=user&password=secret123",
			expected: "password=%5BFILTERED%5D&username=user",
		},
		{
			name:     "tokenフィールドをマスク",
			data:     "email=test@example.com&reset_token=abc123",
			expected: "email=test%40example.com&reset_token=%5BFILTERED%5D",
		},
		{
			name:     "secretフィールドをマスク",
			data:     "api_secret=mysecret&name=test",
			expected: "api_secret=%5BFILTERED%5D&name=test",
		},
		{
			name:     "複数のセンシティブフィールドをマスク",
			data:     "password=pass123&api_token=token123&client_secret=sec123",
			expected: "api_token=%5BFILTERED%5D&client_secret=%5BFILTERED%5D&password=%5BFILTERED%5D",
		},
		{
			name:     "大文字小文字を区別しない",
			data:     "PASSWORD=secret&Token=abc&SECRET=xyz",
			expected: "PASSWORD=%5BFILTERED%5D&SECRET=%5BFILTERED%5D&Token=%5BFILTERED%5D",
		},
		{
			name:     "センシティブでないフィールドは変更しない",
			data:     "username=user&email=test@example.com",
			expected: "username=user&email=test@example.com",
		},
		{
			name:     "空のデータ",
			data:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &sentry.Event{
				Request: &sentry.Request{
					Data: tt.data,
				},
			}

			result := beforeSend(event, nil)

			if result.Request.Data != tt.expected {
				t.Errorf("Data = %q、期待値 = %q", result.Request.Data, tt.expected)
			}
		})
	}
}

func TestBeforeSend_FiltersQueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "tokenパラメータをマスク",
			query:    "page=1&token=secret123",
			expected: "page=1&token=%5BFILTERED%5D",
		},
		{
			name:     "keyパラメータをマスク",
			query:    "id=123&api_key=mykey",
			expected: "api_key=%5BFILTERED%5D&id=123",
		},
		{
			name:     "複数のセンシティブパラメータをマスク",
			query:    "access_token=abc&secret_key=xyz&page=1",
			expected: "access_token=%5BFILTERED%5D&page=1&secret_key=%5BFILTERED%5D",
		},
		{
			name:     "大文字小文字を区別しない",
			query:    "TOKEN=secret&API_KEY=mykey",
			expected: "API_KEY=%5BFILTERED%5D&TOKEN=%5BFILTERED%5D",
		},
		{
			name:     "センシティブでないパラメータは変更しない",
			query:    "page=1&limit=10&sort=created_at",
			expected: "page=1&limit=10&sort=created_at",
		},
		{
			name:     "空のクエリ",
			query:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &sentry.Event{
				Request: &sentry.Request{
					QueryString: tt.query,
				},
			}

			result := beforeSend(event, nil)

			if result.Request.QueryString != tt.expected {
				t.Errorf("QueryString = %q、期待値 = %q", result.Request.QueryString, tt.expected)
			}
		})
	}
}

func TestBeforeSend_FiltersTags(t *testing.T) {
	t.Parallel()

	// アプリケーションのイベントハンドラーがslog属性をタグへ載せるため、
	// 構造化属性としてログに載ったPII (例: "email") がここでマスクされることを
	// 検証する。
	tests := []struct {
		name     string
		tags     map[string]string
		expected map[string]string
	}{
		{
			name: "emailタグをマスク",
			tags: map[string]string{
				"email":   "user@example.com",
				"user_id": "123",
			},
			expected: map[string]string{
				"email":   "[FILTERED]",
				"user_id": "123",
			},
		},
		{
			name: "部分一致でマスク",
			tags: map[string]string{
				"user_email": "user@example.com",
			},
			expected: map[string]string{
				"user_email": "[FILTERED]",
			},
		},
		{
			name: "password/token/secretタグをマスク",
			tags: map[string]string{
				"password":      "raw-password",
				"api_token":     "token-value",
				"client_secret": "secret-value",
			},
			expected: map[string]string{
				"password":      "[FILTERED]",
				"api_token":     "[FILTERED]",
				"client_secret": "[FILTERED]",
			},
		},
		{
			name: "大文字小文字を区別しない",
			tags: map[string]string{
				"Email": "user@example.com",
			},
			expected: map[string]string{
				"Email": "[FILTERED]",
			},
		},
		{
			name: "センシティブでないタグは変更しない",
			tags: map[string]string{
				"user_id":       "123",
				"annict_source": "other_subsystem",
			},
			expected: map[string]string{
				"user_id":       "123",
				"annict_source": "other_subsystem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &sentry.Event{Tags: tt.tags}

			result := beforeSend(event, nil)

			for key, expectedValue := range tt.expected {
				if result.Tags[key] != expectedValue {
					t.Errorf("タグ%s = %q、期待値 = %q", key, result.Tags[key], expectedValue)
				}
			}
		})
	}
}

func TestBeforeSend_HandlesNilTags(t *testing.T) {
	t.Parallel()

	event := &sentry.Event{Tags: nil}

	if result := beforeSend(event, nil); result == nil {
		t.Error("Tagsがnilでもイベントは保持されるべき")
	}
}

func TestBeforeSend_HandlesNilRequest(t *testing.T) {
	t.Parallel()

	event := &sentry.Event{
		Request: nil,
	}

	result := beforeSend(event, nil)

	if result.Request != nil {
		t.Error("nilのRequestはnilのままであるべき")
	}
}

func TestBeforeSend_HandlesInvalidData(t *testing.T) {
	t.Parallel()

	event := &sentry.Event{
		Request: &sentry.Request{
			Data: "%invalid-data%",
		},
	}

	result := beforeSend(event, nil)

	if result.Request.Data != "[FILTERED]" {
		t.Errorf("無効なデータの表示 = %q、期待値 = [FILTERED]", result.Request.Data)
	}
}

func TestBeforeSend_DropsIgnorableErrors(t *testing.T) {
	t.Parallel()

	// クライアント切断・runtime中断由来のエラーを持つイベントは破棄される
	// (beforeSendがnilを返す) ことを検証する。
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "context.Canceledは破棄",
			err:  context.Canceled,
		},
		{
			name: "context.Canceledのラップは破棄",
			err:  fmt.Errorf("接続エラー: %w", context.Canceled),
		},
		{
			name: "http.ErrAbortHandlerは破棄",
			err:  http.ErrAbortHandler,
		},
		{
			name: "http.ErrAbortHandlerのラップは破棄",
			err:  fmt.Errorf("ハンドラー中断: %w", http.ErrAbortHandler),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &sentry.Event{
				Request: &sentry.Request{},
			}
			hint := &sentry.EventHint{OriginalException: tt.err}

			if result := beforeSend(event, hint); result != nil {
				t.Errorf("無視対象のエラーのイベント = %+v、期待値 = nil", result)
			}
		})
	}
}

func TestBeforeSend_KeepsNonIgnorableErrors(t *testing.T) {
	t.Parallel()

	// 通常のエラーは引き続きSentryに届くことを検証する。
	event := &sentry.Event{
		Request: &sentry.Request{},
	}
	hint := &sentry.EventHint{OriginalException: errors.New("通常のエラー")}

	if result := beforeSend(event, hint); result == nil {
		t.Error("無視対象でないエラーはイベントを保持すべき")
	}
}

func TestShouldDropError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nilは破棄しない", err: nil, want: false},
		{name: "context.Canceledは破棄", err: context.Canceled, want: true},
		{name: "http.ErrAbortHandlerは破棄", err: http.ErrAbortHandler, want: true},
		{name: "ラップされたcontext.Canceledは破棄", err: fmt.Errorf("ラップ元のエラー = %w", context.Canceled), want: true},
		{name: "通常のエラーは破棄しない", err: errors.New("通常のエラー"), want: false},
		{name: "context.DeadlineExceededは破棄しない", err: context.DeadlineExceeded, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := shouldDropError(tt.err); got != tt.want {
				t.Errorf("shouldDropError(%v) = %v、期待値 = %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestFilterRequestHeaders_NilHeaders(t *testing.T) {
	t.Parallel()

	req := &sentry.Request{
		Headers: nil,
	}

	filterRequestHeaders(req)

	if req.Headers != nil {
		t.Error("nilのHeadersはnilのままであるべき")
	}
}
