package db_person

import (
	"bytes"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_people"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// perPageはRailsのDb::PeopleController#indexのページ件数 (.per(100)) に合わせる。
const perPage int32 = 100

// IndexはAnnict DB管理画面の人物一覧ページ (GET /db/people) を描画する。
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := parsePageParam(r)

	// 人物一覧は公開のため、閲覧者のロールで操作列を出し分ける。
	user := middleware.GetUserFromContext(ctx)

	result, err := h.getDBPeopleUC.Execute(ctx, usecase.GetDBPeopleInput{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		slog.ErrorContext(ctx, "DB人物一覧の取得エラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	pagination := viewmodel.NewPagination(int(page), int(result.TotalCount), int(perPage), indexCanonicalPath(1))

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, indexCanonicalPath(page))
	meta.SetDBTitle(ctx, "db_people_index_title")

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_people.Index(db_people.IndexPageData{
			People:      viewmodel.NewDBPersonListItems(result.People),
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
		slog.ErrorContext(ctx, "DB人物一覧レスポンスの書き込みに失敗", "error", err)
	}
}

// parsePageParamはページ番号のクエリパラメータを読み取り、欠落・無効な値のときは1を返す。
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

// indexCanonicalPathは人物一覧ページの代表パスを生成する。現在の表示のog:urlと、
// ページネーションのリンクが伸ばす起点 (page = 1) の双方で使う。作品一覧と同じく
// 1ページ目はパラメータなしで書き、共有されたリンクが持つ未知のパラメータは載せない。
func indexCanonicalPath(page int32) string {
	if page > 1 {
		return "/db/people?page=" + strconv.FormatInt(int64(page), 10)
	}
	return "/db/people"
}
