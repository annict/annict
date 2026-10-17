package legal_document

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/legaldoc"
)

func newTestHandler() *Handler {
	return NewHandler(&config.Config{Env: "test", Domain: "test.annict.com"}, nil)
}

// TestShowは、一覧表に登録した文書をその言語でページ全体に描画することを検証する。
func TestShow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expected    []string
		notExpected []string
		// hasLanguageVersionsは日本語版と英語版の両方を一覧表に登録した文書の現行の版かどうか。
		// falseのときは、結び付ける相手の言語版が無いため、hreflangと言語切替リンクを出力しないことを確かめる
		hasLanguageVersions bool
	}{
		{
			name: "利用規約",
			path: "/terms",
			expected: []string{
				`<html lang="ja">`,
				"<title>利用規約 | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/terms">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<h1 class="text-2xl font-bold antialiased">利用規約</h1>`,
				// 本文の前に改定の告知ブロックを置き、境界の後に本文の条が続く
				`<h2 id="2026年10月15日の改定について">2026年10月15日の改定について</h2>`,
				"<hr>",
				`<h2 id="第1条-適用">第1条 (適用)</h2>`,
				`<a href="/terms/api">API利用規約</a>`,
				// 特定商取引法に基づく表記とサポーターの申込み画面から、サポーターの条へリンクする
				`<h2 id="第11条-サポーター">第11条 (サポーター)</h2>`,
				`<a href="/terms/archive/20140310">2014年3月10日施行版</a>`,
				// 英語版と相互にhreflangを出力し、x-defaultは正本の日本語版を指す
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/terms">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/terms">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/terms">`,
				`<nav aria-label="言語">`,
				`href="/terms" lang="ja" hreflang="ja" aria-current="page">日本語</a>`,
				`href="/en/terms" lang="en" hreflang="en">English</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "利用規約の英語版",
			path: "/en/terms",
			expected: []string{
				`<html lang="en">`,
				"<title>Terms of Service | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/en/terms">`,
				`<meta property="og:locale" content="en_US">`,
				`<h1 class="text-2xl font-bold antialiased">Terms of Service</h1>`,
				`<h2 id="about-the-revision-of-october-15-2026">About the Revision of October 15, 2026</h2>`,
				"<hr>",
				`<h2 id="article-1-application">Article 1 (Application)</h2>`,
				// 日本語版を正本とする旨と、準拠法が日本法である旨を明記する
				"the Japanese version prevails.",
				"These Terms are governed by the laws of Japan.",
				// ほかの文書へのリンクは英語版を指す
				`<a href="/en/terms/api">API Terms of Use</a>`,
				`<a href="/en/privacy">Privacy Policy</a>`,
				// サポーターの申込み画面から、サポーターの条へリンクする
				`<h2 id="article-11-annict-supporters">Article 11 (Annict Supporters)</h2>`,
				// 改定前の版は日本語版だけを残すため、日本語版のアーカイブページを指す
				`<a href="/terms/archive/20140310">Version effective March 10, 2014</a>`,
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/terms">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/terms">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/terms">`,
				`<nav aria-label="Language">`,
				`href="/terms" lang="ja" hreflang="ja">日本語</a>`,
				`href="/en/terms" lang="en" hreflang="en" aria-current="page">English</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "API利用規約",
			path: "/terms/api",
			expected: []string{
				`<html lang="ja">`,
				"<title>API利用規約 | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/terms/api">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<h1 class="text-2xl font-bold antialiased">API利用規約</h1>`,
				// 本文の前に新設の告知ブロックを置き、境界の後に本文の章が続く
				`<h2 id="本規約の新設について">本規約の新設について</h2>`,
				"<hr>",
				`<h2 id="第1条-適用範囲">第1条 (適用範囲)</h2>`,
				`<a href="/terms">利用規約</a>`,
				// レート制限の数値は規約に書かず、開発者向けドキュメントへリンクする
				`<a href="https://wikino.app/s/annict/pages/357">開発者向けヘルプ</a>`,
				`<h2 id="附則">附則</h2>`,
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/terms/api">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/terms/api">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/terms/api">`,
				`href="/terms/api" lang="ja" hreflang="ja" aria-current="page">日本語</a>`,
				`href="/en/terms/api" lang="en" hreflang="en">English</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "API利用規約の英語版",
			path: "/en/terms/api",
			expected: []string{
				`<html lang="en">`,
				"<title>API Terms of Use | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/en/terms/api">`,
				`<meta property="og:locale" content="en_US">`,
				`<h1 class="text-2xl font-bold antialiased">API Terms of Use</h1>`,
				`<h2 id="about-the-establishment-of-these-terms">About the Establishment of These Terms</h2>`,
				"<hr>",
				`<h2 id="article-1-scope">Article 1 (Scope)</h2>`,
				"the Japanese version prevails.",
				"These Terms are governed by the laws of Japan.",
				`<a href="/en/terms">Terms of Service</a>`,
				// 開発者向けヘルプは日本語のページしか無いため、言語を添える
				`<a href="https://wikino.app/s/annict/pages/357">developer help</a> (in Japanese)`,
				`<h2 id="supplementary-provisions">Supplementary Provisions</h2>`,
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/terms/api">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/terms/api">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/terms/api">`,
				`<nav aria-label="Language">`,
				`href="/terms/api" lang="ja" hreflang="ja">日本語</a>`,
				`href="/en/terms/api" lang="en" hreflang="en" aria-current="page">English</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "プライバシーポリシー",
			path: "/privacy",
			expected: []string{
				`<html lang="ja">`,
				"<title>プライバシーポリシー | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/privacy">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<h1 class="text-2xl font-bold antialiased">プライバシーポリシー</h1>`,
				`<h2 id="1-取り扱う情報">1. 取り扱う情報</h2>`,
				`<a href="/privacy/archive/20190216">2019年2月16日施行版</a>`,
				`<a href="/privacy/archive/20140310">2014年3月10日施行版</a>`,
				// 英語版と相互にhreflangを出力し、x-defaultは正本の日本語版を指す
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/privacy">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/privacy">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/privacy">`,
				`<nav aria-label="言語">`,
				`href="/privacy" lang="ja" hreflang="ja" aria-current="page">日本語</a>`,
				`href="/en/privacy" lang="en" hreflang="en">English</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "プライバシーポリシーの英語版",
			path: "/en/privacy",
			expected: []string{
				`<html lang="en">`,
				"<title>Privacy Policy | Annict</title>",
				`<link rel="canonical" href="https://test.annict.com/en/privacy">`,
				`<meta property="og:locale" content="en_US">`,
				`<h1 class="text-2xl font-bold antialiased">Privacy Policy</h1>`,
				`<h2 id="1-information-we-handle">1. Information We Handle</h2>`,
				// 日本語版を正本とする旨と、準拠法が日本法である旨を明記する
				"the Japanese version prevails.",
				"This Policy is governed by the laws of Japan",
				// 改定前の版は日本語版だけを残すため、日本語版のアーカイブページを指す
				`<a href="/privacy/archive/20190216">Version effective February 16, 2019</a>`,
				`<a href="/privacy/archive/20140310">Version effective March 10, 2014</a>`,
				`<link rel="alternate" hreflang="ja" href="https://test.annict.com/privacy">` +
					`<link rel="alternate" hreflang="en" href="https://test.annict.com/en/privacy">` +
					`<link rel="alternate" hreflang="x-default" href="https://test.annict.com/privacy">`,
				`<nav aria-label="Language">`,
				`href="/privacy" lang="ja" hreflang="ja">日本語</a>`,
				`href="/en/privacy" lang="en" hreflang="en" aria-current="page">English</a>`,
				// フッターの文書リンクも英語版を指す
				`href="/en/privacy">Privacy Policy</a>`,
			},
			notExpected: []string{
				`name="robots"`,
			},
			hasLanguageVersions: true,
		},
		{
			name: "利用規約の2014年3月10日施行版",
			path: "/terms/archive/20140310",
			expected: []string{
				"<title>利用規約 (2014年3月10日施行版) | Annict</title>",
				`<meta name="robots" content="noindex">`,
				"この版の適用期間は2014年3月10日から2026年11月14日までです。",
				`<a class="link" href="/terms">現行の版を見る</a>`,
				`<h2 id="第13条-秘密保持">第13条 (秘密保持)</h2>`,
			},
			notExpected: []string{
				`rel="canonical"`,
			},
		},
		{
			name: "プライバシーポリシーの2014年3月10日施行版",
			path: "/privacy/archive/20140310",
			expected: []string{
				"<title>プライバシーポリシー (2014年3月10日施行版) | Annict</title>",
				`<meta name="robots" content="noindex">`,
				"この版の適用期間は2014年3月10日から2019年2月15日までです。",
				`<a class="link" href="/privacy">現行の版を見る</a>`,
				`<h2 id="1-収集する利用者情報及び収集方法">1. 収集する利用者情報及び収集方法</h2>`,
			},
			notExpected: []string{
				`rel="canonical"`,
			},
		},
		{
			name: "プライバシーポリシーの2019年2月16日施行版",
			path: "/privacy/archive/20190216",
			expected: []string{
				`<html lang="ja">`,
				"<title>プライバシーポリシー (2019年2月16日施行版) | Annict</title>",
				`<meta name="robots" content="noindex">`,
				"この版の適用期間は2019年2月16日から2026年10月13日までです。",
				`<a class="link" href="/privacy">現行の版を見る</a>`,
				`<h2 id="1-収集する利用者情報および収集方法">1. 収集する利用者情報および収集方法</h2>`,
			},
			// 古い文面が検索結果に出ないようnoindexにし、正規URLも宣言しない
			notExpected: []string{
				`rel="canonical"`,
			},
		},
		{
			name: "特定商取引法に基づく表記",
			path: "/legal",
			expected: []string{
				`<html lang="ja">`,
				"<title>特定商取引法に基づく表記 | Annict</title>",
				`<meta name="robots" content="noindex,nofollow,noarchive">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<h1 class="text-2xl font-bold antialiased">特定商取引法に基づく表記</h1>`,
				"<dt>販売事業者</dt>",
				"クレジットカード (JCBを含む)、Apple Pay、Linkによる決済",
				`<dd><a href="/faq">よくある質問</a>の対応ブラウザの項目をご確認下さい</dd>`,
				`<a href="/terms#%E7%AC%AC11%E6%9D%A1-%E3%82%B5%E3%83%9D%E3%83%BC%E3%82%BF%E3%83%BC">利用規約の第11条 (サポーター)</a>`,
			},
			// noindexのページに正規URLを宣言すると矛盾したシグナルになるため、canonicalを出力しない
			notExpected: []string{
				`rel="canonical"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			newTestHandler().Show(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", ct)
			}

			body := rr.Body.String()
			for _, expected := range tt.expected {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
			for _, notExpected := range tt.notExpected {
				if strings.Contains(body, notExpected) {
					t.Errorf("レスポンスに含まれてはいけない文字列 = %q", notExpected)
				}
			}

			if tt.hasLanguageVersions {
				return
			}
			// 英語版を一覧表に登録するまでは、結び付ける相手の言語版が無いためhreflangを出力しない
			if strings.Contains(body, "hreflang") {
				t.Error("英語版の無い文書でhreflangが出力された")
			}
			// 切り替える相手の言語版が無いため、言語切替リンクも表示しない
			if strings.Contains(body, `aria-label="言語"`) {
				t.Error("英語版の無い文書で言語切替リンクが表示された")
			}
		})
	}
}

