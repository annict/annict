// Package viewmodelはビューモデル変換機能を提供します
package viewmodel

import (
	"context"
	"path"
	"strings"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// PageMetaはページのメタ情報を保持します
type PageMeta struct {
	Title       string // ページタイトル (<title>タグ、og:title、twitter:title用)
	Description string // ページ説明 (description、og:description、twitter:description用)
	OGType      string // og:typeの値 ("website", "article"など)
	// CanonicalURLはそのページ自身の絶対URL。canonicalとog:urlの両方に出力し、
	// 両者が常に一致するようにする。
	CanonicalURL string
	OGImage      string // og:imageの値
	// PreconnectOriginsはそのページが確実にリクエストする第三者オリジンの一覧。
	// それぞれを <link rel="preconnect"> として出力し、リクエスト自体が見つかる前に
	// ブラウザがDNS解決とTCP / TLSのハンドシェイクを済ませられるようにする。宣言するのは
	// 実際にリクエストするページだけ。接触しないオリジンへのヒントは接続を無駄に使うため。
	PreconnectOrigins []string
}

// turnstileOriginはcomponents.Turnstileがウィジェットのスクリプトを読み込むオリジン。
const turnstileOrigin = "https://challenges.cloudflare.com"

const (
	// TitleSuffixはすべての文書タイトルの末尾に付き、ブラウザのタブがどのサイトの
	// ページかを示せるようにする。DBTitleSuffixはAnnict DB管理画面に対して同じ役割を
	// 果たし、タブや履歴で公開画面と区別できるようにする。
	//
	// いずれも公開しているのは、PageMetaの外でタイトルを組み立てるページがあるため
	// (共通のHTTPエラーページはPageMetaを持たずに描画する)。リテラルをここに集約して
	// おくことで、サイト名の変更が1箇所で済む。
	TitleSuffix   = " | Annict"
	DBTitleSuffix = " | Annict DB"
)

// DefaultPageMetaはrequestPathで配信されるページの既定のメタ情報を返す。タイトルと
// 説明はコンテキストから検出した言語に従い、タイトルには " | Annict" サフィックスが付く。
//
// requestPathは描画するページの代表パス。これが自己参照のcanonical URLになり、ハンドラー
// ごとに代入しなくても全ページにcanonicalが入る。通常のハンドラーはr.URL.Pathを渡し、
// 未知・追跡・機密クエリ (パスワードリセットの使い捨てトークンなど) を意図的に落とす。パスを
// 自前で組み立てるのは2つの場合。ページ自身のGETパスがリクエストパスと異なる場合 (POST
// やPATCHのエンドポイントが再描画するフォーム) と、既知のクエリパラメータが内容を変える
// 場合 (フィルタやページネーション) で、後者は呼び出し側がパース・正規化した値だけを付ける。
func DefaultPageMeta(ctx context.Context, cfg *config.Config, requestPath string) PageMeta {
	ogImageURL := cfg.AppURL() + "/static/images/og-image.png"
	title := i18n.T(ctx, "default_title") + TitleSuffix
	return PageMeta{
		Title:        title,
		Description:  i18n.T(ctx, "default_description"),
		OGType:       "website",
		CanonicalURL: cfg.AppURL() + canonicalPath(requestPath),
		OGImage:      ogImageURL,
	}
}

// canonicalPathはrequestPathのパス部分を正規化し、末尾スラッシュなど同じページを表す
// パス表現を1つの代表URLへ揃える。クエリ文字列はそのまま通す。path.Cleanは引数全体を
// パスとして扱うため、クエリ内の区切りやドットセグメントまで書き換えてしまうため。
func canonicalPath(requestPath string) string {
	pathPart, query, hasQuery := strings.Cut(requestPath, "?")
	cleaned := path.Clean(pathPart)
	if !hasQuery {
		return cleaned
	}
	return cleaned + "?" + query
}

// SetTitleは公開ページのタイトルを設定し、TitleSuffixを付ける。templateDataは翻訳に
// そのまま渡すため、対象のリソースを名指しするタイトルは、呼び出し側で文字列連結せず
// プレースホルダーで書ける。
func (p *PageMeta) SetTitle(ctx context.Context, titleKey string, templateData ...map[string]any) {
	p.Title = i18n.T(ctx, titleKey, templateData...) + TitleSuffix
}

// SetDBTitleはAnnict DB管理画面のページのタイトルを設定し、DBTitleSuffixを付ける。
// ブラウザのタブや履歴で公開画面と区別できるようにするため。templateDataはSetTitleと
// 同じく翻訳へそのまま渡す。
func (p *PageMeta) SetDBTitle(ctx context.Context, titleKey string, templateData ...map[string]any) {
	p.Title = i18n.T(ctx, titleKey, templateData...) + DBTitleSuffix
}

// AddTurnstilePreconnectはTurnstileのウィジェットを描画するページで、そのオリジンを
// preconnectの対象として宣言します。components.TurnstileはsiteKeyがあるときだけ
// ウィジェットとスクリプトを描画するため、siteKeyが空ならヒントも出しません。
func (p *PageMeta) AddTurnstilePreconnect(siteKey string) {
	if siteKey == "" {
		return
	}
	p.PreconnectOrigins = append(p.PreconnectOrigins, turnstileOrigin)
}

// SetTitleWithoutSuffixはタイトルを設定します (サフィックスなし)
// トップページなど、サフィックスが不要なページで使用します
func (p *PageMeta) SetTitleWithoutSuffix(ctx context.Context, titleKey string) {
	p.Title = i18n.T(ctx, titleKey)
}
