package db_person_archive

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

// CreateはAnnict DB管理画面で人物をアーカイブする (POST /db/people/:id/archive)。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	if _, err := h.archivePersonUC.Execute(ctx, usecase.ArchivePersonInput{
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
		slog.ErrorContext(ctx, "人物のアーカイブに失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_person_archived"))

	http.Redirect(w, r, returnPath(r), http.StatusSeeOther)
}
