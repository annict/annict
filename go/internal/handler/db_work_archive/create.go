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

// Create archives (unpublishes) a work in the Annict DB admin UI
// (POST /db/works/:id/archive).
//
// [Ja] Annict DB 管理画面で作品を非公開 (アーカイブ) にする (POST /db/works/:id/archive)。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := h.archiveWorkUC.Execute(ctx, usecase.ArchiveWorkInput{WorkID: model.WorkID(id)}); err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "作品の非公開に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_archived"))
	http.Redirect(w, r, "/db/works", http.StatusSeeOther)
}
