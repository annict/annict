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

// 新しいStripeSubscriberを作成できることをテストする。
func TestStripeSubscriberRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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
		t.Errorf("StripeCustomerID = %s、期待値 = %s", subscriber.StripeCustomerID, params.StripeCustomerID)
	}
	if subscriber.StripeStatus != "active" {
		t.Errorf("StripeStatus = %s、期待値 = %s", subscriber.StripeStatus, "active")
	}
}

// IDでStripeSubscriberを取得できることをテストする。
func TestStripeSubscriberRepository_GetByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_test_getbyid").
		WithStripeStatus("active").
		Build()

	// IDで取得する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.ID != subscriberID {
		t.Errorf("ID = %d、期待値 = %d", subscriber.ID, subscriberID)
	}
	if subscriber.StripeCustomerID != "cus_test_getbyid" {
		t.Errorf("StripeCustomerID = %s、期待値 = %s", subscriber.StripeCustomerID, "cus_test_getbyid")
	}
}

// 存在しないIDの場合に (nil, nil) が返ることをテストする。
func TestStripeSubscriberRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	subscriber, err := repo.GetByID(context.Background(), 99999)
	if err != nil {
		t.Fatalf("未存在時はエラーではなく (nil, nil) が期待されます: %v", err)
	}
	if subscriber != nil {
		t.Errorf("未存在時はnilが期待されますが、値が返されました: %+v", subscriber)
	}
}

// Stripe顧客IDでStripeSubscriberを取得できることをテストする。
func TestStripeSubscriberRepository_GetByStripeCustomerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成する。
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_unique_customer").
		Build()

	// Stripe顧客IDで取得する。
	subscriber, err := repo.GetByStripeCustomerID(context.Background(), "cus_unique_customer")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeCustomerID != "cus_unique_customer" {
		t.Errorf("StripeCustomerID = %s、期待値 = %s", subscriber.StripeCustomerID, "cus_unique_customer")
	}
}

// 存在しない顧客IDの場合に (nil, nil) が返ることをテストする。
func TestStripeSubscriberRepository_GetByStripeCustomerID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	subscriber, err := repo.GetByStripeCustomerID(context.Background(), "cus_nonexistent")
	if err != nil {
		t.Fatalf("未存在時はエラーではなく (nil, nil) が期待されます: %v", err)
	}
	if subscriber != nil {
		t.Errorf("未存在時はnilが期待されますが、値が返されました: %+v", subscriber)
	}
}

// StripeサブスクリプションIDでStripeSubscriberを取得できることをテストする。
func TestStripeSubscriberRepository_GetByStripeSubscriptionID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成する。
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeSubscriptionID("sub_unique_subscription").
		Build()

	// StripeサブスクリプションIDで取得する。
	subscriber, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_unique_subscription")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeSubscriptionID != "sub_unique_subscription" {
		t.Errorf("StripeSubscriptionID = %s、期待値 = %s", subscriber.StripeSubscriptionID, "sub_unique_subscription")
	}
}

// 存在しないサブスクリプションIDの場合に (nil, nil) が返ることをテストする。
func TestStripeSubscriberRepository_GetByStripeSubscriptionID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	subscriber, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_nonexistent")
	if err != nil {
		t.Fatalf("未存在時はエラーではなく (nil, nil) が期待されます: %v", err)
	}
	if subscriber != nil {
		t.Errorf("未存在時はnilが期待されますが、値が返されました: %+v", subscriber)
	}
}

// サブスクライバー情報を更新できることをテストする。
func TestStripeSubscriberRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		WithStripePriceID("price_monthly").
		Build()

	// 更新する。
	now := time.Now()
	newPeriodEnd := now.AddDate(1, 0, 0)
	err := repo.Update(context.Background(), query.UpdateStripeSubscriberParams{
		ID:                       int64(subscriberID),
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

	// 更新後のデータを確認する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripePriceID != "price_yearly" {
		t.Errorf("StripePriceID = %s、期待値 = %s", subscriber.StripePriceID, "price_yearly")
	}
}

// ステータスのみを更新できることをテストする。
func TestStripeSubscriberRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	// ステータスを更新する。
	err := repo.UpdateStatus(context.Background(), query.UpdateStripeSubscriberStatusParams{
		ID:           int64(subscriberID),
		StripeStatus: "canceled",
	})
	if err != nil {
		t.Fatalf("Stripeサブスクライバーのステータス更新に失敗: %v", err)
	}

	// 更新後のデータを確認する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeStatus != "canceled" {
		t.Errorf("StripeStatus = %s、期待値 = %s", subscriber.StripeStatus, "canceled")
	}
}

// アクティブ判定が正しく動作することをテストする。
func TestStripeSubscriberRepository_IsActive(t *testing.T) {
	t.Parallel()

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
			name:     "past_due状態はアクティブ (猶予期間)",
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
				t.Errorf("IsActive() = %v、期待値 = %v (status: %s)", result, tc.expected, tc.status)
			}
		})
	}
}
