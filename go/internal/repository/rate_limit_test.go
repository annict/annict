package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestRateLimitRepository_Incrementは同じキーと窓の組では件数が積み上がり、
// 窓やキーが違えば別に数えられることをテスト
func TestRateLimitRepository_Increment(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewRateLimitRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	key := testutil.UniqueRateLimitPrefix(t)
	windowStart := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	for want := 1; want <= 3; want++ {
		got, err := repo.Increment(ctx, key, windowStart)
		if err != nil {
			t.Fatalf("Incrementに失敗: %v", err)
		}
		if got != want {
			t.Errorf("%d回目の件数 = %d、期待値 = %d", want, got, want)
		}
	}

	got, err := repo.Increment(ctx, key, windowStart.Add(time.Minute))
	if err != nil {
		t.Fatalf("次の窓のIncrementに失敗: %v", err)
	}
	if got != 1 {
		t.Errorf("次の窓の件数 = %d、期待値 = 1", got)
	}

	got, err = repo.Increment(ctx, key+":other", windowStart)
	if err != nil {
		t.Fatalf("別のキーのIncrementに失敗: %v", err)
	}
	if got != 1 {
		t.Errorf("別のキーの件数 = %d、期待値 = 1", got)
	}
}

// TestRateLimitRepository_DeleteExpiredは指定日時より前に始まった窓だけを削除することをテスト
func TestRateLimitRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewRateLimitRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	key := testutil.UniqueRateLimitPrefix(t)
	cutoff := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	expired := cutoff.Add(-time.Minute)

	for _, windowStart := range []time.Time{expired, cutoff} {
		if _, err := repo.Increment(ctx, key, windowStart); err != nil {
			t.Fatalf("Incrementに失敗: %v", err)
		}
	}

	if err := repo.DeleteExpired(ctx, cutoff); err != nil {
		t.Fatalf("DeleteExpiredに失敗: %v", err)
	}

	if exists := rateLimitExists(t, tx, key, expired); exists {
		t.Error("cutoffより前に始まった窓が削除されていない")
	}
	if exists := rateLimitExists(t, tx, key, cutoff); !exists {
		t.Error("cutoffちょうどに始まった窓が削除された")
	}
}

func rateLimitExists(t *testing.T, tx *sql.Tx, key string, windowStart time.Time) bool {
	t.Helper()

	var exists bool
	err := tx.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM rate_limits WHERE key = $1 AND window_start = $2)",
		key, windowStart,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("rate_limitsの確認に失敗: %v", err)
	}
	return exists
}
