package db_person

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

// UpdateはAnnict DB管理画面の人物編集フォームの1回の送信 (PATCH /db/people/{id}) を
// 処理する。
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	personID, ok := parsePersonIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	input := usecase.UpdatePersonInput{
		PersonID:          personID,
		User:              middleware.GetUserFromContext(ctx),
		Name:              r.FormValue("name"),
		NameKana:          r.FormValue("name_kana"),
		NameEn:            r.FormValue("name_en"),
		Nickname:          r.FormValue("nickname"),
		NicknameEn:        r.FormValue("nickname_en"),
		Gender:            r.FormValue("gender"),
		BloodType:         r.FormValue("blood_type"),
		PrefectureID:      r.FormValue("prefecture_id"),
		Birthday:          r.FormValue("birthday"),
		Height:            r.FormValue("height"),
		URL:               r.FormValue("url"),
		URLEn:             r.FormValue("url_en"),
		WikipediaURL:      r.FormValue("wikipedia_url"),
		WikipediaURLEn:    r.FormValue("wikipedia_url_en"),
		TwitterUsername:   r.FormValue("twitter_username"),
		TwitterUsernameEn: r.FormValue("twitter_username_en"),
		UpdatedAt:         r.FormValue("updated_at"),
	}

	output, err := h.updatePersonUC.Execute(ctx, input)
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
				// フォームを開いてから本送信までの間に、他者がその人物を書いた。送信された
				// 値と、冒頭に述べた競合の説明、そして保存済みの値を並べてフォームが返るため、
				// 編集者が両者を見比べて判断できる。自動マージは行わない。
				conflict := model.NewValidationError()
				conflict.AddGlobal(ae.UserMsg)
				h.renderRejectedUpdate(w, r, input, editFormState{
					Status:     http.StatusConflict,
					FormErrors: conflict,
					Conflict:   true,
				})
			default:
				slog.ErrorContext(ctx, ae.LogString())
				httperror.InternalServerError(w, r)
			}
			return
		}
		slog.ErrorContext(ctx, "人物の更新に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// 送信が成功したら同じ人物の編集フォームに着地する。RailsのDb::PeopleController#update
	// (db_edit_person_path) と同じ遷移で、編集者が保存された内容をその場で確かめられるため。
	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_person_updated"))
	http.Redirect(w, r, editPath(output.PersonID), http.StatusSeeOther)
}

// renderRejectedUpdateは適用されなかった送信に対して編集フォームを再描画し、送信された値を
// 保ったまま、適用を止めた理由を述べる。
func (h *Handler) renderRejectedUpdate(
	w http.ResponseWriter,
	r *http.Request,
	input usecase.UpdatePersonInput,
	state editFormState,
) {
	formInput := viewmodel.NewDBPersonFormInputFromSubmit(input)
	state.FormInput = &formInput
	h.renderEdit(w, r, input.PersonID, state)
}
