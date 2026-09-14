// Package httperrorはハンドラー共通のローカライズ済みHTTPエラーレスポンスを描画する。
package httperror

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/layouts"
	errorpages "github.com/annict/annict/go/internal/templates/pages/errors"
)

// NotFoundPathはリソースが見つからない場合の全画面エラーページのパスです。
const NotFoundPath = "/errors/not-found"

// ForbiddenPathは権限不足の全画面エラーページのパスです。
const ForbiddenPath = "/errors/forbidden"

// InvalidCSRFTokenPathは無効なCSRFトークンの全画面エラーページのパスです。
const InvalidCSRFTokenPath = "/errors/invalid-csrf-token"

// InternalServerErrorPathは内部エラーの全画面エラーページのパスです。
const InternalServerErrorPath = "/errors/internal-server-error"

// NotFoundは共通の404ページを描画する。存在しないものへの要求への応答と、
// NotFoundPathが返すページ本体の両方を兼ねる。文言をInternalServerErrorと分けているのは、
// 読み手に求める次の行動が違うため (時間を置いて同じ操作をやり直すのではなく、一覧に戻って
// 実際の状態を見る)。
func NotFound(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, NotFoundPath)
	render(w, r, http.StatusNotFound, "error_not_found_title", "error_not_found_message")
}

// Forbiddenは共通の403ページを描画する。要求は理解されたうえで拒否されたことを述べる
// (404では表せない)。リソースの存在自体を隠す意味があるのは存在が秘密である場合だけで、
// ここが応じる画面は閲覧者から見えるリンクから辿り着くものであるため。拒否した要求への応答と、
// ForbiddenPathが返すページ本体の両方を兼ねる。
func Forbidden(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, ForbiddenPath)
	render(w, r, http.StatusForbidden, "error_forbidden_title", "error_forbidden_message")
}

// InvalidCSRFTokenは、フォーム送信のCSRF検証に失敗したときに出す403ページを描画する。
// 拒否した送信への応答本文と、InvalidCSRFTokenPathが返すページ本体の両方を兼ねる。
// 認可の失敗と同じ403を使うのは、起きたことがどちらも「要求を理解したうえでサーバーが
// 応じない」で同じであるため。422はバリデーションエラーでフォームを再描画するときに
// ハンドラーが既に使っている状態のため、同じステータスが2種類の応答の形を持つことになる。
// 文言をForbiddenと分けているのは、読み手に求める次の行動が違うため (権限の獲得ではなく、
// ページを開き直しての再送信)。
func InvalidCSRFToken(w http.ResponseWriter, r *http.Request) {
	redirectHTMXToPage(w, r, InvalidCSRFTokenPath)
	render(w, r, http.StatusForbidden, "error_invalid_csrf_token_title", "error_invalid_csrf_token_message")
}

// redirectHTMXToPageは、HTMXリクエストにpagePathの全画面エラーページへの遷移を指示する。
// htmxは204と304以外のレスポンスをスワップし、hx-targetを指定していないhx-deleteの
// スワップ先はリクエスト元自身になるため、指示しなければ完全な文書が押したボタンの中へ挿入される。
// ステータスと本文には手を加えないため、通常のフォーム送信が受け取るものは変わらず、区別できて
// はならない応答どうしが区別できるようになることもない。
//
// pagePath自体を配信する場合だけは遷移先が無いため、そのページを要求したリクエストに同じページ
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

// BadGatewayは共通の502ページを描画する。リクエストが依存する上流に到達できなかった
// 場合に使う。再試行について述べる内容は500ページと同じ。読み手にはどちらも打つ手が無く、
// 違うのはアプリケーションのどちら側が失敗したかだけであるため。
//
// ここで唯一専用のページを持たないのは、本関数が応じるのがRailsの配信するパスに対する
// リバースプロキシの失敗だけであるのに対し、文書をリクエスト元の要素へスワップしてしまう
// HTMXリクエストはGoの配信するページから発行されるものだけであるため。
func BadGateway(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusBadGateway, "error_bad_gateway_title", "error_bad_gateway_message")
}

// InternalServerErrorは内部エラーを公開せず、共通の500ページを描画する。失敗した要求
// への応答と、InternalServerErrorPathが返すページ本体の両方を兼ねる。
//
// 専用ルートで配信する場合もページは500を名乗るため、描画に成功したルートがサーバーの失敗を
// 報告することになる。一方でここを200にすると、本文が読み手に失敗を伝えているのに要求は
// 成功したと述べることになる。ステータスは検索エンジンがこのページを索引から外す根拠であり、
// クライアントが操作の成否を読み取る先でもある。本ルートに到達するのは実際に500が起きた要求
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
