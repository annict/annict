package legaldoc

import (
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantOK   bool
		wantKey  Key
		wantLang string
	}{
		{
			name:     "利用規約の日本語版",
			path:     "/terms",
			wantOK:   true,
			wantKey:  KeyTerms,
			wantLang: LangJa,
		},
		{
			name:     "プライバシーポリシーの日本語版",
			path:     "/privacy",
			wantOK:   true,
			wantKey:  KeyPrivacyPolicy,
			wantLang: LangJa,
		},
		{
			name:     "特定商取引法に基づく表記",
			path:     "/legal",
			wantOK:   true,
			wantKey:  KeySpecifiedCommercialTransactions,
			wantLang: LangJa,
		},
		{
			name:     "プライバシーポリシーの英語版",
			path:     "/en/privacy",
			wantOK:   true,
			wantKey:  KeyPrivacyPolicy,
			wantLang: LangEn,
		},
		{
			name:     "利用規約の英語版",
			path:     "/en/terms",
			wantOK:   true,
			wantKey:  KeyTerms,
			wantLang: LangEn,
		},
		{
			name:   "英語版を作らない特定商取引法に基づく表記",
			path:   "/en/legal",
			wantOK: false,
		},
		{
			name:     "API利用規約の日本語版",
			path:     "/terms/api",
			wantOK:   true,
			wantKey:  KeyAPITerms,
			wantLang: LangJa,
		},
		{
			name:     "API利用規約の英語版",
			path:     "/en/terms/api",
			wantOK:   true,
			wantKey:  KeyAPITerms,
			wantLang: LangEn,
		},
		{
			name:   "英語版を残していない利用規約の版",
			path:   "/en/terms/archive/20140310",
			wantOK: false,
		},
		{
			name:   "末尾にスラッシュを付けたパス",
			path:   "/terms/",
			wantOK: false,
		},
		{
			name:   "空文字",
			path:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, ok := Lookup(tt.path)
			if ok != tt.wantOK {
				t.Fatalf("Lookup(%q)の結果の有無 = %v、期待値 = %v", tt.path, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}

			if doc.Key != tt.wantKey {
				t.Errorf("Key = %q、期待値 = %q", doc.Key, tt.wantKey)
			}
			if doc.Lang != tt.wantLang {
				t.Errorf("Lang = %q、期待値 = %q", doc.Lang, tt.wantLang)
			}
			if doc.Path != tt.path {
				t.Errorf("Path = %q、期待値 = %q", doc.Path, tt.path)
			}
			if doc.HTML == "" {
				t.Error("HTMLが空だった")
			}
		})
	}
}

func TestPaths(t *testing.T) {
	t.Parallel()

	// ルーティングに登録したパスは、どれもLookupで文書か、LookupVersionで改定前の版を引ける
	for _, path := range Paths() {
		_, ok := Lookup(path)
		if !ok {
			_, _, ok = LookupVersion(path)
		}
		if !ok {
			t.Errorf("%qで文書も改定前の版も見つからなかった", path)
		}
	}
}

// newVersionedDocumentsは、改定前の版を2つ持つ利用規約を登録した一覧表を描画して返す。
// 改定前の版を一覧表に登録した後の振る舞いを確かめるために使う
func newVersionedDocuments(t *testing.T) []Document {
	t.Helper()

	fsys := fstest.MapFS{
		"documents/terms.ja.md":           {Data: []byte("## 現行の版")},
		"documents/terms.v20140310.ja.md": {Data: []byte("## 最初の版")},
		"documents/terms.v20200401.ja.md": {Data: []byte("## 2つ目の版")},
		"documents/privacy_policy.ja.md":  {Data: []byte("## 改定前の版の無い文書")},
	}
	docs := []Document{
		{
			Key:         KeyTerms,
			Lang:        LangJa,
			Path:        "/terms",
			EnactedOn:   date(2014, time.March, 10),
			RevisedOn:   date(2026, time.October, 1),
			EffectiveOn: date(2026, time.November, 1),
			Versions: []Version{
				{EffectiveOn: date(2014, time.March, 10)},
				{EffectiveOn: date(2020, time.April, 1)},
			},
			file: "terms.ja.md",
		},
		{
			Key:         KeyPrivacyPolicy,
			Lang:        LangJa,
			Path:        "/privacy",
			EnactedOn:   date(2014, time.August, 3),
			EffectiveOn: date(2014, time.August, 3),
			file:        "privacy_policy.ja.md",
		},
	}

	if err := prepareDocuments(fsys, docs); err != nil {
		t.Fatalf("prepareDocuments()のエラー = %v", err)
	}

	return docs
}

