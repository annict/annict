package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCastRepository_GetByWorkIDs は作品IDのリストに紐づくキャストを取得できることをテスト
func TestCastRepository_GetByWorkIDs(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 指定した作品のキャストを取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewCastRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("作品A").Build()
		testutil.NewCastBuilder(t, tx, workID).
			WithCharacterName("キャラクター1").
			WithPersonName("声優1").
			Build()

		casts, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{workID})
		if err != nil {
			t.Fatalf("GetByWorkIDs() error = %v", err)
		}

		if len(casts) != 1 {
			t.Fatalf("len(casts) = %d, want 1", len(casts))
		}
		if casts[0].WorkID != workID {
			t.Errorf("WorkID = %v, want %v", casts[0].WorkID, workID)
		}
		if casts[0].CharacterName != "キャラクター1" {
			t.Errorf("CharacterName = %q, want %q", casts[0].CharacterName, "キャラクター1")
		}
		if casts[0].PersonName != "声優1" {
			t.Errorf("PersonName = %q, want %q", casts[0].PersonName, "声優1")
		}
		if casts[0].ID == 0 {
			t.Error("CastID がゼロ値です")
		}
	})

	t.Run("正常系: workIDs が空の場合は空のスライスを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewCastRepository(queries)

		casts, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{})
		if err != nil {
			t.Fatalf("GetByWorkIDs() error = %v", err)
		}
		if len(casts) != 0 {
			t.Errorf("len(casts) = %d, want 0", len(casts))
		}
	})

	t.Run("正常系: 該当する作品が存在しない場合は空のスライスを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewCastRepository(queries)

		casts, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{999999999})
		if err != nil {
			t.Fatalf("GetByWorkIDs() error = %v", err)
		}
		if len(casts) != 0 {
			t.Errorf("len(casts) = %d, want 0", len(casts))
		}
	})
}

// TestCastRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestCastRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewCastRepository(queries)

	repoWithTx := repo.WithTx(tx)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("作品B").Build()
	testutil.NewCastBuilder(t, tx, workID).
		WithCharacterName("キャラクターB").
		WithPersonName("声優B").
		Build()

	casts, err := repoWithTx.GetByWorkIDs(context.Background(), []model.WorkID{workID})
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでGetByWorkIDsに失敗: %v", err)
	}
	if len(casts) != 1 {
		t.Errorf("len(casts) = %d, want 1", len(casts))
	}
}
