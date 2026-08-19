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
	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	h.renderEdit(w, r, episodeID, editFormState{Status: http.StatusOK})
}

// editFormState is the part of the edit page that depends on the submission: nothing when the
// form is opened, and the submitted values with what was wrong about them after a rejected
// submit. FormInput is nil in the former case, where the page renders the stored values.
//
// [Ja] editFormState は編集ページのうち送信に依存する部分。フォームを開いたときは空で、送信が
// 却下された後は送信された値とその問題点を持つ。前者では FormInput が nil で、ページは保存済み
// の値を描画する。
type editFormState struct {
	Status     int
	FormErrors *model.ValidationError
	FormInput  *viewmodel.DBEpisodeFormInput
	// Conflict states that the submit was refused because someone else had written the episode
	// first. The page then shows the stored values beside the submitted ones and carries the
	// stored version, so the editor can compare the two and, if they decide their values are
	// the ones to keep, submit again against the row they have just seen. A submit refused for
	// any other reason echoes back the version it was made against, since nothing about the
	// stored row was shown.
	//
	// [Ja] Conflict は、他者が先にそのエピソードを書いたために送信が却下されたことを表す。
	// ページは保存済みの値を送信された値と並べて表示し、保存済みの版を運ぶ。編集者が両者を
	// 見比べ、自分の値を残すと判断したなら、いま見た行に対して送信し直せるようにするため。
	// それ以外の理由で却下された送信は、保存済みの行について何も示していないため、送信が前提と
	// した版をそのまま返す。
	Conflict bool
}

// renderEdit renders the edit page. Edit serves it, and Update re-renders it with the submitted
// values when they are rejected, so both go through here and cannot drift apart in how they
// describe the same page.
//
// [Ja] renderEdit は編集ページを描画する。Edit はこれを配信し、Update は送信された値が却下され
// たときにこれで再描画する。同じページの説明が両者でずれないよう、双方がここを通る。
func (h *Handler) renderEdit(w http.ResponseWriter, r *http.Request, episodeID model.EpisodeID, state editFormState) {
	ctx := r.Context()

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
	episodeIdentifier := viewmodel.DBEpisodeIdentifier(ctx, output.Episode)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, editPath(output.Episode.ID))
	setEditTitle(ctx, &meta, episodeIdentifier, workName)

	// The stored values open the form; a rejected submit replaces them with what was typed so
	// the editor corrects the input instead of retyping it.
	//
	// [Ja] フォームは保存済みの値で開き、却下された送信では入力された内容に差し替える。編集者が
	// 入力を打ち直さずに手直しできるようにするため。
	stored := viewmodel.NewDBEpisodeFormInputFromEpisode(output.Episode)
	formInput := stored
	var conflictCurrent *viewmodel.DBEpisodeFormInput
	if state.FormInput != nil {
		formInput = *state.FormInput
	}
	if state.Conflict {
		conflictCurrent = &stored
		formInput.UpdatedAt = stored.UpdatedAt
	}

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.Edit(db_episodes.EditPageData{
			EpisodeID:       viewmodel.EpisodeID(output.Episode.ID),
			WorkID:          viewmodel.WorkID(output.Work.ID),
			WorkName:        workName,
			NoEpisodes:      output.Work.NoEpisodes,
			CSRFToken:       middleware.GetCSRFToken(r, h.sessionManager),
			FormErrors:      viewmodel.NewFormErrors(state.FormErrors),
			FormInput:       formInput,
			ConflictCurrent: conflictCurrent,
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
