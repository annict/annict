package layouts

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/viewmodel"
)

// TestSimple_Rendering Simpleレイアウトが正常にレンダリングされることを確認
func TestSimple_Rendering(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	// i18nミドルウェアを経由してコンテキストを取得
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja")

	var ctx context.Context
	i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}))
	i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

	meta := viewmodel.DefaultPageMeta(ctx, cfg)
	meta.SetTitle(ctx, "test_page_title")

	// テストコンテンツ
	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div class=\"test-content\">Test Content</div>"))
		return err
	})

	// レンダリング
	var buf bytes.Buffer
	err := Simple(ctx, meta, nil, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 必要な要素が含まれているか確認
	checks := []string{
		"<!doctype html>",
		"<html lang=\"ja\">",
		"<head>",
		"<body class=\"min-h-screen flex items-center justify-center\">",
		"Test Content",
	}

	for _, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
		}
	}

	// ヘッダーやフッターは含まれないはず
	notExpected := []string{
		"<header",
		"<nav",
		"<footer",
	}

	for _, unexpected := range notExpected {
		if strings.Contains(html, unexpected) {
			t.Errorf("HTMLに含まれてはいけない要素が含まれています: %q", unexpected)
		}
	}
}

// TestSimple_WithFlash フラッシュメッセージが表示されることを確認
func TestSimple_WithFlash(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	// i18nミドルウェアを経由してコンテキストを取得
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja")

	var ctx context.Context
	i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}))
	i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

	meta := viewmodel.DefaultPageMeta(ctx, cfg)

	flash := &session.Flash{
		Type:    session.FlashError,
		Message: "エラーが発生しました",
	}

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	var buf bytes.Buffer
	err := Simple(ctx, meta, flash, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// フラッシュメッセージが含まれているか
	if !strings.Contains(html, "エラーが発生しました") {
		t.Error("フラッシュメッセージが表示されていません")
	}

	// フラッシュのtoaster要素が含まれているか
	if !strings.Contains(html, "toaster") {
		t.Error("フラッシュのtoaster要素が表示されていません")
	}
}

// TestSimple_WithoutFlash フラッシュメッセージがnilの場合の表示を確認
func TestSimple_WithoutFlash(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	// i18nミドルウェアを経由してコンテキストを取得
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja")

	var ctx context.Context
	i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}))
	i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

	meta := viewmodel.DefaultPageMeta(ctx, cfg)

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	var buf bytes.Buffer
	err := Simple(ctx, meta, nil, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// フラッシュのtoaster要素は含まれないはず
	if strings.Contains(html, "toaster") {
		t.Error("フラッシュメッセージがnilの場合、toaster要素は表示されないはず")
	}
}

// TestSimple_I18n 国際化対応が正しく動作することを確認
func TestSimple_I18n(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	tests := []struct {
		name           string
		acceptLanguage string
		langAttr       string
	}{
		{"日本語", "ja", "lang=\"ja\""},
		{"英語", "en", "lang=\"en\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// i18nミドルウェアを経由してコンテキストを取得
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Language", tt.acceptLanguage)

			var ctx context.Context
			i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx = r.Context()
			}))
			i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

			meta := viewmodel.DefaultPageMeta(ctx, cfg)

			content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				_, err := w.Write([]byte("<div>Content</div>"))
				return err
			})

			var buf bytes.Buffer
			err := Simple(ctx, meta, nil, "v1.0.0", content).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// 言語属性が正しく設定されているか
			if !strings.Contains(html, tt.langAttr) {
				t.Errorf("言語属性が正しく設定されていません: 期待=%s", tt.langAttr)
			}
		})
	}
}

// TestSimple_AssetVersion アセットバージョンが正しく設定されることを確認
func TestSimple_AssetVersion(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	// i18nミドルウェアを経由してコンテキストを取得
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja")

	var ctx context.Context
	i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}))
	i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

	meta := viewmodel.DefaultPageMeta(ctx, cfg)

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	assetVersion := "v1.2.3"

	var buf bytes.Buffer
	err := Simple(ctx, meta, nil, assetVersion, content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// アセットバージョンがURLに含まれているか
	expectedCSS := "/static/css/style.css?v=" + assetVersion
	expectedJS := "/static/js/main.js?v=" + assetVersion

	if !strings.Contains(html, expectedCSS) {
		t.Errorf("CSSのアセットバージョンが正しく設定されていません: 期待=%s", expectedCSS)
	}

	if !strings.Contains(html, expectedJS) {
		t.Errorf("JSのアセットバージョンが正しく設定されていません: 期待=%s", expectedJS)
	}
}
