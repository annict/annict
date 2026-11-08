package usecase

import (
	"context"
	"math"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetPopularWorksUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: キャスト・スタッフ情報を作品ごとにグルーピングして取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		// 人気作品はDB全体の上位30件のため、並行するほかのテストがコミットした作品に
		// 押し出されないよう、ウォッチャー数を最大にして必ず上位に入れる。
		workID1 := testutil.NewWorkBuilder(t, tx).
			WithTitle("人気アニメ1").
			WithSeason(2024, testutil.SeasonSpring).
			WithWatchersCount(math.MaxInt32).
			Build()
		testutil.NewWorkImageBuilder(t, tx, workID1).Build()
		testutil.NewCastBuilder(t, tx, workID1).WithCharacterName("キャラクター1").WithPersonName("声優1").Build()
		testutil.NewCastBuilder(t, tx, workID1).WithCharacterName("キャラクター2").WithPersonName("声優2").Build()
		testutil.NewStaffBuilder(t, tx, workID1).WithName("監督1").WithRole("director").Build()
		testutil.NewStaffBuilder(t, tx, workID1).WithName("脚本1").WithRole("series_composition").Build()

		workID2 := testutil.NewWorkBuilder(t, tx).
			WithTitle("人気アニメ2").
			WithSeason(2024, testutil.SeasonSummer).
			WithWatchersCount(math.MaxInt32).
			Build()
		testutil.NewWorkImageBuilder(t, tx, workID2).Build()
		testutil.NewCastBuilder(t, tx, workID2).WithCharacterName("キャラクター3").WithPersonName("声優3").Build()

		uc := NewGetPopularWorksUsecase(
			repository.NewWorkRepository(queries),
			repository.NewCastRepository(queries),
			repository.NewStaffRepository(queries),
		)

		result, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if result == nil {
			t.Fatal("resultがnilだった")
		}

		var work1, work2 *model.Work
		for _, w := range result.Works {
			switch w.ID {
			case workID1:
				work1 = w
			case workID2:
				work2 = w
			}
		}

		if work1 == nil {
			t.Fatalf("作品1 (ID=%d) が結果に含まれていません", workID1)
		}
		if len(work1.Casts) != 2 {
			t.Errorf("作品1のキャスト数 = %d、期待値 = 2", len(work1.Casts))
		}
		if len(work1.Staffs) != 2 {
			t.Errorf("作品1のスタッフ数 = %d、期待値 = 2", len(work1.Staffs))
		}
		for _, c := range work1.Casts {
			if c.WorkID != workID1 {
				t.Errorf("作品1にグルーピングされたキャストのWorkID = %v、期待値 = %v", c.WorkID, workID1)
			}
		}

		if work2 == nil {
			t.Fatalf("作品2 (ID=%d) が結果に含まれていません", workID2)
		}
		if len(work2.Casts) != 1 {
			t.Errorf("作品2のキャスト数 = %d、期待値 = 1", len(work2.Casts))
		}
		if len(work2.Staffs) != 0 {
			t.Errorf("作品2のスタッフ数 = %d、期待値 = 0 (キャスト・スタッフがなくても結果に含まれる)", len(work2.Staffs))
		}
	})
}
