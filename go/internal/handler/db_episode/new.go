package db_episode

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_episodes"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New renders the bulk-create form for a work's episodes in the Annict DB admin UI
// (GET /db/works/:work_id/episodes/new).
//
// [Ja] Annict DB 管理画面の、ある作品のエピソード一括作成フォーム
// (GET /db/works/:work_id/episodes/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	workID, ok := parseWorkIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	h.renderNew(w, r, workID, newFormState{Status: http.StatusOK})
}

// newFormState is the part of the bulk-create page that depends on the submission: nothing
// on a fresh form, and the rejected lines with their messages after a failed submit.
//
// [Ja] newFormState は一括作成ページのうち送信に依存する部分。新規フォームでは空で、送信が
// 失敗した後は却下された行とそのメッセージを持つ。
type newFormState struct {
	Status     int
	FormErrors *model.ValidationError
	Rows       string
}

// renderNew renders the bulk-create page. New serves it, and Create re-renders it with the
// submitted lines when they are rejected, so both go through here and cannot drift apart in
// how they describe the same page.
//
// [Ja] renderNew は一括作成ページを描画する。New はこれを配信し、Create は送信された行が
// 却下されたときにこれで再描画する。同じページの説明が両者でずれないよう、双方がここを通る。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, workID model.WorkID, state newFormState) {
	ctx := r.Context()

	output, err := h.getDBEpisodeNewUC.Execute(ctx, usecase.GetDBEpisodeNewInput{WorkID: workID})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DBエピソード一括作成フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	user := middleware.GetUserFromContext(ctx)
	workName := viewmodel.DBEpisodeListWorkName(output.Work.Title)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, newPath(output.Work.ID))
	setNewTitle(ctx, &meta, workName)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.New(db_episodes.NewPageData{
			WorkID:         viewmodel.WorkID(output.Work.ID),
			WorkName:       workName,
			NoEpisodes:     output.Work.NoEpisodes,
			CSRFToken:      middleware.GetCSRFToken(r, h.sessionManager),
			FormErrors:     viewmodel.NewFormErrors(state.FormErrors),
			Rows:           state.Rows,
			IsAdmin:        middleware.IsAdmin(user),
			ManualCreation: viewmodel.NewDBEpisodeManualCreationRestriction(output.ManualCreationState),
		}),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(state.Status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "DBエピソード一括作成フォームのレスポンスの書き込みに失敗", "error", err)
	}
}

// setNewTitle gives meta a document title that identifies the work whose episodes are being
// created. A work with no name to show falls back to the generic form title, matching the
// page heading.
//
// [Ja] setNewTitle は meta に、エピソードを作成する対象の作品を識別できる文書タイトルを設定
// する。表示できる名前が無い作品では、ページ見出しと同じく汎用のフォームタイトルへフォール
// バックする。
func setNewTitle(ctx context.Context, meta *viewmodel.PageMeta, workName string) {
	if workName == "" {
		meta.SetDBTitle(ctx, "db_episodes_new_title")
		return
	}

	meta.SetDBTitle(ctx, "db_episodes_new_document_title", map[string]any{"WorkTitle": workName})
}

// newPath builds the representative GET path of the bulk-create form. New serves the page at
// this path and Create re-renders the same page from POST /db/works/:work_id/episodes, so
// both take their og:url from here rather than from the request path.
//
// [Ja] newPath は一括作成フォームの代表 GET パスを生成する。New はこのパスでページを配信し、
// Create は同じページを POST /db/works/:work_id/episodes から再描画するため、双方とも
// リクエストパスではなくここから og:url を取る。
func newPath(workID model.WorkID) string {
	return fmt.Sprintf("/db/works/%d/episodes/new", int64(workID))
}
