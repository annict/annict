package db_work_revision

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_work_revisions"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageは作品の履歴の1ページに並べる版の数。
const perPage int32 = 30

// IndexはAnnict DB管理画面の作品の履歴 (GET /db/works/:id/revisions) を描画する。
// ページ送りで薄いページが増えるため、検索エンジンにはインデックスさせない。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}
	page := parsePage(r)

	output, err := h.getDBWorkRevisionsUC.Execute(ctx, usecase.GetDBWorkRevisionsInput{
		WorkID:  model.WorkID(id),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DB作品の履歴の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	basePath := fmt.Sprintf("/db/works/%d/revisions", id)
	pagination := viewmodel.NewPagination(int(page), int(output.TotalCount), int(perPage), basePath)
	canonicalPath := basePath
	if page > 1 {
		canonicalPath = pagination.PageURL(int(page))
	}

	work := viewmodel.NewDBWorkRevisionWork(output.Work)
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, canonicalPath)
	if work.Title == "" {
		meta.SetDBTitle(ctx, "db_work_revisions_index_title")
	} else {
		meta.SetDBTitle(ctx, "db_work_revisions_index_document_title", map[string]any{"WorkTitle": work.Title})
	}
	meta.NoIndex = true

	viewer := middleware.GetUserFromContext(ctx)
	revisions := viewmodel.NewDBRevisionsFromAnimeRevisions(ctx, output.Revisions, viewmodel.DBRevisionLocation(viewer), h.imageHelper, viewmodel.NewDBRevisionIDNames(output.NumberFormats, nil))

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_work_revisions.Index(db_work_revisions.IndexPageData{
			Work:        work,
			Revisions:   revisions,
			Pagination:  pagination,
			IsCommitter: middleware.IsCommitter(viewer),
		}),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "DB作品の履歴のレスポンスの書き込みに失敗", "error", err)
	}
}

// parsePageはクエリパラメータpageを正のページ番号として読み、欠落・無効な値のときは1を返す。
func parsePage(r *http.Request) int32 {
	v, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil || v < 1 {
		return 1
	}
	return int32(v)
}
