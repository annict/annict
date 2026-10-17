// Package legaldocは利用規約・プライバシーポリシーなどの法的文書を埋め込み、HTMLに描画して提供する。
//
// 文書はリポジトリで管理するMarkdownで、パッケージの初期化時に1回だけHTMLへ描画して保持する。
// どの文書のどの言語版を配信するかは文書の一覧表 (documents) で管理し、表に登録した言語版だけを配信の対象にする。
// 改定前の版も一覧表に登録し、文書のパスに /archive/{施行日} を付けたアーカイブページで配信する。
//
// 本パッケージはアプリケーション内の他のパッケージに依存しないleafに保つ。
package legaldoc

import (
	"embed"
	"fmt"
	"io/fs"
	"time"
)

// 文書のMarkdownを埋め込み
//
//go:embed documents/*.md
var documentsFS embed.FS

// Keyは文書の種類を表す
type Key string

const (
	KeyTerms         Key = "terms"
	KeyAPITerms      Key = "api_terms"
	KeyPrivacyPolicy Key = "privacy_policy"
	// KeySpecifiedCommercialTransactionsは特定商取引法に基づく表記
	KeySpecifiedCommercialTransactions Key = "specified_commercial_transactions"
)

// 文書の言語。internal/i18nの言語コードと同じ値にする
const (
	LangJa = "ja"
	LangEn = "en"
)

// Documentは文書の1つの言語版を表す
type Document struct {
	Key  Key
	Lang string
	// Pathは文書を配信するURLのパス
	Path string
	// EnactedOnは制定日。
	// 契約の条項ではない法定の表示 (特定商取引法に基づく表記) は、制定日・改定日・施行日を持たずゼロ値にする
	EnactedOn time.Time
	// RevisedOnは改定日。一度も改定していない文書ではゼロ値
	RevisedOn time.Time
	// EffectiveOnは現行の版の施行日
	EffectiveOn time.Time
	// HTMLは本文のMarkdownを描画したHTML
	HTML string
	// Versionsは改定前の版の一覧で、施行日の古い順に並べる。
	// 改定するときは、改定前のMarkdownを `{文書}.v{施行日}.{言語}.md` の名前で残して末尾に足す
	Versions []Version

	// fileは本文のMarkdownのファイル名
	file string
}

// Versionは文書の言語版の改定前の版を表す
type Version struct {
	// EffectiveOnはその版の施行日。一覧表にはこの日付だけを書き、ほかの値は初期化時に埋める
	EffectiveOn time.Time
	// EndedOnは適用終了日。次の版 (最後の版では現行の版) の施行日の前日
	EndedOn time.Time
	// Pathはアーカイブページのパス。文書のパスに /archive/{施行日} を付ける
	Path string
	// HTMLは本文のMarkdownを描画したHTML
	HTML string
}

// versionIDLayoutは版の識別子 (アーカイブページのパスと版のファイル名に使う施行日) の書式
const versionIDLayout = "20060102"

// jstは文書の日付を日本時間の暦日として扱うためのタイムゾーン
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, jst)
}

// documentsは配信する文書の一覧表
var documents = []Document{
	{
		Key:         KeyTerms,
		Lang:        LangJa,
		Path:        "/terms",
		EnactedOn:   date(2014, time.March, 10),
		RevisedOn:   date(2026, time.October, 15),
		EffectiveOn: date(2026, time.November, 15),
		file:        "terms.ja.md",
		Versions: []Version{
			{EffectiveOn: date(2014, time.March, 10)},
		},
	},
	{
		Key:         KeyAPITerms,
		Lang:        LangJa,
		Path:        "/terms/api",
		EnactedOn:   date(2026, time.October, 15),
		EffectiveOn: date(2026, time.November, 15),
		file:        "api_terms.ja.md",
	},
	{
		Key:         KeyPrivacyPolicy,
		Lang:        LangJa,
		Path:        "/privacy",
		EnactedOn:   date(2014, time.March, 10),
		RevisedOn:   date(2026, time.October, 14),
		EffectiveOn: date(2026, time.October, 14),
		file:        "privacy_policy.ja.md",
		Versions: []Version{
			{EffectiveOn: date(2014, time.March, 10)},
			{EffectiveOn: date(2019, time.February, 16)},
		},
	},
	// 英語版は日本語版の翻訳のため、制定日・改定日・施行日は日本語版と同じにする。
	// 改定前の版は日本語版だけを残し、英語版の附則からは日本語版のアーカイブページへリンクする
	{
		Key:         KeyTerms,
		Lang:        LangEn,
		Path:        "/en/terms",
		EnactedOn:   date(2014, time.March, 10),
		RevisedOn:   date(2026, time.October, 15),
		EffectiveOn: date(2026, time.November, 15),
		file:        "terms.en.md",
	},
	{
		Key:         KeyAPITerms,
		Lang:        LangEn,
		Path:        "/en/terms/api",
		EnactedOn:   date(2026, time.October, 15),
		EffectiveOn: date(2026, time.November, 15),
		file:        "api_terms.en.md",
	},
	{
		Key:         KeyPrivacyPolicy,
		Lang:        LangEn,
		Path:        "/en/privacy",
		EnactedOn:   date(2014, time.March, 10),
		RevisedOn:   date(2026, time.October, 14),
		EffectiveOn: date(2026, time.October, 14),
		file:        "privacy_policy.en.md",
	},
	{
		Key:  KeySpecifiedCommercialTransactions,
		Lang: LangJa,
		Path: "/legal",
		file: "specified_commercial_transactions.ja.md",
	},
}

