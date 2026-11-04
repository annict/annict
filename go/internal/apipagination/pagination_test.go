package apipagination_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"testing"

	"github.com/annict/annict/go/internal/apipagination"
)

func TestLimit(t *testing.T) {
	t.Parallel()

	int32Ptr := func(v int32) *int32 { return &v }

	tests := []struct {
		name      string
		requested *int32
		want      int32
	}{
		{name: "省略時は既定の件数", requested: nil, want: apipagination.DefaultLimit},
		{name: "上限以下の値はそのまま", requested: int32Ptr(1), want: 1},
		{name: "上限ちょうどはそのまま", requested: int32Ptr(apipagination.MaxLimit), want: apipagination.MaxLimit},
		{name: "上限を超える値は上限に切り詰める", requested: int32Ptr(apipagination.MaxLimit + 1), want: apipagination.MaxLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := apipagination.Limit(tt.requested); got != tt.want {
				t.Errorf("Limit() = %d、期待値 = %d", got, tt.want)
			}
		})
	}
}

type testCursor struct {
	Number *int32 `json:"n"`
}

func TestEncodeCursor_DecodeCursor(t *testing.T) {
	t.Parallel()

	number := int32(42)
	cursor, err := apipagination.EncodeCursor(testCursor{Number: &number})
	if err != nil {
		t.Fatalf("EncodeCursor()のエラー = %v", err)
	}

	var got testCursor
	if err := apipagination.DecodeCursor(cursor, &got); err != nil {
		t.Fatalf("DecodeCursor()のエラー = %v", err)
	}
	if got.Number == nil || *got.Number != number {
		t.Errorf("復号した値 = %v、期待値 = %d", got.Number, number)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		cursor string
	}{
		{name: "base64urlとして読めない", cursor: "!!!"},
		{name: "JSONとして読めない", cursor: "bm90IGpzb24"},           // "not json"
		{name: "知らないキーを含む", cursor: "eyJ4IjoxfQ"},              // {"x":1}
		{name: "値の後ろに余分な内容が続く", cursor: "eyJuIjoxfXsibiI6Mn0"}, // {"n":1}{"n":2}
		{name: "型が合わない", cursor: "eyJuIjoiMSJ9"},               // {"n":"1"}
		{name: "オブジェクトの後ろに閉じ角括弧がある", cursor: base64.RawURLEncoding.EncodeToString([]byte(`{"n":1}]`))},
		{name: "オブジェクトの後ろに閉じ波括弧がある", cursor: base64.RawURLEncoding.EncodeToString([]byte(`{"n":1}}`))},
		{name: "閉じ括弧の後ろに文字列がある", cursor: base64.RawURLEncoding.EncodeToString([]byte(`{"n":1}]garbage`))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got testCursor
			if err := apipagination.DecodeCursor(tt.cursor, &got); !errors.Is(err, apipagination.ErrInvalidCursor) {
				t.Errorf("DecodeCursor()のエラー = %v、期待値 = ErrInvalidCursor", err)
			}
		})
	}
}

func TestDecodeCursor_TrailingWhitespace(t *testing.T) {
	t.Parallel()

	cursor := base64.RawURLEncoding.EncodeToString([]byte("{\"n\":42} \n\t"))
	var got testCursor
	if err := apipagination.DecodeCursor(cursor, &got); err != nil {
		t.Fatalf("DecodeCursor()のエラー = %v", err)
	}
	if got.Number == nil || *got.Number != 42 {
		t.Errorf("復号した値 = %v、期待値 = 42", got.Number)
	}
}

func TestNextLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requestURL string
		want       string
	}{
		{
			name:       "カーソルを足し、ほかのクエリを引き継ぐ",
			requestURL: "https://example.com/api/v2/anime/works?limit=5",
			want:       `</api/v2/anime/works?cursor=next-cursor&limit=5>; rel="next"`,
		},
		{
			name:       "前のページのカーソルを差し替える",
			requestURL: "/api/v2/anime/works?cursor=prev-cursor",
			want:       `</api/v2/anime/works?cursor=next-cursor>; rel="next"`,
		},
		{
			name:       "エスケープされたパスをそのまま保つ",
			requestURL: "/api/v2/anime/a%2Fb/episodes",
			want:       `</api/v2/anime/a%2Fb/episodes?cursor=next-cursor>; rel="next"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tt.requestURL)
			if err != nil {
				t.Fatalf("url.Parse()のエラー = %v", err)
			}
			ctx := apipagination.WithRequestURL(context.Background(), u)
			if got := apipagination.NextLink(ctx, "next-cursor"); got != tt.want {
				t.Errorf("NextLink() = %q、期待値 = %q", got, tt.want)
			}
		})
	}

	t.Run("リクエストのURLが無ければ空文字列", func(t *testing.T) {
		t.Parallel()

		if got := apipagination.NextLink(context.Background(), "next-cursor"); got != "" {
			t.Errorf("NextLink() = %q、期待値 = 空文字列", got)
		}
	})
}
