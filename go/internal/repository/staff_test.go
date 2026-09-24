package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestStaffRepository_GetByWorkIDsは作品IDのリストに紐づくスタッフを取得できることをテスト
func TestStaffRepository_GetByWorkIDs(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 指定した作品のスタッフを取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewStaffRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("作品A").Build()
		testutil.NewStaffBuilder(t, tx, workID).WithName("監督A").WithRole("director").Build()

		staffs, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{workID})
		if err != nil {
			t.Fatalf("GetByWorkIDs()のエラー = %v", err)
		}

		if len(staffs) != 1 {
			t.Fatalf("len(staffs) = %d、期待値 = 1", len(staffs))
		}
		if staffs[0].WorkID != workID {
			t.Errorf("WorkID = %v、期待値 = %v", staffs[0].WorkID, workID)
		}
		if staffs[0].Name != "監督A" {
			t.Errorf("Name = %q、期待値 = %q", staffs[0].Name, "監督A")
		}
		if staffs[0].Role != "director" {
			t.Errorf("Role = %q、期待値 = %q", staffs[0].Role, "director")
		}
		if staffs[0].ID == 0 {
			t.Error("StaffIDがゼロ値です")
		}
	})

	t.Run("正常系: role='other' のスタッフは除外される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewStaffRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("作品B").Build()
		testutil.NewStaffBuilder(t, tx, workID).WithName("監督B").WithRole("director").Build()
		testutil.NewStaffBuilder(t, tx, workID).WithName("その他担当").WithRole("other").Build()

		staffs, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{workID})
		if err != nil {
			t.Fatalf("GetByWorkIDs()のエラー = %v", err)
		}

		if len(staffs) != 1 {
			t.Fatalf("len(staffs) = %d、期待値 = 1", len(staffs))
		}
		if staffs[0].Role != "director" {
			t.Errorf("Role = %q、期待値 = %q (otherは除外されるべき)", staffs[0].Role, "director")
		}
	})

	t.Run("正常系: workIDsが空の場合は空のスライスを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewStaffRepository(queries)

		staffs, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{})
		if err != nil {
			t.Fatalf("GetByWorkIDs()のエラー = %v", err)
		}
		if len(staffs) != 0 {
			t.Errorf("len(staffs) = %d、期待値 = 0", len(staffs))
		}
	})

	t.Run("正常系: 該当する作品が存在しない場合は空のスライスを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewStaffRepository(queries)

		staffs, err := repo.GetByWorkIDs(context.Background(), []model.WorkID{999999999})
		if err != nil {
			t.Fatalf("GetByWorkIDs()のエラー = %v", err)
		}
		if len(staffs) != 0 {
			t.Errorf("len(staffs) = %d、期待値 = 0", len(staffs))
		}
	})
}

// TestStaffRepository_WithTxはWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestStaffRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewStaffRepository(queries)

	repoWithTx := repo.WithTx(tx)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("作品C").Build()
	testutil.NewStaffBuilder(t, tx, workID).WithName("脚本C").WithRole("series_composition").Build()

	staffs, err := repoWithTx.GetByWorkIDs(context.Background(), []model.WorkID{workID})
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでGetByWorkIDsに失敗: %v", err)
	}
	if len(staffs) != 1 {
		t.Errorf("len(staffs) = %d、期待値 = 1", len(staffs))
	}
}
