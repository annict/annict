// Package viewmodelはビューモデル変換機能を提供します
package viewmodel

import (
	"context"
	"path"
	"slices"
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
	// LocaleはURLを代表する言語 (i18nの言語コード) で、og:localeに出力する。
	// URLで言語が決まるページでは、その言語を設定する。
	Locale string
	// LanguageVersionsは同じ内容の言語版の一覧で、ページ自身の版も含める。
	// AlternateLinksを通して <link rel="alternate" hreflang> に出力する。
	LanguageVersions []AlternateLink
	// NoIndexがtrueのとき <meta name="robots"> にnoindexを出力し、
	// 検索エンジンにページをインデックスさせない。canonicalも出力しない。
	NoIndex bool
	// NoFollowがtrueのとき <meta name="robots"> にnofollowを出力し、
	// 検索エンジンにページ内のリンクをたどらせない。
	NoFollow bool
	// NoArchiveがtrueのとき <meta name="robots"> にnoarchiveを出力し、
	// 検索エンジンにページのキャッシュを表示させない。
	NoArchive bool
}

// AlternateLinkは <link rel="alternate" hreflang> の1件を表す。
type AlternateLink struct {
	HrefLang string // hreflangの値 (BCP 47の言語タグ、またはx-default)
	URL      string // 言語版の絶対URL
}

// turnstileOriginはcomponents.Turnstileがウィジェットのスクリプトを読み込むオリジン。
const turnstileOrigin = "https://challenges.cloudflare.com"

// xDefaultHrefLangは、どの言語版にも合致しない利用者に向けた版を示すhreflangの値。
const xDefaultHrefLang = "x-default"

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
//
// Localeはリクエストのロケールではなく既定言語にする。URLが言語を持たないページは、
// Accept-Languageを送らないクローラーには既定言語で描画されるため、そのURLを代表する内容の
// 言語は既定言語になる。URLで言語が決まるページは、呼び出し側でその言語に上書きする。
func DefaultPageMeta(ctx context.Context, cfg *config.Config, requestPath string) PageMeta {
	ogImageURL := cfg.AppURL() + "/static/images/og-image.png"
	title := i18n.T(ctx, "default_title") + TitleSuffix
	return PageMeta{
		Title:        title,
		Description:  i18n.T(ctx, "default_description"),
		OGType:       "website",
		CanonicalURL: cfg.AppURL() + canonicalPath(requestPath),
		OGImage:      ogImageURL,
		Locale:       i18n.DefaultLang,
	}
}

// OGLocaleはLocaleをog:localeの値 (言語_地域の形) にして返す。翻訳を持たない言語や
// 未設定のときは既定言語の値を返す。
func (p PageMeta) OGLocale() string {
	switch p.Locale {
	case i18n.LangEn:
		return "en_US"
	default:
		return "ja_JP"
	}
}

// RobotsはNoIndex・NoFollow・NoArchiveのうちtrueのものを、<meta name="robots"> のcontentの
// 形 (カンマ区切り) にして返す。どれもfalseのときは空文字を返し、<meta name="robots"> を出力しない。
func (p PageMeta) Robots() string {
	var directives []string
	if p.NoIndex {
		directives = append(directives, "noindex")
	}
	if p.NoFollow {
		directives = append(directives, "nofollow")
	}
	if p.NoArchive {
		directives = append(directives, "noarchive")
	}

	return strings.Join(directives, ",")
}

// AddLanguageVersionは同じ内容の言語版をhreflangの対象に加える。pathはDefaultPageMetaの
// requestPathと同じ規則で正規化してから絶対URLにする。hreflangの指す先はcanonicalと同じ
// 正規URLでなければ、検索エンジンが言語版どうしを結び付けられないため。
func (p *PageMeta) AddLanguageVersion(cfg *config.Config, lang, path string) {
	p.LanguageVersions = append(p.LanguageVersions, AlternateLink{
		HrefLang: lang,
		URL:      cfg.AppURL() + canonicalPath(path),
	})
}

// AlternateLinksは <link rel="alternate" hreflang> として出力するリンクを、言語版の一覧の
// 末尾にx-defaultを加えて返す。x-defaultは日本語版を正本とする方針に合わせて既定言語の版を
// 指す。言語版が1つしかないページでは何も返さない。結び付ける相手の言語版が無いhreflangは
// 検索結果の出し分けに使われず、意味を持たないため。
func (p PageMeta) AlternateLinks() []AlternateLink {
	if len(p.LanguageVersions) < 2 {
		return nil
	}

	links := slices.Clone(p.LanguageVersions)
	for _, version := range p.LanguageVersions {
		if version.HrefLang == i18n.DefaultLang {
			links = append(links, AlternateLink{HrefLang: xDefaultHrefLang, URL: version.URL})
			break
		}
	}

	return links
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
