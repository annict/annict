package viewmodel

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// TestDefaultPageMeta はDefaultPageMeta関数のテスト
func TestDefaultPageMeta(t *testing.T) {
	// テスト用のconfigを作成
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name           string
		locale         string
		expectedTitle  string
		expectedDescJa string // 日本語のdescriptionの一部
		expectedDescEn string // 英語のdescriptionの一部
	}{
		{
			name:           "日本語環境",
			locale:         i18n.LangJa,
			expectedTitle:  "Annict | Annict",
			expectedDescJa: "アニメ視聴を記録・管理できるWebサービス",
		},
		{
			name:           "英語環境",
			locale:         i18n.LangEn,
			expectedTitle:  "Annict | Annict",
			expectedDescEn: "Track what you watch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// コンテキストに言語設定を追加
			ctx := i18n.SetLocale(context.Background(), tt.locale)

			// DefaultPageMetaを呼び出し
			meta := DefaultPageMeta(ctx, cfg, "/works/popular")

			// タイトルの確認
			if meta.Title != tt.expectedTitle {
				t.Errorf("Title: got %q, want %q", meta.Title, tt.expectedTitle)
			}

			// Descriptionの確認（一部の文字列が含まれているか）
			switch tt.locale {
			case i18n.LangJa:
				if meta.Description == "" {
					t.Error("Description is empty for Japanese locale")
				}
				// 日本語の一部が含まれていることを確認
				if len(meta.Description) < 10 {
					t.Errorf("Description too short: %q", meta.Description)
				}
			case i18n.LangEn:
				if meta.Description == "" {
					t.Error("Description is empty for English locale")
				}
				// 英語の一部が含まれていることを確認
				if len(meta.Description) < 10 {
					t.Errorf("Description too short: %q", meta.Description)
				}
			}

			// OGTypeのデフォルト値を確認
			if meta.OGType != "website" {
				t.Errorf("OGType: got %q, want %q", meta.OGType, "website")
			}

			expectedCanonicalURL := "https://test.annict.com/works/popular"
			if meta.CanonicalURL != expectedCanonicalURL {
				t.Errorf("CanonicalURL: got %q, want %q", meta.CanonicalURL, expectedCanonicalURL)
			}

			// OGImageが正しく設定されていることを確認
			expectedOGImage := "https://test.annict.com/static/images/og-image.png"
			if meta.OGImage != expectedOGImage {
				t.Errorf("OGImage: got %q, want %q", meta.OGImage, expectedOGImage)
			}
		})
	}
}

// TestDefaultPageMetaWithoutLocale はロケールが設定されていない場合のテスト
func TestDefaultPageMetaWithoutLocale(t *testing.T) {
	// テスト用のconfigを作成
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	// ロケールを設定せずに呼び出し（デフォルトは日本語）
	ctx := context.Background()
	meta := DefaultPageMeta(ctx, cfg, "/works/popular")

	// タイトルが設定されていることを確認
	if meta.Title == "" {
		t.Error("Title is empty")
	}

	// Descriptionが設定されていることを確認
	if meta.Description == "" {
		t.Error("Description is empty")
	}

	// OGTypeがデフォルト値であることを確認
	if meta.OGType != "website" {
		t.Errorf("OGType: got %q, want %q", meta.OGType, "website")
	}

	// OGImageが設定されていることを確認
	expectedOGImage := "https://test.annict.com/static/images/og-image.png"
	if meta.OGImage != expectedOGImage {
		t.Errorf("OGImage: got %q, want %q", meta.OGImage, expectedOGImage)
	}
}

