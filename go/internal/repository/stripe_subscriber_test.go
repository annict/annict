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

// TestStripeSubscriberRepository_Create verifies that a new StripeSubscriber can be created.
//
// [Ja] 新しい StripeSubscriber を作成できることをテストする。
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
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, params.StripeCustomerID)
	}
	if subscriber.StripeStatus != "active" {
		t.Errorf("StripeStatusが一致しません: got %s, want %s", subscriber.StripeStatus, "active")
	}
}

// TestStripeSubscriberRepository_GetByID verifies that a StripeSubscriber can be fetched by ID.
//
// [Ja] ID で StripeSubscriber を取得できることをテストする。
func TestStripeSubscriberRepository_GetByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// Create test data.
	//
	// [Ja] テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_test_getbyid").
		WithStripeStatus("active").
		Build()

	// Fetch by ID.
	//
	// [Ja] ID で取得する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.ID != subscriberID {
		t.Errorf("IDが一致しません: got %d, want %d", subscriber.ID, subscriberID)
	}
	if subscriber.StripeCustomerID != "cus_test_getbyid" {
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, "cus_test_getbyid")
	}
}

// TestStripeSubscriberRepository_GetByID_NotFound verifies that (nil, nil) is returned for a nonexistent ID.
//
// [Ja] 存在しない ID の場合に (nil, nil) が返ることをテストする。
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
		t.Errorf("未存在時は nil が期待されますが、値が返されました: %+v", subscriber)
	}
}

// TestStripeSubscriberRepository_GetByStripeCustomerID verifies that a StripeSubscriber can be fetched by Stripe customer ID.
//
// [Ja] Stripe 顧客 ID で StripeSubscriber を取得できることをテストする。
func TestStripeSubscriberRepository_GetByStripeCustomerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// Create test data.
	//
	// [Ja] テストデータを作成する。
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeCustomerID("cus_unique_customer").
		Build()

	// Fetch by Stripe customer ID.
	//
	// [Ja] Stripe 顧客 ID で取得する。
	subscriber, err := repo.GetByStripeCustomerID(context.Background(), "cus_unique_customer")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeCustomerID != "cus_unique_customer" {
		t.Errorf("StripeCustomerIDが一致しません: got %s, want %s", subscriber.StripeCustomerID, "cus_unique_customer")
	}
}

// TestStripeSubscriberRepository_GetByStripeCustomerID_NotFound verifies that (nil, nil) is returned for a nonexistent customer ID.
//
// [Ja] 存在しない顧客 ID の場合に (nil, nil) が返ることをテストする。
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
		t.Errorf("未存在時は nil が期待されますが、値が返されました: %+v", subscriber)
	}
}

// TestStripeSubscriberRepository_GetByStripeSubscriptionID verifies that a StripeSubscriber can be fetched by Stripe subscription ID.
//
// [Ja] Stripe サブスクリプション ID で StripeSubscriber を取得できることをテストする。
func TestStripeSubscriberRepository_GetByStripeSubscriptionID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// Create test data.
	//
	// [Ja] テストデータを作成する。
	testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeSubscriptionID("sub_unique_subscription").
		Build()

	// Fetch by Stripe subscription ID.
	//
	// [Ja] Stripe サブスクリプション ID で取得する。
	subscriber, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_unique_subscription")
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeSubscriptionID != "sub_unique_subscription" {
		t.Errorf("StripeSubscriptionIDが一致しません: got %s, want %s", subscriber.StripeSubscriptionID, "sub_unique_subscription")
	}
}

// TestStripeSubscriberRepository_GetByStripeSubscriptionID_NotFound verifies that (nil, nil) is returned for a nonexistent subscription ID.
//
// [Ja] 存在しないサブスクリプション ID の場合に (nil, nil) が返ることをテストする。
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
		t.Errorf("未存在時は nil が期待されますが、値が返されました: %+v", subscriber)
	}
}

// TestStripeSubscriberRepository_Update verifies that subscriber information can be updated.
//
// [Ja] サブスクライバー情報を更新できることをテストする。
func TestStripeSubscriberRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// Create test data.
	//
	// [Ja] テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		WithStripePriceID("price_monthly").
		Build()

	// Update.
	//
	// [Ja] 更新する。
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

	// Verify the data after the update.
	//
	// [Ja] 更新後のデータを確認する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripePriceID != "price_yearly" {
		t.Errorf("StripePriceIDが更新されていません: got %s, want %s", subscriber.StripePriceID, "price_yearly")
	}
}

// TestStripeSubscriberRepository_UpdateStatus verifies that only the status can be updated.
//
// [Ja] ステータスのみを更新できることをテストする。
func TestStripeSubscriberRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeSubscriberRepository(queries)

	// Create test data.
	//
	// [Ja] テストデータを作成する。
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	// Update the status.
	//
	// [Ja] ステータスを更新する。
	err := repo.UpdateStatus(context.Background(), query.UpdateStripeSubscriberStatusParams{
		ID:           int64(subscriberID),
		StripeStatus: "canceled",
	})
	if err != nil {
		t.Fatalf("Stripeサブスクライバーのステータス更新に失敗: %v", err)
	}

	// Verify the data after the update.
	//
	// [Ja] 更新後のデータを確認する。
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("更新後のStripeサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber == nil {
		t.Fatal("サブスクライバーが見つかりませんでした")
	}
	if subscriber.StripeStatus != "canceled" {
		t.Errorf("StripeStatusが更新されていません: got %s, want %s", subscriber.StripeStatus, "canceled")
	}
}

// TestStripeSubscriberRepository_IsActive verifies that the active check works correctly.
//
// [Ja] アクティブ判定が正しく動作することをテストする。
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
