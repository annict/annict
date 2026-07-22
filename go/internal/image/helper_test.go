package image

import (
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

const testImageDataJSON = `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`

func newTestHelper() *Helper {
	return NewHelper(&config.Config{
		Env:              "test",
		ImgproxyEndpoint: "http://localhost:18080",
		ImgproxyKey:      "test-key",
		ImgproxySalt:     "test-salt",
		S3BucketName:     "test-bucket",
	})
}

// TestWorkImageHeight verifies the portrait work-image ratio: 3 wide to 4 tall.
//
// [Ja] TestWorkImageHeight は作品画像の縦長比率 (横 3 : 縦 4) を検証する。
func TestWorkImageHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		width int
		want  int
	}{
		{width: 70, want: 93},
		{width: 280, want: 373},
		{width: 600, want: 800},
	}

	for _, tt := range tests {
		if got := WorkImageHeight(tt.width); got != tt.want {
			t.Errorf("WorkImageHeight(%d) = %d, want %d", tt.width, got, tt.want)
		}
	}
}

// TestGenerateImgproxyURL_ResizesWithFit is the guard for the resizing mode: work images
// must be resized with "fit", which preserves the source aspect ratio, and never with
// "fill", which crops a landscape image's top and bottom away to fill the 3:4 box.
//
// [Ja] TestGenerateImgproxyURL_ResizesWithFit はリサイズ方式の歯止め。作品画像は元の
// アスペクト比を保つ "fit" でリサイズしなければならず、3:4 の枠を埋めるために横長画像の
// 上下を切り落とす "fill" を使ってはならない。
func TestGenerateImgproxyURL_ResizesWithFit(t *testing.T) {
	t.Parallel()

	url := newTestHelper().GenerateImgproxyURL("s3://test-bucket/shrine/a.jpg", 70, "jpg")

	if !strings.Contains(url, "resize:fit:70:93:0") {
		t.Errorf("URL に resize:fit:70:93:0 が含まれるべきです: %s", url)
	}

	// "resize:fill:" would crop; "fill-down" belongs to the 1:1 avatar path only.
	//
	// [Ja] "resize:fill:" は切り抜きになる。"fill-down" は 1:1 のアバター専用。
	if strings.Contains(url, "resize:fill") {
		t.Errorf("作品画像で fill 系のリサイズを使ってはいけません: %s", url)
	}
}

// TestGenerateImgproxyURL_Format verifies that a non-jpg format is appended as a format
// option while jpg (the default output) is not.
//
// [Ja] TestGenerateImgproxyURL_Format は jpg 以外のフォーマットが format オプションとして
// 付与され、既定の出力である jpg では付与されないことを検証する。
func TestGenerateImgproxyURL_Format(t *testing.T) {
	t.Parallel()

	h := newTestHelper()
	originalURL := "s3://test-bucket/shrine/a.jpg"

	webp := h.GenerateImgproxyURL(originalURL, 70, "webp")
	if !strings.Contains(webp, "format:webp") || !strings.HasSuffix(webp, ".webp") {
		t.Errorf("webp の URL = %s, want format:webp と .webp 拡張子", webp)
	}

	jpg := h.GenerateImgproxyURL(originalURL, 70, "jpg")
	if strings.Contains(jpg, "format:") || !strings.HasSuffix(jpg, ".jpg") {
		t.Errorf("jpg の URL = %s, want format オプション無しと .jpg 拡張子", jpg)
	}
}

// TestGenerateImgproxyURL_EmptyOriginal verifies that an empty source URL yields an empty
// result rather than a signed URL pointing at nothing.
//
// [Ja] TestGenerateImgproxyURL_EmptyOriginal は元 URL が空のとき、実体を指さない署名付き
// URL ではなく空文字列を返すことを検証する。
func TestGenerateImgproxyURL_EmptyOriginal(t *testing.T) {
	t.Parallel()

	if got := newTestHelper().GenerateImgproxyURL("", 70, "jpg"); got != "" {
		t.Errorf("GenerateImgproxyURL(\"\") = %q, want 空文字列", got)
	}
}

// TestGetWorkImageURL verifies that image_data is resolved to a signed imgproxy URL, and
// that absent or unparsable data yields "" so callers can fall back to the placeholder.
//
// [Ja] TestGetWorkImageURL は image_data が署名付きの imgproxy URL に解決されること、
// データが無い / 壊れている場合は呼び出し側がプレースホルダーへ退避できるよう "" を返すことを
// 検証する。
func TestGetWorkImageURL(t *testing.T) {
	t.Parallel()

	h := newTestHelper()

	if got := h.GetWorkImageURL(testImageDataJSON, 70, "jpg"); !strings.HasPrefix(got, "http://localhost:18080/") {
		t.Errorf("GetWorkImageURL() = %q, want imgproxy エンドポイント始まりの URL", got)
	}

	tests := []struct {
		name          string
		imageDataJSON string
	}{
		{name: "空の image_data", imageDataJSON: ""},
		{name: "壊れた JSON", imageDataJSON: "{"},
		{name: "id が空", imageDataJSON: `{"master":{"id":"","storage":"store"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := h.GetWorkImageURL(tt.imageDataJSON, 70, "jpg"); got != "" {
				t.Errorf("GetWorkImageURL() = %q, want 空文字列", got)
			}
		})
	}
}

// TestGetSrcSet verifies the 1x/2x descriptors and that the 2x entry is generated at twice
// the width, so a high-density display gets a sharper image of the same box.
//
// [Ja] TestGetSrcSet は 1x/2x の記述子と、2x が 2 倍幅で生成されることを検証する。
// 高精細ディスプレイで同じ枠のより鮮明な画像が使われるようにするため。
func TestGetSrcSet(t *testing.T) {
	t.Parallel()

	srcSet := newTestHelper().GetSrcSet("s3://test-bucket/shrine/a.jpg", 70, "webp")

	if !strings.Contains(srcSet, " 1x, ") || !strings.HasSuffix(srcSet, " 2x") {
		t.Errorf("GetSrcSet() = %q, want 1x と 2x の記述子", srcSet)
	}
	if !strings.Contains(srcSet, "resize:fit:70:93:0") || !strings.Contains(srcSet, "resize:fit:140:186:0") {
		t.Errorf("GetSrcSet() = %q, want 70x93 (1x) と 140x186 (2x) のリサイズ", srcSet)
	}

	if got := newTestHelper().GetSrcSet("", 70, "webp"); got != "" {
		t.Errorf("GetSrcSet(\"\") = %q, want 空文字列", got)
	}
}

// TestGetAvatarImageURL verifies that avatars keep their own 1:1 fill-down treatment, so
// changing the work-image resizing mode does not leak into them.
//
// [Ja] TestGetAvatarImageURL はアバターが 1:1 の fill-down のままであることを検証する。
// 作品画像のリサイズ方式の変更が波及していないことを担保する。
func TestGetAvatarImageURL(t *testing.T) {
	t.Parallel()

	url := newTestHelper().GetAvatarImageURL(testImageDataJSON, 50, "webp")

	if !strings.Contains(url, "resize:fill-down:50:50:0") {
		t.Errorf("アバターの URL に resize:fill-down:50:50:0 が含まれるべきです: %s", url)
	}
}
