package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSettingRepository_Create は設定を正常に作成し、Modelとして返却されることをテスト
func TestSettingRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSettingRepository(queries)

	userID := createUserWithoutRelations(t, queries, "setting-create@example.com", "settingcreateuser")

	setting, err := repo.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if setting == nil {
		t.Fatal("settingがnilです")
	}
	if setting.ID == 0 {
		t.Error("SettingIDがゼロ値です")
	}
	if setting.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", setting.UserID, userID)
	}
	// CreateクエリでINSERT時に明示的に true を指定している
	if !setting.PrivacyPolicyAgreed {
		t.Error("PrivacyPolicyAgreedがtrueではありません")
	}
	// CreateクエリでINSERT時に明示的に true を指定している
	if !setting.HideRecordBody {
		t.Error("HideRecordBodyがtrueではありません")
	}
	if setting.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
	if setting.UpdatedAt.IsZero() {
		t.Error("UpdatedAtがゼロ値です")
	}
}

// TestSettingRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestSettingRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewSettingRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "setting-withtx@example.com", "settingwithtxuser")

	setting, err := repoWithTx.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if setting.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", setting.UserID, userID)
	}
}
