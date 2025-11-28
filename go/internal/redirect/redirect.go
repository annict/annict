// Package redirect はリダイレクトURLのバリデーションを提供する
package redirect

import "strings"

// ValidateBackURL は back パラメータの値が安全かどうかを検証する
// 安全な場合は true を返し、危険な場合は false を返す
//
// オープンリダイレクト攻撃を防ぐため、以下のルールでバリデーションを行う：
// - 空文字は無効
// - "/" で始まらない場合は無効（相対パスのみ許可）
// - "//" で始まる場合は無効（プロトコル相対URL）
func ValidateBackURL(backURL string) bool {
	// 空文字の場合は無効
	if backURL == "" {
		return false
	}

	// "/" で始まらない場合は無効（相対パスのみ許可）
	if !strings.HasPrefix(backURL, "/") {
		return false
	}

	// "//" で始まる場合は無効（プロトコル相対URL）
	if strings.HasPrefix(backURL, "//") {
		return false
	}

	return true
}

// GetSafeRedirectURL は安全なリダイレクトURLを返す
// backURL が無効な場合はデフォルトURL（"/"）を返す
func GetSafeRedirectURL(backURL string) string {
	if ValidateBackURL(backURL) {
		return backURL
	}
	return "/"
}