// TestPrepareDocuments_Versionsは、改定前の版のMarkdownを描画し、パスと適用終了日を埋めることを検証する。
func TestPrepareDocuments_Versions(t *testing.T) {
	t.Parallel()

	versions := newVersionedDocuments(t)[0].Versions

	tests := []struct {
		name        string
		version     Version
		wantPath    string
		wantEndedOn time.Time
		wantHTML    string
	}{
		{
			name:     "次の版がある版は、次の版の施行日の前日まで適用される",
			version:  versions[0],
			wantPath: "/terms/archive/20140310",
			// 次の版の施行日 (2020-04-01) の前日。月をまたぐ
			wantEndedOn: date(2020, time.March, 31),
			wantHTML:    "<h2 id=\"最初の版\">最初の版</h2>\n",
		},
		{
			name:     "最後の版は、現行の版の施行日の前日まで適用される",
			version:  versions[1],
			wantPath: "/terms/archive/20200401",
			// 改定後の文書を施行日の前から公開していても、改定前の版は施行日の前日まで適用される
			wantEndedOn: date(2026, time.October, 31),
			wantHTML:    "<h2 id=\"2つ目の版\">2つ目の版</h2>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.version.Path != tt.wantPath {
				t.Errorf("Path = %q、期待値 = %q", tt.version.Path, tt.wantPath)
			}
			if !tt.version.EndedOn.Equal(tt.wantEndedOn) {
				t.Errorf("EndedOn = %s、期待値 = %s", tt.version.EndedOn, tt.wantEndedOn)
			}
			if tt.version.HTML != tt.wantHTML {
				t.Errorf("HTML = %q、期待値 = %q", tt.version.HTML, tt.wantHTML)
			}
		})
	}
}

// TestPrepareDocuments_MissingVersionFileは、一覧表に登録した版のMarkdownが無いときにエラーを返すことを検証する。
// 起動時にpanicさせ、配信できない版の登録に気づけるようにするため。
func TestPrepareDocuments_MissingVersionFile(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"documents/terms.ja.md": {Data: []byte("## 現行の版")},
	}
	docs := []Document{
		{
			Key:         KeyTerms,
			Lang:        LangJa,
			Path:        "/terms",
			EffectiveOn: date(2026, time.November, 1),
			Versions:    []Version{{EffectiveOn: date(2014, time.March, 10)}},
			file:        "terms.ja.md",
		},
	}

	err := prepareDocuments(fsys, docs)
	if err == nil {
		t.Fatal("版のMarkdownが無いのにエラーにならなかった")
	}
	if !strings.Contains(err.Error(), "terms.v20140310.ja.md") {
		t.Errorf("エラー = %q に版のファイル名が含まれていない", err)
	}
}

func TestLookupVersion(t *testing.T) {
	t.Parallel()

	docs := newVersionedDocuments(t)

	tests := []struct {
		name            string
		path            string
		wantOK          bool
		wantEffectiveOn time.Time
	}{
		{
			name:            "改定前の版",
			path:            "/terms/archive/20140310",
			wantOK:          true,
			wantEffectiveOn: date(2014, time.March, 10),
		},
		{
			name:            "2つ目の改定前の版",
			path:            "/terms/archive/20200401",
			wantOK:          true,
			wantEffectiveOn: date(2020, time.April, 1),
		},
		{
			name:   "現行の版の施行日はアーカイブページにしない",
			path:   "/terms/archive/20261101",
			wantOK: false,
		},
		{
			name:   "一覧表に無い施行日",
			path:   "/terms/archive/20150101",
			wantOK: false,
		},
		{
			name:   "改定前の版の無い文書",
			path:   "/privacy/archive/20140803",
			wantOK: false,
		},
		{
			name:   "現行の版のパス",
			path:   "/terms",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version, doc, ok := lookupVersion(docs, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("lookupVersion(%q)の結果の有無 = %v、期待値 = %v", tt.path, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}

			if !version.EffectiveOn.Equal(tt.wantEffectiveOn) {
				t.Errorf("EffectiveOn = %s、期待値 = %s", version.EffectiveOn, tt.wantEffectiveOn)
			}
			// 現行の版へのリンクのため、版を持つ文書も返す
			if doc.Path != "/terms" {
				t.Errorf("文書のPath = %q、期待値 = %q", doc.Path, "/terms")
			}
		})
	}
}

