package seed

import (
	"context"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestGetRandomEpisodesForKnownTotal_NoProgressは、件数取得とロック付きSELECTの間にある
// READ COMMITTEDの競合窓を検証する。空の一時episodesテーブルで、正の件数を得た後に全行が
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
		t.Fatalf("一時episodesテーブルの作成に失敗: %v", err)
	}

	_, err := getRandomEpisodesForKnownTotal(context.Background(), tx, 1, 1)
	if !errors.Is(err, errNoEpisodesAvailable) {
		t.Fatalf("エラー = %v、期待値 = errNoEpisodesAvailable", err)
	}
}
