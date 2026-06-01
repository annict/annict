// Command commentlint checks that bilingual code comments use the [Ja]
// marker correctly. The marker introduces the
// Japanese translation block, which must sit below a corresponding English
// block and appear at most once per comment group. The tool targets the
// recurring *misuse* of the marker rather than enforcing full bilingual
// coverage, which keeps false positives near zero.
//
// Two modes:
//
//	commentlint [paths...]              checks the [Ja]-on-non-Japanese rule
//	                                    across the whole tree (default: ".").
//	commentlint -base=<ref> [paths...]  also checks the marker-placement rules,
//	                                    limited to lines added since <ref>.
//
// .go files are parsed via go/parser so that "//" inside string literals is
// never mistaken for a comment; .templ files (not valid Go) are scanned by line.
//
// [Ja] commentlint コマンドは、コードコメントの英日併記で
// `[Ja]` マーカーが正しく使われているかをチェックする。`[Ja]` は日本語訳ブロックの
// 冒頭を示すマーカーで、対応する英語ブロックの下に置き、1 コメント群に 1 つだけ
// 付ける。全併記の強制ではなく、再発している「マーカーの誤用」のみを対象にする
// ことで誤検出をほぼゼロに保つ。
//
// モードは 2 つ:
//
//	commentlint [paths...]              「[Ja] が日本語を含まない行に付く」誤用を
//	                                    ツリー全体で検査する (既定は ".")。
//	commentlint -base=<ref> [paths...]  マーカー配置の規則も検査する。<ref> 以降に
//	                                    追加された行に限定する。
//
// .go は go/parser で解析し、文字列リテラル中の "//" をコメントと誤認しない。
// .templ (Go として不正) は行単位で走査する。
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// reJapanese matches any Hiragana, Katakana, or Han (kanji) rune.
	// [Ja] reJapanese はひらがな・カタカナ・漢字のいずれかにマッチする。
	reJapanese = regexp.MustCompile(`[\p{Hiragana}\p{Katakana}\p{Han}]`)
	// reLatin matches an ASCII Latin letter.
	// [Ja] reLatin は ASCII のラテン文字にマッチする。
	reLatin = regexp.MustCompile(`[A-Za-z]`)
	// reGenerated matches the standard "generated; do not edit" header.
	// [Ja] reGenerated は「生成物・編集禁止」の定型ヘッダーにマッチする。
	reGenerated = regexp.MustCompile(`^//\s*Code generated .* DO NOT EDIT\.$`)
)

// commentLine is a single physical line of a comment with its 1-based line number.
// [Ja] commentLine はコメントの 1 物理行と、その 1 始まりの行番号。
type commentLine struct {
	line int
	text string
}

// finding is one detected violation.
// [Ja] finding は検出した違反 1 件。
type finding struct {
	file string
	line int
	cond int
	msg  string
}

const (
	msgCond1 = "[Ja] marker on a line with no Japanese text; it must lead the Japanese translation, not the English block / [Ja] マーカーが日本語を含まない行に付いている"
	msgCond2 = "[Ja] marker has no English block above it; write the English block first, then [Ja] / [Ja] の上に対応する英語ブロックが無い"
	msgCond3 = "[Ja] marker appears more than once in one comment block; mark only the first line of the Japanese block / [Ja] マーカーが 1 コメント群に複数ある"
)

func main() {
	base := flag.String("base", "", "git ref to diff against; enables marker-placement checks limited to added lines / 差分の基準 git ref。追加行に限定したマーカー配置検査を有効化する")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	findings, err := run(roots, *base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "commentlint:", err)
		os.Exit(2)
	}
	if len(findings) == 0 {
		return
	}
	for _, f := range findings {
		fmt.Printf("%s:%d: %s\n", f.file, f.line, f.msg)
	}
	fmt.Fprintf(os.Stderr, "\ncommentlint: %d bilingual [Ja] marker violation(s)\n", len(findings))
	os.Exit(1)
}

