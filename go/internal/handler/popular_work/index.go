package popular_work

import (
	"log/slog"
	"net/http"

	authMiddleware "github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/templates/layouts"
	"github.com/annict/annict/internal/templates/pages/works"
	"github.com/annict/annict/internal/viewmodel"
)

// Index GET /works/popular - 人気作品一覧を表示
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. RepositoryからModelを取得（Domain/Infrastructure層）
	modelWorks, err := h.workRepo.GetPopularWorksWithDetails(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "人気作品の取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. ModelをViewModelに変換（Presentation層内の変換）
	viewWorks := viewmodel.NewWorksFromModelDetails(modelWorks, h.imageHelper)

	// コンテキストからユーザー情報を取得
	user := authMiddleware.GetUserFromContext(ctx)

	// フラッシュメッセージを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)

	// ページメタ情報を準備
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "popular_anime") // "人気アニメ | Annict" / "Popular Anime | Annict"

	// 3. テンプレートにViewModelを渡す（templ使用）
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Default(
		ctx,
		meta,
		user,
		flash,
		h.cfg.GetAssetVersion(),
		works.Popular(ctx, viewWorks),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
