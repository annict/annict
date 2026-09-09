package db_work_unarchive

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// dbWorkUnarchiveNewPath builds the representative GET path of a work's publish confirmation
// page. It is built from the parsed work ID rather than from the request path so that links
// spelling the ID differently (a leading zero, say) still resolve to one representative URL.
//
// [Ja] dbWorkUnarchiveNewPath は作品の公開確認ページの代表 GET パスを生成する。リクエスト
// パスではなくパース済みの作品 ID から組み立てることで、ID の表記が違うリンク (先頭ゼロなど)
// でも 1 つの代表 URL に収まるようにする。
func dbWorkUnarchiveNewPath(id model.WorkID) string {
	return fmt.Sprintf("/db/works/%d/unarchive/new", int64(id))
}

// New renders the publish-confirmation page in the Annict DB admin UI
// (GET /db/works/:id/unarchive/new).
//
// [Ja] Annict DB 管理画面の公開確認ページ (GET /db/works/:id/unarchive/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBWorkUnarchiveNewUC.Execute(ctx, usecase.GetDBWorkUnarchiveNewInput{
		User:   middleware.GetUserFromContext(ctx),
		WorkID: model.WorkID(id),
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
				return
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
				return
			}
		}
		slog.ErrorContext(ctx, "公開確認画面の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)
	workName := strings.TrimSpace(output.Work.Title)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbWorkUnarchiveNewPath(output.Work.ID))
	setNewTitle(ctx, &meta, workName)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.UnarchiveNew(db_works.UnarchiveNewPageData{
			CSRFToken: csrfToken,
			WorkID:    viewmodel.WorkID(output.Work.ID),
			Title:     workName,
			ReturnTo:  returnPath(r),
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
		slog.ErrorContext(ctx, "DB作品公開確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}

// setNewTitle gives meta a document title that starts with the page name, followed by the work
// once it has a display name. A work without one leaves the page name standing alone, which is
// also what the heading shows, so the two never disagree on whether the target can be named.
//
// [Ja] setNewTitle は meta に、画面名から始まり、表示名があれば作品が続く文書タイトルを
// 設定する。表示名が無い作品では画面名だけになり、見出しの表示とも揃う。対象を名指しできる
// かどうかの判断が両者で食い違わないようにするため。
func setNewTitle(ctx context.Context, meta *viewmodel.PageMeta, workName string) {
	if workName == "" {
		meta.SetDBTitle(ctx, "db_works_unarchive_new_title")
		return
	}

	meta.SetDBTitle(ctx, "db_works_unarchive_new_document_title", map[string]any{"WorkTitle": workName})
}

// dbWorkListPath is where the confirmation returns the reader when the link names no return_to:
// the work list of the Annict DB admin UI.
//
// [Ja] dbWorkListPath はリンクが return_to を伴わないときに確認画面が読み手を戻す先。Annict DB
// 管理画面の作品一覧。
const dbWorkListPath = "/db/works"

// returnPath is the listing the confirmation sends the reader back to, both from its cancel link
// and from the form it submits. It carries the listing the reader came from so neither leaving
// nor completing the action drops them on a list they did not ask for, and an absent or
// non-Annict-DB value falls back to the work list.
//
// [Ja] returnPath は確認画面がキャンセルリンクと送信フォームの双方で読み手を戻す一覧。読み手が
// 来た一覧を持ち回ることで、操作をやめた場合も完了した場合も、読み手が求めていない一覧に着地
// させない。値が無い場合や Annict DB のパスでない場合は作品一覧にフォールバックする。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbWorkListPath)
}
