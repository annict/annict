package worker

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/password_reset"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
	"github.com/riverqueue/river"
)

func TestCleanupExpiredTokensWorker(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("cleanup_test@example.com").
		WithUsername("cleanup_test_user").
		Build()

	ctx := context.Background()

	// テストトークンを生成
	token1, _ := password_reset.GenerateToken()
	token1Digest := password_reset.HashToken(token1)

	token2, _ := password_reset.GenerateToken()
	token2Digest := password_reset.HashToken(token2)

	token3, _ := password_reset.GenerateToken()
	token3Digest := password_reset.HashToken(token3)

	// トークン1: 48時間前に期限切れ（削除対象）
	_, err := queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: token1Digest,
		ExpiresAt:   time.Now().Add(-48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークン1の作成に失敗: %v", err)
	}

	// トークン2: 30時間前に使用済み（削除対象）
	token2Row, err := queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: token2Digest,
		ExpiresAt:   time.Now().Add(1 * time.Hour), // 有効期限は未来
	})
	if err != nil {
		t.Fatalf("トークン2の作成に失敗: %v", err)
	}
	// トークンを使用済みにする（30時間前）
	_, err = tx.ExecContext(ctx, "UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2",
		time.Now().Add(-30*time.Hour), token2Row.ID)
	if err != nil {
		t.Fatalf("トークン2の使用済み設定に失敗: %v", err)
	}

	// トークン3: 有効なトークン（削除対象外）
	_, err = queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: token3Digest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークン3の作成に失敗: %v", err)
	}

	// ワーカーを作成
	worker := NewCleanupExpiredTokensWorker(queries)

	// ジョブを実行
	job := &river.Job[CleanupExpiredTokensArgs]{
		Args: CleanupExpiredTokensArgs{},
	}

	err = worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// トークン1とトークン2が削除されていることを確認
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	// トークン3のみが残っているはず
	if len(tokens) != 1 {
		t.Errorf("トークン数が正しくありません: got %d, want 1", len(tokens))
	}

	if len(tokens) > 0 && tokens[0].TokenDigest != token3Digest {
		t.Errorf("残っているトークンが正しくありません: got %s, want %s", tokens[0].TokenDigest, token3Digest)
	}
}

func TestCleanupExpiredTokensWorker_NoTokens(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	ctx := context.Background()

	// ワーカーを作成
	worker := NewCleanupExpiredTokensWorker(queries)

	// ジョブを実行（トークンが存在しない状態）
	job := &river.Job[CleanupExpiredTokensArgs]{
		Args: CleanupExpiredTokensArgs{},
	}

	err := worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// エラーなく完了すればOK
}

func TestCleanupExpiredTokensWorker_RecentlyExpired(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("cleanup_recently_expired@example.com").
		WithUsername("cleanup_recently_expired_user").
		Build()

	ctx := context.Background()

	// テストトークンを生成
	token1, _ := password_reset.GenerateToken()
	token1Digest := password_reset.HashToken(token1)

	token2, _ := password_reset.GenerateToken()
	token2Digest := password_reset.HashToken(token2)

	// トークン1: 12時間前に期限切れ（削除対象外: 24時間以内）
	_, err := queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: token1Digest,
		ExpiresAt:   time.Now().Add(-12 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークン1の作成に失敗: %v", err)
	}

	// トークン2: 30時間前に期限切れ（削除対象）
	_, err = queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: token2Digest,
		ExpiresAt:   time.Now().Add(-30 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークン2の作成に失敗: %v", err)
	}

	// ワーカーを作成
	worker := NewCleanupExpiredTokensWorker(queries)

	// ジョブを実行
	job := &river.Job[CleanupExpiredTokensArgs]{
		Args: CleanupExpiredTokensArgs{},
	}

	err = worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	// トークン1のみが残っているはず
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	if len(tokens) != 1 {
		t.Errorf("トークン数が正しくありません: got %d, want 1", len(tokens))
	}

	if len(tokens) > 0 && tokens[0].TokenDigest != token1Digest {
		t.Errorf("残っているトークンが正しくありません: got %s, want %s", tokens[0].TokenDigest, token1Digest)
	}
}
