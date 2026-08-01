package db_episode

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_episodes"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPage matches the Rails Db::EpisodesController#index page size (.per(100)).
//
// [Ja] perPage は Rails の Db::EpisodesController#index のページ件数 (.per(100)) に合わせる。
const perPage int32 = 100

// Index renders a work's episode list page in the Annict DB admin UI.
// (GET /db/works/:work_id/episodes).
//
// [Ja] Annict DB 管理画面の、ある作品のエピソード一覧ページ
// (GET /db/works/:work_id/episodes) を描画する。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	workID, err := strconv.ParseInt(chi.URLParam(r, "work_id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	page := parsePageParam(r)

	output, err := h.getDBEpisodesUC.Execute(ctx, usecase.GetDBEpisodesInput{
		WorkID:  model.WorkID(workID),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DBエピソード一覧の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// A page number past the end is sent to the last page that holds rows. Rendering it as
	// asked would leave the reader on an empty list whose title and og:url name a page the
	// work does not have, and with no pagination to return from: the list template drops the
	// pagination along with the table when there is nothing to show. The redirect also keeps
	// the page number reaching the title, the og:url and the pagination within range, so the
	// three always name the same page.
	//
	// [Ja] 末尾を超えるページ番号は、行を持つ最後のページへ送る。要求どおり描画すると、
	// タイトルと og:url がその作品には無いページ番号を名乗る空の一覧に着地し、しかも戻る
	// 導線が残らない (一覧テンプレートは表示するものが無いとき、テーブルと一緒に
	// ページネーションも落とすため)。このリダイレクトにより、タイトル・og:url・
	// ページネーションへ渡るページ番号が常に範囲内に収まり、三者が同じページを名乗る。
	if lastPage := lastPageNumber(output.TotalCount, perPage); int64(page) > lastPage {
		http.Redirect(w, r, indexPath(output.Work.ID, lastPage), http.StatusFound)
		return
	}

	// The pagination links and the canonical URL describe the same list, so both come from
	// the same path builder: the pagination base path is this view without a page number,
	// and Pagination.PageURL adds the page back per link.
	//
	// [Ja] ページネーションのリンクと canonical URL は同じ一覧を指すため、どちらも同じパス
	// 生成を通す。ページネーションの起点はページ番号を外したこの表示で、ページ番号は
	// Pagination.PageURL がリンクごとに付け直す。
	pagination := viewmodel.NewPagination(int(page), int(output.TotalCount), int(perPage), indexPath(output.Work.ID, 1))

	workName := viewmodel.DBEpisodeListWorkName(output.Work.Title)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, indexPath(output.Work.ID, int64(page)))
	setIndexTitle(ctx, &meta, workName, page)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.Index(db_episodes.IndexPageData{
			WorkID:     viewmodel.WorkID(output.Work.ID),
			WorkName:   workName,
			NoEpisodes: output.Work.NoEpisodes,
			Episodes:   viewmodel.NewDBEpisodeListItems(output.Episodes),
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
		slog.ErrorContext(ctx, "DBエピソード一覧レスポンスの書き込みに失敗", "error", err)
	}
}

// setIndexTitle gives meta a document title that identifies the work and, after the first
// page, the page number. A work with no name to show falls back to the generic episode-list
// title, matching the page heading.
//
// [Ja] setIndexTitle は meta に、作品を識別でき、2 ページ目以降ではページ番号も含む文書
// タイトルを設定する。表示できる名前が無い作品では、ページ見出しと同じく汎用のエピソード
// 一覧タイトルへフォールバックする。
func setIndexTitle(ctx context.Context, meta *viewmodel.PageMeta, workName string, page int32) {
	if workName == "" {
		if page > 1 {
			meta.SetDBTitle(ctx, "db_episodes_index_title_paginated", map[string]any{"Page": page})
			return
		}
		meta.SetDBTitle(ctx, "db_episodes_index_title")
		return
	}

	templateData := map[string]any{"WorkTitle": workName, "Page": page}
	if page > 1 {
		meta.SetDBTitle(ctx, "db_episodes_index_document_title_paginated", templateData)
		return
	}
	meta.SetDBTitle(ctx, "db_episodes_index_document_title", templateData)
}

// lastPageNumber returns the highest 1-based page number that still holds rows. A work with
// no episodes yet has one page all the same, so its empty list stays reachable at the path
// without a page number instead of redirecting to a page zero.
//
// The result is an int64 because the count it divides is one. Narrowing it here would trade a
// correct number for a conversion whose safety has to be argued, and both consumers (the
// comparison against the requested page and indexPath) take the wider type anyway.
//
// [Ja] lastPageNumber は行を持つ 1 始まりの最大ページ番号を返す。エピソードがまだ無い作品も
// 1 ページは持つ扱いとし、空の一覧が 0 ページ目へリダイレクトされずページ番号なしのパスで
// 見られるようにする。
//
// 割る対象の件数が int64 のため、戻り値も int64 とする。ここで幅を狭めると、正しい値が
// 「安全性を論じる必要のある変換」に変わるだけで、利用側 (要求ページとの比較と indexPath) は
// どちらも広いほうの型を受け取れる。
func lastPageNumber(totalCount int64, perPage int32) int64 {
	if totalCount <= 0 || perPage <= 0 {
		return 1
	}

	return (totalCount + int64(perPage) - 1) / int64(perPage)
}

// parsePageParam reads the 1-based page number from the query string, falling back to the
// first page when it is missing or not a positive number.
//
// [Ja] parsePageParam はクエリ文字列から 1 始まりのページ番号を読み取る。欠落や正の数でない
// 値のときは 1 ページ目にフォールバックする。
func parsePageParam(r *http.Request) int32 {
	s := r.URL.Query().Get("page")
	if s == "" {
		return 1
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil || v < 1 {
		return 1
	}
	return int32(v)
}

// indexPath builds the representative path of a work's episode list page, used as the og:url
// of the current view and, with page 1, as the base path the pagination links extend. The
// first page is written without the parameter so that the list and its ?page=1 form share one
// representative path. Only the page number goes in, so tracking parameters a shared link
// happens to carry stay out of og:url and out of the pagination links.
//
// [Ja] indexPath はある作品のエピソード一覧ページの代表パスを生成する。現在の表示の og:url
// と、page に 1 を渡してページネーションのリンクが伸ばす起点の双方で使う。1 ページ目は
// パラメータなしで書き、一覧と ?page=1 の形が 1 つの代表パスを共有するようにする。含めるのは
// ページ番号だけとし、共有されたリンクがたまたま持つトラッキングパラメータが og:url にも
// ページネーションのリンクにも載らないようにする。
func indexPath(workID model.WorkID, page int64) string {
	path := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if page > 1 {
		return fmt.Sprintf("%s?page=%d", path, page)
	}
	return path
}
