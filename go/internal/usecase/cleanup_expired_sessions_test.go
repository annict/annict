package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// insertSessionsはsession_idの接頭辞とupdated_atを共有するセッション行をcount件
// 挿入する。1文でまとめて挿入するのは、クリーンアップのバッチサイズを超える件数を必要と
// するテストがあり、1行ずつの挿入では件数が多すぎるため。
func insertSessions(t *testing.T, tx *sql.Tx, prefix string, updatedAt time.Time, count int) {
	t.Helper()

	_, err := tx.Exec(
		`INSERT INTO sessions (session_id, data, created_at, updated_at)
		 SELECT $1 || i::text, '{}'::jsonb, $2, $2 FROM generate_series(1, $3) AS i`,
		prefix, updatedAt, count,
	)
	if err != nil {
		t.Fatalf("セッションの作成に失敗: %v", err)
	}
}

// countSessionsはsession_idが指定した接頭辞を持つセッション行の件数を返す。
func countSessions(t *testing.T, tx *sql.Tx, prefix string) int {
	t.Helper()

	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM sessions WHERE session_id LIKE $1 || '%'`, prefix).Scan(&count); err != nil {
		t.Fatalf("セッションの件数取得に失敗: %v", err)
	}

	return count
}

func TestCleanupExpiredSessionsUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: カットオフより古いセッションだけが削除される", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupExpiredSessionsUsecase(repository.NewSessionRepository(queries))

		cutoff := time.Now().Add(-model.SessionMaxAge)
		insertSessions(t, tx, "cleanup-sessions-old-", cutoff.Add(-time.Hour), 3)
		insertSessions(t, tx, "cleanup-sessions-fresh-", cutoff.Add(time.Hour), 2)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		if got := countSessions(t, tx, "cleanup-sessions-old-"); got != 0 {
			t.Errorf("期限切れセッションの件数 = %d、期待値 = 0", got)
		}
		if got := countSessions(t, tx, "cleanup-sessions-fresh-"); got != 2 {
			t.Errorf("有効なセッションの件数 = %d、期待値 = 2", got)
		}
	})

	t.Run("正常系: バッチサイズを超える件数でもすべて削除される", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupExpiredSessionsUsecase(repository.NewSessionRepository(queries))

		cutoff := time.Now().Add(-model.SessionMaxAge)
		insertSessions(t, tx, "cleanup-sessions-batch-", cutoff.Add(-time.Hour), cleanupExpiredSessionsBatchSize+1)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		if got := countSessions(t, tx, "cleanup-sessions-batch-"); got != 0 {
			t.Errorf("期限切れセッションの件数 = %d、期待値 = 0", got)
		}
	})

	t.Run("異常系: 削除エラーを原因付きで返す", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupExpiredSessionsUsecase(repository.NewSessionRepository(queries))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := uc.Execute(ctx)
		if err == nil {
			t.Fatal("Execute()のエラー = nil、期待値 = エラーあり")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Execute()のエラー = %v、期待値 = context.Canceled", err)
		}
	})
}
