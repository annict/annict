package viewmodel

import "github.com/annict/annict/go/internal/image"

// NoWorkImagePath is the static placeholder shown for works that have no registered
// image. It is a plain asset rather than an imgproxy URL, so it has no size variants.
//
// [Ja] NoWorkImagePath は画像が登録されていない作品に表示する静的なプレースホルダー画像の
// パス。imgproxy の URL ではなく素のアセットのため、サイズ違いの派生は持たない。
const NoWorkImagePath = "/static/images/no-work-image.png"

// WorkImage resolves the thumbnail URLs of one work at whatever width the caller needs,
// falling back to NoWorkImagePath when the work has no registered image. Templates hold
// this instead of a pre-rendered URL so the same work can be drawn at different sizes.
//
// [Ja] WorkImage は 1 作品のサムネイル URL を、呼び出し側が必要とする幅で解決する。作品に
// 画像が登録されていない場合は NoWorkImagePath にフォールバックする。テンプレートは生成済み
// URL ではなくこの型を持つため、同じ作品を異なるサイズで描画できる。
type WorkImage struct {
	imageDataJSON string
	helper        *image.Helper
}

// NewWorkImage builds a WorkImage from the raw work_images.image_data JSON. A nil helper
// is allowed (tests and callers without imgproxy configured) and yields the placeholder.
//
// [Ja] NewWorkImage は work_images.image_data の生 JSON から WorkImage を組み立てる。
// helper が nil でもよく (テストや imgproxy 未設定の呼び出し元)、その場合は
// プレースホルダーになる。
func NewWorkImage(imageDataJSON string, helper *image.Helper) WorkImage {
	return WorkImage{imageDataJSON: imageDataJSON, helper: helper}
}

// Exists reports whether the work has a registered image. Callers use it to decide
// between the real thumbnail's presentation and the placeholder's.
//
// [Ja] Exists は作品に画像が登録されているかを返す。呼び出し側は実サムネイルと
// プレースホルダーのどちらの見せ方をするかの判断に使う。
func (i WorkImage) Exists() bool {
	return i.originalURL() != ""
}

// URL returns the imgproxy URL for the given width, or NoWorkImagePath when the work
// has no image.
//
// [Ja] URL は指定幅の imgproxy URL を返す。作品に画像が無い場合は NoWorkImagePath を返す。
func (i WorkImage) URL(width int, format string) string {
	if i.helper == nil {
		return NoWorkImagePath
	}

	if url := i.helper.GetWorkImageURL(i.imageDataJSON, width, format); url != "" {
		return url
	}

	return NoWorkImagePath
}

// SrcSet returns the 1x/2x srcset for the given width, or "" when the work has no image.
// The placeholder is a single fixed asset, so an empty srcset is correct there: the
// caller omits the <source> elements and lets the <img> src carry it.
//
// [Ja] SrcSet は指定幅の 1x/2x srcset を返す。作品に画像が無い場合は "" を返す。
// プレースホルダーは単一の固定アセットなので、そこでは空の srcset が正しい (呼び出し側は
// <source> を出さず、<img> の src に任せる)。
func (i WorkImage) SrcSet(width int, format string) string {
	if i.helper == nil {
		return ""
	}

	return i.helper.GetSrcSet(i.originalURL(), width, format)
}

// Height returns the height of the image box at the given width.
//
// [Ja] Height は指定幅における画像の枠の高さを返す。
func (i WorkImage) Height(width int) int {
	return image.WorkImageHeight(width)
}

func (i WorkImage) originalURL() string {
	if i.helper == nil {
		return ""
	}

	return i.helper.ExtractImageURL(i.imageDataJSON)
}
