package syobocal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
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

// TestClient_LookupSubTitles_Batchesは、件数上限を超えても100件以下ずつ問い合わせ、
// すべてのバッチのサブタイトルを返すことを検証する。
func TestClient_LookupSubTitles_Batches(t *testing.T) {
	t.Parallel()

	for _, count := range []int{100, 101, 250} {
		t.Run(strconv.Itoa(count)+"件", func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				ids := strings.Split(r.URL.Query().Get("TID"), ",")
				if len(ids) > 100 {
					t.Errorf("リクエストのTID = %d件、期待値 = 100件以下", len(ids))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_, _ = w.Write([]byte(`<TitleLookupResponse><Result><Code>200</Code></Result><TitleItems>`))
				for _, id := range ids {
					_, _ = fmt.Fprintf(w, `<TitleItem><TID>%s</TID><SubTitles>*01*サブタイトル%s</SubTitles></TitleItem>`, id, id)
				}
				_, _ = w.Write([]byte(`</TitleItems></TitleLookupResponse>`))
			})
			tids := make([]int32, count)
			for i := range tids {
				tids[i] = int32(i + 1) // #nosec G115 -- 最大250件のテストデータ
			}
			got, err := client.LookupSubTitles(context.Background(), tids)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != count || int(calls.Load()) != (count+99)/100 {
				t.Errorf("結果 = %d件、問い合わせ = %d回、期待値 = %d件・%d回", len(got), calls.Load(), count, (count+99)/100)
			}
			for _, tid := range tids {
				if got[tid][1] != "サブタイトル"+strconv.Itoa(int(tid)) {
					t.Errorf("TID%dのサブタイトル = %q", tid, got[tid][1])
				}
			}
		})
	}
}

// TestClient_LookupSubTitles_BatchFailureは、途中のバッチが404なら残りを取得し、
// エラーなら部分的な結果を返さず終了することを検証する。
func TestClient_LookupSubTitles_BatchFailure(t *testing.T) {
	t.Parallel()

	for _, code := range []int{404, 500} {
		t.Run(strconv.Itoa(code)+"の結果", func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 2 {
					_, _ = fmt.Fprintf(w, `<TitleLookupResponse><Result><Code>%d</Code></Result></TitleLookupResponse>`, code)
					return
				}
				ids := strings.Split(r.URL.Query().Get("TID"), ",")
				_, _ = w.Write([]byte(`<TitleLookupResponse><Result><Code>200</Code></Result><TitleItems>`))
				for _, id := range ids {
					_, _ = fmt.Fprintf(w, `<TitleItem><TID>%s</TID><SubTitles>*01*サブタイトル</SubTitles></TitleItem>`, id)
				}
				_, _ = w.Write([]byte(`</TitleItems></TitleLookupResponse>`))
			})
			tids := make([]int32, 201)
			for i := range tids {
				tids[i] = int32(i + 1) // #nosec G115 -- 最大201件のテストデータ
			}
			got, err := client.LookupSubTitles(context.Background(), tids)
			if code == 500 {
				if err == nil || got != nil || calls.Load() != 2 {
					t.Errorf("結果 = %v、エラー = %v、問い合わせ = %d回、期待値 = 結果なし・エラー・2回", got, err, calls.Load())
				}
				return
			}
			if err != nil || len(got) != 101 || got[201][1] != "サブタイトル" || calls.Load() != 3 {
				t.Errorf("結果 = %d件、エラー = %v、問い合わせ = %d回、期待値 = 101件・3回", len(got), err, calls.Load())
			}
		})
	}
}
