package db_work

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPage matches the Rails Db::WorksController#index page size (.per(100)).
//
// [Ja] perPage は Rails の Db::WorksController#index のページ件数 (.per(100)) に合わせる。
const perPage int32 = 100

// Index renders the work list page in the Annict DB admin UI (GET /db/works).
//
// [Ja] Annict DB 管理画面の作品一覧ページ (GET /db/works) を描画する。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parseIntParam(r, "page", 1)
	filterNoEpisodes := r.URL.Query().Get("filter_no_episodes") == "1"
	filterNoImage := r.URL.Query().Get("filter_no_image") == "1"
	filterNoSeason := r.URL.Query().Get("filter_no_season") == "1"
	filterNoSlots := r.URL.Query().Get("filter_no_slots") == "1"
	seasonSlugs := r.URL.Query()["season_slugs"]
	seasonYears, seasonNames := viewmodel.ParseSeasonSlugs(seasonSlugs)

	result, err := h.getDBWorksUC.Execute(ctx, usecase.GetDBWorksInput{
		FilterNoEpisodes: filterNoEpisodes,
		FilterNoImage:    filterNoImage,
		FilterNoSeason:   filterNoSeason,
		FilterNoSlots:    filterNoSlots,
		SeasonYears:      seasonYears,
		SeasonNames:      seasonNames,
		Page:             page,
		PerPage:          perPage,
	})
	if err != nil {
		slog.ErrorContext(ctx, "DB作品一覧の取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	seasonFilterOptions := viewmodel.NewSeasonFilterOptions(ctx, seasonSlugs)
	canonicalSeasonSlugs := make([]string, 0, len(seasonSlugs))
	for _, option := range seasonFilterOptions {
		if option.Selected {
			canonicalSeasonSlugs = append(canonicalSeasonSlugs, option.Slug)
		}
	}

	// The pagination links and the canonical URL describe the same list, so both come from one
	// set of parsed parameters: the pagination base path is this view without a page number,
	// and Pagination.PageURL adds the page back per link.
	//
	// [Ja] ページネーションのリンクと canonical URL は同じ一覧を指すため、どちらもパース済みの
	// 同じパラメータから組み立てる。ページネーションの起点はページ番号を外したこの表示で、
	// ページ番号は Pagination.PageURL がリンクごとに付け直す。
	canonicalParams := indexCanonicalParams{
		filterNoEpisodes: filterNoEpisodes,
		filterNoImage:    filterNoImage,
		filterNoSeason:   filterNoSeason,
		filterNoSlots:    filterNoSlots,
		seasonSlugs:      canonicalSeasonSlugs,
	}
	pagination := viewmodel.NewPagination(int(page), int(result.TotalCount), int(perPage), indexCanonicalPath(canonicalParams))

	canonicalParams.page = page
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, indexCanonicalPath(canonicalParams))
	meta.SetDBTitle(ctx, "db_works_index_title")

	worksVM := viewmodel.NewDBWorkListItems(ctx, result.Works, h.imageHelper)

	// The work list is public, so resolve the viewer's role to gate the action column: the
	// edit / unpublish / publish actions require committer, delete requires admin. The CSRF
	// token rides in the X-CSRF-Token header of the htmx DELETE actions (publish / delete).
	//
	// [Ja] 作品一覧は公開のため、閲覧者のロールを解決して操作列を出し分ける。編集・非公開・
	// 公開の操作は committer、削除は admin を要する。CSRF トークンは htmx の DELETE 操作
	// (公開 / 削除) の X-CSRF-Token ヘッダーで送る。
	user := middleware.GetUserFromContext(ctx)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.Index(db_works.IndexPageData{
			Works:               worksVM,
			Pagination:          pagination,
			FilterNoEpisodes:    filterNoEpisodes,
			FilterNoImage:       filterNoImage,
			FilterNoSeason:      filterNoSeason,
			FilterNoSlots:       filterNoSlots,
			SeasonFilterOptions: seasonFilterOptions,
			IsCommitter:         middleware.IsCommitter(user),
			IsAdmin:             middleware.IsAdmin(user),
			CSRFToken:           middleware.GetCSRFToken(r, h.sessionManager),
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// parseIntParam reads a positive int32 query parameter and falls back to defaultValue when missing or invalid.
//
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

// indexCanonicalParams carries the request parameters that decide which works the list page
// shows.
//
// [Ja] indexCanonicalParams は一覧ページに並ぶ作品を決めるリクエストパラメータを運ぶ。
type indexCanonicalParams struct {
	page             int32
	filterNoEpisodes bool
	filterNoImage    bool
	filterNoSeason   bool
	filterNoSlots    bool
	seasonSlugs      []string
}

// indexCanonicalPath builds the representative path of the work list page, used both as the
// og:url of the current view and, with the page number left out, as the base path the
// pagination links extend. Every parameter here changes which works the page lists, so all of
// them belong in the representative URL: opening a link to a filtered page 3 then shows that
// same view instead of the top of the whole list. Only these known parameters go in, so that
// tracking parameters a shared link happens to carry stay out of og:url and out of the
// pagination links. The first page is written without the parameter so that /db/works and
// /db/works?page=1 share one representative path. The season slugs come from the selected
// server-defined options, which removes invalid values and duplicates and gives the selected
// set a stable order.
//
// [Ja] indexCanonicalPath は作品一覧ページの代表パスを生成する。現在の表示の og:url と、
// ページ番号を外してページネーションのリンクが伸ばす起点の双方で使う。ここで扱うパラメータは
// いずれもページに並ぶ作品を変えるため、すべて代表 URL に含める。そうすることで、絞り込んだ
// 3 ページ目へのリンクを開いた人は一覧の先頭ではなく同じ表示を見る。含めるのはこれらの既知の
// パラメータだけとし、共有されたリンクがたまたま持つトラッキングパラメータが og:url にも
// ページネーションのリンクにも載らないようにする。1 ページ目はパラメータなしで書き、
// /db/works と /db/works?page=1 が 1 つの代表パスを共有するようにする。シーズンのスラッグは
// 選択済みのサーバー定義の選択肢から取り、不正値と重複を除いて、選択集合を安定した順序に
// 揃える。
func indexCanonicalPath(params indexCanonicalParams) string {
	q := url.Values{}
	if params.filterNoEpisodes {
		q.Set("filter_no_episodes", "1")
	}
	if params.filterNoImage {
		q.Set("filter_no_image", "1")
	}
	if params.filterNoSeason {
		q.Set("filter_no_season", "1")
	}
	if params.filterNoSlots {
		q.Set("filter_no_slots", "1")
	}
	for _, slug := range params.seasonSlugs {
		q.Add("season_slugs", slug)
	}
	if params.page > 1 {
		q.Set("page", strconv.FormatInt(int64(params.page), 10))
	}

	encoded := q.Encode()
	if encoded == "" {
		return "/db/works"
	}
	return "/db/works?" + encoded
}
