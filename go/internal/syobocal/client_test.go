package syobocal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClientは、httptestのサーバーへリクエストを送るクライアントを返す。
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{baseURL: server.URL, httpClient: server.Client()}
}

func TestClient_LookupSubTitles_ParsesSubTitles(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Command"); got != "TitleLookup" {
			t.Errorf("Command = %q、期待値 = TitleLookup", got)
		}
		if got := r.URL.Query().Get("TID"); got != "6309,7666" {
			t.Errorf("TID = %q、期待値 = 6309,7666", got)
		}
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><TitleLookupResponse>` +
			`<Result><Code>200</Code><Message></Message></Result><TitleItems>` +
			`<TitleItem id="6309"><TID>6309</TID><Title>SPY×FAMILY</Title>` +
			"<SubTitles>*01*オペレーション〈梟(ストリクス)〉\r\n*02*妻役を確保せよ\r\n*03*\r\nメモ\r\n*12*ペンギン*パーク</SubTitles></TitleItem>" +
			`<TitleItem id="7666"><TID>7666</TID><Title>幼馴染とはラブコメにならない</Title><SubTitles></SubTitles></TitleItem>` +
			`</TitleItems></TitleLookupResponse>`))
	})

	got, err := client.LookupSubTitles(context.Background(), []int32{6309, 7666})
	if err != nil {
		t.Fatalf("LookupSubTitles()のエラー = %v", err)
	}

	subTitles := got[6309]
	want := map[int32]string{
		1:  "オペレーション〈梟(ストリクス)〉",
		2:  "妻役を確保せよ",
		12: "ペンギン*パーク",
	}
	if len(subTitles) != len(want) {
		t.Fatalf("got[6309] = %v、期待値 = %v", subTitles, want)
	}
	for number, title := range want {
		if subTitles[number] != title {
			t.Errorf("got[6309][%d] = %q、期待値 = %q", number, subTitles[number], title)
		}
	}
	if _, ok := got[7666]; ok {
		t.Errorf("サブタイトルの無い作品が結果に含まれた: %v", got[7666])
	}
}

func TestClient_LookupSubTitles_NotFound(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><TitleLookupResponse>` +
			`<Result><Code>404</Code><Message>条件に一致するデータは存在しません</Message></Result>` +
			`<TitleItems></TitleItems></TitleLookupResponse>`))
	})

	got, err := client.LookupSubTitles(context.Background(), []int32{99999999})
	if err != nil {
		t.Fatalf("LookupSubTitles()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("結果 = %v、期待値 = 空", got)
	}
}

func TestClient_LookupSubTitles_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "HTTPのステータスが200でない",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
		},
		{
			name: "結果のコードがエラー",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`<TitleLookupResponse><Result><Code>500</Code><Message>error</Message></Result></TitleLookupResponse>`))
			},
		},
		{
			name: "XMLとして読めない",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`<html>`))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t, tt.handler)
			if _, err := client.LookupSubTitles(context.Background(), []int32{1}); err == nil {
				t.Error("エラーを期待したが、nilだった")
			}
		})
	}
}

func TestClient_LookupSubTitles_EmptyTIDs(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("TIDが無いのにリクエストを送った")
	})

	got, err := client.LookupSubTitles(context.Background(), nil)
	if err != nil {
		t.Fatalf("LookupSubTitles()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("結果 = %v、期待値 = 空", got)
	}
}
