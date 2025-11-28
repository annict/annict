package viewmodel

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
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
			meta := DefaultPageMeta(ctx, cfg)

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

			// OGURLが空であることを確認（デフォルト値）
			if meta.OGURL != "" {
				t.Errorf("OGURL: got %q, want empty string", meta.OGURL)
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
	meta := DefaultPageMeta(ctx, cfg)

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
			meta := DefaultPageMeta(ctx, cfg)
			meta.SetTitle(ctx, tt.titleKey)

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
			meta := DefaultPageMeta(ctx, cfg)
			meta.SetTitleWithoutSuffix(ctx, tt.titleKey)

			if meta.Title != tt.expectedTitle {
				t.Errorf("Title: got %q, want %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}
