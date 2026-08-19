package httperror

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		locale        string
		wantTitle     string
		wantMessage   string
		wantBackLabel string
	}{
		{
			name:          "日本語",
			locale:        "ja",
			wantTitle:     "ページが見つかりません",
			wantMessage:   "ページが移動または削除された可能性があります。",
			wantBackLabel: "ホームに戻る",
		},
		{
			name:          "英語",
			locale:        "en",
			wantTitle:     "Page not found",
			wantMessage:   "The page may have been moved or deleted.",
			wantBackLabel: "Back to Home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/missing", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()

			NotFound(rr, req)

			if got := rr.Header().Get("HX-Redirect"); got != "" {
				t.Errorf("HX-Redirect = %q, want empty", got)
			}
			assertErrorResponse(t, rr, http.StatusNotFound, tt.locale, tt.wantTitle, tt.wantMessage, tt.wantBackLabel)
		})
	}
}

// TestNotFound_HTMXRedirect fixes that an HTMX request is sent to the standalone page. A row
// deleted in another tab leaves a button that answers 404, and the DB lists issue hx-delete
// without hx-target, so the document would otherwise be swapped into the button.
//
// [Ja] TestNotFound_HTMXRedirect は、HTMX リクエストが全画面ページへ送られることを固定する。
// 別タブで先に消された行のボタンは 404 を受け取り、DB 一覧の発行する hx-delete は hx-target を
// 指定していないため、そのままでは文書が押したボタンの中にスワップされる。
func TestNotFound_HTMXRedirect(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodDelete, "/db/episodes/1", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	NotFound(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != NotFoundPath {
		t.Errorf("HX-Redirect = %q, want %q", got, NotFoundPath)
	}
	if got := rr.Header().Get("Location"); got != "" {
		t.Errorf("Location = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusNotFound,
		"ja",
		"ページが見つかりません",
		"ページが移動または削除された可能性があります。",
		"ホームに戻る",
	)
}

// TestNotFound_HTMXRequestOnPageIsNotRedirected fixes that the page itself never answers with
// HX-Redirect. It is the target of the redirect a missing resource sends, so pointing an HTMX
// request at the path it already asked for would say nothing.
//
// [Ja] TestNotFound_HTMXRequestOnPageIsNotRedirected は、ページ本体が HX-Redirect を返さない
// ことを固定する。ここは存在しないリソースへの応答が送るリダイレクトの遷移先であり、HTMX
// リクエストに対して既に要求されているパスへの遷移を指示しても意味がないため。
func TestNotFound_HTMXRequestOnPageIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, NotFoundPath, nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	NotFound(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusNotFound,
		"ja",
		"ページが見つかりません",
		"ページが移動または削除された可能性があります。",
		"ホームに戻る",
	)
}

func TestForbidden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		locale      string
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "日本語",
			locale:      "ja",
			wantTitle:   "アクセスできません",
			wantMessage: "この操作を行う権限がありません。",
		},
		{
			name:        "英語",
			locale:      "en",
			wantTitle:   "Access denied",
			wantMessage: "You do not have permission to perform this operation.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/db/works/1/episodes", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()

			Forbidden(rr, req)

			if got := rr.Header().Get("HX-Redirect"); got != "" {
				t.Errorf("HX-Redirect = %q, want empty", got)
			}
			wantBackLabel := "ホームに戻る"
			if tt.locale == "en" {
				wantBackLabel = "Back to Home"
			}
			assertErrorResponse(t, rr, http.StatusForbidden, tt.locale, tt.wantTitle, tt.wantMessage, wantBackLabel)
		})
	}
}

