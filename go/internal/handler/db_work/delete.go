package db_work

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/usecase"
)

// dbWorkListPath is where a work write lands when the submit names no return_to: the work list
// of the Annict DB admin UI.
//
// [Ja] dbWorkListPath は送信が return_to を伴わないときに作品の書き込みが着地する先。Annict DB
// 管理画面の作品一覧。
const dbWorkListPath = "/db/works"

// Delete soft-deletes a work in the Annict DB admin UI (DELETE /db/works/:id). It is reached
// both from the work list's delete button (guarded by a confirmation alert) and from the
// delete-confirmation screen; admin authorization is enforced by the RequireAdmin middleware on
// the route and repeated by the usecase.
//
// [Ja] Annict DB 管理画面で作品をソフトデリートする (DELETE /db/works/:id)。作品一覧の削除
// ボタン (確認アラートで保護) と削除確認画面の双方から到達する。admin 認可はルートの
// RequireAdmin middleware で強制し、UseCase でも繰り返す。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	if _, err := h.deleteWorkUC.Execute(ctx, usecase.DeleteWorkInput{
		User:   middleware.GetUserFromContext(ctx),
		WorkID: model.WorkID(id),
	}); err != nil {
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
				return
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
				return
			}
		}
		slog.ErrorContext(ctx, "作品の削除に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_deleted"))

	returnTo := returnPath(r)

	// htmx follows fetch redirects transparently, so a 303 would swap the list page into the
	// clicked button instead of navigating. For htmx requests (the work list's delete button)
	// return HX-Redirect so htmx does a full navigation to the list; the flash set above is
	// shown on the followed GET. Non-htmx clients keep the plain 303 redirect.
	//
	// [Ja] htmx は fetch のリダイレクトを透過的に追うため、303 だと一覧ページが押した
	// ボタンにスワップされ遷移しない。htmx リクエスト (作品一覧の削除ボタン) には HX-Redirect
	// を返して一覧へフル遷移させる。上で設定した flash は遷移後の GET で表示される。
	// 非 htmx クライアントには従来どおり 303 を返す。
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", returnTo)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

// returnPath is the listing a work write returns the reader to. A confirmation screen carries
// the listing the reader came from, so the write returns them to it instead of a list they did
// not ask for. return_to is read from the request as a whole, which covers both the
// confirmation form field and a link query string, and an absent or non-Annict-DB value falls
// back to the work list. The work list's own buttons submit no return_to and keep landing there.
//
// [Ja] returnPath は作品の書き込みのあとに読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、書き込み後はそこへ戻す。読み手が求めていない一覧に着地させないため。return_to は
// リクエスト全体から読むので、確認フォームのフィールドとリンクのクエリ文字列の双方を扱える。
// 値が無い場合や Annict DB のパスでない場合は作品一覧にフォールバックする。作品一覧のボタンは
// return_to を送らないので従来どおりそこに着地する。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbWorkListPath)
}
