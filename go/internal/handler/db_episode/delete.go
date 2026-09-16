package db_episode

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DeleteはAnnict DB管理画面でエピソードをソフトデリートする (DELETE /db/episodes/:id)。Railsの
// Db::EpisodesController#destroyと同じく確認アラートのみ (非公開と違い確認画面は挟まない)。
// admin認可はルートのRequireAdmin middlewareで強制し、UseCaseでも繰り返す。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.deleteEpisodeUC.Execute(ctx, usecase.DeleteEpisodeInput{
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
		slog.ErrorContext(ctx, "エピソードの削除に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// 削除が成功したらその作品のエピソード一覧に着地する。Railsのdestroyアクション
	// (db_episode_list_path) と同じ遷移で、管理者が次に確認するのは残った行であるため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episode_deleted"))

	// htmxはfetchのリダイレクトを透過的に追うため、303だと一覧ページが押したボタンに
	// スワップされ遷移しない。htmxリクエスト (エピソード一覧の削除ボタン) にはHX-Redirectを
	// 返して一覧へフル遷移させる。上で設定したflashは遷移後のGETで表示される。非htmx
	// クライアントには従来どおり303を返す。エピソード一覧の操作列が踏襲する作品一覧の削除ボタン
	// と同じ形である。
	listPath := indexPath(output.WorkID, 1)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", listPath)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, listPath, http.StatusSeeOther)
}
