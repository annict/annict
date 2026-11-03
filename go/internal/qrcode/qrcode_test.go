package qrcode_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/qrcode"
)

// otpauthURLは認証アプリへの登録に使うotpauth URIと同じ長さ・形の内容。
const otpauthURL = "otpauth://totp/Annict:alice?algorithm=SHA1&digits=6&issuer=Annict&period=30&secret=JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

// TestEncodeは、余白を除いた1辺のモジュール数と、左上の位置検出パターンを塗るpathを返すことを検証する。
// 位置検出パターンは7モジュールの正方形の枠のため、1行目は横に7つ連続した暗いモジュールになる。
func TestEncode(t *testing.T) {
	t.Parallel()

	code, err := qrcode.Encode(otpauthURL)
	if err != nil {
		t.Fatalf("Encode()のエラー = %v", err)
	}

	// otpauth URIの長さでは、バージョンは1辺21〜177モジュールの範囲に収まり、4モジュールずつ増える。
	if code.Size < 21 || (code.Size-21)%4 != 0 {
		t.Errorf("Size = %d、QRコードのバージョンに対応する1辺のモジュール数を期待", code.Size)
	}
	if !strings.HasPrefix(code.Path, "M0 0h7v1h-7z") {
		t.Errorf("Path = %q、左上の位置検出パターンの1行目で始まることを期待", code.Path)
	}
	if strings.Contains(code.Path, "h0") {
		t.Errorf("Path = %q、幅0の矩形を含まないことを期待", code.Path)
	}

	wantViewBox := "-4 -4 " + strconv.Itoa(code.Size+8) + " " + strconv.Itoa(code.Size+8)
	if got := code.ViewBox(); got != wantViewBox {
		t.Errorf("ViewBox() = %q、期待値 = %q", got, wantViewBox)
	}
}

// TestEncode_Deterministicは、同じ内容からは同じQRコードを返すことを検証する。
// QRコードが内容だけで決まり、描くたびに読み取り結果が揺れないことを確かめるため。
func TestEncode_Deterministic(t *testing.T) {
	t.Parallel()

	first, err := qrcode.Encode(otpauthURL)
	if err != nil {
		t.Fatalf("Encode()のエラー = %v", err)
	}
	second, err := qrcode.Encode(otpauthURL)
	if err != nil {
		t.Fatalf("Encode()のエラー = %v", err)
	}

	if *first != *second {
		t.Error("同じ内容から異なるQRコードが返った")
	}
}

// TestEncode_TooLongは、QRコードに収まらない長さの内容をエラーにすることを検証する。
func TestEncode_TooLong(t *testing.T) {
	t.Parallel()

	if _, err := qrcode.Encode(strings.Repeat("a", 5000)); err == nil {
		t.Error("収まらない長さの内容でエラーが返らなかった")
	}
}
