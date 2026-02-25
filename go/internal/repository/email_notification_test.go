package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestEmailNotificationRepository_Create はメール通知設定を正常に作成できることをテスト
func TestEmailNotificationRepository_Create(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewEmailNotificationRepository(queries)

	userID := createUserWithoutRelations(t, queries, "emailnotif-create@example.com", "emailnotifcreateuser")

	unsubscriptionKey := "test-unsub-key-123"
	err := repo.Create(context.Background(), userID, unsubscriptionKey)
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	// 作成されたメール通知設定を確認
	notification, err := queries.GetEmailNotificationByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("メール通知設定の取得に失敗: %v", err)
	}

	if notification.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", notification.UserID, userID)
	}
	if notification.UnsubscriptionKey != unsubscriptionKey {
		t.Errorf("UnsubscriptionKeyが一致しません: got %v, want %v", notification.UnsubscriptionKey, unsubscriptionKey)
	}
}

// TestEmailNotificationRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestEmailNotificationRepository_WithTx(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db)
	repo := repository.NewEmailNotificationRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "emailnotif-withtx@example.com", "emailnotifwithtxuser")

	unsubscriptionKey := "withtx-unsub-key-456"
	err := repoWithTx.Create(context.Background(), userID, unsubscriptionKey)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	// トランザクション内で取得できることを確認
	notification, err := queriesWithTx.GetEmailNotificationByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("トランザクション内でメール通知設定の取得に失敗: %v", err)
	}

	if notification.UnsubscriptionKey != unsubscriptionKey {
		t.Errorf("UnsubscriptionKeyが一致しません: got %v, want %v", notification.UnsubscriptionKey, unsubscriptionKey)
	}
}
