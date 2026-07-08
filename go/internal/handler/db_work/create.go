package db_work

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Create processes the work creation request in the Annict DB admin UI (POST /db/works).
//
// [Ja] Annict DB 管理画面の作品作成リクエスト (POST /db/works) を処理する。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := usecase.CreateWorkInput{WorkFormInput: parseWorkForm(r)}

	output, err := h.createWorkUC.Execute(ctx, input)
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewWithErrors(w, r, input, ve)
			return
		}
		slog.ErrorContext(ctx, "作品の作成に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_created"))

	// Redirect to the just-created work's edit page, matching the Rails create action
	// (db_edit_work_path) and the Update handler, so the editor lands on the new work to
	// keep filling in its details.
	//
	// [Ja] 作成直後の作品の編集ページへリダイレクトする。Rails の create アクション
	// (db_edit_work_path) や Update ハンドラーと同じ遷移で、作成した作品で編集者がそのまま
	// 詳細を入力し続けられるようにする。
	http.Redirect(w, r, fmt.Sprintf("/db/works/%d/edit", output.WorkID), http.StatusSeeOther)
}

// renderNewWithErrors re-renders the new-work form with validation errors and the previously submitted values.
//
// [Ja] バリデーションエラーと送信済みの入力値を保持したまま新規作成フォームを再描画する。
func (h *Handler) renderNewWithErrors(w http.ResponseWriter, r *http.Request, input usecase.CreateWorkInput, formErrors *model.ValidationError) {
	ctx := r.Context()

	optionsResult, err := h.getDBWorkFormOptionsUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats)
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_new_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.New(db_works.NewPageData{
			CSRFToken:   csrfToken,
			FormOptions: formOptions,
			FormErrors:  formErrors,
			FormInput:   viewmodel.NewDBWorkFormInput(input.WorkFormInput),
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
	}
}
