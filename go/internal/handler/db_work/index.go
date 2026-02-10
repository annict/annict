package db_work

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/viewmodel"
)

const perPage int32 = 30

// Index GET /db/works - DB管理画面の作品一覧を表示
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// クエリパラメータを取得
	page := parseIntParam(r, "page", 1)
	filterNoEpisodes := r.URL.Query().Get("filter_no_episodes") == "1"
	filterNoImage := r.URL.Query().Get("filter_no_image") == "1"
	filterNoSeason := r.URL.Query().Get("filter_no_season") == "1"

	params := repository.DBWorkListParams{
		FilterNoEpisodes: filterNoEpisodes,
		FilterNoImage:    filterNoImage,
		FilterNoSeason:   filterNoSeason,
		Page:             page,
		PerPage:          perPage,
	}

	// 作品一覧と総数を取得
	works, err := h.workRepo.ListForDB(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "DB作品一覧の取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	totalCount, err := h.workRepo.CountForDB(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "DB作品総数の取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// ページネーション用のBasePathを構築（フィルタパラメータを維持）
	basePath := buildBasePath(r.URL)

	// ページネーション情報を作成
	pagination := viewmodel.NewPagination(int(page), int(totalCount), int(perPage), basePath)

	// フラッシュメッセージを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)

	// ページメタ情報を準備
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_title")

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		ctx,
		meta,
		flash,
		h.cfg.GetAssetVersion(),
		db_works.Index(db_works.IndexPageData{
			Works:            works,
			Pagination:       pagination,
			FilterNoEpisodes: filterNoEpisodes,
			FilterNoImage:    filterNoImage,
			FilterNoSeason:   filterNoSeason,
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// parseIntParam はクエリパラメータから整数値を取得します
func parseIntParam(r *http.Request, name string, defaultValue int32) int32 {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultValue
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil || v < 1 {
		return defaultValue
	}
	return int32(v)
}

// buildBasePath はページネーション用のBasePathを構築します
// ページパラメータを除いた現在のURLを返します
func buildBasePath(u *url.URL) string {
	q := u.Query()
	q.Del("page")
	result := u.Path
	if encoded := q.Encode(); encoded != "" {
		result += "?" + encoded
	}
	return result
}
