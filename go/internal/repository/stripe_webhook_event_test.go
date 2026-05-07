package repository_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestStripeWebhookEventRepository_Create は新しいWebhookイベントを作成できることをテスト
func TestStripeWebhookEventRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	now := time.Now()
	params := repository.CreateStripeWebhookEventParams{
		StripeEventID:   "evt_test_create",
		StripeEventType: "customer.subscription.created",
		StripePayload:   json.RawMessage(`{"id": "evt_test_create", "type": "customer.subscription.created"}`),
		Status:          model.WebhookEventStatusPending.String(),
		ReceivedAt:      now,
	}

	event, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Webhookイベントの作成に失敗: %v", err)
	}

	if event.ID == 0 {
		t.Error("IDが設定されていません")
	}
	if event.StripeEventID != params.StripeEventID {
		t.Errorf("StripeEventIDが一致しません: got %s, want %s", event.StripeEventID, params.StripeEventID)
	}
	if event.StripeEventType != params.StripeEventType {
		t.Errorf("StripeEventTypeが一致しません: got %s, want %s", event.StripeEventType, params.StripeEventType)
	}
	if event.Status != model.WebhookEventStatusPending.String() {
		t.Errorf("Statusが一致しません: got %s, want %s", event.Status, model.WebhookEventStatusPending.String())
	}
}

// TestStripeWebhookEventRepository_GetByStripeEventID はStripeイベントIDで取得できることをテスト
func TestStripeWebhookEventRepository_GetByStripeEventID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_unique_event").
		Build()

	// StripeイベントIDで取得
	event, err := repo.GetByStripeEventID(context.Background(), "evt_unique_event")
	if err != nil {
		t.Fatalf("Webhookイベントの取得に失敗: %v", err)
	}

	if event.StripeEventID != "evt_unique_event" {
		t.Errorf("StripeEventIDが一致しません: got %s, want %s", event.StripeEventID, "evt_unique_event")
	}
}

// TestStripeWebhookEventRepository_GetByStripeEventID_NotFound は存在しないイベントIDの場合エラーが返ることをテスト
func TestStripeWebhookEventRepository_GetByStripeEventID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	_, err := repo.GetByStripeEventID(context.Background(), "evt_nonexistent")
	if err == nil {
		t.Error("存在しないイベントIDでエラーが返されるべきです")
	}
	if err != sql.ErrNoRows {
		t.Errorf("予期しないエラー: got %v, want %v", err, sql.ErrNoRows)
	}
}

// TestStripeWebhookEventRepository_UpdateStatus はステータスを更新できることをテスト
func TestStripeWebhookEventRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	eventID := testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_update_status").
		WithStatus(model.WebhookEventStatusPending.String()).
		Build()

	// ステータスを更新
	now := time.Now()
	err := repo.UpdateStatus(context.Background(), query.UpdateStripeWebhookEventStatusParams{
		ID:           int64(eventID),
		Status:       model.WebhookEventStatusProcessed.String(),
		ErrorMessage: sql.NullString{},
		ProcessedAt:  sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("ステータス更新に失敗: %v", err)
	}

	// 更新後のデータを確認
	event, err := repo.GetByStripeEventID(context.Background(), "evt_update_status")
	if err != nil {
		t.Fatalf("更新後のイベント取得に失敗: %v", err)
	}

	if event.Status != model.WebhookEventStatusProcessed.String() {
		t.Errorf("Statusが更新されていません: got %s, want %s", event.Status, model.WebhookEventStatusProcessed.String())
	}
	if !event.ProcessedAt.Valid {
		t.Error("ProcessedAtが設定されていません")
	}
}

// TestStripeWebhookEventRepository_MarkAsProcessed は処理完了としてマークできることをテスト
func TestStripeWebhookEventRepository_MarkAsProcessed(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	eventID := testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_mark_processed").
		Build()

	// 処理完了としてマーク
	err := repo.MarkAsProcessed(context.Background(), eventID)
	if err != nil {
		t.Fatalf("MarkAsProcessedに失敗: %v", err)
	}

	// 更新後のデータを確認
	event, err := repo.GetByStripeEventID(context.Background(), "evt_mark_processed")
	if err != nil {
		t.Fatalf("更新後のイベント取得に失敗: %v", err)
	}

	if event.Status != model.WebhookEventStatusProcessed.String() {
		t.Errorf("Statusが更新されていません: got %s, want %s", event.Status, model.WebhookEventStatusProcessed.String())
	}
	if !event.ProcessedAt.Valid {
		t.Error("ProcessedAtが設定されていません")
	}
}

// TestStripeWebhookEventRepository_MarkAsFailed は処理失敗としてマークできることをテスト
func TestStripeWebhookEventRepository_MarkAsFailed(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	eventID := testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_mark_failed").
		Build()

	// 処理失敗としてマーク
	errorMsg := "処理中にエラーが発生しました"
	err := repo.MarkAsFailed(context.Background(), eventID, errorMsg)
	if err != nil {
		t.Fatalf("MarkAsFailedに失敗: %v", err)
	}

	// 更新後のデータを確認
	event, err := repo.GetByStripeEventID(context.Background(), "evt_mark_failed")
	if err != nil {
		t.Fatalf("更新後のイベント取得に失敗: %v", err)
	}

	if event.Status != model.WebhookEventStatusFailed.String() {
		t.Errorf("Statusが更新されていません: got %s, want %s", event.Status, model.WebhookEventStatusFailed.String())
	}
	if !event.ErrorMessage.Valid {
		t.Error("ErrorMessageが設定されていません")
	}
	if event.ErrorMessage.String != errorMsg {
		t.Errorf("ErrorMessageが一致しません: got %s, want %s", event.ErrorMessage.String, errorMsg)
	}
	if !event.ProcessedAt.Valid {
		t.Error("ProcessedAtが設定されていません")
	}
}

// TestStripeWebhookEventRepository_MarkAsSkipped は処理スキップとしてマークできることをテスト
func TestStripeWebhookEventRepository_MarkAsSkipped(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	eventID := testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_mark_skipped").
		Build()

	// 処理スキップとしてマーク
	err := repo.MarkAsSkipped(context.Background(), eventID)
	if err != nil {
		t.Fatalf("MarkAsSkippedに失敗: %v", err)
	}

	// 更新後のデータを確認
	event, err := repo.GetByStripeEventID(context.Background(), "evt_mark_skipped")
	if err != nil {
		t.Fatalf("更新後のイベント取得に失敗: %v", err)
	}

	if event.Status != model.WebhookEventStatusSkipped.String() {
		t.Errorf("Statusが更新されていません: got %s, want %s", event.Status, model.WebhookEventStatusSkipped.String())
	}
	if !event.ProcessedAt.Valid {
		t.Error("ProcessedAtが設定されていません")
	}
}

// TestStripeWebhookEventRepository_Exists はイベントの存在確認ができることをテスト
func TestStripeWebhookEventRepository_Exists(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewStripeWebhookEventRepository(queries)

	// テストデータを作成
	testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID("evt_exists_check").
		Build()

	testCases := []struct {
		name          string
		stripeEventID string
		expected      bool
	}{
		{
			name:          "存在するイベントID",
			stripeEventID: "evt_exists_check",
			expected:      true,
		},
		{
			name:          "存在しないイベントID",
			stripeEventID: "evt_nonexistent",
			expected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := repo.Exists(context.Background(), tc.stripeEventID)
			if err != nil {
				t.Fatalf("Exists()でエラーが発生: %v", err)
			}
			if exists != tc.expected {
				t.Errorf("Exists() = %v, want %v", exists, tc.expected)
			}
		})
	}
}
