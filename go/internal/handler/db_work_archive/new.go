package db_work_archive

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

// New renders the archive-confirmation page in the Annict DB admin UI
// (GET /db/works/:id/archive/new).
//
// [Ja] Annict DB 管理画面の非公開確認ページ (GET /db/works/:id/archive/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	output, err := h.getDBWorkArchiveNewUC.Execute(ctx, usecase.GetDBWorkArchiveNewInput{WorkID: model.WorkID(id)})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "非公開確認画面の取得に失敗", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetDBTitle(ctx, "db_works_archive_new_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.ArchiveNew(db_works.ArchiveNewPageData{
			CSRFToken: csrfToken,
			WorkID:    viewmodel.WorkID(output.Work.ID),
			Title:     output.Work.Title,
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
