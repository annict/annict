package legaldoc

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// markdownは文書の描画に使うMarkdownの変換器。
// Rails版 (GitHub Markup + CommonMarkerのHARDBREAKS) と見た目を揃えるため、ハードラップ・表・URLとメールアドレスの
// 自動リンクを有効にする。
// 特定商取引法に基づく表記の項目と内容の対を、Rails版と同じdl・dt・ddで描画するため、定義リストも有効にする。
// 文書は信頼できるソースだが、書き手が意図しないマークアップを持ち込まないよう生HTMLは出力しない (html.WithUnsafeを指定しない)
var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Linkify,
		extension.DefinitionList,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
	),
)

// renderMarkdownはMarkdownをHTMLに描画する
func renderMarkdown(source []byte) (string, error) {
	// 見出しIDの重複は文書ごとに判定するため、描画のたびに生成器を作る
	ctx := parser.NewContext(parser.WithIDs(newHeadingIDs()))

	var buf bytes.Buffer
	if err := markdown.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		return "", fmt.Errorf("goldmarkでの変換に失敗: %w", err)
	}

	return buf.String(), nil
}

// headingIDsは見出しのテキストから見出しIDを生成する (goldmarkのparser.IDsの実装)。
// goldmarkの既定の生成器はASCIIの英数字以外を捨てるため、「第1条 (適用)」が「1-」になり、
// 数字を含まない日本語の見出しは出現順の連番 (heading-1など) になって、見出しを足すと後続のIDがずれる。
// 文書の章へのリンクを読める形で保てるよう、Unicodeの文字と数字を残す
type headingIDs struct {
	used map[string]bool
}

func newHeadingIDs() *headingIDs {
	return &headingIDs{used: map[string]bool{}}
}

// GenerateはGitHubの見出しアンカーと同じ規則でIDを作る。
// 文字・数字・アンダースコア・ハイフンを小文字にして残し、半角スペースをハイフンにし、ほかの記号を除く。
// 同じIDが既にあるときは末尾に連番を付ける
func (s *headingIDs) Generate(value []byte, _ ast.NodeKind) []byte {
	var b strings.Builder
	for _, r := range strings.TrimSpace(string(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsDigit(r) || unicode.Is(unicode.Pc, r) || r == '-':
			b.WriteRune(unicode.ToLower(r))
		case r == ' ':
			b.WriteRune('-')
		}
	}

	base := b.String()
	if base == "" {
		base = "heading"
	}

	id := base
	for i := 1; s.used[id]; i++ {
		id = fmt.Sprintf("%s-%d", base, i)
	}
	s.used[id] = true

	return []byte(id)
}

// Putは使用済みのIDを登録する
func (s *headingIDs) Put(value []byte) {
	s.used[string(value)] = true
}
