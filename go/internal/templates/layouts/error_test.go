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

// errorTestContextはi18nミドルウェアがacceptLanguageから解決したはずのロケールを
// 持つコンテキストを返す。ミドルウェア自体はinternal/middlewareにあり、同パッケージは
// internal/httperror経由で本パッケージに依存するため、importせずに解決処理を再現している。
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

// TestError_Renderingは文書の枠を検証する。タイトルにサイトのサフィックスが付くこと、
// 本文が中央寄せのmain領域へ入ること、本文が頼るスタイルがインラインで出ることを確認する。
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
		// 本文のクラスが依存するスタイルは文書と一緒に配信される。
		"<style>",
		".error-card {",
		".error-link {",
		// テーマスクリプトはクラスを切り替えるだけなので、両テーマの定義を持つ。
		".dark {",
	}

	for _, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
		}
	}
}

// TestError_IsSelfContainedは、別途ビルド・デプロイされるものを一切リクエストしないこと
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

// TestError_ThemeScriptは共有のテーマスクリプトがここでも実行されることを検証する。
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

// TestError_DarkThemeLinkContrastは、ダークテーマのリンク配色が通常サイズの文字向けに
// 選んだWCAG AA適合の組み合わせを維持することを検証する。
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
			t.Errorf("AA適合のリンク配色が含まれていません: %q", expected)
		}
	}
}

// TestError_ColorSchemeは、ブラウザUIとネイティブ部品のために、文書が対応する両テーマを
// headの早い位置とCSSの双方で宣言することを検証する。
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

// TestError_FullHeightFollowsVisibleViewportは、中央寄せの領域の高さが可視ビューポート
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
		t.Error("フォールバックのmin-height: 100vhが含まれていません")
	}
	if dynamic == -1 {
		t.Error("min-height: 100dvhが含まれていません")
	}
	if fallback != -1 && dynamic != -1 && fallback > dynamic {
		t.Error("フォールバックの100vhは100dvhより前に置く必要があります")
	}
}

// TestError_I18nは文書の言語が解決済みのロケールに追随することを検証する。
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
