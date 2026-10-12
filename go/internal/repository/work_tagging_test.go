package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestWorkTaggingRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkTaggingRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	anotherWorkID := testutil.NewWorkBuilder(t, tx).Build()
	workTagID := testutil.CreateTestWorkTag(t, tx)
	anotherWorkTagID := testutil.CreateTestWorkTag(t, tx)
	if _, err := tx.ExecContext(ctx, "UPDATE work_tags SET work_taggings_count = 5 WHERE id IN ($1, $2)", workTagID, anotherWorkTagID); err != nil {
		t.Fatalf("タグの作品数の設定に失敗: %v", err)
	}
	// 同じタグを2つの作品に付けた分は、タグの作品数を2減らす。
	testutil.CreateTestWorkTagging(t, tx, userID, workID, workTagID)
	testutil.CreateTestWorkTagging(t, tx, userID, anotherWorkID, workTagID)
	testutil.CreateTestWorkTagging(t, tx, userID, workID, anotherWorkTagID)
	// 他人のタグ付けは残す。
	otherWorkTaggingID := testutil.CreateTestWorkTagging(t, tx, otherUserID, workID, workTagID)

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{2, 1} {
		got, err := repo.DeleteByUserID(ctx, userID, 2)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "work_taggings", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのタグ付けの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "work_taggings", "id = $1", otherWorkTaggingID); got != 1 {
		t.Errorf("他人のタグ付けの件数 = %d、期待値 = 1", got)
	}
	for _, tt := range []struct {
		id   int64
		want int32
	}{{id: workTagID, want: 3}, {id: anotherWorkTagID, want: 4}} {
		var count int32
		if err := tx.QueryRowContext(ctx, "SELECT work_taggings_count FROM work_tags WHERE id = $1", tt.id).Scan(&count); err != nil {
			t.Fatalf("タグの作品数の取得に失敗: %v", err)
		}
		if count != tt.want {
			t.Errorf("タグ %d の作品数 = %d、期待値 = %d", tt.id, count, tt.want)
		}
	}
}
