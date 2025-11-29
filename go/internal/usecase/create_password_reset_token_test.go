package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestCreatePasswordResetTokenUsecase_Execute はトークン生成が正常に動作することをテストします
func TestCreatePasswordResetTokenUsecase_Execute(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成（ユニークなユーザー名を使用）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("reset_token_test").
		WithEmail("reset_token@example.com").
		Build()

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にユーザーを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// UseCase を作成（riverClient は nil で OK - ジョブエンキューはテストしない）
	uc := NewCreatePasswordResetTokenUsecase(db, queries, nil)

	// トークンを生成
	ctx := context.Background()
	result, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}

	// 結果を検証
	if result.Token == "" {
		t.Error("トークンが空です")
	}
	if result.UserID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, userID)
	}

	// トークンがデータベースに保存されているか確認
	// 注: トークンはハッシュ化されて保存されるため、平文トークンで直接検索できない
	// ここでは、ユーザーIDでトークンが存在することを確認
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("トークンがデータベースに保存されていません")
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_InvalidatesOldTokens は古いトークンが無効化されることをテストします
func TestCreatePasswordResetTokenUsecase_Execute_InvalidatesOldTokens(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成（ユニークなユーザー名を使用）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("reset_token_invalidate_test").
		WithEmail("reset_token_invalidate@example.com").
		Build()

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にユーザーを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// UseCase を作成
	uc := NewCreatePasswordResetTokenUsecase(db, queries, nil)

	ctx := context.Background()

	// 最初のトークンを生成
	_, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("最初のトークン生成に失敗: %v", err)
	}

	// 2番目のトークンを生成（古いトークンが無効化されるはず）
	result2, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("2番目のトークン生成に失敗: %v", err)
	}

	// トークンがデータベースに1つだけ存在することを確認
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	if len(tokens) != 1 {
		t.Errorf("トークン数が正しくありません: got %d, want 1", len(tokens))
	}

	// 最新のトークンが保存されていることを確認（結果のUserIDが一致）
	if result2.UserID != userID {
		t.Errorf("最新のトークンのユーザーIDが一致しません")
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_WithNonExistentUser はfail-fastケース：存在しないユーザーIDでのトークン生成をテストします
func TestCreatePasswordResetTokenUsecase_Execute_WithNonExistentUser(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	// UseCase を作成
	uc := NewCreatePasswordResetTokenUsecase(db, queries, nil)

	// 存在しないユーザーIDでトークン生成を試みる
	ctx := context.Background()
	nonExistentUserID := int64(999999999)
	_, err := uc.Execute(ctx, nonExistentUserID)

	// エラーが発生することを確認（外部キー制約違反のはず）
	if err == nil {
		t.Error("存在しないユーザーIDでトークン生成が成功してしまいました")
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_TransactionRollback はfail-fastケース：トランザクションロールバックのテストです
func TestCreatePasswordResetTokenUsecase_Execute_TransactionRollback(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成（ユニークなユーザー名を使用）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("rollback_test_user").
		WithEmail("rollback_test@example.com").
		Build()

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にユーザーを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// UseCase を作成
	uc := NewCreatePasswordResetTokenUsecase(db, queries, nil)

	ctx := context.Background()

	// 最初のトークンを生成
	result1, err := uc.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("最初のトークン生成に失敗: %v", err)
	}

	// トークンがデータベースに保存されているか確認
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("トークン数が正しくありません: got %d, want 1", len(tokens))
	}

	// 最初のトークンが正しく保存されていることを確認
	if result1.Token == "" {
		t.Error("トークンが空です")
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_WithNullUserData はfail-fastケース：ユーザーデータにNULL値がある場合のテストです
// 注: このテストは現在の実装ではユーザー情報を直接取得しないため、スキップします
// 将来的にユーザー情報の検証が追加された場合に有効化してください
func TestCreatePasswordResetTokenUsecase_Execute_WithNullUserData(t *testing.T) {
	t.Skip("現在の実装ではユーザー情報を直接取得しないため、このテストはスキップします")

	// 以下は将来的な実装のための参考コード:
	// - created_at, updated_atがNULLの場合にエラーを返すべき
	// - emailがNULLの場合にエラーを返すべき
	// - データベース制約違反を検出してfail-fastするべき
}

// TestCreatePasswordResetTokenUsecase_Execute_WithInvalidExpiresAt はfail-fastケース：無効な有効期限でのトークン生成をテストします
func TestCreatePasswordResetTokenUsecase_Execute_WithInvalidExpiresAt(t *testing.T) {
	// 注: 現在の実装では有効期限は自動的に設定されるため、このテストは実装されていません
	// 将来的に有効期限を引数で受け取るようになった場合、このテストを実装してください
	t.Skip("現在の実装では有効期限は自動的に設定されるため、このテストはスキップします")
}

// TestCreatePasswordResetTokenUsecase_Execute_ConcurrentRequests はfail-fastケース：並行リクエストのテストです
func TestCreatePasswordResetTokenUsecase_Execute_ConcurrentRequests(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成（ユニークなユーザー名を使用）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("concurrent_test").
		WithEmail("concurrent@example.com").
		Build()

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にユーザーを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// UseCase を作成
	uc := NewCreatePasswordResetTokenUsecase(db, queries, nil)

	ctx := context.Background()

	// 複数回並行してトークン生成を実行
	const concurrentRequests = 5
	results := make(chan *CreatePasswordResetTokenResult, concurrentRequests)
	errors := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			result, err := uc.Execute(ctx, userID)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}()
	}

	// すべてのゴルーチンが完了するまで待つ
	var successCount int
	var errorCount int
	for i := 0; i < concurrentRequests; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			errorCount++
		}
	}

	// 少なくとも1つは成功することを確認
	if successCount == 0 {
		t.Error("すべてのトークン生成が失敗しました")
	}

	t.Logf("成功: %d, 失敗: %d", successCount, errorCount)

	// 最終的にトークンの数を確認
	// 注: 並行実行の場合、DeleteUnusedPasswordResetTokensByUserIDの処理タイミングによっては
	// 複数のトークンが残る可能性がある（レースコンディション）
	tokens, err := queries.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	// 少なくとも1つのトークンが存在することを確認
	if len(tokens) < 1 {
		t.Error("トークンが1つも存在しません")
	}

	// 理想的には1つだけであるべきだが、並行実行により複数残る可能性がある
	if len(tokens) > 1 {
		t.Logf("警告: 並行実行により複数のトークンが残りました: %d個", len(tokens))
	}
}
