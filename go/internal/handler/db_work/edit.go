package db_work

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Edit renders the work edit form page in the Annict DB admin UI (GET /db/works/:id/edit).
//
// [Ja] Annict DB 管理画面の作品編集フォームページ (GET /db/works/:id/edit) を描画する。
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	output, err := h.getDbWorkEditUC.Execute(ctx, usecase.GetDbWorkEditInput{WorkID: model.WorkID(id)})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "DB作品編集フォームの取得に失敗", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, output.NumberFormats)
	formInput := viewmodel.NewDBWorkFormInputFromWork(output.Work)
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_edit_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.Edit(db_works.EditPageData{
			CSRFToken:   csrfToken,
			WorkID:      viewmodel.WorkID(output.Work.ID),
			FormOptions: formOptions,
			FormInput:   formInput,
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
