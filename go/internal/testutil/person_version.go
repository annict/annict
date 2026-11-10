package testutil

import (
	"database/sql"
	"testing"
	"time"
)

// rowQuerierは1行を読むクエリを実行できるもの (*sql.Txと *sql.DB)。
type rowQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// PersonVersionは人物の現在のupdated_atを、確認フォームで送る書式にする。テスト用のトランザクションでも、
// 共有テストDBでも読める。
func PersonVersion(t *testing.T, q rowQuerier, id int64) string {
	t.Helper()
	var updatedAt time.Time
	if err := q.QueryRow(`SELECT updated_at FROM people WHERE id = $1`, id).Scan(&updatedAt); err != nil {
		t.Fatalf("人物の版の取得に失敗: %v", err)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano)
}
