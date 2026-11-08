// Package apipaginationはREST API V2の一覧のカーソルページネーションを扱う。
// カーソルの符号化・復号、1ページの件数の決定、次のページを指す `Link` ヘッダーの組み立てを担う。
package apipagination

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
)

const (
	// DefaultLimitは `limit` を省略したときの1ページの件数
	DefaultLimit int32 = 20
	// MaxLimitは1ページの件数の上限。上限を超える要求はエラーにせず上限に切り詰める
	MaxLimit int32 = 100
)

// cursorParamはカーソルを受け取るクエリパラメーターの名前
const cursorParam = "cursor"

// ErrInvalidCursorは、カーソルがこのAPIの発行したものとして読めないことを表すエラー
var ErrInvalidCursor = errors.New("カーソルを読めません")

// Limitはリクエストの `limit` から1ページの件数を決める。
// 省略時はDefaultLimit、MaxLimitを超える値はMaxLimitにする。
// 1未満の値はOpenAPI記述の `minimum` でリクエストの検証が拒否するため、ここには来ない
func Limit(requested *int32) int32 {
	if requested == nil {
		return DefaultLimit
	}
	return min(*requested, MaxLimit)
}

// EncodeCursorは並び順のキーを持つ値を、利用者に中身を解釈させない不透明な文字列にする。
// 値はJSONにしてからbase64url (パディング無し) で表し、URLのクエリにそのまま載せられるようにする
func EncodeCursor(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("カーソルの符号化に失敗: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeCursorはEncodeCursorで作ったカーソルをvへ復号する。
// 利用者が書き換えたカーソルを受け付けないよう、vに無いキーを含むものはErrInvalidCursorにする
func DecodeCursor(cursor string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return ErrInvalidCursor
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return ErrInvalidCursor
	}
	// 次の値が読める場合も構文エラーになる場合も拒否し、末尾が空白だけなら受け付ける
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidCursor
	}
	return nil
}

type requestURLKey struct{}

// WithRequestURLは、NextLinkが次のページのURLの元にするリクエストのURLをcontextに入れる。
// oapi-codegenのstrict serverのハンドラーはリクエストを受け取らないため、ルーターの構成でこれを呼ぶ
func WithRequestURL(ctx context.Context, u *url.URL) context.Context {
	return context.WithValue(ctx, requestURLKey{}, u)
}

// NextLinkは次のページを指す `Link` ヘッダーの値 (RFC 8288) を返す。
// リクエストのURLのクエリの `cursor` だけを差し替え、絞り込みや `limit` はそのまま引き継ぐ。
// URLはパスから始まる相対参照にし、リクエストのURLを基準に解決させる (RFC 8288 §3.1)。
// リクエストのURLがcontextに無い場合は空文字列を返す
func NextLink(ctx context.Context, cursor string) string {
	u, ok := ctx.Value(requestURLKey{}).(*url.URL)
	if !ok || u == nil {
		return ""
	}

	query := u.Query()
	query.Set(cursorParam, cursor)
	next := url.URL{Path: u.Path, RawPath: u.RawPath, RawQuery: query.Encode()}
	return "<" + next.String() + `>; rel="next"`
}
