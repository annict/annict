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
		WorkFormInput: parseWorkForm(r),
	}

	output, err := h.updateWorkUC.Execute(ctx, input)
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderEditWithErrors(w, r, input, ve)
			return
		}
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeResourceNotFound {
			httperror.NotFound(w, r)
			return
		}
		slog.ErrorContext(ctx, "作品の更新に失敗しました", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_updated"))
	http.Redirect(w, r, dbWorkEditPath(output.WorkID), http.StatusSeeOther)
}

// renderEditWithErrors re-renders the work edit form with validation errors and the previously submitted values.
//
// [Ja] バリデーションエラーと送信済みの入力値を保持したまま作品編集フォームを再描画する。
func (h *Handler) renderEditWithErrors(w http.ResponseWriter, r *http.Request, input usecase.UpdateWorkInput, formErrors *model.ValidationError) {
	ctx := r.Context()

	optionsResult, err := h.getDBWorkFormOptionsUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats)
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
			WorkTitle:   input.Title,
			FormOptions: formOptions,
			FormErrors:  formErrors,
			FormInput:   viewmodel.NewDBWorkFormInput(input.WorkFormInput),
		}),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "DB作品編集フォームのレスポンスの書き込みに失敗", "error", err)
	}
}
