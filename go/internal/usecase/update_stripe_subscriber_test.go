package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestUpdateStripeSubscriberUsecase_Execute(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// リポジトリの作成
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// テスト用のStripeSubscriberを作成
	subscriber := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeSubscriptionID("sub_test_update_123").
		WithStripeStatus("active").
		WithStripePriceID("price_monthly").
		BuildWithResult()

	// Usecaseの作成
	uc := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	tests := []struct {
		name      string
		input     UpdateStripeSubscriberInput
		wantError bool
	}{
		{
			name: "サブスクリプション状態を更新できる",
			input: UpdateStripeSubscriberInput{
				StripeSubscriptionID:     subscriber.StripeSubscriptionID,
				StripePriceID:            "price_yearly",
				StripeStatus:             "active",
				StripeCurrentPeriodStart: time.Now(),
				StripeCurrentPeriodEnd:   time.Now().AddDate(1, 0, 0),
				StripeCancelAt:           sql.NullTime{},
				StripeCanceledAt:         sql.NullTime{},
			},
			wantError: false,
		},
		{
			name: "past_dueステータスに更新できる",
			input: UpdateStripeSubscriberInput{
				StripeSubscriptionID:     subscriber.StripeSubscriptionID,
				StripePriceID:            "price_yearly",
				StripeStatus:             "past_due",
				StripeCurrentPeriodStart: time.Now(),
				StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
				StripeCancelAt:           sql.NullTime{},
				StripeCanceledAt:         sql.NullTime{},
			},
			wantError: false,
		},
		{
			name: "キャンセル予定日を設定できる",
			input: UpdateStripeSubscriberInput{
				StripeSubscriptionID:     subscriber.StripeSubscriptionID,
				StripePriceID:            "price_monthly",
				StripeStatus:             "active",
				StripeCurrentPeriodStart: time.Now(),
				StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
				StripeCancelAt: sql.NullTime{
					Time:  time.Now().AddDate(0, 1, 0),
					Valid: true,
				},
				StripeCanceledAt: sql.NullTime{},
			},
			wantError: false,
		},
		{
			name: "存在しないサブスクリプションIDはエラー",
			input: UpdateStripeSubscriberInput{
				StripeSubscriptionID:     "sub_nonexistent",
				StripePriceID:            "price_monthly",
				StripeStatus:             "active",
				StripeCurrentPeriodStart: time.Now(),
				StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
			},
			wantError: true,
		},
		{
			name: "無効なステータスはエラー",
			input: UpdateStripeSubscriberInput{
				StripeSubscriptionID:     subscriber.StripeSubscriptionID,
				StripePriceID:            "price_monthly",
				StripeStatus:             "invalid_status",
				StripeCurrentPeriodStart: time.Now(),
				StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := uc.Execute(ctx, tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、nilが返されました")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			// 更新後の値を確認
			if result.StripeSubscriber.StripePriceID != tt.input.StripePriceID {
				t.Errorf("StripePriceID: got %s, want %s", result.StripeSubscriber.StripePriceID, tt.input.StripePriceID)
			}
			if result.StripeSubscriber.StripeStatus != tt.input.StripeStatus {
				t.Errorf("StripeStatus: got %s, want %s", result.StripeSubscriber.StripeStatus, tt.input.StripeStatus)
			}
		})
	}
}

func TestUpdateStripeSubscriberUsecase_ExecuteDelete(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// リポジトリの作成
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// Usecaseの作成
	uc := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

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

		result, err := uc.ExecuteDelete(ctx, input)
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

		result, err := uc.ExecuteDelete(ctx, input)
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

		_, err := uc.ExecuteDelete(ctx, input)
		if err == nil {
			t.Error("エラーが期待されましたが、nilが返されました")
		}
	})
}
