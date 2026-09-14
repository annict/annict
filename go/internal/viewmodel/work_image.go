package viewmodel

import "github.com/annict/annict/go/internal/image"

// NoWorkImagePathは画像が登録されていない作品に表示する静的なプレースホルダー画像の
// パス。imgproxyのURLではなく素のアセットのため、サイズ違いの派生は持たない。
const NoWorkImagePath = "/static/images/no-work-image.png"

// WorkImageは1作品のサムネイルURLを、呼び出し側が必要とする幅で解決する。作品に
// 画像が登録されていない場合はNoWorkImagePathにフォールバックする。テンプレートは生成済み
// URLではなくこの型を持つため、同じ作品を異なるサイズで描画できる。
type WorkImage struct {
	imageDataJSON string
	helper        *image.Helper
}

// NewWorkImageはwork_images.image_dataの生JSONからWorkImageを組み立てる。
// helperがnilでもよく (テストやimgproxy未設定の呼び出し元)、その場合は
// プレースホルダーになる。
func NewWorkImage(imageDataJSON string, helper *image.Helper) WorkImage {
	return WorkImage{imageDataJSON: imageDataJSON, helper: helper}
}

// Existsは作品に画像が登録されているかを返す。呼び出し側は実サムネイルと
// プレースホルダーのどちらの見せ方をするかの判断に使う。
func (i WorkImage) Exists() bool {
	return i.originalURL() != ""
}

// URLは指定幅のimgproxy URLを返す。作品に画像が無い場合はNoWorkImagePathを返す。
func (i WorkImage) URL(width int, format string) string {
	if i.helper == nil {
		return NoWorkImagePath
	}

	if url := i.helper.GetWorkImageURL(i.imageDataJSON, width, format); url != "" {
		return url
	}

	return NoWorkImagePath
}

// SrcSetは指定幅の1x/2x srcsetを返す。作品に画像が無い場合は "" を返す。
// プレースホルダーは単一の固定アセットなので、そこでは空のsrcsetが正しい (呼び出し側は
// <source> を出さず、<img> のsrcに任せる)。
func (i WorkImage) SrcSet(width int, format string) string {
	if i.helper == nil {
		return ""
	}

	return i.helper.GetSrcSet(i.originalURL(), width, format)
}

// Heightは指定幅における画像の枠の高さを返す。
func (i WorkImage) Height(width int) int {
	return image.WorkImageHeight(width)
}

func (i WorkImage) originalURL() string {
	if i.helper == nil {
		return ""
	}

	return i.helper.ExtractImageURL(i.imageDataJSON)
}