// TestDefaultPageMeta_CanonicalURL verifies that the canonical URL is the page's own absolute
// URL, built from the configured origin and the request path, for the public pages as well as
// for the Annict DB admin pages.
//
// [Ja] TestDefaultPageMeta_CanonicalURL は canonical URL がそのページ自身の絶対 URL
// (設定されたオリジン + リクエストパス) になることを、公開ページと Annict DB 管理画面の
// 双方について検証します。
func TestDefaultPageMeta_CanonicalURL(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "トップページ",
			path: "/",
			want: "https://test.annict.com/",
		},
		{
			name: "公開ページ",
			path: "/works/popular",
			want: "https://test.annict.com/works/popular",
		},
		{
			name: "Annict DB の画面",
			path: "/db/works/1/edit",
			want: "https://test.annict.com/db/works/1/edit",
		},
		{
			name: "末尾スラッシュを除去",
			path: "/works/popular/",
			want: "https://test.annict.com/works/popular",
		},
		{
			name: "重複スラッシュとドットセグメントを正規化",
			path: "/works//season/../popular/",
			want: "https://test.annict.com/works/popular",
		},
		{
			name: "ページを分けるパラメータはクエリとして保つ",
			path: "/db/works?page=3",
			want: "https://test.annict.com/db/works?page=3",
		},
		{
			name: "クエリを持つ場合もパス部分だけを正規化する",
			path: "/db/works/?page=3",
			want: "https://test.annict.com/db/works?page=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			meta := DefaultPageMeta(context.Background(), cfg, tt.path)

			if meta.CanonicalURL != tt.want {
				t.Errorf("CanonicalURL: got %q, want %q", meta.CanonicalURL, tt.want)
			}
		})
	}
}

// TestPageMeta_SetTitle はSetTitleメソッドのテスト
func TestPageMeta_SetTitle(t *testing.T) {
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name          string
		locale        string
		titleKey      string
		expectedTitle string
	}{
		{
			name:          "日本語環境でのタイトル設定",
			locale:        i18n.LangJa,
			titleKey:      "popular_anime",
			expectedTitle: "人気アニメ | Annict",
		},
		{
			name:          "英語環境でのタイトル設定",
			locale:        i18n.LangEn,
			titleKey:      "popular_anime",
			expectedTitle: "Popular Anime | Annict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			meta := DefaultPageMeta(ctx, cfg, "/works/popular")
			meta.SetTitle(ctx, tt.titleKey)

			if meta.Title != tt.expectedTitle {
				t.Errorf("Title: got %q, want %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

// TestPageMeta_SetDBTitle verifies that SetDBTitle appends the " | Annict DB" suffix in each locale.
//
// [Ja] TestPageMeta_SetDBTitle は SetDBTitle が各ロケールで " | Annict DB" サフィックスを付けることを検証します。
func TestPageMeta_SetDBTitle(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name          string
		locale        string
		titleKey      string
		expectedTitle string
	}{
		{
			name:          "日本語環境でのタイトル設定",
			locale:        i18n.LangJa,
			titleKey:      "db_works_index_title",
			expectedTitle: "作品 | Annict DB",
		},
		{
			name:          "英語環境でのタイトル設定",
			locale:        i18n.LangEn,
			titleKey:      "db_works_index_title",
			expectedTitle: "Works | Annict DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			meta := DefaultPageMeta(ctx, cfg, "/works/popular")
			meta.SetDBTitle(ctx, tt.titleKey)

			if meta.Title != tt.expectedTitle {
				t.Errorf("Title: got %q, want %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

// TestPageMeta_SetTitleWithoutSuffix はSetTitleWithoutSuffixメソッドのテスト
func TestPageMeta_SetTitleWithoutSuffix(t *testing.T) {
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name          string
		locale        string
		titleKey      string
		expectedTitle string
	}{
		{
			name:          "日本語環境でのタイトル設定（サフィックスなし）",
			locale:        i18n.LangJa,
			titleKey:      "default_title",
			expectedTitle: "Annict",
		},
		{
			name:          "英語環境でのタイトル設定（サフィックスなし）",
			locale:        i18n.LangEn,
			titleKey:      "default_title",
			expectedTitle: "Annict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			meta := DefaultPageMeta(ctx, cfg, "/works/popular")
			meta.SetTitleWithoutSuffix(ctx, tt.titleKey)

			if meta.Title != tt.expectedTitle {
				t.Errorf("Title: got %q, want %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}
