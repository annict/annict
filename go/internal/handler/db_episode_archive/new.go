package db_episode_archive

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

// New renders the archive-confirmation page of a single episode in the Annict DB admin UI
// (GET /db/episodes/:id/archive/new).
//
// [Ja] Annict DB 管理画面の単一エピソードの非公開確認ページ
// (GET /db/episodes/:id/archive/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBEpisodeArchiveNewUC.Execute(ctx, usecase.GetDBEpisodeArchiveNewInput{EpisodeID: episodeID})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "DBエピソード非公開確認ページの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	workName := viewmodel.DBEpisodeListWorkName(output.Work.Title)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, newPath(output.Episode.ID))
	setNewTitle(ctx, &meta, viewmodel.DBEpisodeIdentifier(ctx, output.Episode), workName)

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_episodes.ArchiveNew(db_episodes.ArchiveNewPageData{
			EpisodeID:   viewmodel.EpisodeID(output.Episode.ID),
			EpisodeName: viewmodel.DBEpisodeName(ctx, output.Episode),
			WorkID:      viewmodel.WorkID(output.Work.ID),
			WorkName:    workName,
			NoEpisodes:  output.Work.NoEpisodes,
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
		slog.ErrorContext(ctx, "DBエピソード非公開確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}

// setNewTitle gives meta a document title that starts with the page name, followed by the
// episode and then the work when it has a name, as the episode edit page does. The episode
// remains when the work has no name, so confirmation pages whose episode labels differ stay
// distinguishable in tabs, history, and assistive technology.
//
// [Ja] setNewTitle は meta に、画面名から始まり、エピソード、名前があれば作品が続く文書
// タイトルを設定する (エピソード編集ページと同じ形)。作品に表示名が無くてもエピソードを残し、
// エピソードのラベルが異なる確認ページをタブ・履歴・支援技術で区別できるようにする。
func setNewTitle(ctx context.Context, meta *viewmodel.PageMeta, episodeIdentifier string, workName string) {
	templateData := map[string]any{"EpisodeIdentifier": episodeIdentifier}
	if workName == "" {
		meta.SetDBTitle(ctx, "db_episodes_archive_new_document_title_without_work", templateData)
		return
	}

	templateData["WorkTitle"] = workName
	meta.SetDBTitle(ctx, "db_episodes_archive_new_document_title", templateData)
}

// newPath builds the representative GET path of an episode's archive-confirmation page, which
// the page takes its og:url from. It is built from the parsed episode ID rather than from the
// request path so that links spelling the ID differently (a leading zero, say) still resolve to
// one representative URL.
//
// [Ja] newPath はエピソードの非公開確認ページの代表 GET パスを生成する。ページはここから
// og:url を取る。リクエストパスではなくパース済みのエピソード ID から組み立てることで、ID の
// 表記が違うリンク (先頭ゼロなど) でも 1 つの代表 URL に収まるようにする。
func newPath(episodeID model.EpisodeID) string {
	return fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID))
}
