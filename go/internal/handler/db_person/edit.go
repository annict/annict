package db_person

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_people"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// EditはAnnict DB管理画面の人物の編集フォーム (GET /db/people/{id}/edit) を描画する。
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	personID, ok := parsePersonIDParam(r)
	if !ok {
		httperror.NotFound(w, r)
		return
	}

	h.renderEdit(w, r, personID, editFormState{Status: http.StatusOK})
}

// editFormStateは編集ページのうち送信に依存する部分。フォームを開いたときは空で、送信が
// 却下された後は送信された値とその問題点を持つ。前者ではFormInputがnilで、ページは保存済み
// の値を描画する。
type editFormState struct {
	Status     int
	FormErrors *model.ValidationError
	FormInput  *viewmodel.DBPersonFormInput
	// Conflictは、他者が先にその人物を書いたために送信が却下されたことを表す。ページは
	// 保存済みの値を送信された値と並べて表示し、保存済みの版を運ぶ。編集者が両者を見比べ、
	// 自分の値を残すと判断したなら、いま見た行に対して送信し直せるようにするため。それ以外の
	// 理由で却下された送信は、保存済みの行について何も示していないため、送信が前提とした版を
	// そのまま返す。
	Conflict bool
}

// renderEditは編集ページを描画する。Editはこれを配信し、Updateは送信された値が却下され
// たときにこれで再描画する。同じページの説明が両者でずれないよう、双方がここを通る。
func (h *Handler) renderEdit(w http.ResponseWriter, r *http.Request, personID model.PersonID, state editFormState) {
	ctx := r.Context()

	output, err := h.getDBPersonEditUC.Execute(ctx, usecase.GetDBPersonEditInput{
		User:     middleware.GetUserFromContext(ctx),
		PersonID: personID,
	})
	if err != nil {
		if ae := model.AsAppError(err); ae != nil {
			switch ae.Code {
			case model.AppErrCodeResourceNotFound:
				httperror.NotFound(w, r)
				return
			case model.AppErrCodeForbidden:
				httperror.Forbidden(w, r)
				return
			}
		}
		slog.ErrorContext(ctx, "DB人物編集フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	target := viewmodel.NewDBPersonEditTarget(output.Person)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, editPath(output.Person.ID))
	setEditTitle(ctx, &meta, target.Name)

	// フォームは保存済みの値で開き、却下された送信では入力された内容に差し替える。編集者が
	// 入力を打ち直さずに手直しできるようにするため。
	stored := viewmodel.NewDBPersonFormInputFromPerson(output.Person)
	formInput := stored
	var conflictCurrent *viewmodel.DBPersonFormInput
	if state.FormInput != nil {
		formInput = *state.FormInput
	}
	if state.Conflict {
		conflictCurrent = &stored
		formInput.UpdatedAt = stored.UpdatedAt
	}

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_people.Edit(db_people.EditPageData{
			Person:          target,
			CSRFToken:       middleware.GetCSRFToken(r, h.sessionManager),
			FormOptions:     viewmodel.NewDBPersonFormOptions(ctx, output.Prefectures),
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
		slog.ErrorContext(ctx, "DB人物編集フォームのレスポンスの書き込みに失敗", "error", err)
	}
}

// setEditTitleはmetaに、画面名と人物の名前からなる文書タイトルを設定する。名前の異なる
// 人物の編集ページを、タブ・履歴・支援技術で区別できるようにするため。
func setEditTitle(ctx context.Context, meta *viewmodel.PageMeta, name string) {
	meta.SetDBTitle(ctx, "db_people_edit_document_title", map[string]any{"Name": name})
}

// editPathは人物編集フォームの代表GETパスを生成する。Editはこのパスでページを配信し、
// Updateは同じページをPATCH /db/people/{id} から再描画するため、双方ともリクエストパス
// ではなくここからog:urlを取る。Updateの保存後のリダイレクト先でもある。
func editPath(personID model.PersonID) string {
	return fmt.Sprintf("/db/people/%d/edit", int64(personID))
}

// parsePersonIDParamはURLの {id} を人物IDとして読む。数値でなければfalseを返し、呼び出し
// 側は存在しない人物と同じく404を返す。
func parsePersonIDParam(r *http.Request) (model.PersonID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.PersonID(id), true
}
