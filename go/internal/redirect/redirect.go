// Package redirectはリダイレクトURLのバリデーションを提供する
package redirect

import "strings"

// ValidateBackURLはbackパラメータの値が安全かどうかを検証する
// 安全な場合はtrueを返し、危険な場合はfalseを返す
//
// オープンリダイレクト攻撃を防ぐため、以下のルールでバリデーションを行う：
// - 空文字は無効
// - "/" で始まらない場合は無効 (相対パスのみ許可)
// - "//" で始まる場合は無効 (プロトコル相対URL)
func ValidateBackURL(backURL string) bool {
	// 空文字の場合は無効
	if backURL == "" {
		return false
	}

	// "/" で始まらない場合は無効 (相対パスのみ許可)
	if !strings.HasPrefix(backURL, "/") {
		return false
	}

	// "//" で始まる場合は無効 (プロトコル相対URL)
	if strings.HasPrefix(backURL, "//") {
		return false
	}

	return true
}

// GetSafeRedirectURLは安全なリダイレクトURLを返す
// backURLが無効な場合はデフォルトURL ("/") を返す
func GetSafeRedirectURL(backURL string) string {
	if ValidateBackURL(backURL) {
		return backURL
	}
	return "/"
}

// ValidateDBReturnURLはAnnict DBの確認画面に渡された戻り先の値が、ブラウザを送り返す
// 先として安全かどうかを返す。ValidateBackURLのオープンリダイレクト対策に加えて、パスが /db
// または /db/ 配下であることを要求する。確認画面に入る導線はAnnict DBの一覧だけであり、それ
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

// GetSafeDBReturnURLはreturnToが安全なAnnict DBのパスならそれを、そうでなければ
// fallbackを返す。fallbackには確認画面が属する画面を渡す。値が無い場合や弾かれた場合でも、
// 読み手が行き先を失わず一覧に着地するようにするため。
func GetSafeDBReturnURL(returnTo string, fallback string) string {
	if ValidateDBReturnURL(returnTo) {
		return returnTo
	}
	return fallback
}
