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

// Edit renders the edit form of a single episode in the Annict DB admin UI
// (GET /db/episodes/:id/edit).
//
// [Ja] Annict DB 管理画面の単一エピソードの編集フォーム (GET /db/episodes/:id/edit) を
// 描画する。
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBEpisodeEditUC.Execute(ctx, usecase.GetDBEpisodeEditInput{EpisodeID: episodeID})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DBエピソード編集フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	workName := viewmodel.DBEpisodeListWorkName(output.Work.Title)
	episodeIdentifier := viewmodel.DBEpisodeEditIdentifier(ctx, output.Episode)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, editPath(output.Episode.ID))
	setEditTitle(ctx, &meta, episodeIdentifier, workName)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.Edit(db_episodes.EditPageData{
			EpisodeID:  viewmodel.EpisodeID(output.Episode.ID),
			WorkID:     viewmodel.WorkID(output.Work.ID),
			WorkName:   workName,
			NoEpisodes: output.Work.NoEpisodes,
			CSRFToken:  middleware.GetCSRFToken(r, h.sessionManager),
			FormInput:  viewmodel.NewDBEpisodeFormInputFromEpisode(output.Episode),
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
		slog.ErrorContext(ctx, "DBエピソード編集フォームのレスポンスの書き込みに失敗", "error", err)
	}
}

// setEditTitle gives meta a document title that starts with the episode's unique identifier,
// followed by the work when it has a name. The identifier remains when the work has no name, so
// every episode edit page can still be distinguished in tabs, history, and assistive technology.
//
// [Ja] setEditTitle は meta に、エピソード固有の識別子から始まり、名前があれば作品が続く文書
// タイトルを設定する。作品に表示名が無くても識別子を残し、タブ、履歴、支援技術で各エピソード
// 編集ページを区別できるようにする。
func setEditTitle(ctx context.Context, meta *viewmodel.PageMeta, episodeIdentifier string, workName string) {
	templateData := map[string]any{"EpisodeIdentifier": episodeIdentifier}
	if workName == "" {
		meta.SetDBTitle(ctx, "db_episodes_edit_document_title_without_work", templateData)
		return
	}

	templateData["WorkTitle"] = workName
	meta.SetDBTitle(ctx, "db_episodes_edit_document_title", templateData)
}

// editPath builds the representative GET path of an episode's edit form, which the page
// takes its og:url from.
//
// [Ja] editPath はエピソード編集フォームの代表 GET パスを生成する。ページはここから og:url
// を取る。
func editPath(episodeID model.EpisodeID) string {
	return fmt.Sprintf("/db/episodes/%d/edit", int64(episodeID))
}
