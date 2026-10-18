package testutil

import (
	"database/sql"
	"testing"
	"time"
)

// PersonVersionは人物の現在のupdated_atを、確認フォームで送る書式にする。
func PersonVersion(t *testing.T, tx *sql.Tx, id int64) string {
	t.Helper()
	var updatedAt time.Time
	if err := tx.QueryRow(`SELECT updated_at FROM people WHERE id = $1`, id).Scan(&updatedAt); err != nil {
		t.Fatalf("人物の版の取得に失敗: %v", err)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano)
}
