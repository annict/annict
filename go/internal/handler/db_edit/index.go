package db_edit

import (
	"bytes"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_edits"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageはAnnict DB全体の編集履歴の1ページに並べる操作の数。各リソースの履歴とそろえる。
const perPage int32 = 30

// basePathはAnnict DB全体の編集履歴のパス。
const basePath = "/db/edits"

// IndexはAnnict DB全体の編集履歴 (GET /db/edits) を描画する。
// ページ送りで薄いページが増えるため、検索エンジンにはインデックスさせない。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := parsePage(r)
	viewer := middleware.GetUserFromContext(ctx)

	output, err := h.getDBEditsUC.Execute(ctx, usecase.GetDBEditsInput{
		User:    viewer,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Annict DBの編集履歴の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	pagination := viewmodel.NewPagination(int(page), int(output.TotalCount), int(perPage), basePath)
	canonicalPath := basePath
	if page > 1 {
		canonicalPath = pagination.PageURL(int(page))
	}

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, canonicalPath)
	meta.SetDBTitle(ctx, "db_edits_index_title")
	meta.NoIndex = true

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_edits.Index(db_edits.IndexPageData{
			Edits:      viewmodel.NewDBEdits(ctx, output.Edits, viewmodel.DBRevisionLocation(viewer), h.imageHelper, viewmodel.NewDBRevisionIDNames(output.NumberFormats, output.Prefectures)),
			Pagination: pagination,
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
		slog.ErrorContext(ctx, "Annict DBの編集履歴のレスポンスの書き込みに失敗", "error", err)
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
