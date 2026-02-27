package db_work

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New GET /db/works/new - DB管理画面の作品新規作成フォームを表示
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// NumberFormatの選択肢を取得
	numberFormats, err := h.numberFormatRepo.ListAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// フォーム用の選択肢を作成
	formOptions := viewmodel.NewDBWorkFormOptions(ctx, numberFormats)

	// CSRFトークンを取得
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	// ページメタ情報を準備
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_new_title")

	// フラッシュメッセージを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		ctx,
		meta,
		flash,
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
