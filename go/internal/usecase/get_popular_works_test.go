package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetPopularWorksUsecase_Execute(t *testing.T) {
	t.Run("正常系: 人気作品を取得できる", func(t *testing.T) {
		db, tx := testutil.SetupTestDB(t)
		queries := query.New(db).WithTx(tx)

		// テストデータを作成
		testutil.NewWorkBuilder(t, tx).
			WithTitle("人気アニメ1").
			WithSeason(2024, testutil.SeasonSpring).
			Build()

		testutil.NewWorkBuilder(t, tx).
			WithTitle("人気アニメ2").
			Build()

		workRepo := repository.NewWorkRepository(queries)
		uc := NewGetPopularWorksUsecase(workRepo)

		result, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if len(result.Works) != 2 {
			t.Errorf("expected 2 works, got %d", len(result.Works))
		}
	})

	t.Run("正常系: 作品がない場合は空のスライスを返す", func(t *testing.T) {
		db, tx := testutil.SetupTestDB(t)
		queries := query.New(db).WithTx(tx)

		workRepo := repository.NewWorkRepository(queries)
		uc := NewGetPopularWorksUsecase(workRepo)

		result, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if len(result.Works) != 0 {
			t.Errorf("expected 0 works, got %d", len(result.Works))
		}
	})
}
