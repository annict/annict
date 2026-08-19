package db_work

import (
	"bytes"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Update processes the work update request in the Annict DB admin UI (PATCH /db/works/:id).
//
// [Ja] Annict DB 管理画面の作品更新リクエスト (PATCH /db/works/:id) を処理する。
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	input := usecase.UpdateWorkInput{
		WorkID:        model.WorkID(id),
		UpdatedAt:     r.FormValue("updated_at"),
		WorkFormInput: parseWorkForm(r),
	}

	output, err := h.updateWorkUC.Execute(ctx, input)
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
			case model.AppErrCodeConflict:
				// Someone else wrote the work between the form being opened and this submit.
				// The form comes back with the submitted values, the conflict stated at the
				// top and the stored values beside them, so the editor compares the two and
				// decides; nothing is merged automatically.
				//
				// [Ja] フォームを開いてから本送信までの間に、他者がその作品を書いた。送信された
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
		slog.ErrorContext(ctx, "作品の更新に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_updated"))
	http.Redirect(w, r, dbWorkEditPath(output.WorkID), http.StatusSeeOther)
}

// editFormState is the part of the re-rendered edit page that depends on the submission: the
// status it comes back with and what was wrong about it.
//
// [Ja] editFormState は再描画する編集ページのうち送信に依存する部分。返すステータスと、送信の
// 何が問題だったかを持つ。
type editFormState struct {
	Status     int
	FormErrors *model.ValidationError
	// Conflict states that the submit was refused because someone else had written the work
	// first. The page then shows the stored values beside the submitted ones and carries the
	// stored version, so the editor can compare the two and, if they decide their values are
	// the ones to keep, submit again against the row they have just seen. A submit refused for
	// any other reason echoes back the version it was made against, since nothing about the
	// stored row was shown.
	//
	// [Ja] Conflict は、他者が先にその作品を書いたために送信が却下されたことを表す。ページは
	// 保存済みの値を送信された値と並べて表示し、保存済みの版を運ぶ。編集者が両者を見比べ、
	// 自分の値を残すと判断したなら、いま見た行に対して送信し直せるようにするため。それ以外の
	// 理由で却下された送信は、保存済みの行について何も示していないため、送信が前提とした版を
	// そのまま返す。
	Conflict bool
}

// renderRejectedUpdate re-renders the work edit form for a submit that was not applied, keeping
// the submitted values and stating what stopped them.
//
// [Ja] renderRejectedUpdate は適用されなかった送信に対して作品編集フォームを再描画し、送信された
// 値を保ったまま、適用を止めた理由を述べる。
func (h *Handler) renderRejectedUpdate(w http.ResponseWriter, r *http.Request, input usecase.UpdateWorkInput, state editFormState) {
	ctx := r.Context()

	formInput := viewmodel.NewDBWorkFormInputFromSubmit(input)

	formOptions, conflictCurrent, ok := h.editFormRejectionData(w, r, input, state, formInput)
	if !ok {
		return
	}

	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbWorkEditPath(input.WorkID))
	meta.SetDBTitle(ctx, "db_works_edit_title")

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.Edit(db_works.EditPageData{
			CSRFToken: csrfToken,
			WorkID:    viewmodel.WorkID(input.WorkID),
			// Validation runs before the work is loaded, so the stored title is not at hand
			// here; the submitted title names the work in the heading instead.
			//
			// [Ja] バリデーションは work の読み込みより前に走るため、ここでは保存済みの
			// タイトルを持たない。見出しでは代わりに送信されたタイトルで作品を示す。
			WorkTitle:       input.Title,
			FormOptions:     formOptions,
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
		slog.ErrorContext(ctx, "DB作品編集フォームのレスポンスの書き込みに失敗", "error", err)
	}
}

// editFormRejectionData loads what the re-rendered page needs beyond the submitted values, and
// reports false once it has written a response of its own. A conflict reads the whole stored
// work, which supplies both the values shown beside the submitted ones and the version a second
// submit has to state; formInput is moved onto that version, so the editor overwrites exactly
// the row they were just shown. Every other rejection needs only the select options, and leaves
// formInput carrying the version it was submitted with.
//
// [Ja] editFormRejectionData は再描画するページが送信された値の他に必要とするものを読み込み、
// 自ら応答を書いた場合に false を返す。競合では保存済みの work 全体を読む。送信された値と並べて
// 表示する値と、2 回目の送信が名乗るべき版の両方をそこから得るため。formInput はその版に載せ替え、
// 編集者がいま示された行だけを上書きするようにする。それ以外の却下では選択欄の選択肢だけを必要と
// し、formInput は送信された版を運んだままにする。
func (h *Handler) editFormRejectionData(
	w http.ResponseWriter,
	r *http.Request,
	input usecase.UpdateWorkInput,
	state editFormState,
	formInput *viewmodel.DBWorkFormInput,
) (viewmodel.DBWorkFormOptions, *viewmodel.DBWorkFormInput, bool) {
	ctx := r.Context()

	if !state.Conflict {
		optionsResult, err := h.getDBWorkFormOptionsUC.Execute(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
			httperror.InternalServerError(w, r)
			return viewmodel.DBWorkFormOptions{}, nil, false
		}

		return viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats), nil, true
	}

	output, err := h.getDBWorkEditUC.Execute(ctx, usecase.GetDBWorkEditInput{WorkID: input.WorkID})
	if err != nil {
		// The work is gone by the time the conflict is reported, so there is nothing left to
		// compare the submit against and no page to send it to again.
		//
		// [Ja] 競合を報告する時点で work が失われている。送信と見比べる対象も、送信し直す先の
		// ページも残っていない。
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return viewmodel.DBWorkFormOptions{}, nil, false
		}
		slog.ErrorContext(ctx, "DB作品編集フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return viewmodel.DBWorkFormOptions{}, nil, false
	}

	stored := viewmodel.NewDBWorkFormInputFromWork(output.Work)
	formInput.UpdatedAt = stored.UpdatedAt

	return viewmodel.NewDBWorkFormOptions(ctx, output.NumberFormats), stored, true
}
