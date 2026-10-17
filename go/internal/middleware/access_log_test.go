package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureDefaultLogはslogの既定ロガーをJSONの出力先に差し替え、テストの終了時に元へ戻す。
// 既定ロガーはプロセス全体で共有されるため、これを呼ぶテストは並列化しない。
func captureDefaultLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	previousLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	var logOutput bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	return &logOutput
}

func TestAccessLog(t *testing.T) {
	logOutput := captureDefaultLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "https://annict.example.com/password/edit?token=secret-reset-token", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.10")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if strings.Contains(logOutput.String(), "secret-reset-token") {
		t.Errorf("ログにクエリ文字列のトークンが含まれている: %s", logOutput.String())
	}

	var entry map[string]any
	if err := json.Unmarshal(logOutput.Bytes(), &entry); err != nil {
		t.Fatalf("ログをJSONとして読めない: %v (ログ = %q)", err, logOutput.String())
	}

	want := map[string]any{
		"msg":        "リクエストを処理しました",
		"method":     http.MethodGet,
		"host":       "annict.example.com",
		"path":       "/password/edit",
		"status":     float64(http.StatusCreated),
		"bytes":      float64(len("ok")),
		"ip_address": "203.0.113.10",
	}
	for key, wantValue := range want {
		if entry[key] != wantValue {
			t.Errorf("ログの%s = %v、期待値 = %v", key, entry[key], wantValue)
		}
	}
	if _, ok := entry["duration"]; !ok {
		t.Error("ログにdurationが含まれていない")
	}
}

func TestAccessLog_HandlerPanics(t *testing.T) {
	logOutput := captureDefaultLog(t)

	handler := AccessLog(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("ハンドラーの失敗")
	}))

	func() {
		defer func() {
			if recover() == nil {
				t.Error("panicが外側へ伝わらなかった")
			}
		}()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/works", nil))
	}()

	if !strings.Contains(logOutput.String(), `"path":"/works"`) {
		t.Errorf("panicしたリクエストがログに残っていない: %q", logOutput.String())
	}
}
