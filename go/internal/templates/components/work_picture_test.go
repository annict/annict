package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/viewmodel"
)

const workPictureImageData = `{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`

func renderWorkPicture(t *testing.T, data WorkPictureData) string {
	t.Helper()

	var buf strings.Builder
	if err := WorkPicture(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	return buf.String()
}

// TestWorkPicture_WithImage verifies that a work with a registered image renders the
// webp / jpeg sources and the caller's alt text and classes, and that it is fitted inside
// the 3:4 slot (object-contain) rather than cropped to fill it.
//
// [Ja] TestWorkPicture_WithImage は画像が登録されている作品が webp / jpeg のソースと
// 呼び出し側の alt・クラスを描画し、枠を埋めるよう切り抜かれるのではなく 3:4 の枠内に
// 収められる (object-contain) ことを検証する。
func TestWorkPicture_WithImage(t *testing.T) {
	t.Parallel()

	html := renderWorkPicture(t, WorkPictureData{
		Image: viewmodel.NewWorkImage(workPictureImageData, testutil.NewTestImageHelper()),
		Width: 70,
		Alt:   "テストアニメ",
		Class: "rounded border",
	})

	expected := []string{
		`type="image/webp"`,
		`type="image/jpeg"`,
		`alt="テストアニメ"`,
		`width="70"`,
		`height="93"`,
		`loading="lazy"`,
		`class="rounded border object-contain"`,
		`style="width:70px;height:93px;"`,
	}

	for _, want := range expected {
		if !strings.Contains(html, want) {
			t.Errorf("期待する文字列が含まれていません: %q", want)
		}
	}

	if strings.Contains(html, viewmodel.NoWorkImagePath) {
		t.Error("画像がある作品ではプレースホルダーを描画してはいけません")
	}
}

// TestWorkPicture_WithoutImage verifies the fallback: the placeholder is rendered without
// <source> elements (it has no size variants), in the same box as a real thumbnail, and
// with an empty alt so it is announced as decorative rather than as a picture of the work.
//
// [Ja] TestWorkPicture_WithoutImage はフォールバックを検証する。プレースホルダーは
// <source> 無し (サイズ違いの派生が無いため) で実サムネイルと同じ枠に描画され、作品を写した
// 画像ではなく装飾として扱われるよう alt は空になる。
func TestWorkPicture_WithoutImage(t *testing.T) {
	t.Parallel()

	html := renderWorkPicture(t, WorkPictureData{
		Image: viewmodel.NewWorkImage("", testutil.NewTestImageHelper()),
		Width: 70,
		Alt:   "テストアニメ",
		Class: "rounded border",
	})

	expected := []string{
		`src="` + viewmodel.NoWorkImagePath + `"`,
		`alt=""`,
		`width="70"`,
		`height="93"`,
		`class="rounded border object-contain"`,
		`style="width:70px;height:93px;"`,
	}

	for _, want := range expected {
		if !strings.Contains(html, want) {
			t.Errorf("期待する文字列が含まれていません: %q", want)
		}
	}

	if strings.Contains(html, "<source") {
		t.Error("プレースホルダーには <source> を描画してはいけません")
	}
}

// TestWorkPicture_WidthDrivesHeight verifies that the height attribute follows the width
// at the work-image ratio, so one component serves different display sizes.
//
// [Ja] TestWorkPicture_WidthDrivesHeight は height 属性が作品画像の比率で width に
// 追従し、1 つのコンポーネントで異なる表示サイズに対応できることを検証する。
func TestWorkPicture_WidthDrivesHeight(t *testing.T) {
	t.Parallel()

	html := renderWorkPicture(t, WorkPictureData{
		Image: viewmodel.NewWorkImage(workPictureImageData, testutil.NewTestImageHelper()),
		Width: 280,
	})

	if !strings.Contains(html, `width="280"`) || !strings.Contains(html, `height="373"`) {
		t.Errorf("width=280 に対して height=373 が描画されるべきです: %s", html)
	}
}

// TestWorkPicture_NoClass verifies that omitting Class leaves only the object-fit class,
// with no stray leading space in the class attribute.
//
// [Ja] TestWorkPicture_NoClass は Class を省略したとき object-fit のクラスだけが残り、
// class 属性の先頭に余分な空白が入らないことを検証する。
func TestWorkPicture_NoClass(t *testing.T) {
	t.Parallel()

	html := renderWorkPicture(t, WorkPictureData{
		Image: viewmodel.NewWorkImage(workPictureImageData, testutil.NewTestImageHelper()),
		Width: 70,
	})

	if !strings.Contains(html, `class="object-contain"`) {
		t.Errorf(`class="object-contain" が描画されるべきです: %s`, html)
	}
}
