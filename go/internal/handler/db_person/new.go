package db_person

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_people"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// newPathは一括登録フォームの代表GETパス。Newはこのパスでページを配信し、Createは同じ
// ページをPOST /db/peopleから再描画するため、双方ともリクエストパスではなくここから
// og:urlを取る。
const newPath = "/db/people/new"

// NewはAnnict DB管理画面の人物の一括登録フォーム (GET /db/people/new) を描画する。
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	h.renderNew(w, r, newFormState{Status: http.StatusOK})
}

// newFormStateは一括登録ページのうち送信に依存する部分。新規フォームでは空で、送信が
// 失敗した後は却下された行とそのメッセージを持つ。
type newFormState struct {
	Status     int
	FormErrors *model.ValidationError
	Rows       string
}

// renderNewは一括登録ページを描画する。Newはこれを配信し、Createは送信された行が
// 却下されたときにこれで再描画する。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, state newFormState) {
	ctx := r.Context()

	if err := h.getDBPersonNewUC.Execute(ctx, usecase.GetDBPersonNewInput{
		User: middleware.GetUserFromContext(ctx),
	}); err != nil {
		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeForbidden {
			httperror.Forbidden(w, r)
			return
		}
		slog.ErrorContext(ctx, "DB人物一括登録フォームの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, newPath)
	meta.SetDBTitle(ctx, "db_people_new_title")

	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_people.New(db_people.NewPageData{
			CSRFToken:  middleware.GetCSRFToken(r, h.sessionManager),
			FormErrors: viewmodel.NewFormErrors(state.FormErrors),
			Rows:       state.Rows,
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
		slog.ErrorContext(ctx, "DB人物一括登録フォームのレスポンスの書き込みに失敗", "error", err)
	}
}
