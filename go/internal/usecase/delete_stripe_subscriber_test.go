package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestDeleteStripeSubscriberUsecase_Execute(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// リポジトリの作成
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// Usecaseの作成
	uc := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	t.Run("サブスクリプションを削除できる", func(t *testing.T) {
		// テスト用のStripeSubscriberを作成
		subscriber := testutil.NewStripeSubscriberBuilder(t, tx).
			WithStripeSubscriptionID("sub_test_delete_1").
			WithStripeStatus("active").
			BuildWithResult()

		ctx := context.Background()
		canceledAt := time.Now()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: subscriber.StripeSubscriptionID,
			StripeCanceledAt:     canceledAt,
		}

		result, err := uc.Execute(ctx, input)
		if err != nil {
			t.Errorf("予期しないエラー: %v", err)
			return
		}

		// ステータスがcanceledに更新されていることを確認
		if result.StripeSubscriber.StripeStatus != string(model.StripeSubscriptionStatusCanceled) {
			t.Errorf("StripeStatus: got %s, want %s", result.StripeSubscriber.StripeStatus, model.StripeSubscriptionStatusCanceled)
		}

		// キャンセル日時が設定されていることを確認
		if !result.StripeSubscriber.StripeCanceledAt.Valid {
			t.Error("StripeCanceledAt: Validがfalseですが、trueが期待されました")
		}
	})

	t.Run("ユーザーとの紐付けが解除される", func(t *testing.T) {
		// テスト用のStripeSubscriberを作成
		subscriber := testutil.NewStripeSubscriberBuilder(t, tx).
			WithStripeSubscriptionID("sub_test_delete_user_link").
			WithStripeStatus("active").
			BuildWithResult()

		// テスト用のユーザーを作成してStripeSubscriberと紐付け
		user := testutil.NewUserBuilder(t, tx).
			WithStripeSubscriberID(&subscriber.ID).
			BuildWithResult()

		ctx := context.Background()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: subscriber.StripeSubscriptionID,
			StripeCanceledAt:     time.Now(),
		}

		result, err := uc.Execute(ctx, input)
		if err != nil {
			t.Errorf("予期しないエラー: %v", err)
			return
		}

		// ユーザーIDが返されていることを確認
		if result.UserID == nil {
			t.Error("UserID: nilが返されましたが、ユーザーIDが期待されました")
			return
		}
		if *result.UserID != user.ID {
			t.Errorf("UserID: got %d, want %d", *result.UserID, user.ID)
		}

		// ユーザーの紐付けが解除されていることを確認
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		if err != nil {
			t.Errorf("ユーザー取得エラー: %v", err)
			return
		}
		if updatedUser.StripeSubscriberID.Valid {
			t.Errorf("StripeSubscriberID: 紐付けが解除されていません (値: %d)", updatedUser.StripeSubscriberID.Int64)
		}
	})

	t.Run("存在しないサブスクリプションIDはエラー", func(t *testing.T) {
		ctx := context.Background()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: "sub_nonexistent_delete",
			StripeCanceledAt:     time.Now(),
		}

		_, err := uc.Execute(ctx, input)
		if err == nil {
			t.Error("エラーが期待されましたが、nilが返されました")
		}
	})
}
