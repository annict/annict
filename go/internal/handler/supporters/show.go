package supporters

import (
	"context"
	"log/slog"
	"net/http"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/templates/layouts"
	supportersTemplate "github.com/annict/annict/go/internal/templates/pages/supporters"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Show GET /supporters - サポーターページを表示
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// コンテキストからユーザー情報を取得
	user := authMiddleware.GetUserFromContext(ctx)
	// サイドバー用のviewmodelに変換
	viewUser := viewmodel.NewUserForSidebar(user, h.imageHelper)

	// フラッシュメッセージを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)

	// クエリパラメータからメッセージ表示フラグを取得
	showSuccessMessage := r.URL.Query().Get("success") == "true"
	showCanceledMessage := r.URL.Query().Get("canceled") == "true"

	// サポーターページのビューモデルを作成
	pageData := viewmodel.SupporterPageData{
		IsLoggedIn:          user != nil,
		Status:              viewmodel.SupporterStatusNone,
		ShowSuccessMessage:  showSuccessMessage,
		ShowCanceledMessage: showCanceledMessage,
	}

	if user != nil {
		pageData = h.buildSupporterPageData(ctx, user, pageData)
	}

	// ページメタ情報を準備
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "supporters_title")

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Default(
		ctx,
		meta,
		viewUser,
		flash,
		h.cfg.GetAssetVersion(),
		supportersTemplate.Show(ctx, pageData),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// buildSupporterPageData はユーザーのサポーター情報からビューモデルを構築します
func (h *Handler) buildSupporterPageData(ctx context.Context, user *repository.User, data viewmodel.SupporterPageData) viewmodel.SupporterPageData {
	var isStripeActive, isGumroadActive bool

	// Stripeサブスクリプションをチェック
	if user.StripeSubscriberID.Valid && h.stripeSubscriberRepo != nil {
		stripeSubscriber, err := h.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
		if err == nil {
			if h.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
				isStripeActive = true
				data.StripeSubscriber = convertStripeSubscriberToView(&stripeSubscriber)
			}
		}
	}

	// Gumroadサブスクリプションをチェック
	if user.GumroadSubscriberID.Valid && h.gumroadSubscriberRepo != nil {
		gumroadSubscriber, err := h.gumroadSubscriberRepo.GetByID(ctx, user.GumroadSubscriberID.Int64)
		if err == nil {
			if h.gumroadSubscriberRepo.IsActive(&gumroadSubscriber) {
				isGumroadActive = true
				data.GumroadSubscriber = convertGumroadSubscriberToView(&gumroadSubscriber)
			}
		}
	}

	// サポーター状態を設定
	switch {
	case isStripeActive && isGumroadActive:
		data.Status = viewmodel.SupporterStatusBoth
	case isStripeActive:
		data.Status = viewmodel.SupporterStatusStripe
	case isGumroadActive:
		data.Status = viewmodel.SupporterStatusGumroad
	default:
		data.Status = viewmodel.SupporterStatusNone
	}

	return data
}

// convertStripeSubscriberToView はStripeサブスクライバーをビューモデルに変換します
func convertStripeSubscriberToView(s *repository.StripeSubscriber) *viewmodel.StripeSubscriberView {
	view := &viewmodel.StripeSubscriberView{
		Status:           s.StripeStatus,
		CurrentPeriodEnd: s.StripeCurrentPeriodEnd,
	}
	if s.StripeCancelAt.Valid {
		view.CancelAt = &s.StripeCancelAt.Time
	}
	return view
}

// convertGumroadSubscriberToView はGumroadサブスクライバーをビューモデルに変換します
func convertGumroadSubscriberToView(s *repository.GumroadSubscriber) *viewmodel.GumroadSubscriberView {
	view := &viewmodel.GumroadSubscriberView{
		CreatedAt: s.GumroadCreatedAt,
	}
	if s.GumroadEndedAt.Valid {
		view.EndedAt = &s.GumroadEndedAt.Time
	}
	return view
}