// TestLookupVersion_Documentsは、一覧表に登録した改定前の版だけをアーカイブページとして配信することを検証する。
func TestLookupVersion_Documents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		wantOK      bool
		wantKey     Key
		wantEndedOn time.Time
	}{
		{
			name:        "プライバシーポリシーの2014年3月10日施行版",
			path:        "/privacy/archive/20140310",
			wantOK:      true,
			wantKey:     KeyPrivacyPolicy,
			wantEndedOn: date(2019, time.February, 15),
		},
		{
			name:        "プライバシーポリシーの2019年2月16日施行版",
			path:        "/privacy/archive/20190216",
			wantOK:      true,
			wantKey:     KeyPrivacyPolicy,
			wantEndedOn: date(2026, time.October, 13),
		},
		{
			name:        "利用規約の2014年3月10日施行版",
			path:        "/terms/archive/20140310",
			wantOK:      true,
			wantKey:     KeyTerms,
			wantEndedOn: date(2026, time.November, 14),
		},
		{
			name:   "一覧表に登録していないプライバシーポリシーの版",
			path:   "/privacy/archive/20140803",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version, doc, ok := LookupVersion(tt.path)
			if ok != tt.wantOK {
				t.Fatalf("LookupVersion(%q)の結果の有無 = %v、期待値 = %v", tt.path, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}

			if doc.Key != tt.wantKey {
				t.Errorf("Key = %q、期待値 = %q", doc.Key, tt.wantKey)
			}
			if !version.EndedOn.Equal(tt.wantEndedOn) {
				t.Errorf("EndedOn = %s、期待値 = %s", version.EndedOn, tt.wantEndedOn)
			}
		})
	}
}

func TestPaths_Versions(t *testing.T) {
	t.Parallel()

	got := paths(newVersionedDocuments(t))
	want := []string{"/terms", "/terms/archive/20140310", "/terms/archive/20200401", "/privacy"}
	if !slices.Equal(got, want) {
		t.Errorf("パス = %q、期待値 = %q", got, want)
	}
}

// TestDocumentFilesは、埋め込んだMarkdownがどれも一覧表の文書か改定前の版として登録されていることを検証する。
// 改定前の版のファイルを残して一覧表への登録を忘れると、アーカイブページが配信されないため。
func TestDocumentFiles(t *testing.T) {
	t.Parallel()

	registered := map[string]bool{}
	for _, doc := range documents {
		registered[doc.file] = true
		for _, version := range doc.Versions {
			registered[string(doc.Key)+".v"+version.EffectiveOn.Format(versionIDLayout)+"."+doc.Lang+".md"] = true
		}
	}

	files, err := fs.Glob(documentsFS, "documents/*.md")
	if err != nil {
		t.Fatalf("埋め込んだMarkdownの一覧の取得に失敗: %v", err)
	}
	for _, file := range files {
		name := strings.TrimPrefix(file, "documents/")
		if !registered[name] {
			t.Errorf("%s: 一覧表に登録されていない", name)
		}
	}
}

func TestLanguageVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       Key
		wantPaths []string
	}{
		{
			name:      "英語版を登録した利用規約",
			key:       KeyTerms,
			wantPaths: []string{"/terms", "/en/terms"},
		},
		{
			name:      "英語版を登録したAPI利用規約",
			key:       KeyAPITerms,
			wantPaths: []string{"/terms/api", "/en/terms/api"},
		},
		{
			name:      "英語版を登録したプライバシーポリシー",
			key:       KeyPrivacyPolicy,
			wantPaths: []string{"/privacy", "/en/privacy"},
		},
		{
			name:      "日本語版だけの特定商取引法に基づく表記",
			key:       KeySpecifiedCommercialTransactions,
			wantPaths: []string{"/legal"},
		},
		{
			name:      "一覧表に無い文書",
			key:       Key("unknown"),
			wantPaths: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			versions := LanguageVersions(tt.key)
			gotPaths := make([]string, 0, len(versions))
			for _, version := range versions {
				if version.Key != tt.key {
					t.Errorf("Key = %q、期待値 = %q", version.Key, tt.key)
				}
				gotPaths = append(gotPaths, version.Path)
			}

			if !slices.Equal(gotPaths, tt.wantPaths) {
				t.Errorf("言語版のパス = %q、期待値 = %q", gotPaths, tt.wantPaths)
			}
		})
	}
}

func TestLocalizedVersion(t *testing.T) {
	t.Parallel()

	// 英語版を一覧表に登録した後の振る舞いも確かめるため、一覧表を差し替えて検証する
	docs := []Document{
		{Key: KeyTerms, Lang: LangJa, Path: "/terms"},
		{Key: KeyTerms, Lang: LangEn, Path: "/en/terms"},
		{Key: KeySpecifiedCommercialTransactions, Lang: LangJa, Path: "/legal"},
	}

	tests := []struct {
		name     string
		key      Key
		lang     string
		wantOK   bool
		wantPath string
	}{
		{
			name:     "日本語版",
			key:      KeyTerms,
			lang:     LangJa,
			wantOK:   true,
			wantPath: "/terms",
		},
		{
			name:     "英語版がある文書の英語版",
			key:      KeyTerms,
			lang:     LangEn,
			wantOK:   true,
			wantPath: "/en/terms",
		},
		{
			name:     "英語版が無い文書は日本語版にフォールバックする",
			key:      KeySpecifiedCommercialTransactions,
			lang:     LangEn,
			wantOK:   true,
			wantPath: "/legal",
		},
		{
			name:     "翻訳を持たない言語は日本語版にフォールバックする",
			key:      KeyTerms,
			lang:     "fr",
			wantOK:   true,
			wantPath: "/terms",
		},
		{
			name:   "一覧表に無い文書",
			key:    KeyPrivacyPolicy,
			lang:   LangJa,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, ok := localizedVersion(docs, tt.key, tt.lang)
			if ok != tt.wantOK {
				t.Fatalf("結果の有無 = %v、期待値 = %v", ok, tt.wantOK)
			}
			if doc.Path != tt.wantPath {
				t.Errorf("Path = %q、期待値 = %q", doc.Path, tt.wantPath)
			}
		})
	}
}

// TestLocalizedVersion_Documentsは、英語を使う利用者に、英語版を登録した文書は英語版を返し、
// 英語版を作らない文書は日本語版を返すことを検証する。
func TestLocalizedVersion_Documents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      Key
		wantLang string
	}{
		{key: KeyTerms, wantLang: LangEn},
		{key: KeyAPITerms, wantLang: LangEn},
		{key: KeyPrivacyPolicy, wantLang: LangEn},
		{key: KeySpecifiedCommercialTransactions, wantLang: LangJa},
	}

	for _, tt := range tests {
		doc, ok := LocalizedVersion(tt.key, LangEn)
		if !ok {
			t.Errorf("LocalizedVersion(%q)で文書が見つからなかった", tt.key)
			continue
		}
		if doc.Lang != tt.wantLang {
			t.Errorf("LocalizedVersion(%q)のLang = %q、期待値 = %q", tt.key, doc.Lang, tt.wantLang)
		}
	}
}

// h2Patternは見出し (章) の開始タグを取り出す
var h2Pattern = regexp.MustCompile(`<h2[\s>]`)

