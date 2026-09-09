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

// ValidateDBReturnURL reports whether a return-to value handed to an Annict DB confirmation
// screen is safe to send a browser back to. On top of the open-redirect rules ValidateBackURL
// applies, the path must be /db or sit under /db/: the confirmation screens are only ever
// entered from an Annict DB listing, so anything else is a value the caller did not produce.
//
// [Ja] ValidateDBReturnURL は Annict DB の確認画面に渡された戻り先の値が、ブラウザを送り返す
// 先として安全かどうかを返す。ValidateBackURL のオープンリダイレクト対策に加えて、パスが /db
// または /db/ 配下であることを要求する。確認画面に入る導線は Annict DB の一覧だけであり、それ
// 以外は呼び出し元が生成していない値であるため。
func ValidateDBReturnURL(returnTo string) bool {
	if !ValidateBackURL(returnTo) {
		return false
	}

	path := returnTo
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}

	return path == "/db" || strings.HasPrefix(path, "/db/")
}

// GetSafeDBReturnURL returns returnTo when it is a safe Annict DB path, and fallback otherwise.
// The fallback is the screen the confirmation belongs to, so a missing or rejected value still
// lands the reader on a list rather than nowhere.
//
// [Ja] GetSafeDBReturnURL は returnTo が安全な Annict DB のパスならそれを、そうでなければ
// fallback を返す。fallback には確認画面が属する画面を渡す。値が無い場合や弾かれた場合でも、
// 読み手が行き先を失わず一覧に着地するようにするため。
func GetSafeDBReturnURL(returnTo string, fallback string) string {
	if ValidateDBReturnURL(returnTo) {
		return returnTo
	}
	return fallback
}
