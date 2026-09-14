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

// DeleteはAnnict DB管理画面でエピソードを再公開 (アーカイブ解除) にする
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

	// htmxはfetchのリダイレクトを透過的に追うため、303だと一覧ページが押したボタンに
	// スワップされ遷移しない。htmxリクエスト (エピソード一覧の公開ボタン) にはHX-Redirectを
	// 返して一覧へフル遷移させる。上で設定したflashは遷移後のGETで表示される。非htmx
	// クライアントには従来どおり303を返す。エピソード一覧の操作列が踏襲する作品一覧の公開ボタン
	// と同じ形である。
	listPath := episodesPath(output.WorkID)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", listPath)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, listPath, http.StatusSeeOther)
}
