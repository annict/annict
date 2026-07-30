// Package viewmodel はビューモデル変換機能を提供します
package viewmodel

import (
	"context"
	"path"
	"strings"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// PageMeta はページのメタ情報を保持します
type PageMeta struct {
	Title       string // ページタイトル（<title>タグ、og:title、twitter:title用）
	Description string // ページ説明（description、og:description、twitter:description用）
	OGType      string // og:typeの値（"website", "article"など）
	// CanonicalURL is the page's own absolute URL. It is rendered as both the canonical link
	// and og:url so the two always agree.
	//
	// [Ja] CanonicalURL はそのページ自身の絶対 URL。canonical と og:url の両方に出力し、
	// 両者が常に一致するようにする。
	CanonicalURL string
	OGImage      string // og:imageの値
}

// DefaultPageMeta returns the default metadata for the page served at requestPath. The title
// and the description follow the language detected from the context, and the title carries
// the " | Annict" suffix.
//
// requestPath is the representative path of the page being rendered. It becomes the
// self-referential canonical URL, so every page gets one without each handler assigning it.
// Most handlers pass r.URL.Path, intentionally dropping unknown, tracking and sensitive query
// parameters such as a one-time password reset token. Two cases call for a hand-built path
// instead: a page whose own GET path differs from the request path (a form re-rendered by its
// POST or PATCH endpoint), and a page whose known query parameters change its content (filters
// or pagination), where the caller appends only values it has parsed and normalized.
//
// [Ja] DefaultPageMeta は requestPath で配信されるページの既定のメタ情報を返す。タイトルと
// 説明はコンテキストから検出した言語に従い、タイトルには " | Annict" サフィックスが付く。
//
// requestPath は描画するページの代表パス。これが自己参照の canonical URL になり、ハンドラー
// ごとに代入しなくても全ページに canonical が入る。通常のハンドラーは r.URL.Path を渡し、
// 未知・追跡・機密クエリ (パスワードリセットの使い捨てトークンなど) を意図的に落とす。パスを
// 自前で組み立てるのは 2 つの場合。ページ自身の GET パスがリクエストパスと異なる場合 (POST
// や PATCH のエンドポイントが再描画するフォーム) と、既知のクエリパラメータが内容を変える
// 場合 (フィルタやページネーション) で、後者は呼び出し側がパース・正規化した値だけを付ける。
func DefaultPageMeta(ctx context.Context, cfg *config.Config, requestPath string) PageMeta {
	ogImageURL := cfg.AppURL() + "/static/images/og-image.png"
	title := i18n.T(ctx, "default_title") + " | Annict"
	return PageMeta{
		Title:        title,
		Description:  i18n.T(ctx, "default_description"),
		OGType:       "website",
		CanonicalURL: cfg.AppURL() + canonicalPath(requestPath),
		OGImage:      ogImageURL,
	}
}

// canonicalPath normalizes the path portion of requestPath so that trailing slashes and
// equivalent path forms share one representative URL, and passes any query string through
// untouched: path.Clean treats its whole argument as a path and would rewrite separators and
// dot segments that appear inside the query.
//
// [Ja] canonicalPath は requestPath のパス部分を正規化し、末尾スラッシュなど同じページを表す
// パス表現を 1 つの代表 URL へ揃える。クエリ文字列はそのまま通す。path.Clean は引数全体を
// パスとして扱うため、クエリ内の区切りやドットセグメントまで書き換えてしまうため。
func canonicalPath(requestPath string) string {
	pathPart, query, hasQuery := strings.Cut(requestPath, "?")
	cleaned := path.Clean(pathPart)
	if !hasQuery {
		return cleaned
	}
	return cleaned + "?" + query
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
