package db_work

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/viewmodel"
)

// dbWorksNewPathは作品新規作成フォームの代表GETパス。Newはこのパスでページを配信し、
// Createは同じページをPOST /db/worksから再描画するため、双方ともリクエストパスではなく
// ここからcanonical URLを取る。
const dbWorksNewPath = "/db/works/new"

// NewはAnnict DB管理画面の作品新規作成フォームページ (GET /db/works/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	optionsResult, err := h.getDBWorkFormOptionsUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats)

	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbWorksNewPath)
	meta.SetDBTitle(ctx, "db_works_new_title")

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.New(db_works.NewPageData{
			CSRFToken:   csrfToken,
			FormOptions: formOptions,
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
		slog.ErrorContext(ctx, "DB作品新規作成フォームのレスポンスの書き込みに失敗", "error", err)
	}
}
