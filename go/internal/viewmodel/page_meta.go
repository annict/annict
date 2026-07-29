// Package viewmodel はビューモデル変換機能を提供します
package viewmodel

import (
	"context"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// PageMeta はページのメタ情報を保持します
type PageMeta struct {
	Title       string // ページタイトル（<title>タグ、og:title、twitter:title用）
	Description string // ページ説明（description、og:description、twitter:description用）
	OGType      string // og:typeの値（"website", "article"など）
	OGURL       string // og:urlの値
	OGImage     string // og:imageの値
}

// DefaultPageMeta はデフォルトのメタ情報を返します
// DetectLanguage()で検出された言語に応じて、タイトルと説明が自動的に切り替わります
// Titleには自動的に " | Annict" サフィックスが付加されます
func DefaultPageMeta(ctx context.Context, cfg *config.Config) PageMeta {
	ogImageURL := cfg.AppURL() + "/static/images/og-image.png"
	title := i18n.T(ctx, "default_title") + " | Annict"
	return PageMeta{
		Title:       title,
		Description: i18n.T(ctx, "default_description"),
		OGType:      "website",
		OGURL:       "",
		OGImage:     ogImageURL,
	}
}

// SetTitle はタイトルを設定します（" | Annict" サフィックス付き）
// 通常のページで使用します
func (p *PageMeta) SetTitle(ctx context.Context, titleKey string) {
	p.Title = i18n.T(ctx, titleKey) + " | Annict"
}

// SetDBTitle sets the title with the " | Annict DB" suffix.
// It is used for the Annict DB admin pages so that they stay distinguishable from
// the public pages in browser tabs and history.
//
// [Ja] SetDBTitle はタイトルを設定します (" | Annict DB" サフィックス付き)。
// Annict DB 管理画面のページで使い、ブラウザのタブや履歴で公開画面と区別できるようにします。
func (p *PageMeta) SetDBTitle(ctx context.Context, titleKey string) {
	p.Title = i18n.T(ctx, titleKey) + " | Annict DB"
}

// SetTitleWithoutSuffix はタイトルを設定します（サフィックスなし）
// トップページなど、サフィックスが不要なページで使用します
func (p *PageMeta) SetTitleWithoutSuffix(ctx context.Context, titleKey string) {
	p.Title = i18n.T(ctx, titleKey)
}
