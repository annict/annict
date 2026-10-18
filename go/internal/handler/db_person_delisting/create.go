package db_person_delisting

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

// CreateはAnnict DB管理画面で人物の掲載を停止する (POST /db/people/:id/delisting)。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	delistMessage := r.FormValue("delist_message")
	if _, err := h.delistPersonUC.Execute(ctx, usecase.DelistPersonInput{
		User:          middleware.GetUserFromContext(ctx),
		PersonID:      model.PersonID(id),
		UpdatedAt:     r.FormValue("updated_at"),
		DelistMessage: delistMessage,
	}); err != nil {
		if model.AsValidationError(err) != nil {
			h.renderNew(w, r, delistingFormState{Status: http.StatusUnprocessableEntity, DelistMessage: &delistMessage})
			return
		}
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
				return
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
				return
			case model.AppErrCodeConflict:
				h.renderNew(w, r, delistingFormState{Status: http.StatusConflict, DelistMessage: &delistMessage})
				return
			}
		}
		slog.ErrorContext(ctx, "人物の掲載停止に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_person_delisted"))
	http.Redirect(w, r, returnPath(r), http.StatusSeeOther)
}