// TestForbidden_HTMXRedirect fixes that an HTMX request is sent to the standalone page. The
// delete and archive buttons on the DB lists are hx-delete without hx-target, so a 403 document
// would otherwise be swapped into the button that was pressed.
//
// [Ja] TestForbidden_HTMXRedirect は、HTMX リクエストが全画面ページへ送られることを固定する。
// DB 一覧の削除・非公開ボタンは hx-target を指定していない hx-delete のため、そのままでは
// 403 の文書が押したボタンの中にスワップされる。
func TestForbidden_HTMXRedirect(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodDelete, "/db/episodes/1", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	Forbidden(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != ForbiddenPath {
		t.Errorf("HX-Redirect = %q, want %q", got, ForbiddenPath)
	}
	if got := rr.Header().Get("Location"); got != "" {
		t.Errorf("Location = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusForbidden,
		"ja",
		"アクセスできません",
		"この操作を行う権限がありません。",
		"ホームに戻る",
	)
}

// TestForbidden_HTMXRequestOnPageIsNotRedirected fixes that the page itself never answers with
// HX-Redirect. It is the target of the redirect a refused request sends, so pointing an HTMX
// request at the path it already asked for would say nothing.
//
// [Ja] TestForbidden_HTMXRequestOnPageIsNotRedirected は、ページ本体が HX-Redirect を返さない
// ことを固定する。ここは拒否された要求が送るリダイレクトの遷移先であり、HTMX リクエストに
// 対して既に要求されているパスへの遷移を指示しても意味がないため。
func TestForbidden_HTMXRequestOnPageIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, ForbiddenPath, nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	Forbidden(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusForbidden,
		"ja",
		"アクセスできません",
		"この操作を行う権限がありません。",
		"ホームに戻る",
	)
}

func TestInvalidCSRFToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		locale      string
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "日本語",
			locale:      "ja",
			wantTitle:   "フォームを送信できませんでした",
			wantMessage: "ページを開いたまま時間が経ったか、別のタブでログインし直した可能性があります。ページを開き直してから、もう一度送信してください。",
		},
		{
			name:        "英語",
			locale:      "en",
			wantTitle:   "Could not submit the form",
			wantMessage: "The page may have been left open too long, or you may have signed in again in another tab. Please reopen the page and submit again.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, InvalidCSRFTokenPath, nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()

			InvalidCSRFToken(rr, req)

			if got := rr.Header().Get("HX-Redirect"); got != "" {
				t.Errorf("HX-Redirect = %q, want empty", got)
			}
			wantBackLabel := "ホームに戻る"
			if tt.locale == "en" {
				wantBackLabel = "Back to Home"
			}
			assertErrorResponse(t, rr, http.StatusForbidden, tt.locale, tt.wantTitle, tt.wantMessage, wantBackLabel)
		})
	}
}

// TestInvalidCSRFToken_HTMXRequestOnPageIsNotRedirected fixes that the page itself never answers
// with HX-Redirect. It is the target of the redirect a rejected submission sends, so pointing an
// HTMX request at the path it already asked for would say nothing.
//
// [Ja] TestInvalidCSRFToken_HTMXRequestOnPageIsNotRedirected は、ページ本体が HX-Redirect を
// 返さないことを固定する。ここは拒否された送信が送るリダイレクトの遷移先であり、HTMX
// リクエストに対して既に要求されているパスへの遷移を指示しても意味がないため。
func TestInvalidCSRFToken_HTMXRequestOnPageIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, InvalidCSRFTokenPath, nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InvalidCSRFToken(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusForbidden,
		"ja",
		"フォームを送信できませんでした",
		"ページを開いたまま時間が経ったか、別のタブでログインし直した可能性があります。ページを開き直してから、もう一度送信してください。",
		"ホームに戻る",
	)
}

func TestInvalidCSRFToken_NonHTMXRequestIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/db/works/1/episodes", nil)
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InvalidCSRFToken(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusForbidden,
		"ja",
		"フォームを送信できませんでした",
		"ページを開いたまま時間が経ったか、別のタブでログインし直した可能性があります。ページを開き直してから、もう一度送信してください。",
		"ホームに戻る",
	)
}

