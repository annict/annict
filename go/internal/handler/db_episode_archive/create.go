package db_episode_archive

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Create archives (unpublishes) an episode in the Annict DB admin UI
// (POST /db/episodes/:id/archive).
//
// [Ja] Annict DB 管理画面でエピソードを非公開 (アーカイブ) にする
// (POST /db/episodes/:id/archive)。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.archiveEpisodeUC.Execute(ctx, usecase.ArchiveEpisodeInput{
		EpisodeID: episodeID,
		User:      middleware.GetUserFromContext(ctx),
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
			default:
				slog.ErrorContext(ctx, ae.LogString())
				httperror.InternalServerError(w, r)
			}
			return
		}
		slog.ErrorContext(ctx, "エピソードの非公開に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episode_archived"))
	http.Redirect(w, r, episodesPath(output.WorkID), http.StatusSeeOther)
}
