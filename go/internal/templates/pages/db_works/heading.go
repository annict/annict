package db_works

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/templates"
)

// headingOrFallback returns the text of a page heading: the trimmed title, falling back to
// the translation of fallbackKey while the title is blank or consists only of whitespace.
// Every /db page that puts a work title in its heading shares this rule, so none of them
// can render an empty <h1>.
//
// [Ja] headingOrFallback はページ見出しのテキストとして、前後の空白を除いたタイトルを返す。
// タイトルが空または空白文字だけのあいだは fallbackKey の翻訳へフォールバックする。
// 作品タイトルを見出しに置く /db の各ページがこの規則を共有するため、どのページも
// 空の <h1> を描画しない。
func headingOrFallback(ctx context.Context, title, fallbackKey string) string {
	if trimmed := strings.TrimSpace(title); trimmed != "" {
		return trimmed
	}

	return templates.T(ctx, fallbackKey)
}