func TestInvalidCSRFToken_HTMXRedirect(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/db/works/1/episodes", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InvalidCSRFToken(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != InvalidCSRFTokenPath {
		t.Errorf("HX-Redirect = %q, want %q", got, InvalidCSRFTokenPath)
	}
	if got := rr.Header().Get("Location"); got != "" {
		t.Errorf("Location = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusForbidden,
		"ja",
		"フォームを送信できませんでした",
		"ページを開いたまま時間が経ったか、別のタブでログインし直した可能性があります。ページを開き直してから、もう一度送信してください。",
		"ホームに戻る",
	)
}

// TestInvalidCSRFToken_DoesNotReuseForbiddenMessage fixes that the CSRF page carries its own copy.
// Forbidden explains a missing permission, which is not what happened and not what the reader
// should act on: the page they submitted from is stale, and resubmitting from a fresh one works.
//
// [Ja] TestInvalidCSRFToken_DoesNotReuseForbiddenMessage は CSRF のページが専用の文言を持つことを
// 固定する。Forbidden の文言は権限不足の説明で、起きたことと読み手が取るべき行動のどちらとも
// 合わない (送信元のページが古いだけで、開き直して送り直せば通る)。
func TestInvalidCSRFToken_DoesNotReuseForbiddenMessage(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, InvalidCSRFTokenPath, nil)
	ctx := i18n.SetLocale(req.Context(), "ja")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	InvalidCSRFToken(rr, req)

	if body := rr.Body.String(); strings.Contains(body, i18n.T(ctx, "error_forbidden_message")) {
		t.Error("CSRF エラーページに権限不足の文言が使われています")
	}
}

func TestBadGateway(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		locale        string
		wantTitle     string
		wantMessage   string
		wantBackLabel string
	}{
		{
			name:          "日本語",
			locale:        "ja",
			wantTitle:     "サービスに接続できません",
			wantMessage:   "しばらくしてから、もう一度お試しください。",
			wantBackLabel: "ホームに戻る",
		},
		{
			name:          "英語",
			locale:        "en",
			wantTitle:     "Cannot connect to the service",
			wantMessage:   "Please try again later.",
			wantBackLabel: "Back to Home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/works", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()

			BadGateway(rr, req)

			assertErrorResponse(t, rr, http.StatusBadGateway, tt.locale, tt.wantTitle, tt.wantMessage, tt.wantBackLabel)
		})
	}
}

func TestInternalServerError(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/broken", nil)
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InternalServerError(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusInternalServerError,
		"ja",
		"問題が発生しました",
		"しばらくしてから、もう一度お試しください。",
		"ホームに戻る",
	)
	if strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Error("500 レスポンスに内部エラーの文言を含めてはいけません")
	}
}

