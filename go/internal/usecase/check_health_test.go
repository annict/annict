package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestCheckHealthUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: DBが正常な場合はhealthyを返す", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		workRepo := repository.NewWorkRepository(queries)
		uc := NewCheckHealthUsecase(workRepo)

		result, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if !result.DBHealthy {
			t.Errorf("expected DBHealthy to be true, got false (error: %s)", result.DBError)
		}
	})
}
