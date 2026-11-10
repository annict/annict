package db_person_revision

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
	"github.com/annict/annict/go/internal/templates/pages/db_person_revisions"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageは人物の履歴の1ページに並べる版の数。作品の履歴とそろえる。
const perPage int32 = 30

// IndexはAnnict DB管理画面の人物の履歴 (GET /db/people/:id/revisions) を描画する。
// ページ送りで薄いページが増えるため、検索エンジンにはインデックスさせない。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}
	page := parsePage(r)
	viewer := middleware.GetUserFromContext(ctx)

	output, err := h.getDBPersonRevisionsUC.Execute(ctx, usecase.GetDBPersonRevisionsInput{
		User:     viewer,
		PersonID: model.PersonID(id),
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DB人物の履歴の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	basePath := fmt.Sprintf("/db/people/%d/revisions", id)
	pagination := viewmodel.NewPagination(int(page), int(output.TotalCount), int(perPage), basePath)
	canonicalPath := basePath
	if page > 1 {
		canonicalPath = pagination.PageURL(int(page))
	}

	person := viewmodel.NewDBPersonRevisionPerson(output.Person)
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, canonicalPath)
	if person.Name == "" {
		meta.SetDBTitle(ctx, "db_person_revisions_index_title")
	} else {
		meta.SetDBTitle(ctx, "db_person_revisions_index_document_title", map[string]any{"PersonName": person.Name})
	}
	meta.NoIndex = true

	revisions := viewmodel.NewDBRevisionsFromPersonRevisions(ctx, output.Revisions, viewmodel.DBRevisionLocation(viewer), h.imageHelper, viewmodel.NewDBRevisionIDNames(nil, output.Prefectures))

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_person_revisions.Index(db_person_revisions.IndexPageData{
			Person:      person,
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
		slog.ErrorContext(ctx, "DB人物の履歴のレスポンスの書き込みに失敗", "error", err)
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
