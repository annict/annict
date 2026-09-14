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

// TestWorkImageHeightは作品画像の縦長比率 (横3 : 縦4) を検証する。
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
			t.Errorf("WorkImageHeight(%d) = %d、期待値 = %d", tt.width, got, tt.want)
		}
	}
}

// TestGenerateImgproxyURL_ResizesWithFitはリサイズ方式の歯止め。作品画像は元の
// アスペクト比を保つ "fit" でリサイズしなければならず、3:4の枠を埋めるために横長画像の
// 上下を切り落とす "fill" を使ってはならない。
func TestGenerateImgproxyURL_ResizesWithFit(t *testing.T) {
	t.Parallel()

	url := newTestHelper().GenerateImgproxyURL("s3://test-bucket/shrine/a.jpg", 70, "jpg")

	if !strings.Contains(url, "resize:fit:70:93:0") {
		t.Errorf("URLにresize:fit:70:93:0が含まれるべきです: %s", url)
	}

	// "resize:fill:" は切り抜きになる。"fill-down" は1:1のアバター専用。
	if strings.Contains(url, "resize:fill") {
		t.Errorf("作品画像でfill系のリサイズを使ってはいけません: %s", url)
	}
}

// TestGenerateImgproxyURL_Formatはjpg以外のフォーマットがformatオプションとして
// 付与され、既定の出力であるjpgでは付与されないことを検証する。
func TestGenerateImgproxyURL_Format(t *testing.T) {
	t.Parallel()

	h := newTestHelper()
	originalURL := "s3://test-bucket/shrine/a.jpg"

	webp := h.GenerateImgproxyURL(originalURL, 70, "webp")
	if !strings.Contains(webp, "format:webp") || !strings.HasSuffix(webp, ".webp") {
		t.Errorf("webpのURL = %s、期待値 = format:webpと.webp拡張子", webp)
	}

	jpg := h.GenerateImgproxyURL(originalURL, 70, "jpg")
	if strings.Contains(jpg, "format:") || !strings.HasSuffix(jpg, ".jpg") {
		t.Errorf("jpgのURL = %s、期待値 = formatオプション無しと.jpg拡張子", jpg)
	}
}

// TestGenerateImgproxyURL_EmptyOriginalは元URLが空のとき、実体を指さない署名付き
// URLではなく空文字列を返すことを検証する。
func TestGenerateImgproxyURL_EmptyOriginal(t *testing.T) {
	t.Parallel()

	if got := newTestHelper().GenerateImgproxyURL("", 70, "jpg"); got != "" {
		t.Errorf("GenerateImgproxyURL(\"\") = %q、期待値 = 空文字列", got)
	}
}

// TestGetWorkImageURLはimage_dataが署名付きのimgproxy URLに解決されること、
// データが無い / 壊れている場合は呼び出し側がプレースホルダーへ退避できるよう "" を返すことを
// 検証する。
func TestGetWorkImageURL(t *testing.T) {
	t.Parallel()

	h := newTestHelper()

	if got := h.GetWorkImageURL(testImageDataJSON, 70, "jpg"); !strings.HasPrefix(got, "http://localhost:18080/") {
		t.Errorf("GetWorkImageURL() = %q、期待値 = imgproxyエンドポイント始まりのURL", got)
	}

	tests := []struct {
		name          string
		imageDataJSON string
	}{
		{name: "空のimage_data", imageDataJSON: ""},
		{name: "壊れたJSON", imageDataJSON: "{"},
		{name: "idが空", imageDataJSON: `{"master":{"id":"","storage":"store"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := h.GetWorkImageURL(tt.imageDataJSON, 70, "jpg"); got != "" {
				t.Errorf("GetWorkImageURL() = %q、期待値 = 空文字列", got)
			}
		})
	}
}

// TestGetSrcSetは1x/2xの記述子と、2xが2倍幅で生成されることを検証する。
// 高精細ディスプレイで同じ枠のより鮮明な画像が使われるようにするため。
func TestGetSrcSet(t *testing.T) {
	t.Parallel()

	srcSet := newTestHelper().GetSrcSet("s3://test-bucket/shrine/a.jpg", 70, "webp")

	if !strings.Contains(srcSet, " 1x, ") || !strings.HasSuffix(srcSet, " 2x") {
		t.Errorf("GetSrcSet() = %q、期待値 = 1xと2xの記述子", srcSet)
	}
	if !strings.Contains(srcSet, "resize:fit:70:93:0") || !strings.Contains(srcSet, "resize:fit:140:186:0") {
		t.Errorf("GetSrcSet() = %q、期待値 = 70x93 (1x) と140x186 (2x) のリサイズ", srcSet)
	}

	if got := newTestHelper().GetSrcSet("", 70, "webp"); got != "" {
		t.Errorf("GetSrcSet(\"\") = %q、期待値 = 空文字列", got)
	}
}

// TestGetAvatarImageURLはアバターが1:1のfill-downのままであることを検証する。
// 作品画像のリサイズ方式の変更が波及していないことを担保する。
func TestGetAvatarImageURL(t *testing.T) {
	t.Parallel()

	url := newTestHelper().GetAvatarImageURL(testImageDataJSON, 50, "webp")

	if !strings.Contains(url, "resize:fill-down:50:50:0") {
		t.Errorf("アバターのURLにresize:fill-down:50:50:0が含まれるべきです: %s", url)
	}
}
