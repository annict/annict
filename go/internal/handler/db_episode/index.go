package db_episode

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_episodes"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageはRailsのDb::EpisodesController#indexのページ件数 (.per(100)) に合わせる。
const perPage int32 = 100

// IndexはAnnict DB管理画面の、ある作品のエピソード一覧ページ
// (GET /db/works/:work_id/episodes) を描画する。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	workID, ok := parseWorkIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	page := parsePageParam(r)

	output, err := h.getDBEpisodesUC.Execute(ctx, usecase.GetDBEpisodesInput{
		WorkID:  workID,
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

	// 末尾を超えるページ番号は、行を持つ最後のページへ送る。要求どおり描画すると、
	// タイトルとog:urlがその作品には無いページ番号を名乗る空の一覧に着地し、しかも戻る
	// 導線が残らない (一覧テンプレートは表示するものが無いとき、テーブルと一緒に
	// ページネーションも落とすため)。このリダイレクトにより、タイトル・og:url・
	// ページネーションへ渡るページ番号が常に範囲内に収まり、三者が同じページを名乗る。
	if lastPage := lastPageNumber(output.TotalCount, perPage); int64(page) > lastPage {
		http.Redirect(w, r, indexPath(output.Work.ID, lastPage), http.StatusFound)
		return
	}

	// ページネーションのリンクとcanonical URLは同じ一覧を指すため、どちらも同じパス
	// 生成を通す。ページネーションの起点はページ番号を外したこの表示で、ページ番号は
	// Pagination.PageURLがリンクごとに付け直す。
	pagination := viewmodel.NewPagination(int(page), int(output.TotalCount), int(perPage), indexPath(output.Work.ID, 1))

	workName := viewmodel.DBEpisodeListWorkName(output.Work.Title)

	// 一覧は公開のため、閲覧者のロールがページの出すリンク・操作を決める。見出しの作成
	// リンクと行ごとの操作列が同じ答えを読むため、ここで1度だけ解決する。
	user := middleware.GetUserFromContext(ctx)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, indexPath(output.Work.ID, int64(page)))
	setIndexTitle(ctx, &meta, workName, page)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.Index(db_episodes.IndexPageData{
			WorkID:     viewmodel.WorkID(output.Work.ID),
			WorkName:   workName,
			NoEpisodes: output.Work.NoEpisodes,
			Generation: viewmodel.NewDBEpisodeGenerationSummary(
				output.Work.ManualEpisodesCount,
				output.PublishedEpisodeCount,
				output.MaxGeneratableEpisodeNumber,
			),
			Episodes:    viewmodel.NewDBEpisodeListItems(output.Episodes),
			Pagination:  pagination,
			IsCommitter: middleware.IsCommitter(user),
			IsAdmin:     middleware.IsAdmin(user),
			CSRFToken:   middleware.GetCSRFToken(r, h.sessionManager),
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

// setIndexTitleはmetaに、作品を識別でき、2ページ目以降ではページ番号も含む文書
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

// lastPageNumberは行を持つ1始まりの最大ページ番号を返す。エピソードがまだ無い作品も
// 1ページは持つ扱いとし、空の一覧が0ページ目へリダイレクトされずページ番号なしのパスで
// 見られるようにする。
//
// 割る対象の件数がint64のため、戻り値もint64とする。ここで幅を狭めると、正しい値が
// 「安全性を論じる必要のある変換」に変わるだけで、利用側 (要求ページとの比較とindexPath) は
// どちらも広いほうの型を受け取れる。
func lastPageNumber(totalCount int64, perPage int32) int64 {
	if totalCount <= 0 || perPage <= 0 {
		return 1
	}

	return (totalCount + int64(perPage) - 1) / int64(perPage)
}

// parsePageParamはクエリ文字列から1始まりのページ番号を読み取る。欠落や正の数でない
// 値のときは1ページ目にフォールバックする。
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

// indexPathはある作品のエピソード一覧ページの代表パスを生成する。現在の表示のog:url
// と、pageに1を渡してページネーションのリンクが伸ばす起点の双方で使う。1ページ目は
// パラメータなしで書き、一覧と ?page=1の形が1つの代表パスを共有するようにする。含めるのは
// ページ番号だけとし、共有されたリンクがたまたま持つトラッキングパラメータがog:urlにも
// ページネーションのリンクにも載らないようにする。
func indexPath(workID model.WorkID, page int64) string {
	path := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if page > 1 {
		return fmt.Sprintf("%s?page=%d", path, page)
	}
	return path
}
