package db_person

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// CreateはAnnict DB管理画面の人物の一括登録の送信 (POST /db/people) を処理する。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows := r.FormValue("rows")

	if _, err := h.createPeopleUC.Execute(ctx, usecase.CreatePeopleInput{
		User: middleware.GetUserFromContext(ctx),
		Rows: rows,
	}); err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNew(w, r, newFormState{
				Status:     http.StatusUnprocessableEntity,
				FormErrors: ve,
				Rows:       rows,
			})
			return
		}
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
			default:
				slog.ErrorContext(ctx, ae.LogString())
				httperror.InternalServerError(w, r)
			}
			return
		}
		slog.ErrorContext(ctx, "人物の一括登録に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// 送信が成功したら人物一覧に着地する。RailsのDb::PeopleController#create
	// (db_person_list_path) と同じ遷移で、作成した人物はIDの降順の先頭に並ぶため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_people_created"))
	http.Redirect(w, r, indexCanonicalPath(1), http.StatusSeeOther)
}
