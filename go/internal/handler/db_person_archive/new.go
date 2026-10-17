package db_person_archive

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_people"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// dbPersonArchiveNewPathは人物のアーカイブ確認ページの代表GETパスを生成する。リクエストパスではなく
// パース済みの人物IDから組み立てることで、IDの表記が違うリンク (先頭ゼロなど) でも1つの
// 代表URLに収まるようにする。
func dbPersonArchiveNewPath(id model.PersonID) string {
	return fmt.Sprintf("/db/people/%d/archive/new", int64(id))
}

// NewはAnnict DB管理画面の人物のアーカイブ確認ページ (GET /db/people/:id/archive/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBPersonArchiveNewUC.Execute(ctx, usecase.GetDBPersonArchiveNewInput{
		User:     middleware.GetUserFromContext(ctx),
		PersonID: model.PersonID(id),
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
		slog.ErrorContext(ctx, "人物のアーカイブ確認画面の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// 名前の異なる人物の確認画面を、タブ・履歴・支援技術で区別できるよう、文書タイトルに
	// 人物の名前を含める (編集ページと同じ)。
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbPersonArchiveNewPath(output.Person.ID))
	meta.SetDBTitle(ctx, "db_people_archive_new_document_title", map[string]any{"Name": output.Person.Name})

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_people.ArchiveNew(db_people.ArchiveNewPageData{
			CSRFToken: middleware.GetCSRFToken(r, h.sessionManager),
			PersonID:  viewmodel.PersonID(output.Person.ID),
			Name:      output.Person.Name,
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
		slog.ErrorContext(ctx, "DB人物アーカイブ確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}
