package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGumroadSubscriberRepository_GetByID はIDでGumroadサブスクライバーを取得できることをテスト
func TestGumroadSubscriberRepository_GetByID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewGumroadSubscriberRepository(queries)

	// テストデータを作成
	subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).
		WithGumroadID("gum_test_getbyid").
		Build()

	// IDで取得
	subscriber, err := repo.GetByID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("Gumroadサブスクライバーの取得に失敗: %v", err)
	}

	if subscriber.ID != subscriberID {
		t.Errorf("IDが一致しません: got %d, want %d", subscriber.ID, subscriberID)
	}
	if subscriber.GumroadID != "gum_test_getbyid" {
		t.Errorf("GumroadIDが一致しません: got %s, want %s", subscriber.GumroadID, "gum_test_getbyid")
	}
}

// TestGumroadSubscriberRepository_GetByID_NotFound は存在しないIDの場合エラーが返ることをテスト
func TestGumroadSubscriberRepository_GetByID_NotFound(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewGumroadSubscriberRepository(queries)

	_, err := repo.GetByID(context.Background(), 99999)
	if err == nil {
		t.Error("存在しないIDでエラーが返されるべきです")
	}
}

// TestGumroadSubscriberRepository_IsActive はアクティブ判定が正しく動作することをテスト
func TestGumroadSubscriberRepository_IsActive(t *testing.T) {
	repo := repository.NewGumroadSubscriberRepository(nil)
	now := time.Now()
	past := now.AddDate(0, 0, -1)
	future := now.AddDate(0, 0, 1)

	testCases := []struct {
		name       string
		subscriber *query.GumroadSubscriber
		expected   bool
	}{
		{
			name: "キャンセルなし、終了なしはアクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{},
				GumroadEndedAt:     sql.NullTime{},
			},
			expected: true,
		},
		{
			name: "キャンセル日時が未来はアクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{Time: future, Valid: true},
				GumroadEndedAt:     sql.NullTime{},
			},
			expected: true,
		},
		{
			name: "終了日時が未来はアクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{},
				GumroadEndedAt:     sql.NullTime{Time: future, Valid: true},
			},
			expected: true,
		},
		{
			name: "キャンセル日時が過去は非アクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{Time: past, Valid: true},
				GumroadEndedAt:     sql.NullTime{},
			},
			expected: false,
		},
		{
			name: "終了日時が過去は非アクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{},
				GumroadEndedAt:     sql.NullTime{Time: past, Valid: true},
			},
			expected: false,
		},
		{
			name: "キャンセル日時が過去、終了日時が未来は非アクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{Time: past, Valid: true},
				GumroadEndedAt:     sql.NullTime{Time: future, Valid: true},
			},
			expected: false,
		},
		{
			name: "キャンセル日時が未来、終了日時が過去は非アクティブ",
			subscriber: &query.GumroadSubscriber{
				GumroadCancelledAt: sql.NullTime{Time: future, Valid: true},
				GumroadEndedAt:     sql.NullTime{Time: past, Valid: true},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := repo.IsActive(tc.subscriber)
			if result != tc.expected {
				t.Errorf("IsActive() = %v, want %v", result, tc.expected)
			}
		})
	}
}
