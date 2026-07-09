package db_work_archive

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Delete re-publishes (un-archives) a work in the Annict DB admin UI
// (DELETE /db/works/:id/archive).
//
// [Ja] Annict DB 管理画面で作品を再公開 (アーカイブ解除) にする (DELETE /db/works/:id/archive)。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := h.unarchiveWorkUC.Execute(ctx, usecase.UnarchiveWorkInput{WorkID: model.WorkID(id)}); err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "作品の再公開に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_published"))
	// htmx follows fetch redirects transparently, so a 303 would swap the list page into the
	// clicked button instead of navigating. For htmx requests (the work list's publish button)
	// return HX-Redirect so htmx does a full navigation to the list; the flash set above is
	// shown on the followed GET. Non-htmx clients keep the plain 303 redirect.
	//
	// [Ja] htmx は fetch のリダイレクトを透過的に追うため、303 だと一覧ページが押した
	// ボタンにスワップされ遷移しない。htmx リクエスト (作品一覧の公開ボタン) には HX-Redirect
	// を返して一覧へフル遷移させる。上で設定した flash は遷移後の GET で表示される。
	// 非 htmx クライアントには従来どおり 303 を返す。
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/db/works")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/db/works", http.StatusSeeOther)
}
