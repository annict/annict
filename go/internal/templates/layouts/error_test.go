package layouts

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// errorTestContext returns a context carrying the locale the i18n middleware would have
// resolved from acceptLanguage. The middleware itself lives in internal/middleware, which
// depends on this package through internal/httperror, so the resolution is reproduced here
// rather than imported.
//
// [Ja] errorTestContext は i18n ミドルウェアが acceptLanguage から解決したはずのロケールを
// 持つコンテキストを返す。ミドルウェア自体は internal/middleware にあり、同パッケージは
// internal/httperror 経由で本パッケージに依存するため、import せずに解決処理を再現している。
func errorTestContext(acceptLanguage string) context.Context {
	req := httptest.NewRequest("GET", "/missing", nil)
	req.Header.Set("Accept-Language", acceptLanguage)

	return i18n.SetLocale(req.Context(), i18n.DetectLanguage(req))
}

func errorTestContent() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<div class="error-card">Content</div>`))
		return err
	})
}

// TestError_Rendering verifies the document shell: the title carries the site suffix, the
// content is placed in the centered main region, and the styles the body relies on are
// inlined.
//
// [Ja] TestError_Rendering は文書の枠を検証する。タイトルにサイトのサフィックスが付くこと、
// 本文が中央寄せの main 領域へ入ること、本文が頼るスタイルがインラインで出ることを確認する。
func TestError_Rendering(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("ページが見つかりません", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	checks := []string{
		"<!doctype html>",
		`<html lang="ja">`,
		"<title>ページが見つかりません" + viewmodel.TitleSuffix + "</title>",
		`<main class="error-page">`,
		"Content",
		// The styles the body's classes depend on ship with the document.
		//
		// [Ja] 本文のクラスが依存するスタイルは文書と一緒に配信される。
		"<style>",
		".error-card {",
		".error-link {",
		// Both themes are covered, since the theme script only toggles the class.
		//
		// [Ja] テーマスクリプトはクラスを切り替えるだけなので、両テーマの定義を持つ。
		".dark {",
	}

	for _, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
		}
	}
}

// TestError_IsSelfContained verifies that the page requests nothing that is built or deployed
// separately: an error page has to render while the rest of the application may be failing.
//
// [Ja] TestError_IsSelfContained は、別途ビルド・デプロイされるものを一切リクエストしないこと
// を検証する。エラーページはアプリケーションの他の部分が失敗しているあいだも描画できる必要が
// あるため。
func TestError_IsSelfContained(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("問題が発生しました", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	notExpected := []string{
		"/static/css/style.css",
		"/static/js/main.js",
		"htmx",
		"<link rel=\"manifest\"",
	}

	for _, unexpected := range notExpected {
		if strings.Contains(html, unexpected) {
			t.Errorf("エラーページが外部アセットを参照しています: %q", unexpected)
		}
	}
}

// TestError_ThemeScript verifies that the shared theme script runs here too, so the error page
// follows the reader's color scheme like every other page.
//
// [Ja] TestError_ThemeScript は共有のテーマスクリプトがここでも実行されることを検証する。
// エラーページも他のページと同じく閲覧者の配色に追随するため。
func TestError_ThemeScript(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("ページが見つかりません", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), `classList.add("dark")`) {
		t.Error("テーマ判定スクリプトが含まれていません")
	}
}

// TestError_DarkThemeLinkContrast verifies that the dark-theme link colors stay at the
// WCAG AA contrast pair selected for normal-sized text.
//
// [Ja] TestError_DarkThemeLinkContrast は、ダークテーマのリンク配色が通常サイズの文字向けに
// 選んだ WCAG AA 適合の組み合わせを維持することを検証する。
func TestError_DarkThemeLinkContrast(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("問題が発生しました", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()
	darkStart := strings.Index(html, ".dark {")
	if darkStart == -1 {
		t.Fatal("ダークテーマのスタイルが含まれていません")
	}
	darkEnd := strings.Index(html[darkStart:], "}")
	if darkEnd == -1 {
		t.Fatal("ダークテーマのスタイルが閉じられていません")
	}
	darkStyles := html[darkStart : darkStart+darkEnd]

	for _, expected := range []string{
		"--error-accent-hover: #de204c;",
		"--error-accent-fg: #ffffff;",
	} {
		if !strings.Contains(darkStyles, expected) {
			t.Errorf("AA 適合のリンク配色が含まれていません: %q", expected)
		}
	}
}

// TestError_ColorScheme verifies that the document declares both supported themes early for
// browser chrome and again in CSS for native controls.
//
// [Ja] TestError_ColorScheme は、ブラウザ UI とネイティブ部品のために、文書が対応する両テーマを
// head の早い位置と CSS の双方で宣言することを検証する。
func TestError_ColorScheme(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("問題が発生しました", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()
	for _, expected := range []string{
		`<meta name="color-scheme" content="light dark">`,
		"color-scheme: light dark;",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("対応配色の宣言が含まれていません: %q", expected)
		}
	}
}

// TestError_FullHeightFollowsVisibleViewport verifies that the centered region is sized by
// the visible viewport height, with the static unit kept ahead of it as the fallback.
//
// [Ja] TestError_FullHeightFollowsVisibleViewport は、中央寄せの領域の高さが可視ビューポート
// に追随し、静的な単位がその手前にフォールバックとして残ることを検証する。
func TestError_FullHeightFollowsVisibleViewport(t *testing.T) {
	t.Parallel()

	ctx := errorTestContext("ja")

	var buf bytes.Buffer
	if err := Error("ページが見つかりません", errorTestContent()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()
	fallback := strings.Index(html, "min-height: 100vh;")
	dynamic := strings.Index(html, "min-height: 100dvh;")

	if fallback == -1 {
		t.Error("フォールバックの min-height: 100vh が含まれていません")
	}
	if dynamic == -1 {
		t.Error("min-height: 100dvh が含まれていません")
	}
	if fallback != -1 && dynamic != -1 && fallback > dynamic {
		t.Error("フォールバックの 100vh は 100dvh より前に置く必要があります")
	}
}

// TestError_I18n verifies that the document language follows the resolved locale.
//
// [Ja] TestError_I18n は文書の言語が解決済みのロケールに追随することを検証する。
func TestError_I18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		acceptLanguage string
		langAttr       string
	}{
		{name: "日本語", acceptLanguage: "ja", langAttr: `lang="ja"`},
		{name: "英語", acceptLanguage: "en", langAttr: `lang="en"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := errorTestContext(tt.acceptLanguage)

			var buf bytes.Buffer
			if err := Error("Page not found", errorTestContent()).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			if !strings.Contains(buf.String(), tt.langAttr) {
				t.Errorf("言語属性が正しく設定されていません: 期待=%s", tt.langAttr)
			}
		})
	}
}