// TestLanguageVersionsMatchは、同じ文書の言語版どうしで、日付と章構成が揃っていることを検証する。
// 英語版は日本語版の翻訳のため、片方だけを改定したり章を足したりすると食い違いに気づけるようにする。
func TestLanguageVersionsMatch(t *testing.T) {
	t.Parallel()

	for _, key := range []Key{KeyTerms, KeyAPITerms, KeyPrivacyPolicy, KeySpecifiedCommercialTransactions} {
		ja, ok := localizedVersion(documents, key, LangJa)
		if !ok {
			t.Fatalf("%q: 日本語版が見つからなかった", key)
		}

		for _, doc := range LanguageVersions(key) {
			if doc.Lang == LangJa {
				continue
			}

			if !doc.EnactedOn.Equal(ja.EnactedOn) || !doc.RevisedOn.Equal(ja.RevisedOn) || !doc.EffectiveOn.Equal(ja.EffectiveOn) {
				t.Errorf("%s: 制定日・改定日・施行日が日本語版 (%s) と一致しない", doc.file, ja.file)
			}

			got := len(h2Pattern.FindAllString(doc.HTML, -1))
			want := len(h2Pattern.FindAllString(ja.HTML, -1))
			if got != want {
				t.Errorf("%s: 見出し (h2) の数 = %d、日本語版の見出しの数 = %d", doc.file, got, want)
			}

			// 改定前の版は日本語版だけを残すため、英語版の附則からは日本語版と同じアーカイブページへリンクする
			for _, version := range ja.Versions {
				if !strings.Contains(doc.HTML, `href="`+version.Path+`"`) {
					t.Errorf("%s: 改定前の版 %q へのリンクが含まれていない", doc.file, version.Path)
				}
			}
		}
	}
}

func TestDocuments(t *testing.T) {
	t.Parallel()

	paths := map[string]bool{}
	versions := map[string]bool{}

	for _, doc := range documents {
		if paths[doc.Path] {
			t.Errorf("パス %q が一覧表に重複して登録されている", doc.Path)
		}
		paths[doc.Path] = true

		version := string(doc.Key) + "." + doc.Lang
		if versions[version] {
			t.Errorf("文書の言語版 %q が一覧表に重複して登録されている", version)
		}
		versions[version] = true

		// 生HTMLは描画しないため、文書に書くと黙って欠落する
		if strings.Contains(doc.HTML, "raw HTML omitted") {
			t.Errorf("%s: 生HTMLが含まれている。Markdownの記法に書き換えること", doc.file)
		}

		// 日付を持たない文書 (特定商取引法に基づく表記) は、制定日・改定日・施行日のどれも持たず、改定前の版も残さない
		if doc.EnactedOn.IsZero() {
			if !doc.RevisedOn.IsZero() || !doc.EffectiveOn.IsZero() {
				t.Errorf("%s: 制定日が無いのに改定日 = %s または施行日 = %s が設定されている", doc.file, doc.RevisedOn, doc.EffectiveOn)
			}
			if len(doc.Versions) > 0 {
				t.Errorf("%s: 制定日が無いのに改定前の版が登録されている", doc.file)
			}
			continue
		}
		if doc.EffectiveOn.Before(doc.EnactedOn) {
			t.Errorf("%s: 施行日 = %s が制定日 = %s より前になっている", doc.file, doc.EffectiveOn, doc.EnactedOn)
		}
		if !doc.RevisedOn.IsZero() && doc.RevisedOn.Before(doc.EnactedOn) {
			t.Errorf("%s: 改定日 = %s が制定日 = %s より前になっている", doc.file, doc.RevisedOn, doc.EnactedOn)
		}

		// 改定前の版は、最初の版が制定日に施行され、施行日の古い順に現行の版の施行日より前に並ぶ
		for i, version := range doc.Versions {
			if i == 0 && !version.EffectiveOn.Equal(doc.EnactedOn) {
				t.Errorf("%s: 最初の版の施行日 = %s が制定日 = %s と一致しない", doc.file, version.EffectiveOn, doc.EnactedOn)
			}
			next := doc.EffectiveOn
			if i+1 < len(doc.Versions) {
				next = doc.Versions[i+1].EffectiveOn
			}
			if !version.EffectiveOn.Before(next) {
				t.Errorf("%s: 版の施行日 = %s が次の版の施行日 = %s より前になっていない", doc.file, version.EffectiveOn, next)
			}
		}
	}
}

