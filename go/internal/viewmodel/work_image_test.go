package viewmodel

import (
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

const testWorkImageData = `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`

// TestWorkImage_WithImageは画像が登録されている作品が、プレースホルダーではなく
// imgproxyのURLと1x/2xのsrcsetに解決されることを検証する。
func TestWorkImage_WithImage(t *testing.T) {
	t.Parallel()

	img := NewWorkImage(testWorkImageData, testutil.NewTestImageHelper())

	if !img.Exists() {
		t.Fatal("Exists() = false、期待値 = true")
	}

	url := img.URL(70, "jpg")
	if url == NoWorkImagePath {
		t.Errorf("URL()がプレースホルダーを返しています: %q", url)
	}
	if !strings.HasSuffix(url, ".jpg") {
		t.Errorf("URL() = %q、期待値 = .jpgで終わるimgproxy URL", url)
	}

	srcSet := img.SrcSet(70, "webp")
	if !strings.Contains(srcSet, " 1x, ") || !strings.HasSuffix(srcSet, " 2x") {
		t.Errorf("SrcSet() = %q、期待値 = 1xと2xを含むsrcset", srcSet)
	}
}

// TestWorkImage_WithoutImageはフォールバックを検証する。image_dataが無い作品は
// 静的なプレースホルダーに解決され、srcsetは生成されない (サイズ違いの派生が無いため)。
func TestWorkImage_WithoutImage(t *testing.T) {
	t.Parallel()

	img := NewWorkImage("", testutil.NewTestImageHelper())

	if img.Exists() {
		t.Error("Exists() = true、期待値 = false")
	}
	if got := img.URL(70, "jpg"); got != NoWorkImagePath {
		t.Errorf("URL() = %q、期待値 = %q", got, NoWorkImagePath)
	}
	if got := img.SrcSet(70, "webp"); got != "" {
		t.Errorf("SrcSet() = %q、期待値 = 空文字列", got)
	}
}

// TestWorkImage_NilHelperはゼロ値のWorkImage (構造体リテラルなど画像ヘルパーが
// 未配線の場合) がpanicせずプレースホルダーにフォールバックすることを検証する。
func TestWorkImage_NilHelper(t *testing.T) {
	t.Parallel()

	img := NewWorkImage(testWorkImageData, nil)

	if img.Exists() {
		t.Error("Exists() = true、期待値 = false")
	}
	if got := img.URL(70, "jpg"); got != NoWorkImagePath {
		t.Errorf("URL() = %q、期待値 = %q", got, NoWorkImagePath)
	}
	if got := img.SrcSet(70, "webp"); got != "" {
		t.Errorf("SrcSet() = %q、期待値 = 空文字列", got)
	}
}

// TestWorkImage_Heightは枠の高さが3:4の作品画像表示枠に従い、呼び出し側の
// width / height属性がその枠を確保する (元画像は枠内へ収められ、枠を埋めるよう切り抜かれない)
// ことを検証する。
func TestWorkImage_Height(t *testing.T) {
	t.Parallel()

	img := NewWorkImage("", nil)

	tests := []struct {
		width int
		want  int
	}{
		{width: 70, want: 93},
		{width: 280, want: 373},
	}

	for _, tt := range tests {
		if got := img.Height(tt.width); got != tt.want {
			t.Errorf("Height(%d) = %d、期待値 = %d", tt.width, got, tt.want)
		}
	}
}
