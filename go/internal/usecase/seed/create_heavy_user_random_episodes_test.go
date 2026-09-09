package seed

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestGetRandomEpisodesForKnownTotal_NoProgress covers the READ COMMITTED window between the
// count and the locking select. An empty temporary episodes table represents all rows being
// deleted after a positive count; the selector must return instead of looping forever.
//
// [Ja] TestGetRandomEpisodesForKnownTotal_NoProgress は、件数取得とロック付き SELECT の間にある
// READ COMMITTED の競合窓を検証する。空の一時 episodes テーブルで、正の件数を得た後に全行が
// 削除された状態を表し、抽選処理が無限ループせずに戻ることを確認する。
func TestGetRandomEpisodesForKnownTotal_NoProgress(t *testing.T) {
	t.Parallel()

	_, tx := testutil.SetupTx(t)
	if _, err := tx.ExecContext(context.Background(), `
		CREATE TEMPORARY TABLE episodes (
			id bigint NOT NULL,
			work_id bigint NOT NULL
		) ON COMMIT DROP
	`); err != nil {
		t.Fatalf("一時 episodes テーブルの作成に失敗: %v", err)
	}

	_, err := getRandomEpisodesForKnownTotal(context.Background(), tx, 1, 1)
	if !errors.Is(err, errNoEpisodesAvailable) {
		t.Fatalf("error = %v, want errNoEpisodesAvailable", err)
	}
}