// run walks the roots and returns the violations to report. In full mode only
// condition 1 (marker on a non-Japanese line) is reported, anywhere in the tree.
// In diff mode all conditions are reported, but only for lines added since base.
//
// [Ja] run は roots を走査し、報告すべき違反を返す。全体モードでは条件 1
// (日本語を含まない行のマーカー) のみをツリー全体で報告する。差分モードでは全条件を
// 報告するが、base 以降に追加された行に限定する。
func run(roots []string, base string) ([]finding, error) {
	diffMode := base != ""

	var added map[string]map[int]bool
	if diffMode {
		var err error
		added, err = addedLines(base)
		if err != nil {
			// Be lenient: if the diff cannot be computed, skip the diff-scoped
			// checks rather than failing the build.
			// [Ja] 差分を計算できない場合はビルドを失敗させず、差分限定の検査を
			// スキップする。
			fmt.Fprintf(os.Stderr, "commentlint: skipping diff checks: %v\n", err)
			return nil, nil
		}
	}

	var all []finding
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if skipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".templ" {
				return nil
			}
			if strings.HasSuffix(path, "_templ.go") {
				return nil
			}

			groups, gerr := commentGroups(path, ext)
			if gerr != nil {
				fmt.Fprintf(os.Stderr, "commentlint: %s: %v\n", path, gerr)
				return nil
			}
			for _, g := range groups {
				for _, f := range checkGroup(g) {
					f.file = path
					if diffMode {
						abs, aerr := filepath.Abs(path)
						if aerr != nil || added[abs] == nil || !added[abs][f.line] {
							continue
						}
					} else if f.cond != 1 {
						continue
					}
					all = append(all, f)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].file != all[j].file {
			return all[i].file < all[j].file
		}
		if all[i].line != all[j].line {
			return all[i].line < all[j].line
		}
		return all[i].cond < all[j].cond
	})
	return all, nil
}

// checkGroup evaluates one comment group against the [Ja] marker rules.
// Only lines whose comment content *begins* with the [Ja] marker count as
// marker lines; a mid-sentence mention of "[Ja]" in English prose (such as this
// tool's own docs) is ignored. At most one finding per marker line is produced
// (condition 1 supersedes condition 2 on the same line); condition 3 is reported
// independently.
//
// [Ja] checkGroup は 1 コメント群を [Ja] マーカー規則で評価する。
// コメント本文が [Ja] マーカーで「始まる」行だけをマーカー行とみなし、英語の地の文中で
// "[Ja]" に言及しているだけの行 (本ツール自身の説明など) は無視する。1 マーカー行あたり
// 最大 1 件 (同じ行では条件 1 が条件 2 に優先する)。条件 3 は独立に報告する。
func checkGroup(lines []commentLine) []finding {
	var fs []finding
	markerCount := 0
	englishSeenAbove := false

	for _, cl := range lines {
		if !markerAtStart(cl.text) {
			if isEnglishText(cl.text) {
				englishSeenAbove = true
			}
			continue
		}

		markerCount++
		if markerCount > 1 {
			fs = append(fs, finding{line: cl.line, cond: 3, msg: msgCond3})
		}

		if !hasJapanese(cl.text) {
			// Condition 1: the marker leads an English (or marker-only) line.
			// [Ja] 条件 1: マーカーが英語 (またはマーカーのみ) の行を先導している。
			fs = append(fs, finding{line: cl.line, cond: 1, msg: msgCond1})
			continue
		}

		// Condition 2: a marker line still needs an English block above it.
		// [Ja] 条件 2: マーカー行でも、その上に英語ブロックが必要。
		if !englishSeenAbove {
			fs = append(fs, finding{line: cl.line, cond: 2, msg: msgCond2})
		}
	}
	return fs
}

// markerAtStart reports whether the comment content begins with the [Ja] marker,
// after stripping the comment leader ("//", "/*", or a "*" continuation) and
// surrounding whitespace. This distinguishes a marker use from a prose mention.
//
// [Ja] markerAtStart は、コメントリーダー ("//"・"/*"・継続行の "*") と前後の空白を
// 取り除いた本文が [Ja] マーカーで始まるかを返す。マーカーとしての使用と地の文中の
// 言及を区別する。
func markerAtStart(text string) bool {
	s := strings.TrimSpace(text)
	for _, leader := range []string{"//", "/*", "*"} {
		if strings.HasPrefix(s, leader) {
			s = strings.TrimSpace(strings.TrimPrefix(s, leader))
			break
		}
	}
	return strings.HasPrefix(s, "[Ja]")
}

// hasJapanese reports whether s contains any kana or kanji.
// [Ja] hasJapanese は s に仮名・漢字が含まれるかを返す。
func hasJapanese(s string) bool { return reJapanese.MatchString(s) }

