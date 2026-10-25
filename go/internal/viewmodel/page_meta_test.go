package viewmodel

import (
	"context"
	"slices"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// TestDefaultPageMetaはDefaultPageMeta関数のテスト
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
				t.Errorf("Title = %q、期待値 = %q", meta.Title, tt.expectedTitle)
			}

			// Descriptionの確認 (一部の文字列が含まれているか)
			switch tt.locale {
			case i18n.LangJa:
				if meta.Description == "" {
					t.Error("日本語ロケールのmeta.Descriptionが空だった")
				}
				// 日本語の一部が含まれていることを確認
				if len(meta.Description) < 10 {
					t.Errorf("meta.Description = %q、期待値 = 10バイト以上", meta.Description)
				}
			case i18n.LangEn:
				if meta.Description == "" {
					t.Error("英語ロケールのmeta.Descriptionが空だった")
				}
				// 英語の一部が含まれていることを確認
				if len(meta.Description) < 10 {
					t.Errorf("meta.Description = %q、期待値 = 10バイト以上", meta.Description)
				}
			}

			// OGTypeのデフォルト値を確認
			if meta.OGType != "website" {
				t.Errorf("OGType = %q、期待値 = %q", meta.OGType, "website")
			}

			expectedCanonicalURL := "https://test.annict.com/works/popular"
			if meta.CanonicalURL != expectedCanonicalURL {
				t.Errorf("CanonicalURL = %q、期待値 = %q", meta.CanonicalURL, expectedCanonicalURL)
			}

			// OGImageが正しく設定されていることを確認
			expectedOGImage := "https://test.annict.com/static/images/og-image.png"
			if meta.OGImage != expectedOGImage {
				t.Errorf("OGImage = %q、期待値 = %q", meta.OGImage, expectedOGImage)
			}
		})
	}
}

// TestDefaultPageMetaWithoutLocaleはロケールが設定されていない場合のテスト
func TestDefaultPageMetaWithoutLocale(t *testing.T) {
	// テスト用のconfigを作成
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	// ロケールを設定せずに呼び出し (デフォルトは日本語)
	ctx := context.Background()
	meta := DefaultPageMeta(ctx, cfg, "/works/popular")

	// タイトルが設定されていることを確認
	if meta.Title == "" {
		t.Error("meta.Titleが空だった")
	}

	// Descriptionが設定されていることを確認
	if meta.Description == "" {
		t.Error("meta.Descriptionが空だった")
	}

	// OGTypeがデフォルト値であることを確認
	if meta.OGType != "website" {
		t.Errorf("OGType = %q、期待値 = %q", meta.OGType, "website")
	}

	// OGImageが設定されていることを確認
	expectedOGImage := "https://test.annict.com/static/images/og-image.png"
	if meta.OGImage != expectedOGImage {
		t.Errorf("OGImage = %q、期待値 = %q", meta.OGImage, expectedOGImage)
	}
}

// TestDefaultPageMeta_CanonicalURLはcanonical URLがそのページ自身の絶対URL
// (設定されたオリジン + リクエストパス) になることを、公開ページとAnnict DB管理画面の
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
			name: "Annict DBの画面",
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
				t.Errorf("CanonicalURL = %q、期待値 = %q", meta.CanonicalURL, tt.want)
			}
		})
	}
}

// TestPageMeta_SetTitleはSetTitleメソッドのテスト
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
				t.Errorf("Title = %q、期待値 = %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

// TestPageMeta_SetDBTitleはSetDBTitleが各ロケールで " | Annict DB" サフィックスを付けることを検証します。
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
				t.Errorf("Title = %q、期待値 = %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

// TestPageMeta_AddTurnstilePreconnectはヒントがウィジェットの描画に追従することを
// 検証する。components.Turnstileはsite keyがあるときだけ描画するため、空のキーでは
// PreconnectOriginsを変えてはならない。
func TestPageMeta_AddTurnstilePreconnect(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	tests := []struct {
		name    string
		siteKey string
		want    []string
	}{
		{
			name:    "site keyがあるときはオリジンを宣言する",
			siteKey: "1x00000000000000000000AA",
			want:    []string{"https://challenges.cloudflare.com"},
		},
		{
			name:    "site keyが空のときは宣言しない",
			siteKey: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), i18n.LangJa)
			meta := DefaultPageMeta(ctx, cfg, "/sign_in")
			meta.AddTurnstilePreconnect(tt.siteKey)

			if len(meta.PreconnectOrigins) != len(tt.want) {
				t.Fatalf("PreconnectOrigins = %v、期待値 = %v", meta.PreconnectOrigins, tt.want)
			}
			for i, want := range tt.want {
				if meta.PreconnectOrigins[i] != want {
					t.Errorf("PreconnectOrigins[%d] = %q、期待値 = %q", i, meta.PreconnectOrigins[i], want)
				}
			}
		})
	}
}

