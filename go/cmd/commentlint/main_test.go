package main

import (
	"regexp"
	"strings"
	"testing"
)

// reMarker matches the [Ja] translation marker. It is only needed by tests,
// so it lives here rather than in the production code.
//
// [Ja] reMarker は [Ja] 翻訳マーカーにマッチする。テストでのみ必要なため、
// 本番コードではなくこちらに置く。
var reMarker = regexp.MustCompile(`\[Ja\]`)

// helper builds a comment group from raw "//" lines, numbering them from 1.
// [Ja] helper は "//" 行から 1 始まりで番号付けしたコメント群を作る。
func group(texts ...string) []commentLine {
	lines := make([]commentLine, len(texts))
	for i, t := range texts {
		lines[i] = commentLine{line: i + 1, text: t}
	}
	return lines
}

// condsOf returns the sorted-ish list of condition numbers in findings.
// [Ja] condsOf は findings に含まれる条件番号の一覧を返す。
func condsOf(fs []finding) []int {
	got := make([]int, len(fs))
	for i, f := range fs {
		got[i] = f.cond
	}
	return got
}

func TestCheckGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		lines     []commentLine
		wantConds []int
	}{
		{
			name:      "correct multi-line block",
			lines:     group("// New post form.", "//", "// [Ja] 新規投稿フォーム。"),
			wantConds: nil,
		},
		{
			name:      "correct one-line pair",
			lines:     group("// Hash the password.", "// [Ja] パスワードをハッシュ化する。"),
			wantConds: nil,
		},
		{
			name:      "correct inline pair on a single line",
			lines:     group("// Hash the password. [Ja] パスワードをハッシュ化する。"),
			wantConds: nil,
		},
		{
			name: "prose mention of [Ja] in English text is not a marker",
			// A [Ja] mention inside English prose must not be treated as a marker.
			// [Ja] 英語の地の文中の [Ja] 言及はマーカー扱いしない。
			lines:     group("// The [Ja] marker leads the Japanese translation block.", "// [Ja] 日本語訳ブロックを先導するマーカー。"),
			wantConds: nil,
		},
		{
			name: "marker on an English line (001/003 inversion)",
			// Japanese leads and the English line carries the marker (the 001/003 inversion).
			// [Ja] 英語行に [Ja] が付く誤り。日本語先・英語に [Ja]。
			lines:     group("// POST /posts は後続タスクで登録する。", "// [Ja] POST /posts is registered in a later task."),
			wantConds: []int{1},
		},
		{
			name: "Japanese-only with stray marker, no English above (002)",
			// No English block; [Ja] is misused as a separator (same shape as 002).
			// [Ja] 英語ブロックが無く [Ja] を区切りに誤用 (002 と同型)。
			lines:     group("// CSRF トークンを設定する。", "// [Ja] /new は RequireAuth 配下のため context 経由で渡る。"),
			wantConds: []int{2},
		},
		{
			name:      "marker-only Japanese comment without any English block",
			lines:     group("// [Ja] 日本語のみのコメント。"),
			wantConds: []int{2},
		},
		{
			name: "Japanese line with a Latin acronym is still Japanese (not an English block)",
			// A Japanese line stays Japanese even when it contains the Latin acronym CSRF.
			// [Ja] ラテン略語 CSRF を含んでも日本語行は英語ブロックにならない。
			lines:     group("// CSRF を検証する。", "// [Ja] これは説明である。"),
			wantConds: []int{2},
		},
		{
			name:      "more than one marker in a group",
			lines:     group("// English line.", "// [Ja] 日本語。", "// [Ja] 二つ目のマーカー。"),
			wantConds: []int{3},
		},
		{
			name: "marker on English line that is also a duplicate",
			// The second marker sits on an English line, so both condition 3 and condition 1 fire.
			// [Ja] 2 つ目のマーカーが英語行 → 条件 3 と条件 1 の両方。
			lines:     group("// English.", "// [Ja] 日本語。", "// [Ja] second english marker."),
			wantConds: []int{3, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := condsOf(checkGroup(tt.lines))
			if !equalInts(got, tt.wantConds) {
				t.Errorf("checkGroup conds = %v, want %v", got, tt.wantConds)
			}
		})
	}
}

func TestCheckGroupReportsLineNumber(t *testing.T) {
	t.Parallel()

	lines := []commentLine{
		{line: 766, text: "// POST /posts は後続タスクで登録する。"},
		{line: 767, text: "// [Ja] POST /posts is registered in a later task."},
	}
	fs := checkGroup(lines)
	if len(fs) != 1 {
		t.Fatalf("got %d findings, want 1", len(fs))
	}
	if fs[0].line != 767 {
		t.Errorf("finding line = %d, want 767", fs[0].line)
	}
}

func TestIsEnglishText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want bool
	}{
		{"Set the CSRF token on the context.", true},
		{"CSRF トークンを設定する。", false}, // Latin acronym in Japanese is not English.
		{"// ", false}, // marker leader only (the "before" slice for "// [Ja] ...").
		{"日本語のみ。", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isEnglishText(tt.text); got != tt.want {
			t.Errorf("isEnglishText(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestGoCommentGroupsIgnoresStringLiterals(t *testing.T) {
	t.Parallel()

	// The "[Ja]" below lives inside a string literal, so the parser must not
	// surface it as a comment (otherwise the tool would flag its own fixtures).
	//
	// [Ja] 下の "[Ja]" は文字列リテラル内にあるため、コメントとして抽出されては
	// ならない (さもないとツール自身のフィクスチャを誤検出する)。
	src := []byte(strings.Join([]string{
		"package sample",
		"",
		"// Greet returns a greeting.",
		"// [Ja] Greet は挨拶を返す。",
		"func Greet() string {",
		"\treturn \"// [Ja] this is not a comment\"",
		"}",
	}, "\n"))

	groups, err := goCommentGroups("sample.go", src)
	if err != nil {
		t.Fatalf("goCommentGroups: %v", err)
	}

	var markerLines int
	for _, g := range groups {
		for _, cl := range g {
			if reMarker.MatchString(cl.text) {
				markerLines++
			}
		}
	}
	if markerLines != 1 {
		t.Errorf("found %d comment lines with [Ja], want 1 (string literal must be ignored)", markerLines)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
