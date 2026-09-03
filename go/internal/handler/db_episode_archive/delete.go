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

// Delete re-publishes (un-archives) an episode in the Annict DB admin UI
// (DELETE /db/episodes/:id/archive). Unlike the archive it takes no confirmation page: the
// episode list submits it directly, as the work list does for its work counterpart.
//
// [Ja] Annict DB 管理画面でエピソードを再公開 (アーカイブ解除) にする
// (DELETE /db/episodes/:id/archive)。非公開と違い確認ページは挟まず、作品一覧が作品側で行うのと
// 同じくエピソード一覧から直接送信する。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.unarchiveEpisodeUC.Execute(ctx, usecase.UnarchiveEpisodeInput{
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
		slog.ErrorContext(ctx, "エピソードの再公開に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episode_published"))

	// htmx follows fetch redirects transparently, so a 303 would swap the list page into the
	// clicked button instead of navigating. For htmx requests (the episode list's publish
	// button) return HX-Redirect so htmx does a full navigation to the list; the flash set above
	// is shown on the followed GET. Non-htmx clients keep the plain 303 redirect. This mirrors
	// the work list's publish button, which the episode list's action column follows.
	//
	// [Ja] htmx は fetch のリダイレクトを透過的に追うため、303 だと一覧ページが押したボタンに
	// スワップされ遷移しない。htmx リクエスト (エピソード一覧の公開ボタン) には HX-Redirect を
	// 返して一覧へフル遷移させる。上で設定した flash は遷移後の GET で表示される。非 htmx
	// クライアントには従来どおり 303 を返す。エピソード一覧の操作列が踏襲する作品一覧の公開ボタン
	// と同じ形である。
	listPath := episodesPath(output.WorkID)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", listPath)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, listPath, http.StatusSeeOther)
}
