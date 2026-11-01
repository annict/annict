package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestUpdateLibraryEntryNextResourcesUsecase_Executeは、実行日 (UTC) の3日前の0時 (UTC) 以降に
// 更新されたライブラリだけを更新し、更新日時に実行時刻を入れることを検証する。ページの区切り方と
// 次のエピソード・次の放送枠の決め方はRepositoryのテストで確かめる。
func TestUpdateLibraryEntryNextResourcesUsecase_Execute(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	// 日本時間の1月10日8時はUTCの1月9日23時のため、下限はUTCの1月6日0時になる。共有DBにほかの
	// テストがコミットした行を対象にしないよう、実行時刻を遠い未来に置く。
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Date(2100, 1, 10, 8, 0, 0, 0, jst)
	updatedFrom := time.Date(2100, 1, 6, 0, 0, 0, 0, time.UTC)

	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	newEntry := func(updatedAt time.Time) int64 {
		userID := testutil.NewUserBuilder(t, tx).Build()
		return testutil.NewLibraryEntryBuilder(t, tx).WithUserID(userID).WithWorkID(workID).WithUpdatedAt(updatedAt).Build()
	}
	atLowerBound := newEntry(updatedFrom)
	beforeLowerBound := newEntry(updatedFrom.Add(-time.Second))

	output, err := NewUpdateLibraryEntryNextResourcesUsecase(repo).Execute(ctx, UpdateLibraryEntryNextResourcesInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.UpdatedCount != 1 {
		t.Errorf("UpdatedCount = %d、期待値 = 1", output.UpdatedCount)
	}

	tests := []struct {
		name              string
		id                int64
		wantNextEpisodeID sql.NullInt64
		wantUpdatedAt     time.Time
	}{
		{name: "下限ちょうどに更新された: 更新する", id: atLowerBound, wantNextEpisodeID: sql.NullInt64{Int64: int64(episodeID), Valid: true}, wantUpdatedAt: now},
		{name: "下限より前に更新された: 更新しない", id: beforeLowerBound, wantUpdatedAt: updatedFrom.Add(-time.Second)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nextEpisodeID sql.NullInt64
			var updatedAt time.Time
			if err := tx.QueryRow("SELECT next_episode_id, updated_at FROM library_entries WHERE id = $1", tt.id).Scan(&nextEpisodeID, &updatedAt); err != nil {
				t.Fatalf("ライブラリエントリの取得に失敗: %v", err)
			}
			if nextEpisodeID != tt.wantNextEpisodeID || !updatedAt.Equal(tt.wantUpdatedAt) {
				t.Errorf("(次のエピソード, 更新日時) = (%v, %v)、期待値 = (%v, %v)", nextEpisodeID, updatedAt, tt.wantNextEpisodeID, tt.wantUpdatedAt)
			}
		})
	}
}

// TestUpdateLibraryEntryNextResourcesUsecase_Execute_StopsWhenContextEndsは、ctxが終わっていたら
// ライブラリを更新せず、エラーにしないで途中でやめたことを返すことを検証する。
func TestUpdateLibraryEntryNextResourcesUsecase_Execute_StopsWhenContextEnds(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(tx)
	now := time.Date(2100, 1, 10, 0, 0, 0, 0, time.UTC)

	workID := testutil.NewWorkBuilder(t, tx).Build()
	testutil.NewEpisodeBuilder(t, tx, workID).Build()
	userID := testutil.NewUserBuilder(t, tx).Build()
	entryID := testutil.NewLibraryEntryBuilder(t, tx).WithUserID(userID).WithWorkID(workID).WithUpdatedAt(now).Build()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	output, err := NewUpdateLibraryEntryNextResourcesUsecase(repo).Execute(ctx, UpdateLibraryEntryNextResourcesInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !output.Interrupted || output.UpdatedCount != 0 {
		t.Errorf("output = %+v、期待値 = 更新0件で途中でやめた", output)
	}

	var nextEpisodeID sql.NullInt64
	if err := tx.QueryRow("SELECT next_episode_id FROM library_entries WHERE id = $1", entryID).Scan(&nextEpisodeID); err != nil {
		t.Fatalf("ライブラリエントリの取得に失敗: %v", err)
	}
	if nextEpisodeID.Valid {
		t.Errorf("次のエピソード = %v、期待値 = NULLのまま", nextEpisodeID)
	}
}