// TestPageMeta_SetTitleWithoutSuffixはSetTitleWithoutSuffixメソッドのテスト
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
			name:          "日本語環境でのタイトル設定 (サフィックスなし)",
			locale:        i18n.LangJa,
			titleKey:      "default_title",
			expectedTitle: "Annict",
		},
		{
			name:          "英語環境でのタイトル設定 (サフィックスなし)",
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
				t.Errorf("Title = %q、期待値 = %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

// TestDefaultPageMeta_LanguageMetaはURLが言語を持たないページの既定値を検証する。
// リクエストのロケールが英語でもLocaleは既定言語のままで、言語版の一覧とnoindexは持たない。
// 既存ページのog:locale・hreflang・canonicalの出力を変えないため。
func TestDefaultPageMeta_LanguageMeta(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	for _, locale := range []string{i18n.LangJa, i18n.LangEn} {
		t.Run(locale, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), locale)
			meta := DefaultPageMeta(ctx, cfg, "/works/popular")

			if meta.Locale != i18n.LangJa {
				t.Errorf("Locale = %q、期待値 = %q", meta.Locale, i18n.LangJa)
			}
			if meta.LanguageVersions != nil {
				t.Errorf("LanguageVersions = %v、期待値 = nil", meta.LanguageVersions)
			}
			if meta.NoIndex {
				t.Error("NoIndex = true、期待値 = false")
			}
		})
	}
}

// TestPageMeta_OGLocaleはLocaleがog:localeの値 (言語_地域の形) に変換されることを検証する。
// 未設定や翻訳を持たない言語は既定言語の値になる。
func TestPageMeta_OGLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "日本語", locale: i18n.LangJa, want: "ja_JP"},
		{name: "英語", locale: i18n.LangEn, want: "en_US"},
		{name: "未設定は既定言語", locale: "", want: "ja_JP"},
		{name: "翻訳を持たない言語は既定言語", locale: "fr", want: "ja_JP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			meta := PageMeta{Locale: tt.locale}

			if got := meta.OGLocale(); got != tt.want {
				t.Errorf("OGLocale() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

// TestPageMeta_Robotsは、NoIndex・NoFollow・NoArchiveのうちtrueのものだけを
// <meta name="robots"> のcontentの形にまとめることを検証する。
func TestPageMeta_Robots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		meta PageMeta
		want string
	}{
		{name: "どれも指定しない", meta: PageMeta{}, want: ""},
		{name: "noindexだけ", meta: PageMeta{NoIndex: true}, want: "noindex"},
		{name: "nofollowだけ", meta: PageMeta{NoFollow: true}, want: "nofollow"},
		{name: "noarchiveだけ", meta: PageMeta{NoArchive: true}, want: "noarchive"},
		{
			name: "すべて指定する",
			meta: PageMeta{NoIndex: true, NoFollow: true, NoArchive: true},
			want: "noindex,nofollow,noarchive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.meta.Robots(); got != tt.want {
				t.Errorf("Robots() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

// TestPageMeta_AddLanguageVersionは言語版のURLが、canonicalと同じ規則で正規化した
// 絶対URLになることを検証する。
func TestPageMeta_AddLanguageVersion(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	var meta PageMeta
	meta.AddLanguageVersion(cfg, i18n.LangJa, "/terms")
	meta.AddLanguageVersion(cfg, i18n.LangEn, "/en/terms/")

	want := []AlternateLink{
		{HrefLang: "ja", URL: "https://test.annict.com/terms"},
		{HrefLang: "en", URL: "https://test.annict.com/en/terms"},
	}
	if !slices.Equal(meta.LanguageVersions, want) {
		t.Errorf("LanguageVersions = %v、期待値 = %v", meta.LanguageVersions, want)
	}
}

// TestPageMeta_AlternateLinksは言語版が2つ以上あるときだけhreflangのリンクを返し、
// 末尾のx-defaultが既定言語の版を指すことを検証する。
func TestPageMeta_AlternateLinks(t *testing.T) {
	t.Parallel()

	jaVersion := AlternateLink{HrefLang: "ja", URL: "https://annict.com/terms"}
	enVersion := AlternateLink{HrefLang: "en", URL: "https://annict.com/en/terms"}
	xDefault := AlternateLink{HrefLang: "x-default", URL: "https://annict.com/terms"}

	tests := []struct {
		name     string
		versions []AlternateLink
		want     []AlternateLink
	}{
		{
			name:     "言語版の一覧が無いときは返さない",
			versions: nil,
			want:     nil,
		},
		{
			name:     "言語版が1つだけのときは返さない",
			versions: []AlternateLink{jaVersion},
			want:     nil,
		},
		{
			name:     "日本語版から見た一覧",
			versions: []AlternateLink{jaVersion, enVersion},
			want:     []AlternateLink{jaVersion, enVersion, xDefault},
		},
		{
			name:     "英語版から見た一覧でもx-defaultは日本語版を指す",
			versions: []AlternateLink{enVersion, jaVersion},
			want:     []AlternateLink{enVersion, jaVersion, xDefault},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			meta := PageMeta{LanguageVersions: tt.versions}

			if got := meta.AlternateLinks(); !slices.Equal(got, tt.want) {
				t.Errorf("AlternateLinks() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
