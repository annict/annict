package db_works

import (
	"context"

	"github.com/annict/annict/go/internal/templates"
)

// headingOrFallbackはページ見出しのテキストとして、ハンドラーが解決した作品の表示名を
// 返す。表示名が空のあいだはfallbackKeyの翻訳へフォールバックする。作品タイトルを見出しに
// 置く /dbの各ページがこの規則を共有するため、どのページも空の <h1> を描画しない。また各
// ページの文書タイトルも同じ解決済みの表示名から組み立てるため、対象を名指しできるかどうかの
// 判断が見出しと揃う。
func headingOrFallback(ctx context.Context, workName, fallbackKey string) string {
	if workName != "" {
		return workName
	}

	return templates.T(ctx, fallbackKey)
}
