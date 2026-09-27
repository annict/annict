package components

import (
	"strings"
	"testing"
)

// menuItemHTMLは指定したパスへリンクするサイドバー項目のマークアップを返す。
func menuItemHTML(t *testing.T, html string, path string) string {
	t.Helper()

	start := strings.Index(html, `<a href="`+path+`"`)
	if start < 0 {
		t.Fatalf("%qへのリンクが描画されていません", path)
	}
	end := strings.Index(html[start:], "</a>")
	if end < 0 {
		t.Fatalf("%qへのリンクが閉じられていません", path)
	}
	return html[start : start+end]
}

// sidebarGroupHTMLは指定した見出しで識別されるサイドバーグループのマークアップを返す。
func sidebarGroupHTML(t *testing.T, html string, headingID string) string {
	t.Helper()

	start := strings.Index(html, `aria-labelledby="`+headingID+`"`)
	if start < 0 {
		t.Fatalf("%qで識別されるグループが描画されていません", headingID)
	}

	const groupEnd = "</ul></div>"
	end := strings.Index(html[start:], groupEnd)
	if end < 0 {
		t.Fatalf("%qで識別されるグループが閉じられていません", headingID)
	}

	return html[start : start+end+len(groupEnd)]
}