// 起動時に全文書を1回だけ描画する。
// 一覧表とMarkdownの不整合はプログラムの誤りのため、起動を止めて気づけるようにする
func init() {
	if err := prepareDocuments(documentsFS, documents); err != nil {
		panic(err.Error())
	}
}

// prepareDocumentsは一覧表の各文書とその改定前の版のMarkdownを描画し、版のパスと適用終了日を埋める
func prepareDocuments(fsys fs.FS, docs []Document) error {
	for i := range docs {
		doc := &docs[i]

		html, err := renderFile(fsys, doc.file)
		if err != nil {
			return err
		}
		doc.HTML = html

		for j := range doc.Versions {
			version := &doc.Versions[j]
			id := version.EffectiveOn.Format(versionIDLayout)

			html, err := renderFile(fsys, fmt.Sprintf("%s.v%s.%s.md", doc.Key, id, doc.Lang))
			if err != nil {
				return err
			}
			version.HTML = html
			version.Path = doc.Path + "/archive/" + id

			// 版は施行日の古い順に並べるため、次の版 (最後の版では現行の版) の施行日の前日まで適用される
			nextEffectiveOn := doc.EffectiveOn
			if j+1 < len(doc.Versions) {
				nextEffectiveOn = doc.Versions[j+1].EffectiveOn
			}
			version.EndedOn = nextEffectiveOn.AddDate(0, 0, -1)
		}
	}

	return nil
}

func renderFile(fsys fs.FS, file string) (string, error) {
	source, err := fs.ReadFile(fsys, "documents/"+file)
	if err != nil {
		return "", fmt.Errorf("法的文書のMarkdownの読み込みに失敗しました: %s: %w", file, err)
	}

	html, err := renderMarkdown(source)
	if err != nil {
		return "", fmt.Errorf("法的文書のMarkdownの描画に失敗しました: %s: %w", file, err)
	}

	return html, nil
}

// LookupはURLのパスから文書の言語版を引く。一覧表に無いパスではfalseを返す
func Lookup(path string) (Document, bool) {
	for _, doc := range documents {
		if doc.Path == path {
			return doc, true
		}
	}

	return Document{}, false
}

// LookupVersionはアーカイブページのパスから、改定前の版とその版を持つ文書 (現行の版) を引く。
// 一覧表に無い版のパスではfalseを返す
func LookupVersion(path string) (Version, Document, bool) {
	return lookupVersion(documents, path)
}

func lookupVersion(docs []Document, path string) (Version, Document, bool) {
	for _, doc := range docs {
		for _, version := range doc.Versions {
			if version.Path == path {
				return version, doc, true
			}
		}
	}

	return Version{}, Document{}, false
}

// Pathsは一覧表に登録した全文書と改定前の版の配信パスを返す。ルーティングの登録に使う
func Paths() []string {
	return paths(documents)
}

func paths(docs []Document) []string {
	var paths []string
	for _, doc := range docs {
		paths = append(paths, doc.Path)
		for _, version := range doc.Versions {
			paths = append(paths, version.Path)
		}
	}

	return paths
}

// LanguageVersionsは文書の種類が同じ言語版を、一覧表の順に返す。
// 英語版を一覧表に登録する前の文書では、日本語版だけを返す
func LanguageVersions(key Key) []Document {
	var versions []Document
	for _, doc := range documents {
		if doc.Key == key {
			versions = append(versions, doc)
		}
	}

	return versions
}

// LocalizedVersionは文書の種類が同じ言語版のうち、langの言語版を返す。
// langの言語版が無いときは正本の日本語版を返す。一覧表に無い文書ではfalseを返す
func LocalizedVersion(key Key, lang string) (Document, bool) {
	return localizedVersion(documents, key, lang)
}

func localizedVersion(docs []Document, key Key, lang string) (Document, bool) {
	var fallback Document
	found := false
	for _, doc := range docs {
		if doc.Key != key {
			continue
		}
		if doc.Lang == lang {
			return doc, true
		}
		if doc.Lang == LangJa {
			fallback = doc
			found = true
		}
	}

	return fallback, found
}
