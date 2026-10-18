package db_person_delisting

import (
	"bytes"
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

// dbPersonDelistingNewPathは人物の掲載停止確認ページの代表GETパスを生成する。リクエストパスでは
// なくパース済みの人物IDから組み立てることで、IDの表記が違うリンク (先頭ゼロなど) でも1つの
// 代表URLに収まるようにする。
func dbPersonDelistingNewPath(id model.PersonID) string {
	return fmt.Sprintf("/db/people/%d/delisting/new", int64(id))
}

// NewはAnnict DB管理画面の人物の掲載停止確認ページ (GET /db/people/:id/delisting/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	h.renderNew(w, r, delistingFormState{Status: http.StatusOK})
}

// delistingFormStateは確認画面のうち送信に依存する部分。確認画面を開いたときはDelistMessageが
// nilで、入力欄は保存済みの対応記録を初期値にする。送信が却下された後は送信された対応記録を持ち、
// 管理者が書いた記録を失わずに送り直せるようにする (人物の編集の再描画と同じ)。
type delistingFormState struct {
	Status        int
	DelistMessage *string
}

// renderNewは掲載停止確認ページを描画する。Newはこれを配信し、Createは送信が却下されたときに
// これで再描画する。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, state delistingFormState) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httperror.NotFound(w, r)
		return
	}

	output, err := h.getDBPersonDelistingNewUC.Execute(ctx, usecase.GetDBPersonDelistingNewInput{
		User:     middleware.GetUserFromContext(ctx),
		PersonID: model.PersonID(id),
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
		slog.ErrorContext(ctx, "人物の掲載停止確認画面の取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	// 名前の異なる人物の確認画面を、タブ・履歴・支援技術で区別できるよう、文書タイトルに
	// 人物の名前を含める (編集ページと同じ)。
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, dbPersonDelistingNewPath(output.Person.ID))
	meta.SetDBTitle(ctx, "db_people_delisting_new_document_title", map[string]any{"Name": output.Person.Name})

	person := viewmodel.NewDBPersonDelistingTarget(output.Person)
	delistMessage := person.DelistMessage
	if state.DelistMessage != nil {
		delistMessage = *state.DelistMessage
	}

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_people.DelistingNew(db_people.DelistingNewPageData{
			CSRFToken:      middleware.GetCSRFToken(r, h.sessionManager),
			Person:         person,
			DelistMessage:  delistMessage,
			ReturnTo:       returnPath(r),
			Conflict:       state.Status == http.StatusConflict,
			InvalidVersion: state.Status == http.StatusUnprocessableEntity,
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
		slog.ErrorContext(ctx, "DB人物掲載停止確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}
