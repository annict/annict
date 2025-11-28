package sentry

import (
	"context"
	"errors"
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
		t.Errorf("Init() with empty DSN should return nil, got %v", err)
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
		t.Error("Init() with invalid DSN should return error")
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
					t.Errorf("ヘッダー %s: got %q, want %q", key, result.Request.Headers[key], expectedValue)
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
				t.Errorf("Data: got %q, want %q", result.Request.Data, tt.expected)
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
				t.Errorf("QueryString: got %q, want %q", result.Request.QueryString, tt.expected)
			}
		})
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
		t.Errorf("無効なデータは[FILTERED]であるべき: got %q", result.Request.Data)
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
