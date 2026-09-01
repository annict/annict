package db_works

import (
	"context"

	"github.com/annict/annict/go/internal/templates"
)

// headingOrFallback returns the text of a page heading: the work's display name as the handler
// resolved it, falling back to the translation of fallbackKey while that name is empty. Every
// /db page that puts a work title in its heading shares this rule, so none of them can render
// an empty <h1>, and each page's document title, built from the same resolved name, agrees with
// the heading on whether the target can be named.
//
// [Ja] headingOrFallback はページ見出しのテキストとして、ハンドラーが解決した作品の表示名を
// 返す。表示名が空のあいだは fallbackKey の翻訳へフォールバックする。作品タイトルを見出しに
// 置く /db の各ページがこの規則を共有するため、どのページも空の <h1> を描画しない。また各
// ページの文書タイトルも同じ解決済みの表示名から組み立てるため、対象を名指しできるかどうかの
// 判断が見出しと揃う。
func headingOrFallback(ctx context.Context, workName, fallbackKey string) string {
	if workName != "" {
		return workName
	}

	return templates.T(ctx, fallbackKey)
}
