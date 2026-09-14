package db_work

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
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// dbWorkEditPathは作品編集フォームの代表GETパスを生成する。Editはこのパスでページを
// 配信し、Updateは同じページをPATCH /db/works/:idから再描画するため、双方ともリクエスト
// パスではなくここからcanonical URLを取る。CreateとUpdateの保存後のリダイレクト先でも
// ある。
func dbWorkEditPath(id model.WorkID) string {
	return fmt.Sprintf("/db/works/%d/edit", int64(id))
}

// setEditTitleはmetaに、画面名から始まり、表示名があれば作品が続く文書タイトルを
// 設定する。表示名が無い作品では画面名だけになり、見出しの表示とも揃う。対象を名指しできる
// かどうかの判断が両者で食い違わないようにするため。
func setEditTitle(ctx context.Context, meta *viewmodel.PageMeta, workName string) {
	if workName == "" {
		meta.SetDBTitle(ctx, "db_works_edit_title")
		return
	}

	meta.SetDBTitle(ctx, "db_works_edit_document_title", map[string]any{"WorkTitle": workName})
}

// EditはAnnict DB管理画面の作品編集フォームページ (GET /db/works/:id/edit) を描画する。
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBWorkEditUC.Execute(ctx, usecase.GetDBWorkEditInput{WorkID: model.WorkID(id)})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DB作品編集フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, output.NumberFormats)
	formInput := viewmodel.NewDBWorkFormInputFromWork(output.Work)
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	workName := strings.TrimSpace(output.Work.Title)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbWorkEditPath(output.Work.ID))
	setEditTitle(ctx, &meta, workName)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.Edit(db_works.EditPageData{
			CSRFToken:   csrfToken,
			WorkID:      viewmodel.WorkID(output.Work.ID),
			WorkTitle:   workName,
			FormOptions: formOptions,
			FormInput:   formInput,
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
		slog.ErrorContext(ctx, "DB作品編集フォームのレスポンスの書き込みに失敗", "error", err)
	}
}
