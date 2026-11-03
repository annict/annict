// Package qrcodeは文字列をQRコードにし、テンプレートがインラインのSVGとして描ける形で返す。
//
// 画像を別のリクエストで配らず、ページに埋め込んで描くためのPresentation層のヘルパー。
// SVGにするのは、表示する大きさを変えても輪郭がぼやけず、端末の画面どうしで読み取りやすいため。
package qrcode

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/boombuler/barcode/qr"
)

// QuietZoneは、読み取り機が符号の範囲を見分けるためにQRコードの周りに空ける余白のモジュール数。
// QRコードの規格 (JIS X 0510) が求める幅にする。
const QuietZone = 4

// CodeはSVGで描くQRコード。座標の単位はモジュール (QRコードの1マス) で、原点は符号の左上の角。
type Code struct {
	// Sizeは余白を除いた1辺のモジュール数。
	Size int
	// Pathは暗いモジュールを塗るSVGのpath要素のd属性の値。
	Path string
}

// ViewBoxは余白を含めた範囲を指すSVGのviewBox属性の値を返す。
func (c *Code) ViewBox() string {
	return fmt.Sprintf("-%d -%d %d %d", QuietZone, QuietZone, c.Size+2*QuietZone, c.Size+2*QuietZone)
}

// Encodeはcontentをquiet zoneを含まないQRコードにする。
//
// 誤り訂正は中程度 (M) にする。画面を別の端末のカメラで写すときの映り込みや傾きに耐えつつ、
// otpauth URIほどの長さの文字列でもモジュールを細かくしすぎないため。
func Encode(content string) (*Code, error) {
	code, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("QRコードのエンコードに失敗: %w", err)
	}

	size := code.Bounds().Dx()
	var path strings.Builder
	for y := range size {
		// 横に連続する暗いモジュールは1つの矩形にまとめ、埋め込むSVGを小さくする。
		for x := 0; x < size; {
			if !isDark(code.At(x, y)) {
				x++
				continue
			}

			start := x
			for x < size && isDark(code.At(x, y)) {
				x++
			}
			width := strconv.Itoa(x - start)
			path.WriteString("M" + strconv.Itoa(start) + " " + strconv.Itoa(y) + "h" + width + "v1h-" + width + "z")
		}
	}

	return &Code{Size: size, Path: path.String()}, nil
}

// isDarkはモジュールの色が暗い (塗る) かを返す。
func isDark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r+g+b == 0
}
