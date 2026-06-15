package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// randomString はテスト用のランダムな文字列を生成します
func randomString(n int) string {
	id := uuid.New().String()
	if n > len(id) {
		n = len(id)
	}
	return id[:n]
}

func TestDeleteStripeSubscriberUsecase_Execute(t *testing.T) {
	t.Parallel()

	// テストDBを取得（トランザクションを使わない統合テスト）
	db := testutil.GetTestDB()
	queries := query.New(db)

	// リポジトリの作成
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// Usecaseの作成
	uc := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	t.Run("サブスクリプションを削除できる", func(t *testing.T) {
		t.Parallel()

		// 一意のIDを生成
		subscriptionID := "sub_test_delete_" + randomString(8)

		// テスト用のStripeSubscriberを作成（コミットされる）
		subscriber, err := stripeSubscriberRepo.Create(context.Background(), query.CreateStripeSubscriberParams{
			StripeCustomerID:         "cus_test_" + randomString(8),
			StripeSubscriptionID:     subscriptionID,
			StripePriceID:            "price_test",
			StripeStatus:             "active",
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}

		// クリーンアップ
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", subscriber.ID)
		})

		ctx := context.Background()
		canceledAt := time.Now()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: subscriptionID,
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
		t.Parallel()

		// 一意のIDを生成
		subscriptionID := "sub_test_delete_user_" + randomString(8)
		customerID := "cus_test_" + randomString(8)
		email := "test_delete_" + randomString(8) + "@example.com"
		username := "testdel" + randomString(6)

		// テスト用のStripeSubscriberを作成
		subscriber, err := stripeSubscriberRepo.Create(context.Background(), query.CreateStripeSubscriberParams{
			StripeCustomerID:         customerID,
			StripeSubscriptionID:     subscriptionID,
			StripePriceID:            "price_test",
			StripeStatus:             "active",
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("StripeSubscriberの作成に失敗: %v", err)
		}

		// テスト用のユーザーを作成してStripeSubscriberと紐付け
		var userID int64
		err = db.QueryRow(`
			INSERT INTO users (
				username, email, role, locale,
				created_at, updated_at,
				encrypted_password, sign_in_count,
				time_zone, allowed_locales, stripe_subscriber_id
			) VALUES (
				$1, $2, 0, 'ja',
				NOW(), NOW(),
				'encrypted_test_password', 0,
				'Asia/Tokyo', ARRAY[]::varchar[], $3
			) RETURNING id
		`, username, email, subscriber.ID).Scan(&userID)
		if err != nil {
			t.Fatalf("ユーザーの作成に失敗: %v", err)
		}

		// クリーンアップ
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", subscriber.ID)
		})

		ctx := context.Background()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: subscriptionID,
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
		if int64(*result.UserID) != userID {
			t.Errorf("UserID: got %d, want %d", *result.UserID, userID)
		}

		// ユーザーの紐付けが解除されていることを確認
		updatedUser, err := userRepo.GetByID(ctx, model.UserID(userID))
		if err != nil {
			t.Errorf("ユーザー取得エラー: %v", err)
			return
		}
		if updatedUser.StripeSubscriberID.Valid {
			t.Errorf("StripeSubscriberID: 紐付けが解除されていません (値: %d)", updatedUser.StripeSubscriberID.Int64)
		}
	})

	t.Run("存在しないサブスクリプションIDはErrStripeSubscriberNotFound", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		input := DeleteStripeSubscriberInput{
			StripeSubscriptionID: "sub_nonexistent_" + randomString(8),
			StripeCanceledAt:     time.Now(),
		}

		_, err := uc.Execute(ctx, input)
		if err == nil {
			t.Fatal("エラーが期待されましたが、nilが返されました")
		}
		if !errors.Is(err, ErrStripeSubscriberNotFound) {
			t.Errorf("ErrStripeSubscriberNotFound が期待されましたが、別のエラーが返されました: %v", err)
		}
	})
}
