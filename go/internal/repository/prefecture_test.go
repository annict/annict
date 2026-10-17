package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestPrefectureRepository_ListAllは、出身地をIDの順に返すことを検証する。
func TestPrefectureRepository_ListAll(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPrefectureRepository(query.New(db).WithTx(tx))

	firstID := testutil.CreateTestPrefecture(t, tx, "北海道")
	secondID := testutil.CreateTestPrefecture(t, tx, "青森県")

	prefectures, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAllでエラー: %v", err)
	}

	indexes := make(map[model.PrefectureID]int, len(prefectures))
	for i, prefecture := range prefectures {
		indexes[prefecture.ID] = i
	}
	firstIndex, firstOK := indexes[firstID]
	secondIndex, secondOK := indexes[secondID]
	if !firstOK || !secondOK {
		t.Fatalf("作成した出身地が含まれていません (北海道: %v, 青森県: %v)", firstOK, secondOK)
	}
	if firstIndex > secondIndex {
		t.Errorf("出身地の順 = 北海道 %d番目 / 青森県 %d番目、期待値 = IDの順", firstIndex, secondIndex)
	}
	if prefectures[firstIndex].Name != "北海道" {
		t.Errorf("Name = %q、期待値 = 北海道", prefectures[firstIndex].Name)
	}
}

// TestPrefectureRepository_ExistsByIDは、登録済みの出身地だけを存在するとみなすことを
// 検証する。
func TestPrefectureRepository_ExistsByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPrefectureRepository(query.New(db).WithTx(tx))

	id := testutil.CreateTestPrefecture(t, tx, "沖縄県")

	tests := []struct {
		name string
		id   model.PrefectureID
		want bool
	}{
		{name: "登録済みの出身地", id: id, want: true},
		{name: "登録されていない出身地", id: id + 1000000, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ExistsByID(context.Background(), tt.id)
			if err != nil {
				t.Fatalf("ExistsByIDでエラー: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExistsByID(%d) = %v、期待値 = %v", tt.id, got, tt.want)
			}
		})
	}
}
