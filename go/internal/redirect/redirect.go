// Package redirectはリダイレクトURLのバリデーションを提供する
package redirect

import (
	"net/url"
	"path"
	"strings"
)

// ValidateBackURLはbackパラメータの値が安全かどうかを検証する
// 安全な場合はtrueを返し、危険な場合はfalseを返す
//
// オープンリダイレクト攻撃を防ぐため、以下のルールでバリデーションを行う：
// - 空文字は無効
// - "/" で始まらない場合は無効 (相対パスのみ許可)
// - "//" で始まる場合は無効 (プロトコル相対URL)
// - バックスラッシュや制御文字を含む値、解析できないURLは無効
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

	// ブラウザはバックスラッシュをスラッシュとして扱うため、外部サイトへの転送を防ぐ。
	if strings.Contains(backURL, `\`) {
		return false
	}

	// ブラウザはURLを解析する前にタブと改行を取り除くため、残すと "/\t/example.org" が "//example.org" になる。
	// url.Parseはフラグメントの中の制御文字を拒否せず、http.Redirectはフラグメントもパスとして正規化して
	// Locationに含めるため、フラグメントを含む値の全体で拒否する。
	if strings.ContainsFunc(backURL, isASCIIControl) {
		return false
	}

	parsed, err := url.Parse(backURL)
	if err != nil {
		return false
	}
	if parsed.Scheme != "" || parsed.Opaque != "" || parsed.Host != "" {
		return false
	}

	return true
}

func isASCIIControl(r rune) bool {
	return r < 0x20 || r == 0x7f
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
// 先として安全かどうかを返す。ValidateBackURLのオープンリダイレクト対策に加えて、値が
// http.Redirectにもブラウザにも書き換えられない形であり、パスがAnnict DBのトップ (/db) か
// その配下であることを要求する。確認画面に入る導線はAnnict DBの一覧だけであり、一覧は
// ブラウザが正規化した後のパスを渡す。それ以外の値は呼び出し元が生成していないため、拒否する。
func ValidateDBReturnURL(returnTo string) bool {
	if !ValidateBackURL(returnTo) {
		return false
	}

	// HTTPヘッダーの処理やブラウザは先頭・末尾の半角スペースを取り除くため、
	// 検査後に別のパスへ変わる値は拒否する。
	if strings.Trim(returnTo, " ") != returnTo {
		return false
	}

	// ブラウザはパーセントエンコードしたドット ("%2e") をドットとして扱い、エンコードしたスラッシュ ("%2f") を
	// 区切りとして扱わないため、Goでデコードしたパスと解釈が食い違う。"?" より前に含む値は拒否する。
	beforeQuery, _, _ := strings.Cut(returnTo, "?")
	lower := strings.ToLower(beforeQuery)
	if strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") {
		return false
	}

	// http.Redirectは "?" より前をフラグメントも含めて、ブラウザは "#" より前を正規化する。
	// どちらにも書き換えられない値に限り、検査したパスとブラウザの最終的な転送先を一致させる。
	p, _, _ := strings.Cut(beforeQuery, "#")
	if !isCleanPath(beforeQuery) || !isCleanPath(p) {
		return false
	}

	return p == "/db" || strings.HasPrefix(p, "/db/")
}

// isCleanPathは、sが末尾のスラッシュを除いてpath.Cleanで変わらないかどうかを返す。
func isCleanPath(s string) bool {
	cleaned := path.Clean(s)
	return cleaned == s || cleaned+"/" == s
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