// TestShow_OverridesRequestLocaleは、英語を使う利用者が日本語版を開いても、
// ページ全体を日本語で描画することを検証する。文書のページの言語はURLで決まるため。
func TestShow_OverridesRequestLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		expectedTitle string
		// englishTitleは同じ文書の英語のタイトル。日本語版のページに含まれないことを確かめる
		englishTitle string
	}{
		{
			name:          "利用規約",
			path:          "/terms",
			expectedTitle: "<title>利用規約 | Annict</title>",
			englishTitle:  "Terms of Service",
		},
		{
			name:          "特定商取引法に基づく表記",
			path:          "/legal",
			expectedTitle: "<title>特定商取引法に基づく表記 | Annict</title>",
			englishTitle:  "Legal Notice under the Act on Specified Commercial Transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), i18n.LangEn))
			rr := httptest.NewRecorder()
			newTestHandler().Show(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, expected := range []string{
				`<html lang="ja">`,
				tt.expectedTitle,
				`<meta property="og:locale" content="ja_JP">`,
			} {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
			if strings.Contains(body, tt.englishTitle) {
				t.Error("日本語版のページに英語の文言が含まれている")
			}
		})
	}
}

// TestShow_NotFoundは、一覧表に登録していないパスで404を返すことを検証する。
func TestShow_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "英語版を残していない利用規約の版", path: "/en/terms/archive/20140310"},
		{name: "英語版を残していないプライバシーポリシーの版", path: "/en/privacy/archive/20190216"},
		{name: "言語コードを重ねたパス", path: "/en/en/terms"},
		{name: "英語版を作らない特定商取引法に基づく表記", path: "/en/legal"},
		{name: "一覧表に登録していない利用規約の版", path: "/terms/archive/20140803"},
		{name: "一覧表に登録していないプライバシーポリシーの版", path: "/privacy/archive/20140803"},
		{name: "改定前の版を残さない特定商取引法に基づく表記", path: "/legal/archive/20140310"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			newTestHandler().Show(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

// TestShow_AllRegisteredDocumentsは、一覧表に登録した全文書のページを描画できることを検証する。
// 文書の種類を一覧表に足したときに、ページの設定 (documentPages) の登録漏れに気づけるようにする。
func TestShow_AllRegisteredDocuments(t *testing.T) {
	t.Parallel()

	for _, path := range legaldoc.Paths() {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			newTestHandler().Show(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
		})
	}
}

// TestShowVersionは、改定前の版をアーカイブページとして、適用期間と現行の版へのリンクを添えて描画することを検証する。
// 一覧表にまだ無い英語版の版も確かめるため、版を直接渡して確かめる。
func TestShowVersion(t *testing.T) {
	t.Parallel()

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	tests := []struct {
		name        string
		doc         legaldoc.Document
		version     legaldoc.Version
		expected    []string
		notExpected []string
	}{
		{
			name: "日本語版",
			doc:  legaldoc.Document{Key: legaldoc.KeyTerms, Lang: legaldoc.LangJa, Path: "/terms"},
			version: legaldoc.Version{
				EffectiveOn: time.Date(2014, time.March, 10, 0, 0, 0, 0, jst),
				EndedOn:     time.Date(2026, time.October, 31, 0, 0, 0, 0, jst),
				Path:        "/terms/archive/20140310",
				HTML:        `<h2 id="第1条-適用">第1条 (適用)</h2>`,
			},
			expected: []string{
				`<html lang="ja">`,
				"<title>利用規約 (2014年3月10日施行版) | Annict</title>",
				`<meta name="robots" content="noindex">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<h1 class="text-2xl font-bold antialiased">利用規約 (2014年3月10日施行版)</h1>`,
				"<h2>この文書は改定前の版です</h2>",
				"この版の適用期間は2014年3月10日から2026年10月31日までです。",
				`<a class="link" href="/terms">現行の版を見る</a>`,
				`<h2 id="第1条-適用">第1条 (適用)</h2>`,
			},
			notExpected: []string{
				"Terms of Service",
			},
		},
		{
			name: "英語版",
			doc:  legaldoc.Document{Key: legaldoc.KeyTerms, Lang: legaldoc.LangEn, Path: "/en/terms"},
			version: legaldoc.Version{
				EffectiveOn: time.Date(2026, time.April, 1, 0, 0, 0, 0, jst),
				EndedOn:     time.Date(2027, time.March, 31, 0, 0, 0, 0, jst),
				Path:        "/en/terms/archive/20260401",
				HTML:        `<h2 id="article-1">Article 1</h2>`,
			},
			expected: []string{
				`<html lang="en">`,
				"<title>Terms of Service (Version Effective April 1, 2026) | Annict</title>",
				`<meta name="robots" content="noindex">`,
				`<meta property="og:locale" content="en_US">`,
				"This version applies from April 1, 2026 through March 31, 2027.",
				`<a class="link" href="/en/terms">View the current version</a>`,
			},
			notExpected: []string{
				"利用規約",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// 文書のページと同じく、リクエストのロケールによらず版の言語で描画する
			locale := i18n.LangEn
			if tt.doc.Lang == legaldoc.LangEn {
				locale = i18n.LangJa
			}
			req := httptest.NewRequest(http.MethodGet, tt.version.Path, nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), locale))
			rr := httptest.NewRecorder()
			newTestHandler().showVersion(rr, req, tt.version, tt.doc)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, expected := range tt.expected {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
			for _, notExpected := range tt.notExpected {
				if strings.Contains(body, notExpected) {
					t.Errorf("レスポンスに含まれてはいけない文字列 = %q", notExpected)
				}
			}

			// 検索結果に古い文面を出さないnoindexのページのため、正規URLを宣言しない
			if strings.Contains(body, `rel="canonical"`) {
				t.Error("改定前の版のページでcanonicalが出力された")
			}
			// 現行の版の言語版と結び付けないため、hreflangと言語切替リンクを出力しない。
			// フッターの文書リンクはページと異なる言語の版を指すときにhreflangを付けるため、<link> と <nav> で確かめる
			if strings.Contains(body, `rel="alternate"`) {
				t.Error("改定前の版のページでhreflangが出力された")
			}
			if strings.Contains(body, `<nav aria-label="言語">`) || strings.Contains(body, `<nav aria-label="Language">`) {
				t.Error("改定前の版のページで言語切替リンクが表示された")
			}
		})
	}
}
