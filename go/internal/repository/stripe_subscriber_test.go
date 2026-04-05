package repository_test

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

// TestStripeSubscriberRepository_Create は新しいStripeサブスクライバーを作成できることをテスト
func TestStripeSubscriberRepository_Create(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	now := time.Now()
	params := query.CreateStripeSubscriberParams{
		StripeCustomerID:         "cus_test_create",
		StripeSubscriptionID:     "sub_test_create",
		StripePriceID:            "price_monthly",
		StripeStatus:             "active",
		StripeCurrentPeriodStart: now,
		StripeCurrentPeriodEnd:   now.AddDate(0, 1, 0),
		StripeCancelAt:           sql.NullTime{},
		StripeCanceledAt:         sql.NullTime{},
	}

	subscriber, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの作成に失敗: %v", err)
	}

	if subscriber.ID == 0 {
		t.Error("IDが設定されていません")
	}
	if subscriber.StripeCustomerID != params.StripeCustomerID {
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, params.StripeCustomerID)
	}
	if subscriber.StripeStatus != "active" {
		t.Errorf("StripeStatusが一致しません: got %s, want %s", subscriber.StripeStatus, "active")
	}
}

// TestStripeSubscriberRepository_GetByID はIDでStripeサブスクライバーを取得できることをテスト
func TestStripeSubscriberRepository_GetByID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_test_getbyid").
		WithStripeStatus("active").
		Build()

	// IDで取得
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.ID != subscriberID {
		t.Errorf("IDが一致しません: got %d, want %d", subscriber.ID, subscriberID)
	}
	if subscriber.StripeCustomerID != "cus_test_getbyid" {
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, "cus_test_getbyid")
	}
}

// TestStripeSubscriberRepository_GetByID_NotFound は存在しないIDの場合エラーが返ることをテスト
func TestStripeSubscriberRepository_GetByID_NotFound(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	_, err := repo.GetByID(context.Background(), 99999)
	if err == nil {
		t.Error("存在しないIDでエラーが返されるべきです")
	}
}

// TestStripeSubscriberRepository_GetByStripeCustomerID はStripe顧客IDで取得できることをテスト
func TestStripeSubscriberRepository_GetByStripeCustomerID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_unique_customer").
		Build()

	// Stripe顧客IDで取得
	subscriber, err := repo.GetByStripeCustomerID(context.Background(), "cus_unique_customer")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.StripeCustomerID != "cus_unique_customer" {
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, "cus_unique_customer")
	}
}

// TestStripeSubscriberRepository_GetByStripeSubscriptionID はStripeサブスクリプションIDで取得できることをテスト
func TestStripeSubscriberRepository_GetByStripeSubscriptionID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeSubscriptionID("sub_unique_subscription").
		Build()

	// StripeサブスクリプションIDで取得
	subscriber, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_unique_subscription")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.StripeSubscriptionID != "sub_unique_subscription" {
		t.Errorf("StripeSubscriptionIDが一致しません: got %s, want %s", subscriber.StripeSubscriptionID, "sub_unique_subscription")
	}
}

// TestStripeSubscriberRepository_Update はサブスクライバー情報を更新できることをテスト
func TestStripeSubscriberRepository_Update(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		WithStripePriceID("price_monthly").
		Build()

	// 更新
	now := time.Now()
	newPeriodEnd := now.AddDate(1, 0, 0)
	err := repo.Update(context.Background(), query.UpdateStripeSubscriberParams{
		ID:                       subscriberID,
		StripePriceID:            "price_yearly",
		StripeStatus:             "active",
		StripeCurrentPeriodStart: now,
		StripeCurrentPeriodEnd:   newPeriodEnd,
		StripeCancelAt:           sql.NullTime{},
		StripeCanceledAt:         sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの更新に失敗: %v", err)
	}

	// 更新後のデータを確認
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.StripePriceID != "price_yearly" {
		t.Errorf("StripePriceIDが更新されていません: got %s, want %s", subscriber.StripePriceID, "price_yearly")
	}
}

// TestStripeSubscriberRepository_UpdateStatus はステータスのみを更新できることをテスト
func TestStripeSubscriberRepository_UpdateStatus(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	// ステータスを更新
	err := repo.UpdateStatus(context.Background(), query.UpdateStripeSubscriberStatusParams{
		ID:           subscriberID,
		StripeStatus: "canceled",
	})
	if err != nil {
		t.Fatalf("Stripeサブスクライバーのステータス更新に失敗: %v", err)
	}

	// 更新後のデータを確認
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.StripeStatus != "canceled" {
		t.Errorf("StripeStatusが更新されていません: got %s, want %s", subscriber.StripeStatus, "canceled")
	}
}

// TestStripeSubscriberRepository_IsActive はアクティブ判定が正しく動作することをテスト
func TestStripeSubscriberRepository_IsActive(t *testing.T) {
	repo := repository.NewStripeSubscriberRepository(nil)

	testCases := []struct {
		name     string
		status   model.StripeSubscriptionStatus
		expected bool
	}{
		{
			name:     "active状態はアクティブ",
			status:   model.StripeSubscriptionStatusActive,
			expected: true,
		},
		{
			name:     "past_due状態はアクティブ（猶予期間）",
			status:   model.StripeSubscriptionStatusPastDue,
			expected: true,
		},
		{
			name:     "canceled状態は非アクティブ",
			status:   model.StripeSubscriptionStatusCanceled,
			expected: false,
		},
		{
			name:     "unpaid状態は非アクティブ",
			status:   model.StripeSubscriptionStatusUnpaid,
			expected: false,
		},
		{
			name:     "trialing状態は非アクティブ",
			status:   model.StripeSubscriptionStatusTrialing,
			expected: false,
		},
		{
			name:     "incomplete状態は非アクティブ",
			status:   model.StripeSubscriptionStatusIncomplete,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subscriber := &model.StripeSubscriber{
				StripeStatus: tc.status.String(),
			}
			result := repo.IsActive(subscriber)
			if result != tc.expected {
				t.Errorf("IsActive() = %v, want %v (status: %s)", result, tc.expected, tc.status)
			}
		})
	}
}
