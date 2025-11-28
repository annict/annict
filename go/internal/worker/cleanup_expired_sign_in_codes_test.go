package worker

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
	"github.com/riverqueue/river"
	"golang.org/x/crypto/bcrypt"
)

func TestCleanupExpiredSignInCodesWorker(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("cleanup_sign_in_test@example.com").
		WithUsername("cleanup_sign_in_test_user").
		Build()

	ctx := context.Background()

	// テストコードを生成
	code1 := "123456"
	code1Digest, _ := bcrypt.GenerateFromPassword([]byte(code1), bcrypt.DefaultCost)

	code2 := "234567"
	code2Digest, _ := bcrypt.GenerateFromPassword([]byte(code2), bcrypt.DefaultCost)

	code3 := "345678"
	code3Digest, _ := bcrypt.GenerateFromPassword([]byte(code3), bcrypt.DefaultCost)

	// コード1: 48時間前に期限切れ（削除対象）
	_, err := queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: string(code1Digest),
		ExpiresAt:  time.Now().Add(-48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("コード1の作成に失敗: %v", err)
	}

	// コード2: 30時間前に使用済み（削除対象）
	code2Row, err := queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: string(code2Digest),
		ExpiresAt:  time.Now().Add(1 * time.Hour), // 有効期限は未来
	})
	if err != nil {
		t.Fatalf("コード2の作成に失敗: %v", err)
	}
	// コードを使用済みにする（30時間前）
	_, err = tx.ExecContext(ctx, "UPDATE sign_in_codes SET used_at = $1 WHERE id = $2",
		time.Now().Add(-30*time.Hour), code2Row.ID)
	if err != nil {
		t.Fatalf("コード2の使用済み設定に失敗: %v", err)
	}

	// コード3: 有効なコード（削除対象外）
	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: string(code3Digest),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("コード3の作成に失敗: %v", err)
	}

	// ワーカーを作成
	worker := NewCleanupExpiredSignInCodesWorker(queries)

	// ジョブを実行
	job := &river.Job[CleanupExpiredSignInCodesArgs]{
		Args: CleanupExpiredSignInCodesArgs{},
	}

	err = worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// コード1とコード2が削除されていることを確認
	// コード3は有効なので残っているはず
	var count int64
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sign_in_codes WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("コード数の取得に失敗: %v", err)
	}

	if count != 1 {
		t.Errorf("コード数が正しくありません: got %d, want 1", count)
	}

	// 残っているコードがコード3であることを確認
	var remainingCodeDigest string
	err = tx.QueryRowContext(ctx,
		"SELECT code_digest FROM sign_in_codes WHERE user_id = $1",
		userID).Scan(&remainingCodeDigest)
	if err != nil {
		t.Fatalf("コードの取得に失敗: %v", err)
	}

	if remainingCodeDigest != string(code3Digest) {
		t.Errorf("残っているコードが正しくありません")
	}
}

func TestCleanupExpiredSignInCodesWorker_NoCodes(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	ctx := context.Background()

	// ワーカーを作成
	worker := NewCleanupExpiredSignInCodesWorker(queries)

	// ジョブを実行（コードが存在しない状態）
	job := &river.Job[CleanupExpiredSignInCodesArgs]{
		Args: CleanupExpiredSignInCodesArgs{},
	}

	err := worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// エラーなく完了すればOK
}

func TestCleanupExpiredSignInCodesWorker_RecentlyExpired(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("cleanup_recently_expired_sign_in@example.com").
		WithUsername("cleanup_recently_expired_sign_in_user").
		Build()

	ctx := context.Background()

	// テストコードを生成
	code1 := "123456"
	code1Digest, _ := bcrypt.GenerateFromPassword([]byte(code1), bcrypt.DefaultCost)

	code2 := "234567"
	code2Digest, _ := bcrypt.GenerateFromPassword([]byte(code2), bcrypt.DefaultCost)

	// コード1: 12時間前に期限切れ（削除対象外: 24時間以内）
	_, err := queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: string(code1Digest),
		ExpiresAt:  time.Now().Add(-12 * time.Hour),
	})
	if err != nil {
		t.Fatalf("コード1の作成に失敗: %v", err)
	}

	// コード2: 30時間前に期限切れ（削除対象）
	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: string(code2Digest),
		ExpiresAt:  time.Now().Add(-30 * time.Hour),
	})
	if err != nil {
		t.Fatalf("コード2の作成に失敗: %v", err)
	}

	// ワーカーを作成
	worker := NewCleanupExpiredSignInCodesWorker(queries)

	// ジョブを実行
	job := &river.Job[CleanupExpiredSignInCodesArgs]{
		Args: CleanupExpiredSignInCodesArgs{},
	}

	err = worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// コード1のみが残っているはず
	var count int64
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sign_in_codes WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("コード数の取得に失敗: %v", err)
	}

	if count != 1 {
		t.Errorf("コード数が正しくありません: got %d, want 1", count)
	}

	// 残っているコードがコード1であることを確認
	var remainingCodeDigest string
	err = tx.QueryRowContext(ctx,
		"SELECT code_digest FROM sign_in_codes WHERE user_id = $1",
		userID).Scan(&remainingCodeDigest)
	if err != nil {
		t.Fatalf("コードの取得に失敗: %v", err)
	}

	if remainingCodeDigest != string(code1Digest) {
		t.Errorf("残っているコードが正しくありません")
	}
}
