package db_work

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageはRailsのDb::WorksController#indexのページ件数 (.per(100)) に合わせる。
const perPage int32 = 100

// IndexはAnnict DB管理画面の作品一覧ページ (GET /db/works) を描画する。
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
		httperror.InternalServerError(w, r)
		return
	}

	seasonFilterOptions := viewmodel.NewSeasonFilterOptions(ctx, seasonSlugs)
	canonicalSeasonSlugs := make([]string, 0, len(seasonSlugs))
	for _, option := range seasonFilterOptions {
		if option.Selected {
			canonicalSeasonSlugs = append(canonicalSeasonSlugs, option.Slug)
		}
	}

	// ページネーションのリンクとcanonical URLは同じ一覧を指すため、どちらもパース済みの
	// 同じパラメータから組み立てる。ページネーションの起点はページ番号を外したこの表示で、
	// ページ番号はPagination.PageURLがリンクごとに付け直す。
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

	// 作品一覧は公開のため、閲覧者のロールを解決して操作列を出し分ける。編集・非公開・
	// 公開の操作はcommitter、削除はadminを要する。CSRFトークンはhtmxのDELETE操作
	// (公開 / 削除) のX-CSRF-Tokenヘッダーで送る。
	user := middleware.GetUserFromContext(ctx)

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
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "DB作品一覧レスポンスの書き込みに失敗", "error", err)
	}
}

// 正のint32のクエリパラメータを読み取り、欠落・無効な値のときはdefaultValueを返す。
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

// indexCanonicalParamsは一覧ページに並ぶ作品を決めるリクエストパラメータを運ぶ。
type indexCanonicalParams struct {
	page             int32
	filterNoEpisodes bool
	filterNoImage    bool
	filterNoSeason   bool
	filterNoSlots    bool
	seasonSlugs      []string
}

// indexCanonicalPathは作品一覧ページの代表パスを生成する。現在の表示のog:urlと、
// ページ番号を外してページネーションのリンクが伸ばす起点の双方で使う。ここで扱うパラメータは
// いずれもページに並ぶ作品を変えるため、すべて代表URLに含める。そうすることで、絞り込んだ
// 3ページ目へのリンクを開いた人は一覧の先頭ではなく同じ表示を見る。含めるのはこれらの既知の
// パラメータだけとし、共有されたリンクがたまたま持つトラッキングパラメータがog:urlにも
// ページネーションのリンクにも載らないようにする。1ページ目はパラメータなしで書き、
// /db/worksと /db/works?page=1が1つの代表パスを共有するようにする。シーズンのスラッグは
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
