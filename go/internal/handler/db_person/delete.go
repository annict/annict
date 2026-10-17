package db_person

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

// dbPersonListPathは送信がreturn_toを伴わないときに人物の削除が着地する先。Annict DB
// 管理画面の人物一覧。
const dbPersonListPath = "/db/people"

// DeleteはAnnict DB管理画面で人物を論理削除する (DELETE /db/people/:id)。人物一覧の削除
// ボタン (確認ダイアログで保護) と削除確認画面の双方から到達する。admin認可はルートの
// RequireAdmin middlewareで強制し、UseCaseでも繰り返す。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	if _, err := h.deletePersonUC.Execute(ctx, usecase.DeletePersonInput{
		User:     middleware.GetUserFromContext(ctx),
		PersonID: model.PersonID(id),
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
		slog.ErrorContext(ctx, "人物の削除に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_person_deleted"))

	returnTo := returnPath(r)

	// htmxはfetchのリダイレクトを透過的に追うため、303だと一覧ページが押した
	// ボタンにスワップされ遷移しない。htmxリクエスト (人物一覧の削除ボタン) にはHX-Redirect
	// を返して一覧へフル遷移させる。上で設定したflashは遷移後のGETで表示される。
	// 非htmxクライアントには従来どおり303を返す。
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", returnTo)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

// returnPathは人物の削除のあとに読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、削除後はそこへ戻す。読み手が求めていない一覧に着地させないため。return_toは
// リクエスト全体から読むので、確認フォームのフィールドとリンクのクエリ文字列の双方を扱える。
// 値が無い場合やAnnict DBのパスでない場合は人物一覧にフォールバックする。人物一覧のボタンは
// return_toを送らないので人物一覧に着地する。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbPersonListPath)
}
