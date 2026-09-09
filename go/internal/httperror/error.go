// Package httperror renders shared, localized HTTP error responses for handlers.
//
// [Ja] Package httperror はハンドラー共通のローカライズ済み HTTP エラーレスポンスを描画する。
package httperror

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/layouts"
	errorpages "github.com/annict/annict/go/internal/templates/pages/errors"
)

// NotFoundPath is the path of the full-page missing resource error.
//
// [Ja] NotFoundPath はリソースが見つからない場合の全画面エラーページのパスです。
const NotFoundPath = "/errors/not-found"

// ForbiddenPath is the path of the full-page authorization error.
//
// [Ja] ForbiddenPath は権限不足の全画面エラーページのパスです。
const ForbiddenPath = "/errors/forbidden"

// InvalidCSRFTokenPath is the path of the full-page invalid CSRF token error.
//
// [Ja] InvalidCSRFTokenPath は無効な CSRF トークンの全画面エラーページのパスです。
const InvalidCSRFTokenPath = "/errors/invalid-csrf-token"

// InternalServerErrorPath is the path of the full-page internal error.
//
// [Ja] InternalServerErrorPath は内部エラーの全画面エラーページのパスです。
const InternalServerErrorPath = "/errors/internal-server-error"

// NotFound renders the shared 404 page. It answers both a request for something that is not
// there and the page at NotFoundPath. The copy is separate from InternalServerError because
// the reader's next action differs: go back to the list and see what is actually there,
// rather than wait and try the same thing again.
//
// [Ja] NotFound は共通の 404 ページを描画する。存在しないものへの要求への応答と、
// NotFoundPath が返すページ本体の両方を兼ねる。文言を InternalServerError と分けているのは、
// 読み手に求める次の行動が違うため (時間を置いて同じ操作をやり直すのではなく、一覧に戻って
// 実際の状態を見る)。
func NotFound(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, NotFoundPath)
	render(w, r, http.StatusNotFound, "error_not_found_title", "error_not_found_message")
}

// Forbidden renders the shared 403 page. It states that the request was understood and
// refused, which 404 would not: hiding the resource only makes sense where its existence is
// itself private, and the pages this serves are reached from links the viewer can see. It
// answers both a refused request and the page at ForbiddenPath.
//
// [Ja] Forbidden は共通の 403 ページを描画する。要求は理解されたうえで拒否されたことを述べる
// (404 では表せない)。リソースの存在自体を隠す意味があるのは存在が秘密である場合だけで、
// ここが応じる画面は閲覧者から見えるリンクから辿り着くものであるため。拒否した要求への応答と、
// ForbiddenPath が返すページ本体の両方を兼ねる。
func Forbidden(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, ForbiddenPath)
	render(w, r, http.StatusForbidden, "error_forbidden_title", "error_forbidden_message")
}

// InvalidCSRFToken renders the 403 page shown when a form submission fails CSRF verification.
// It serves both as the body of a rejected submission and as the page at InvalidCSRFTokenPath.
// It keeps the 403 that authorization failures use, because what happened is the same in both:
// the server understood the request and refuses to act on it. 422 would name the status the
// handlers already use for a form re-rendered with validation errors, so the same status would
// carry two different response shapes. The copy is separate from Forbidden because the reader's
// next action differs: reload the page and submit again, rather than obtain a permission.
//
// [Ja] InvalidCSRFToken は、フォーム送信の CSRF 検証に失敗したときに出す 403 ページを描画する。
// 拒否した送信への応答本文と、InvalidCSRFTokenPath が返すページ本体の両方を兼ねる。
// 認可の失敗と同じ 403 を使うのは、起きたことがどちらも「要求を理解したうえでサーバーが
// 応じない」で同じであるため。422 はバリデーションエラーでフォームを再描画するときに
// ハンドラーが既に使っている状態のため、同じステータスが 2 種類の応答の形を持つことになる。
// 文言を Forbidden と分けているのは、読み手に求める次の行動が違うため (権限の獲得ではなく、
// ページを開き直しての再送信)。
func InvalidCSRFToken(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, InvalidCSRFTokenPath)
	render(w, r, http.StatusForbidden, "error_invalid_csrf_token_title", "error_invalid_csrf_token_message")
}

