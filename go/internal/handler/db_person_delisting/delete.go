package db_person_delisting

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_people"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// DeleteはAnnict DB管理画面で人物の掲載停止を解除する (DELETE /db/people/:id/delisting)。
// 掲載停止解除の確認画面のフォームから到達する。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	if _, err := h.undelistPersonUC.Execute(ctx, usecase.UndelistPersonInput{
		User:      middleware.GetUserFromContext(ctx),
		PersonID:  model.PersonID(id),
		UpdatedAt: r.FormValue("updated_at"),
	}); err != nil {
		if model.AsValidationError(err) != nil {
			h.renderUndelisting(w, r, model.PersonID(id), http.StatusUnprocessableEntity)
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
				h.renderUndelisting(w, r, model.PersonID(id), http.StatusConflict)
				return
			}
		}
		slog.ErrorContext(ctx, "人物の掲載停止の解除に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_person_undelisted"))
	http.Redirect(w, r, returnPath(r), http.StatusSeeOther)
}

// renderUndelistingは掲載停止解除確認画面を、却下した送信のステータスで再描画する。解除の確認後に
// 人物が変わった場合は、現在の人物と対応記録を示す。
func (h *Handler) renderUndelisting(w http.ResponseWriter, r *http.Request, id model.PersonID, status int) {
	ctx := r.Context()
	output, err := h.getDBPersonUndelistingNewUC.Execute(ctx, usecase.GetDBPersonUndelistingNewInput{
		User:     middleware.GetUserFromContext(ctx),
		PersonID: id,
	})
	if err != nil {
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
		slog.ErrorContext(ctx, "人物の掲載停止解除確認画面の再取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, fmt.Sprintf("/db/people/%d/undelisting/new", int64(id)))
	meta.SetDBTitle(ctx, "db_people_undelisting_new_document_title", map[string]any{"Name": output.Person.Name})
	component := layouts.Db(meta, h.cfg.GetAssetVersion(), db_people.UndelistingNew(db_people.UndelistingNewPageData{
		CSRFToken:      middleware.GetCSRFToken(r, h.sessionManager),
		Person:         viewmodel.NewDBPersonDelistingTarget(output.Person),
		ReturnTo:       returnPath(r),
		Conflict:       status == http.StatusConflict,
		InvalidVersion: status == http.StatusUnprocessableEntity,
	}))
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "DB人物掲載停止解除確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}