// TestInternalServerError_HTMXRedirect fixes that an HTMX request is sent to the standalone page.
// The delete and archive buttons on the DB lists are hx-delete without hx-target, so a 500
// document would otherwise be swapped into the button that was pressed.
//
// [Ja] TestInternalServerError_HTMXRedirect は、HTMX リクエストが全画面ページへ送られることを
// 固定する。DB 一覧の削除・非公開ボタンは hx-target を指定していない hx-delete のため、そのまま
// では 500 の文書が押したボタンの中にスワップされる。
func TestInternalServerError_HTMXRedirect(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodDelete, "/db/episodes/1", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InternalServerError(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != InternalServerErrorPath {
		t.Errorf("HX-Redirect = %q, want %q", got, InternalServerErrorPath)
	}
	if got := rr.Header().Get("Location"); got != "" {
		t.Errorf("Location = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusInternalServerError,
		"ja",
		"問題が発生しました",
		"しばらくしてから、もう一度お試しください。",
		"ホームに戻る",
	)
}

// TestInternalServerError_HTMXRequestOnPageIsNotRedirected fixes that the page itself never
// answers with HX-Redirect. It is the target of the redirect a failed request sends, so pointing
// an HTMX request at the path it already asked for would say nothing.
//
// [Ja] TestInternalServerError_HTMXRequestOnPageIsNotRedirected は、ページ本体が HX-Redirect を
// 返さないことを固定する。ここは失敗した要求への応答が送るリダイレクトの遷移先であり、HTMX
// リクエストに対して既に要求されているパスへの遷移を指示しても意味がないため。
func TestInternalServerError_HTMXRequestOnPageIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, InternalServerErrorPath, nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InternalServerError(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
	assertErrorResponse(
		t,
		rr,
		http.StatusInternalServerError,
		"ja",
		"問題が発生しました",
		"しばらくしてから、もう一度お試しください。",
		"ホームに戻る",
	)
}

// TestBadGateway_HTMXRequestIsNotRedirected fixes that 502 is the one error without a page of its
// own. It has no route to send an HTMX request to, so it answers in place and is swapped like any
// other response. That is intended rather than an omission: 502 comes from the reverse proxy on
// paths Rails serves, while the requests that would swap into their own element are issued by the
// pages Go serves.
//
// [Ja] TestBadGateway_HTMXRequestIsNotRedirected は、502 が専用のページを持たない唯一のエラーで
// あることを固定する。HTMX リクエストの送り先となるルートを持たないため、その場で応答し、他の
// レスポンスと同じようにスワップされる。これは漏れではなく意図した線引きで、502 が返るのは
// Rails の配信するパスに対するリバースプロキシの失敗であるのに対し、リクエスト元自身へ
// スワップしてしまう要求は Go の配信するページから発行されるため。
func TestBadGateway_HTMXRequestIsNotRedirected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodDelete, "/db/episodes/1", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	BadGateway(rr, req)

	if got := rr.Header().Get("HX-Redirect"); got != "" {
		t.Errorf("HX-Redirect = %q, want empty", got)
	}
}

// TestNoindexIsDeclaredOnEveryResponse fixes that every error response tells crawlers not to
// index it. The /errors/* routes answer GET requests and robots.txt disallows only /db/, so
// these are crawlable URLs carrying none of the site's content. Their 4xx and 5xx statuses keep
// them out of a search index already, and this is the declaration in the page that says the same.
//
// [Ja] TestNoindexIsDeclaredOnEveryResponse は、どのエラーレスポンスもクローラーに索引しない
// よう伝えることを固定する。/errors/* のルートは GET に応じており、robots.txt が禁じているのは
// /db/ だけのため、これらはサイトの中身を持たないままクロール可能な URL として存在している。
// 索引から外れること自体は 4xx / 5xx のステータスで既に満たされており、本宣言はそれをページ側
// でも述べるもの。
func TestNoindexIsDeclaredOnEveryResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		render func(http.ResponseWriter, *http.Request)
	}{
		{name: "404", render: NotFound},
		{name: "認可の 403", render: Forbidden},
		{name: "CSRF の 403", render: InvalidCSRFToken},
		{name: "502", render: BadGateway},
		{name: "500", render: InternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/db/works", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
			rr := httptest.NewRecorder()

			tt.render(rr, req)

			if !strings.Contains(rr.Body.String(), `<meta name="robots" content="noindex">`) {
				t.Error("noindex の宣言が含まれていません")
			}
		})
	}
}

func assertErrorResponse(
	t *testing.T,
	rr *httptest.ResponseRecorder,
	wantStatus int,
	wantLocale string,
	wantTitle string,
	wantMessage string,
	wantBackLabel string,
) {
	t.Helper()

	if rr.Code != wantStatus {
		t.Errorf("status = %d, want %d", rr.Code, wantStatus)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		`<html lang="` + wantLocale + `">`,
		"<title>" + wantTitle + " | Annict</title>",
		"<h1",
		wantTitle,
		wantMessage,
		`href="/"`,
		wantBackLabel,
		`class="error-link"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response body does not contain %q", expected)
		}
	}
}
