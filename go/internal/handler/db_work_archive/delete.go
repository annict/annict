package db_work_archive

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DeleteはAnnict DB管理画面で作品を再公開 (アーカイブ解除) にする (DELETE /db/works/:id/archive)。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	if _, err := h.unarchiveWorkUC.Execute(ctx, usecase.UnarchiveWorkInput{
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
		slog.ErrorContext(ctx, "作品の再公開に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_published"))

	returnTo := returnPath(r)

	// htmxはfetchのリダイレクトを透過的に追うため、303だと一覧ページが押した
	// ボタンにスワップされ遷移しない。htmxリクエスト (作品一覧の公開ボタン) にはHX-Redirect
	// を返して一覧へフル遷移させる。上で設定したflashは遷移後のGETで表示される。
	// 非htmxクライアントには従来どおり303を返す。
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", returnTo)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}
