package storage

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// deleteRequestはDeleteObjects APIのリクエスト本文。
type deleteRequest struct {
	Objects []struct {
		Key string `xml:"Key"`
	} `xml:"Object"`
}

// newTestAdapterは、DeleteObjects APIのリクエストを記録してresponseBodyを返すサーバーに
// 接続したAdapterを作成する。
func newTestAdapter(t *testing.T, responseBody string) (*Adapter, *[]deleteRequest) {
	t.Helper()

	var requests []deleteRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasPrefix(r.URL.Path, "/test-bucket") || !r.URL.Query().Has("delete") {
			t.Errorf("想定外のリクエスト: %s %s", r.Method, r.URL)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("リクエスト本文の読み込みに失敗: %v", err)
		}
		var req deleteRequest
		if err := xml.Unmarshal(body, &req); err != nil {
			t.Fatalf("リクエスト本文の解析に失敗: %v", err)
		}
		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, responseBody)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "test-access-key", "test-secret-key", "auto")
	return NewAdapter(client, "test-bucket"), &requests
}

func TestAdapter_DeleteObjects(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 上限ごとに分けて削除する", func(t *testing.T) {
		t.Parallel()

		adapter, requests := newTestAdapter(t, `<DeleteResult></DeleteResult>`)
		keys := make([]string, deleteObjectsLimit+1)
		for i := range keys {
			keys[i] = "shrine/profile/" + strings.Repeat("a", i%5+1)
		}

		if err := adapter.DeleteObjects(context.Background(), keys); err != nil {
			t.Fatalf("DeleteObjects()のエラー = %v", err)
		}
		if len(*requests) != 2 {
			t.Fatalf("リクエスト数 = %d、期待値 = 2", len(*requests))
		}
		if got := len((*requests)[0].Objects); got != deleteObjectsLimit {
			t.Errorf("1回目のオブジェクト数 = %d、期待値 = %d", got, deleteObjectsLimit)
		}
		if got := (*requests)[1].Objects; len(got) != 1 || got[0].Key != keys[deleteObjectsLimit] {
			t.Errorf("2回目のオブジェクト = %v、期待値 = [%s]", got, keys[deleteObjectsLimit])
		}
	})

	t.Run("正常系: キーが無ければリクエストしない", func(t *testing.T) {
		t.Parallel()

		adapter, requests := newTestAdapter(t, `<DeleteResult></DeleteResult>`)
		if err := adapter.DeleteObjects(context.Background(), nil); err != nil {
			t.Fatalf("DeleteObjects()のエラー = %v", err)
		}
		if len(*requests) != 0 {
			t.Errorf("リクエスト数 = %d、期待値 = 0", len(*requests))
		}
	})

	t.Run("異常系: オブジェクトごとの失敗をエラーにする", func(t *testing.T) {
		t.Parallel()

		adapter, _ := newTestAdapter(t, `<DeleteResult><Error><Key>shrine/a.png</Key><Code>AccessDenied</Code><Message>Access Denied</Message></Error></DeleteResult>`)
		err := adapter.DeleteObjects(context.Background(), []string{"shrine/a.png"})
		if err == nil || !strings.Contains(err.Error(), "shrine/a.png") || !strings.Contains(err.Error(), "AccessDenied") {
			t.Errorf("DeleteObjects()のエラー = %v、期待値 = キーとコードを含むエラー", err)
		}
	})
}
