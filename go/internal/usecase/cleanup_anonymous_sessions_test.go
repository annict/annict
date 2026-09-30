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

// anonymousSessionDataは未ログインのセッションが持つ典型的なdata。
const anonymousSessionData = `{"_csrf_token": "token"}`

// loggedInSessionDataはログイン済みのセッションが持つdata。
const loggedInSessionData = `{"warden.user.user.key": [[1], "salt"], "_csrf_token": "token"}`

// insertSessionsWithDataはsession_idの接頭辞、data、updated_atを共有するセッション行を
// count件挿入する。1文でまとめて挿入する理由はinsertSessionsと同じ。
func insertSessionsWithData(t *testing.T, tx *sql.Tx, prefix, data string, updatedAt time.Time, count int) {
	t.Helper()

	_, err := tx.Exec(
		`INSERT INTO sessions (session_id, data, created_at, updated_at)
		 SELECT $1 || i::text, $2::jsonb, $3, $3 FROM generate_series(1, $4) AS i`,
		prefix, data, updatedAt, count,
	)
	if err != nil {
		t.Fatalf("セッションの作成に失敗: %v", err)
	}
}

func TestCleanupAnonymousSessionsUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 最終アクセスから1日を過ぎた未ログインのセッションだけが削除される", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupAnonymousSessionsUsecase(repository.NewSessionRepository(queries), CleanupAnonymousSessionsTimeLimit)

		now := time.Now()
		cutoff := now.Add(-model.AnonymousSessionMaxAge)
		expiredCutoff := now.Add(-model.SessionMaxAge)
		insertSessionsWithData(t, tx, "cleanup-anonymous-old-", anonymousSessionData, cutoff.Add(-time.Hour), 3)
		insertSessionsWithData(t, tx, "cleanup-anonymous-logged-in-", loggedInSessionData, cutoff.Add(-time.Hour), 2)
		insertSessionsWithData(t, tx, "cleanup-anonymous-fresh-", anonymousSessionData, cutoff.Add(time.Hour), 2)
		// 30日より古い行は期限切れセッションのクリーンアップに任せる。
		insertSessionsWithData(t, tx, "cleanup-anonymous-expired-", anonymousSessionData, expiredCutoff.Add(-time.Hour), 2)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		if got := countSessions(t, tx, "cleanup-anonymous-old-"); got != 0 {
			t.Errorf("1日を過ぎた未ログインのセッションの件数 = %d、期待値 = 0", got)
		}
		if got := countSessions(t, tx, "cleanup-anonymous-logged-in-"); got != 2 {
			t.Errorf("ログイン済みのセッションの件数 = %d、期待値 = 2", got)
		}
		if got := countSessions(t, tx, "cleanup-anonymous-fresh-"); got != 2 {
			t.Errorf("1日以内の未ログインのセッションの件数 = %d、期待値 = 2", got)
		}
		if got := countSessions(t, tx, "cleanup-anonymous-expired-"); got != 2 {
			t.Errorf("30日より古いセッションの件数 = %d、期待値 = 2", got)
		}
	})

	t.Run("正常系: バッチサイズを超える件数でもログイン済みの行を読み飛ばしてすべて削除される", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupAnonymousSessionsUsecase(repository.NewSessionRepository(queries), CleanupAnonymousSessionsTimeLimit)

		cutoff := time.Now().Add(-model.AnonymousSessionMaxAge)
		// ログイン済みの行を未ログインの行より古くし、最初のバッチの読み始めに置く。
		insertSessionsWithData(t, tx, "cleanup-anonymous-skip-logged-in-", loggedInSessionData, cutoff.Add(-2*time.Hour), 2)
		insertSessionsWithData(t, tx, "cleanup-anonymous-batch-", anonymousSessionData, cutoff.Add(-time.Hour), cleanupAnonymousSessionsBatchSize+1)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		if got := countSessions(t, tx, "cleanup-anonymous-batch-"); got != 0 {
			t.Errorf("未ログインのセッションの件数 = %d、期待値 = 0", got)
		}
		if got := countSessions(t, tx, "cleanup-anonymous-skip-logged-in-"); got != 2 {
			t.Errorf("ログイン済みのセッションの件数 = %d、期待値 = 2", got)
		}
	})

	t.Run("正常系: 同じupdated_atの行がバッチの境目で分かれても取りこぼさない", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupAnonymousSessionsUsecase(repository.NewSessionRepository(queries), CleanupAnonymousSessionsTimeLimit)

		// 最初のバッチが古い行 (バッチサイズ - 1件) と境目の行の1件目までを削除し、境目の
		// 残りの行を次のバッチに持ち越すよう配置する。
		cutoff := time.Now().Add(-model.AnonymousSessionMaxAge)
		older := cutoff.Add(-2 * time.Hour)
		boundary := cutoff.Add(-time.Hour)
		insertSessionsWithData(t, tx, "cleanup-anonymous-older-", anonymousSessionData, older, cleanupAnonymousSessionsBatchSize-1)
		insertSessionsWithData(t, tx, "cleanup-anonymous-boundary-", anonymousSessionData, boundary, 2)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		if got := countSessions(t, tx, "cleanup-anonymous-older-"); got != 0 {
			t.Errorf("古い未ログインのセッションの件数 = %d、期待値 = 0", got)
		}
		if got := countSessions(t, tx, "cleanup-anonymous-boundary-"); got != 0 {
			t.Errorf("境目の未ログインのセッションの件数 = %d、期待値 = 0", got)
		}
	})

	t.Run("正常系: 上限時間を超えたら次のバッチに入らずに打ち切り、エラーにしない", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		// 最初のバッチを終えた時点で必ず上限を超えるよう、ごく短い上限を渡す。
		uc := NewCleanupAnonymousSessionsUsecase(repository.NewSessionRepository(queries), time.Nanosecond)

		updatedAt := time.Now().Add(-model.AnonymousSessionMaxAge - time.Hour)
		insertSessionsWithData(t, tx, "cleanup-anonymous-limit-", anonymousSessionData, updatedAt, cleanupAnonymousSessionsBatchSize+1)

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Executeに失敗: %v", err)
		}

		// 最初の1バッチだけを削除し、残りは次回に持ち越す。
		if got := countSessions(t, tx, "cleanup-anonymous-limit-"); got != 1 {
			t.Errorf("未ログインのセッションの件数 = %d、期待値 = 1", got)
		}
	})

	t.Run("異常系: 削除エラーを原因付きで返す", func(t *testing.T) {
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewCleanupAnonymousSessionsUsecase(repository.NewSessionRepository(queries), CleanupAnonymousSessionsTimeLimit)

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