// redirectHTMXToPage tells an HTMX request to navigate to the standalone error page at pagePath.
// htmx swaps every response except 204 and 304, and an hx-delete without hx-target swaps into the
// element that issued it, so without this the whole document would be placed inside the button
// that was clicked. The status and body are left alone, so what a plain form submission receives
// is unchanged and responses that must stay indistinguishable still are.
//
// Serving pagePath itself is the one case with nowhere to navigate to, so the request that asked
// for the page is never told to go there again.
//
// [Ja] redirectHTMXToPage は、HTMX リクエストに pagePath の全画面エラーページへの遷移を指示する。
// htmx は 204 と 304 以外のレスポンスをスワップし、hx-target を指定していない hx-delete の
// スワップ先はリクエスト元自身になるため、指示しなければ完全な文書が押したボタンの中へ挿入される。
// ステータスと本文には手を加えないため、通常のフォーム送信が受け取るものは変わらず、区別できて
// はならない応答どうしが区別できるようになることもない。
//
// pagePath 自体を配信する場合だけは遷移先が無いため、そのページを要求したリクエストに同じページ
// への遷移を指示することはしない。
func redirectHTMXToPage(w http.ResponseWriter, r *http.Request, pagePath string) {
	if r.URL.Path == pagePath {
		return
	}

	if r.Header.Get("HX-Request") != "true" {
		return
	}

	w.Header().Set("HX-Redirect", pagePath)
}

// BadGateway renders the shared 502 page, used when an upstream the request depends on could
// not be reached. It says the same thing the 500 page says about retrying, because the reader
// can act on neither: what differs is only which side of the application failed.
//
// It is the one error here without a page of its own, because it is returned only by the
// reverse proxy for paths Rails serves, while the HTMX requests that would swap a document
// into the element that issued them come from the pages Go serves.
//
// [Ja] BadGateway は共通の 502 ページを描画する。リクエストが依存する上流に到達できなかった
// 場合に使う。再試行について述べる内容は 500 ページと同じ。読み手にはどちらも打つ手が無く、
// 違うのはアプリケーションのどちら側が失敗したかだけであるため。
//
// ここで唯一専用のページを持たないのは、本関数が応じるのが Rails の配信するパスに対する
// リバースプロキシの失敗だけであるのに対し、文書をリクエスト元の要素へスワップしてしまう
// HTMX リクエストは Go の配信するページから発行されるものだけであるため。
func BadGateway(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusBadGateway, "error_bad_gateway_title", "error_bad_gateway_message")
}

// InternalServerError renders the shared 500 page without exposing the underlying error. It
// answers both a failed request and the page at InternalServerErrorPath.
//
// The page reports 500 on its own route too, so a route that rendered successfully names a
// server failure. Answering 200 there would instead say the request succeeded while the body
// tells the reader it did not, which is what keeps the page out of a search index and what a
// client reads to decide whether the operation went through. The route is reached only by a
// request an actual 500 sent there, so the status it reports tracks real failures.
//
// [Ja] InternalServerError は内部エラーを公開せず、共通の 500 ページを描画する。失敗した要求
// への応答と、InternalServerErrorPath が返すページ本体の両方を兼ねる。
//
// 専用ルートで配信する場合もページは 500 を名乗るため、描画に成功したルートがサーバーの失敗を
// 報告することになる。一方でここを 200 にすると、本文が読み手に失敗を伝えているのに要求は
// 成功したと述べることになる。ステータスは検索エンジンがこのページを索引から外す根拠であり、
// クライアントが操作の成否を読み取る先でもある。本ルートに到達するのは実際に 500 が起きた要求
// が送り込まれた場合だけなので、報告されるステータスは実際の失敗に対応する。
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, InternalServerErrorPath)
	render(w, r, http.StatusInternalServerError, "error_internal_server_title", "error_internal_server_message")
}

func render(w http.ResponseWriter, r *http.Request, status int, titleKey string, messageKey string) {
	ctx := r.Context()
	title := i18n.T(ctx, titleKey)
	backLink := &errorpages.BackLink{
		URL:  "/",
		Text: i18n.T(ctx, "error_back_to_home"),
	}
	component := layouts.Error(
		title,
		errorpages.HTTPError(title, i18n.T(ctx, messageKey), backLink),
	)

	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "HTTPエラーページのレンダリングに失敗", "status", status, "error", err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "HTTPエラーレスポンスの書き込みに失敗", "status", status, "error", err)
	}
}
