package db_episode

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Update processes one submit of the episode edit form in the Annict DB admin UI
// (PATCH /db/episodes/:id).
//
// [Ja] Annict DB 管理画面のエピソード編集フォームの 1 回の送信 (PATCH /db/episodes/:id) を
// 処理する。
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	episodeID, ok := parseEpisodeIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	input := usecase.UpdateEpisodeInput{
		EpisodeID:  episodeID,
		User:       middleware.GetUserFromContext(ctx),
		Number:     r.FormValue("number"),
		RawNumber:  r.FormValue("raw_number"),
		SortNumber: r.FormValue("sort_number"),
		Title:      r.FormValue("title"),
		TitleEn:    r.FormValue("title_en"),
		UpdatedAt:  r.FormValue("updated_at"),
	}

	output, err := h.updateEpisodeUC.Execute(ctx, input)
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderRejectedUpdate(w, r, input, editFormState{
				Status:     http.StatusUnprocessableEntity,
				FormErrors: ve,
			})
			return
		}
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
			case model.AppErrCodeConflict:
				// Someone else wrote the episode between the form being opened and this
				// submit. The form comes back with the submitted values, the conflict stated
				// at the top and the stored values beside them, so the editor compares the two
				// and decides; nothing is merged automatically.
				//
				// [Ja] フォームを開いてから本送信までの間に、他者がそのエピソードを書いた。
				// 送信された値と、冒頭に述べた競合の説明、そして保存済みの値を並べてフォームが
				// 返るため、編集者が両者を見比べて判断できる。自動マージは行わない。
				conflict := model.NewValidationError()
				conflict.AddGlobal(ae.UserMsg)
				h.renderRejectedUpdate(w, r, input, editFormState{
					Status:     http.StatusConflict,
					FormErrors: conflict,
					Conflict:   true,
				})
			case model.AppErrCodeBusy:
				// The submit was not applied because another write held a row it needed, not
				// because the stored episode disagrees with it. Nothing was written and the
				// version the form carries still matches, so the same submit succeeds once the
				// other writer commits: the form comes back with the submitted values and that
				// version, and says to send it again. 503 rather than 409 states that the
				// request was refused for a temporary condition, and Retry-After puts a number
				// on how brief it is.
				//
				// [Ja] 送信が適用されなかったのは、必要な行を他の書き込みが保持していたためで
				// あり、保存済みのエピソードと食い違ったためではない。何も書かれておらず、
				// フォームが運ぶ版も一致したままなので、相手が commit すれば同じ送信で成功する。
				// フォームは送信された値とその版を保ったまま返り、もう一度送るよう伝える。409 で
				// はなく 503 にするのは、一時的な状況で拒否したことを表すため。Retry-After は
				// その短さに数値を与える。
				busy := model.NewValidationError()
				busy.AddGlobal(ae.UserMsg)
				w.Header().Set("Retry-After", "1")
				h.renderRejectedUpdate(w, r, input, editFormState{
					Status:     http.StatusServiceUnavailable,
					FormErrors: busy,
				})
			default:
				slog.ErrorContext(ctx, ae.LogString())
				httperror.InternalServerError(w, r)
			}
			return
		}
		slog.ErrorContext(ctx, "エピソードの更新に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// A successful submit lands on the work's episode list, matching the Rails update action
	// (db_episode_list_path): the edited row in its place among the others is what the editor
	// checks next.
	//
	// [Ja] 送信が成功したらその作品のエピソード一覧に着地する。Rails の update アクション
	// (db_episode_list_path) と同じ遷移で、編集者が次に確認するのは他の行と並んだ編集後の行で
	// あるため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_episode_updated"))
	http.Redirect(w, r, indexPath(output.WorkID, 1), http.StatusSeeOther)
}

// renderRejectedUpdate re-renders the edit form for a submit that was not applied, keeping the
// submitted values and stating what stopped them.
//
// [Ja] renderRejectedUpdate は適用されなかった送信に対して編集フォームを再描画し、送信された値を
// 保ったまま、適用を止めた理由を述べる。
func (h *Handler) renderRejectedUpdate(
	w http.ResponseWriter,
	r *http.Request,
	input usecase.UpdateEpisodeInput,
	state editFormState,
) {
	formInput := viewmodel.NewDBEpisodeFormInputFromSubmit(input)
	state.FormInput = &formInput
	h.renderEdit(w, r, input.EpisodeID, state)
}
