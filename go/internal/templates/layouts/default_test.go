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
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/viewmodel"
)

// TestDefault_Rendering Defaultレイアウトが正常にレンダリングされることを確認
func TestDefault_Rendering(t *testing.T) {
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
	err := Default(ctx, meta, nil, nil, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 必要な要素が含まれているか確認
	checks := []string{
		"<!doctype html>",
		"<html lang=\"ja\">",
		"<head>",
		"<body class=\"min-h-screen flex flex-col bg-gray-50\">",
		"<header class=\"bg-white shadow-sm\">",
		"<nav class=\"container mx-auto px-4\">",
		"Annict</a>",
		"<main class=\"flex-1 container mx-auto px-4 py-8\">",
		"<footer class=\"bg-gray-100 mt-auto\">",
		"Test Content",
		"&copy; 2024 Annict (Go Version)",
	}

	for _, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
		}
	}
}

// TestDefault_WithUser ユーザー情報が正しく表示されることを確認
func TestDefault_WithUser(t *testing.T) {
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

	// テストユーザー
	user := &query.GetUserByIDRow{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	var buf bytes.Buffer
	err := Default(ctx, meta, user, nil, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// ユーザー名が表示されているか
	if !strings.Contains(html, "testuser") {
		t.Error("ユーザー名が表示されていません")
	}

	// サインアウトリンクが表示されているか
	if !strings.Contains(html, "/sign_out") {
		t.Error("サインアウトリンクが表示されていません")
	}

	// サインインリンクは表示されないはず
	if strings.Contains(html, "/sign_in") {
		t.Error("ログイン中はサインインリンクは表示されないはず")
	}
}

// TestDefault_WithoutUser 未ログイン時の表示を確認
func TestDefault_WithoutUser(t *testing.T) {
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
	err := Default(ctx, meta, nil, nil, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// サインインリンクが表示されているか
	if !strings.Contains(html, "/sign_in") {
		t.Error("サインインリンクが表示されていません")
	}

	// サインアウトリンクは表示されないはず
	if strings.Contains(html, "/sign_out") {
		t.Error("未ログイン時はサインアウトリンクは表示されないはず")
	}
}

// TestDefault_WithFlash フラッシュメッセージが表示されることを確認
func TestDefault_WithFlash(t *testing.T) {
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
		Type:    session.FlashSuccess,
		Message: "操作が成功しました",
	}

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	var buf bytes.Buffer
	err := Default(ctx, meta, nil, flash, "v1.0.0", content).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// フラッシュメッセージが含まれているか
	if !strings.Contains(html, "操作が成功しました") {
		t.Error("フラッシュメッセージが表示されていません")
	}
}

// TestDefault_I18n 国際化対応が正しく動作することを確認
func TestDefault_I18n(t *testing.T) {
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
			err := Default(ctx, meta, nil, nil, "v1.0.0", content).Render(ctx, &buf)
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
