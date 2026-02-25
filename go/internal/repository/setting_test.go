package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSettingRepository_Create は設定を正常に作成できることをテスト
func TestSettingRepository_Create(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSettingRepository(queries)

	userID := createUserWithoutRelations(t, queries, "setting-create@example.com", "settingcreateuser")

	err := repo.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	// 作成された設定を確認
	setting, err := queries.GetSettingByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("設定の取得に失敗: %v", err)
	}

	if setting.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", setting.UserID, userID)
	}
}

// TestSettingRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestSettingRepository_WithTx(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db)
	repo := repository.NewSettingRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "setting-withtx@example.com", "settingwithtxuser")

	err := repoWithTx.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	// トランザクション内で取得できることを確認
	setting, err := queriesWithTx.GetSettingByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("トランザクション内で設定の取得に失敗: %v", err)
	}

	if setting.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", setting.UserID, userID)
	}
}