// tagPatternはHTMLの開始タグの要素名を取り出す
var tagPattern = regexp.MustCompile(`<([a-z][a-z0-9]*)[\s>/]`)

// lookupHTMLは、パスから現行の版の文書か改定前の版を引き、描画したHTMLを返す
func lookupHTML(t *testing.T, path string) string {
	t.Helper()

	if doc, ok := Lookup(path); ok {
		return doc.HTML
	}
	if version, _, ok := LookupVersion(path); ok {
		return version.HTML
	}
	t.Fatalf("%qで文書も改定前の版も見つからなかった", path)

	return ""
}

// TestDocumentsMatchRailsは、Go版の描画がRails版と同じ構造 (見出しレベル・表・リスト・定義リスト・改行) になることを確かめる。
// Rails版から移したあとに改定した文書は、Rails版と同じ文面の改定前の版で確かめる
func TestDocumentsMatchRails(t *testing.T) {
	t.Parallel()

	// 要素ごとの個数の期待値は、Rails版 (GitHub Markup + CommonMarkerのHARDBREAKS) の描画結果から数えた値。
	// brだけは、句点改行に合わせて文ごとに改行した地の文の分 (利用規約2・プライバシーポリシー7) をRails版の個数に足している。
	// 特定商取引法に基づく表記はRails版のビュー (ERB) から数え、aだけは自動リンクでリンクになるメールアドレスの分を足している。
	// さらに、利用規約のサポーターの条との整合で改めた欄 (支払時期・返品・キャンセル) で増やした改行 (br 5) とリンク (a 2) の分を足している
	tests := []struct {
		name       string
		path       string
		wantCounts map[string]int
		wantLinks  []string
	}{
		{
			name: "利用規約の2014年3月10日施行版",
			path: "/terms/archive/20140310",
			wantCounts: map[string]int{
				"a":  1,
				"br": 2,
				"h2": 19,
				"li": 78,
				"ol": 18,
				"p":  61,
			},
			wantLinks: []string{
				`<a href="https://annict.com/privacy">https://annict.com/privacy</a>`,
			},
		},
		{
			name: "プライバシーポリシーの2019年2月16日施行版",
			path: "/privacy/archive/20190216",
			wantCounts: map[string]int{
				"a":     2,
				"br":    8,
				"h2":    8,
				"li":    32,
				"ol":    9,
				"p":     31,
				"table": 1,
				"tbody": 1,
				"td":    8,
				"th":    2,
				"thead": 1,
				"tr":    5,
				"ul":    1,
			},
			wantLinks: []string{
				`<a href="https://annict.com/settings/account">アカウント設定ページ</a>`,
				`<a href="mailto:hello@annict.com">hello@annict.com</a>`,
			},
		},
		{
			name: "特定商取引法に基づく表記",
			path: "/legal",
			wantCounts: map[string]int{
				"a":  4,
				"br": 6,
				"dd": 12,
				"dl": 1,
				"dt": 12,
			},
			wantLinks: []string{
				`<dd><a href="mailto:hello@annict.com">hello@annict.com</a></dd>`,
				`<dd><a href="/faq">よくある質問</a>の対応ブラウザの項目をご確認下さい</dd>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := lookupHTML(t, tt.path)

			gotCounts := map[string]int{}
			for _, m := range tagPattern.FindAllStringSubmatch(html, -1) {
				gotCounts[m[1]]++
			}

			for tag, want := range tt.wantCounts {
				if gotCounts[tag] != want {
					t.Errorf("<%s>の個数 = %d、期待値 = %d", tag, gotCounts[tag], want)
				}
			}
			for tag, got := range gotCounts {
				if _, ok := tt.wantCounts[tag]; !ok {
					t.Errorf("Rails版に無い<%s>が%d個ある", tag, got)
				}
			}

			for _, link := range tt.wantLinks {
				if !strings.Contains(html, link) {
					t.Errorf("リンク %q が含まれていない", link)
				}
			}
		})
	}
}
