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

// Delete soft-deletes an episode in the Annict DB admin UI (DELETE /db/episodes/:id). Like the
// Rails Db::EpisodesController#destroy it is guarded by a confirmation alert only (no
// confirmation screen, unlike the archive); admin authorization is enforced by the RequireAdmin
// middleware on the route and repeated by the usecase.
//
// [Ja] Annict DB 管理画面でエピソードをソフトデリートする (DELETE /db/episodes/:id)。Rails の
// Db::EpisodesController#destroy と同じく確認アラートのみ (非公開と違い確認画面は挟まない)。
// admin 認可はルートの RequireAdmin middleware で強制し、UseCase でも繰り返す。
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

	// A successful delete lands on the work's episode list, matching the Rails destroy action
	// (db_episode_list_path): the remaining rows are what the administrator checks next.
	//
	// [Ja] 削除が成功したらその作品のエピソード一覧に着地する。Rails の destroy アクション
	// (db_episode_list_path) と同じ遷移で、管理者が次に確認するのは残った行であるため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episode_deleted"))

	// htmx follows fetch redirects transparently, so a 303 would swap the list page into the
	// clicked button instead of navigating. For htmx requests (the episode list's delete button)
	// return HX-Redirect so htmx does a full navigation to the list; the flash set above is shown
	// on the followed GET. Non-htmx clients keep the plain 303 redirect. This mirrors the work
	// list's delete button, which the episode list's action column follows.
	//
	// [Ja] htmx は fetch のリダイレクトを透過的に追うため、303 だと一覧ページが押したボタンに
	// スワップされ遷移しない。htmx リクエスト (エピソード一覧の削除ボタン) には HX-Redirect を
	// 返して一覧へフル遷移させる。上で設定した flash は遷移後の GET で表示される。非 htmx
	// クライアントには従来どおり 303 を返す。エピソード一覧の操作列が踏襲する作品一覧の削除ボタン
	// と同じ形である。
	listPath := indexPath(output.WorkID, 1)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", listPath)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, listPath, http.StatusSeeOther)
}
