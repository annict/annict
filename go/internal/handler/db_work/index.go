package db_work

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

const perPage int32 = 30

// Index renders the work list page in the Annict DB admin UI (GET /db/works).
// [Ja] Annict DB 管理画面の作品一覧ページ (GET /db/works) を描画する。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parseIntParam(r, "page", 1)
	filterNoEpisodes := r.URL.Query().Get("filter_no_episodes") == "1"
	filterNoImage := r.URL.Query().Get("filter_no_image") == "1"
	filterNoSeason := r.URL.Query().Get("filter_no_season") == "1"

	result, err := h.listDbWorksUC.Execute(ctx, usecase.ListDbWorksInput{
		FilterNoEpisodes: filterNoEpisodes,
		FilterNoImage:    filterNoImage,
		FilterNoSeason:   filterNoSeason,
		Page:             page,
		PerPage:          perPage,
	})
	if err != nil {
		slog.ErrorContext(ctx, "DB作品一覧の取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	basePath := buildBasePath(r.URL)
	pagination := viewmodel.NewPagination(int(page), int(result.TotalCount), int(perPage), basePath)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_index_title")

	worksVM := viewmodel.NewDBWorkListItems(ctx, result.Works)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.Index(db_works.IndexPageData{
			Works:            worksVM,
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

// parseIntParam reads a positive int32 query parameter and falls back to defaultValue when missing or invalid.
// [Ja] 正の int32 のクエリパラメータを読み取り、欠落・無効な値のときは defaultValue を返す。
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

// buildBasePath returns the current URL with the `page` query parameter stripped, suitable as the base path for pagination links.
// [Ja] ページネーションリンクの起点として使えるよう、現在の URL から `page` クエリパラメータだけを除いたパスを返す。
func buildBasePath(u *url.URL) string {
	q := u.Query()
	q.Del("page")
	result := u.Path
	if encoded := q.Encode(); encoded != "" {
		result += "?" + encoded
	}
	return result
}
