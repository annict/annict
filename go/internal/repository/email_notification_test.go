package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestEmailNotificationRepository_Create はメール通知設定を正常に作成し、Modelとして返却されることをテスト
func TestEmailNotificationRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewEmailNotificationRepository(queries)

	userID := createUserWithoutRelations(t, queries, "emailnotif-create@example.com", "emailnotifcreateuser")

	unsubscriptionKey := "test-unsub-key-123"
	notification, err := repo.Create(context.Background(), userID, unsubscriptionKey)
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if notification == nil {
		t.Fatal("notificationがnilです")
	}
	if notification.ID == 0 {
		t.Error("EmailNotificationIDがゼロ値です")
	}
	if notification.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", notification.UserID, userID)
	}
	if notification.UnsubscriptionKey != unsubscriptionKey {
		t.Errorf("UnsubscriptionKeyが一致しません: got %v, want %v", notification.UnsubscriptionKey, unsubscriptionKey)
	}
	// CreateクエリでINSERT時に明示的に true を指定している
	if !notification.EventFollowedUser {
		t.Error("EventFollowedUserがtrueではありません")
	}
	// CreateクエリでINSERT時に明示的に true を指定している
	if !notification.EventLikedEpisodeRecord {
		t.Error("EventLikedEpisodeRecordがtrueではありません")
	}
	if notification.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
	if notification.UpdatedAt.IsZero() {
		t.Error("UpdatedAtがゼロ値です")
	}
}

// TestEmailNotificationRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestEmailNotificationRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewEmailNotificationRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "emailnotif-withtx@example.com", "emailnotifwithtxuser")

	unsubscriptionKey := "withtx-unsub-key-456"
	notification, err := repoWithTx.Create(context.Background(), userID, unsubscriptionKey)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if notification.UnsubscriptionKey != unsubscriptionKey {
		t.Errorf("UnsubscriptionKeyが一致しません: got %v, want %v", notification.UnsubscriptionKey, unsubscriptionKey)
	}
}