// isEnglishText reports whether s looks like English: it has Latin letters and
// no Japanese. A Japanese sentence containing a Latin acronym (e.g. "CSRF") is
// therefore not treated as English.
//
// [Ja] isEnglishText は s が英語に見えるか (ラテン文字を含み日本語を含まない) を返す。
// "CSRF" のようなラテン略語を含む日本語文は英語扱いにしない。
func isEnglishText(s string) bool {
	return reLatin.MatchString(s) && !reJapanese.MatchString(s)
}

// commentGroups extracts comment groups from a file. Generated files yield none.
// [Ja] commentGroups はファイルからコメント群を抽出する。生成物は空を返す。
func commentGroups(path, ext string) ([][]commentLine, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isGenerated(src) {
		return nil, nil
	}
	if ext == ".templ" {
		return templCommentGroups(src), nil
	}
	return goCommentGroups(path, src)
}

// goCommentGroups uses go/parser so that "//" inside string literals is ignored.
// [Ja] goCommentGroups は go/parser を使い、文字列リテラル中の "//" を無視する。
func goCommentGroups(path string, src []byte) ([][]commentLine, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var groups [][]commentLine
	for _, cg := range f.Comments {
		var lines []commentLine
		for _, c := range cg.List {
			start := fset.Position(c.Pos()).Line
			for i, sub := range strings.Split(c.Text, "\n") {
				lines = append(lines, commentLine{line: start + i, text: sub})
			}
		}
		groups = append(groups, lines)
	}
	return groups, nil
}

// templCommentGroups groups maximal runs of full-line "//" comments.
// [Ja] templCommentGroups は行頭 "//" コメントの連続を 1 群にまとめる。
func templCommentGroups(src []byte) [][]commentLine {
	var groups [][]commentLine
	var cur []commentLine
	flush := func() {
		if len(cur) > 0 {
			groups = append(groups, cur)
			cur = nil
		}
	}
	for i, raw := range strings.Split(string(src), "\n") {
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(t, "//") {
			cur = append(cur, commentLine{line: i + 1, text: t})
			continue
		}
		flush()
	}
	flush()
	return groups
}

// isGenerated reports whether src carries the standard generated-file header.
// [Ja] isGenerated は src が定型の生成物ヘッダーを持つかを返す。
func isGenerated(src []byte) bool {
	sc := bufio.NewScanner(bytes.NewReader(src))
	for n := 0; sc.Scan() && n < 20; n++ {
		if reGenerated.MatchString(strings.TrimSpace(sc.Text())) {
			return true
		}
	}
	return false
}

// skipDir reports whether a directory should not be walked.
// [Ja] skipDir は走査しないディレクトリかどうかを返す。
func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "bin", "tmp", "static":
		return true
	default:
		return false
	}
}

// addedLines maps absolute file paths to the set of line numbers added since
// base, parsed from "git diff --unified=0 base...HEAD".
//
// [Ja] addedLines は "git diff --unified=0 base...HEAD" を解析し、base 以降に
// 追加された行番号の集合を絶対パスごとに返す。
func addedLines(base string) (map[string]map[int]bool, error) {
	root, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	root = strings.TrimSpace(root)

	diff, err := gitOutput("diff", "--unified=0", "--no-color", base+"...HEAD")
	if err != nil {
		return nil, err
	}

	reFile := regexp.MustCompile(`^\+\+\+ b/(.*)$`)
	reHunk := regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

	result := map[string]map[int]bool{}
	var curAbs string
	sc := bufio.NewScanner(strings.NewReader(diff))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := reFile.FindStringSubmatch(line); m != nil {
			abs, aerr := filepath.Abs(filepath.Join(root, m[1]))
			if aerr != nil {
				curAbs = ""
				continue
			}
			curAbs = abs
			if result[curAbs] == nil {
				result[curAbs] = map[int]bool{}
			}
			continue
		}
		if m := reHunk.FindStringSubmatch(line); m != nil && curAbs != "" {
			start, _ := strconv.Atoi(m[1])
			count := 1
			if m[2] != "" {
				count, _ = strconv.Atoi(m[2])
			}
			for n := 0; n < count; n++ {
				result[curAbs][start+n] = true
			}
		}
	}
	return result, sc.Err()
}

// gitOutput runs a git command and returns its stdout.
// [Ja] gitOutput は git コマンドを実行し標準出力を返す。
func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
