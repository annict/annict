package db_episode

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Create processes a bulk-create submit for a work's episodes in the Annict DB admin UI
// (POST /db/works/:work_id/episodes).
//
// [Ja] Annict DB 管理画面の、ある作品のエピソード一括作成の送信 (POST
// /db/works/:work_id/episodes) を処理する。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	workID, ok := parseWorkIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	rows := r.FormValue("rows")

	if _, err := h.createEpisodesUC.Execute(ctx, usecase.CreateEpisodesInput{
		WorkID: workID,
		User:   middleware.GetUserFromContext(ctx),
		Rows:   rows,
	}); err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNew(w, r, workID, newFormState{
				Status:     http.StatusUnprocessableEntity,
				FormErrors: ve,
				Rows:       rows,
			})
			return
		}
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
			default:
				slog.ErrorContext(ctx, ae.LogString())
				httperror.InternalServerError(w, r)
			}
			return
		}
		slog.ErrorContext(ctx, "エピソードの一括作成に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// A successful submit lands on the work's episode list, matching the Rails create action
	// (db_episode_list_path): the created rows are what the editor checks next.
	//
	// [Ja] 送信が成功したらその作品のエピソード一覧に着地する。Rails の create アクション
	// (db_episode_list_path) と同じ遷移で、編集者が次に確認するのは作成された行であるため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episodes_created"))
	http.Redirect(w, r, indexPath(workID, 1), http.StatusSeeOther)
}
