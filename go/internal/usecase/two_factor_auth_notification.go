package usecase

import (
	"context"
	"log/slog"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// enqueueTwoFactorAuthNotificationは、二要素認証の設定の変更を本人に知らせるメールのジョブを投入する。
//
// 設定の変更をコミットしたあとに呼ぶ。送信先とロケールは、その時点のユーザーのメールアドレスと言語にする。
// 投入に失敗しても設定の変更は取り消さずログに残すだけにする (パスワードリセットのメールと同じ)。
// 通知を送れないことよりも、利用者の操作が失敗したように見えることのほうが混乱を招くため。
func enqueueTwoFactorAuthNotification(
	ctx context.Context,
	d *dispatcher.Dispatcher,
	userRepo *repository.UserRepository,
	userID model.UserID,
	event model.TwoFactorAuthNotificationEvent,
) {
	if d == nil {
		slog.WarnContext(ctx, "Dispatcherが設定されていないため、二要素認証の通知メールのジョブを投入できませんでした", "user_id", userID, "event", event)
		return
	}

	user, err := userRepo.GetActiveByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "二要素認証の通知メールの送信先の取得に失敗しました", "error", err, "user_id", userID, "event", event)
		return
	}

	if err := d.EnqueueTwoFactorAuthNotificationEmail(ctx, user.Email, user.Locale, string(event)); err != nil {
		slog.ErrorContext(ctx, "二要素認証の通知メールのジョブの投入に失敗しました", "error", err, "user_id", userID, "event", event)
		return
	}

	slog.InfoContext(ctx, "二要素認証の通知メールのジョブを投入しました", "user_id", userID, "event", event)
}
