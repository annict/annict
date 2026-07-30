package db_work

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/viewmodel"
)

// dbWorksNewPath is the representative GET path of the new-work form. New serves the page at
// this path and Create re-renders the same page from POST /db/works, so both take their
// canonical URL from here rather than from the request path.
//
// [Ja] dbWorksNewPath は作品新規作成フォームの代表 GET パス。New はこのパスでページを配信し、
// Create は同じページを POST /db/works から再描画するため、双方ともリクエストパスではなく
// ここから canonical URL を取る。
const dbWorksNewPath = "/db/works/new"

// New renders the new-work form page in the Annict DB admin UI (GET /db/works/new).
//
// [Ja] Annict DB 管理画面の作品新規作成フォームページ (GET /db/works/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	optionsResult, err := h.getDBWorkFormOptionsUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats)

	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbWorksNewPath)
	meta.SetDBTitle(ctx, "db_works_new_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.New(db_works.NewPageData{
			CSRFToken:   csrfToken,
			FormOptions: formOptions,
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
