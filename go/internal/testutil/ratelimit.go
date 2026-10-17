package testutil

import (
	"strings"
	"testing"
)

// UniqueRateLimitPrefixは呼び出し元テストの名前から派生する一意なプレフィックスを返す。
// レート制限のキー (またはハンドラーがキーに組み込むemail / RemoteAddrなどの構成値) を
// 本プレフィックスで組み立てる。各テストはトランザクション内で数えるため結果は互いに
// 見えないが、`rate_limits` の (key, window_start) の一意制約は未コミットの行どうしでも
// 効くため、並列テストが同じキーを使うと一方が他方のロールバックまで待たされる。
// 戻り値はキーとメールアドレスのローカルパートのどちらでも安全に使える文字のみを含む:
// [A-Za-z0-9_] 以外の文字 (サブテストの区切り、空白、マルチバイト文字など) はすべて
// '_' に置換するため、日本語や記号を含むサブテスト名でも安全なprefixを得られる。
func UniqueRateLimitPrefix(t *testing.T) string {
	t.Helper()
	// [A-Za-z0-9_] 以外のruneをすべて '_' に置換し、キーの一部としても
	// メールアドレスのローカルパートとしても安全な文字列にする。
	// サブテストの '/' (例: "TestFoo/case_a") もこのルールに含まれる。
	name := t.Name()
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
