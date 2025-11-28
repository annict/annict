package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

func TestSendSignInCodeUsecase_Execute(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// 別のトランザクションでテストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("sign_in_code_test_user_1").
		WithEmail("sign_in_code_test_1@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成（Riverクライアントはnil）
	uc := NewSendSignInCodeUsecase(db, queries, nil)

	ctx := context.Background()

	// Execute を実行
	result, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 結果の検証
	if result.UserID != userID {
		t.Errorf("UserID: got %d, want %d", result.UserID, userID)
	}

	// コードが6桁の数字であることを確認
	if len(result.Code) != 6 {
		t.Errorf("Code length: got %d, want 6", len(result.Code))
	}
	for _, c := range result.Code {
		if c < '0' || c > '9' {
			t.Errorf("Code contains non-digit character: %c", c)
		}
	}

	// データベースに保存されたコードを取得
	savedCode, err := queries.GetValidSignInCode(ctx, userID)
	if err != nil {
		t.Fatalf("GetValidSignInCode failed: %v", err)
	}

	// コードダイジェストが bcrypt でハッシュ化されていることを確認
	if !auth.VerifyCode(result.Code, savedCode.CodeDigest) {
		t.Error("Code verification failed")
	}

	// 有効期限が 15 分後に設定されていることを確認（許容誤差: 10秒）
	expectedExpiry := time.Now().Add(15 * time.Minute)
	timeDiff := savedCode.ExpiresAt.Sub(expectedExpiry)
	if timeDiff < -10*time.Second || timeDiff > 10*time.Second {
		t.Errorf("ExpiresAt: got %v, want around %v (diff: %v)", savedCode.ExpiresAt, expectedExpiry, timeDiff)
	}

	// 試行回数が0であることを確認
	if savedCode.Attempts != 0 {
		t.Errorf("Attempts: got %d, want 0", savedCode.Attempts)
	}

	// 使用済みフラグがnilであることを確認
	if savedCode.UsedAt.Valid {
		t.Error("UsedAt should be NULL")
	}
}

func TestSendSignInCodeUsecase_Execute_InvalidatesOldCodes(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// 別のトランザクションでテストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("sign_in_code_test_user_2").
		WithEmail("sign_in_code_test_2@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	uc := NewSendSignInCodeUsecase(db, queries, nil)

	ctx := context.Background()

	// 1回目のコード生成
	_, err = uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	// 2回目のコード生成（古いコードは無効化されるはず）
	result2, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("Second Execute failed: %v", err)
	}

	// 最新のコードのみが有効であることを確認
	savedCode, err := queries.GetValidSignInCode(ctx, userID)
	if err != nil {
		t.Fatalf("GetValidSignInCode failed: %v", err)
	}

	// 2回目のコードが保存されていることを確認
	if !auth.VerifyCode(result2.Code, savedCode.CodeDigest) {
		t.Error("Second code verification failed")
	}
}

func TestSendSignInCodeUsecase_Execute_NonExistentUser(t *testing.T) {
	t.Parallel()

	// テスト用DBとトランザクションをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// ユースケースを作成
	uc := NewSendSignInCodeUsecase(db, queries, nil)

	ctx := context.Background()

	// 存在しないユーザーIDでExecuteを実行
	_, err := uc.Execute(ctx, 999999)
	if err == nil {
		t.Error("Execute should fail for non-existent user")
	}
}
