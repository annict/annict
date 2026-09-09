package viewmodel

import (
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

const testWorkImageData = `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`

// TestWorkImage_WithImage verifies that a work with a registered image resolves to
// imgproxy URLs and a 1x/2x srcset rather than the placeholder.
//
// [Ja] TestWorkImage_WithImage は画像が登録されている作品が、プレースホルダーではなく
// imgproxy の URL と 1x/2x の srcset に解決されることを検証する。
func TestWorkImage_WithImage(t *testing.T) {
	t.Parallel()

	img := NewWorkImage(testWorkImageData, testutil.NewTestImageHelper())

	if !img.Exists() {
		t.Fatal("Exists() = false, want true")
	}

	url := img.URL(70, "jpg")
	if url == NoWorkImagePath {
		t.Errorf("URL() がプレースホルダーを返しています: %q", url)
	}
	if !strings.HasSuffix(url, ".jpg") {
		t.Errorf("URL() = %q, want .jpg で終わる imgproxy URL", url)
	}

	srcSet := img.SrcSet(70, "webp")
	if !strings.Contains(srcSet, " 1x, ") || !strings.HasSuffix(srcSet, " 2x") {
		t.Errorf("SrcSet() = %q, want 1x と 2x を含む srcset", srcSet)
	}
}

// TestWorkImage_WithoutImage verifies the fallback: a work with no image_data resolves
// to the static placeholder and produces no srcset (the placeholder has no size variants).
//
// [Ja] TestWorkImage_WithoutImage はフォールバックを検証する。image_data が無い作品は
// 静的なプレースホルダーに解決され、srcset は生成されない (サイズ違いの派生が無いため)。
func TestWorkImage_WithoutImage(t *testing.T) {
	t.Parallel()

	img := NewWorkImage("", testutil.NewTestImageHelper())

	if img.Exists() {
		t.Error("Exists() = true, want false")
	}
	if got := img.URL(70, "jpg"); got != NoWorkImagePath {
		t.Errorf("URL() = %q, want %q", got, NoWorkImagePath)
	}
	if got := img.SrcSet(70, "webp"); got != "" {
		t.Errorf("SrcSet() = %q, want 空文字列", got)
	}
}

// TestWorkImage_NilHelper verifies that a zero-value WorkImage (no image helper wired,
// as in struct literals) falls back to the placeholder instead of panicking.
//
// [Ja] TestWorkImage_NilHelper はゼロ値の WorkImage (構造体リテラルなど画像ヘルパーが
// 未配線の場合) が panic せずプレースホルダーにフォールバックすることを検証する。
func TestWorkImage_NilHelper(t *testing.T) {
	t.Parallel()

	img := NewWorkImage(testWorkImageData, nil)

	if img.Exists() {
		t.Error("Exists() = true, want false")
	}
	if got := img.URL(70, "jpg"); got != NoWorkImagePath {
		t.Errorf("URL() = %q, want %q", got, NoWorkImagePath)
	}
	if got := img.SrcSet(70, "webp"); got != "" {
		t.Errorf("SrcSet() = %q, want 空文字列", got)
	}
}

// TestWorkImage_Height verifies that the box height follows the 3:4 work-image display box,
// so the caller's width / height attributes reserve that box (the source is fitted inside it,
// never cropped to fill it).
//
// [Ja] TestWorkImage_Height は枠の高さが 3:4 の作品画像表示枠に従い、呼び出し側の
// width / height 属性がその枠を確保する (元画像は枠内へ収められ、枠を埋めるよう切り抜かれない)
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
			t.Errorf("Height(%d) = %d, want %d", tt.width, got, tt.want)
		}
	}
}
