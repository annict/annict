package db_work

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Delete soft-deletes a work in the Annict DB admin UI (DELETE /db/works/:id). Like the
// Rails Db::WorksController#destroy it is guarded by a confirmation alert only (no
// confirmation screen); admin authorization is enforced by the RequireAdmin middleware on
// the route.
//
// [Ja] Annict DB 管理画面で作品をソフトデリートする (DELETE /db/works/:id)。Rails の
// Db::WorksController#destroy と同じく確認アラートのみ (確認画面は挟まない)。admin 認可は
// ルートの RequireAdmin middleware で強制する。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := h.deleteWorkUC.Execute(ctx, usecase.DeleteWorkInput{WorkID: model.WorkID(id)}); err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "作品の削除に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_deleted"))
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
		w.Header().Set("HX-Redirect", "/db/works")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/db/works", http.StatusSeeOther)
}
