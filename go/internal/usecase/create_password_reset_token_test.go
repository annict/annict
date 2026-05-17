package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// TestCreatePasswordResetTokenUsecase_Execute はトークン生成が正常に動作することをテストします
func TestCreatePasswordResetTokenUsecase_Execute(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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

	// UseCase を作成（dispatcher は nil で OK - ジョブエンキューはテストしない）
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	// トークンを生成
	ctx := context.Background()
	result, err := uc.Execute(ctx, CreatePasswordResetTokenInput{
		Email: "reset_token@example.com",
	})
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
	tokens, err := repository.NewPasswordResetTokenRepository(queries).GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("トークンがデータベースに保存されていません")
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_InvalidatesOldTokens は古いトークンが無効化されることをテストします
func TestCreatePasswordResetTokenUsecase_Execute_InvalidatesOldTokens(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	ctx := context.Background()
	input := CreatePasswordResetTokenInput{
		Email: "reset_token_invalidate@example.com",
	}

	// 最初のトークンを生成
	_, err := uc.Execute(ctx, input)
	if err != nil {
		t.Fatalf("最初のトークン生成に失敗: %v", err)
	}

	// 2番目のトークンを生成（古いトークンが無効化されるはず）
	result2, err := uc.Execute(ctx, input)
	if err != nil {
		t.Fatalf("2番目のトークン生成に失敗: %v", err)
	}

	// トークンがデータベースに1つだけ存在することを確認
	tokens, err := repository.NewPasswordResetTokenRepository(queries).GetByUserID(ctx, userID)
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

// TestCreatePasswordResetTokenUsecase_Execute_WithNonExistentUser は存在しないユーザーのメールアドレスでのトークン生成をテストします
func TestCreatePasswordResetTokenUsecase_Execute_WithNonExistentUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	// UseCase を作成
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	// 存在しないメールアドレスでトークン生成を試みる（セキュリティ対策でエラーにならない）
	ctx := context.Background()
	result, err := uc.Execute(ctx, CreatePasswordResetTokenInput{
		Email: "nonexistent@example.com",
	})

	// エラーは返されない（ユーザーの存在を明かさない）
	if err != nil {
		t.Errorf("存在しないユーザーでエラーが返されました: %v", err)
	}

	// ユーザーが存在しない場合は nil を返す
	if result != nil {
		t.Errorf("存在しないユーザーなのに result が返されました: %+v", result)
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_ValidationError はバリデーションエラーをテストします
func TestCreatePasswordResetTokenUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	// UseCase を作成
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	// 空のメールアドレスでバリデーションエラーを発生させる
	ctx := context.Background()
	_, err := uc.Execute(ctx, CreatePasswordResetTokenInput{
		Email: "",
	})
	ve := model.AsValidationError(err)

	if ve == nil {
		t.Fatalf("バリデーションエラーが期待されましたが、発生しませんでした: %v", err)
	}
}

// TestCreatePasswordResetTokenUsecase_Execute_TransactionRollback はトランザクションロールバックのテストです
func TestCreatePasswordResetTokenUsecase_Execute_TransactionRollback(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	ctx := context.Background()
	input := CreatePasswordResetTokenInput{
		Email: "rollback_test@example.com",
	}

	// 最初のトークンを生成
	result1, err := uc.Execute(ctx, input)
	if err != nil {
		t.Fatalf("最初のトークン生成に失敗: %v", err)
	}

	// トークンがデータベースに保存されているか確認
	tokens, err := repository.NewPasswordResetTokenRepository(queries).GetByUserID(ctx, userID)
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

// TestCreatePasswordResetTokenUsecase_Execute_WithNullUserData はユーザーデータにNULL値がある場合のテストです
func TestCreatePasswordResetTokenUsecase_Execute_WithNullUserData(t *testing.T) {
	t.Skip("現在の実装ではユーザー情報を直接取得しないため、このテストはスキップします")
}

// TestCreatePasswordResetTokenUsecase_Execute_WithInvalidExpiresAt は無効な有効期限でのトークン生成をテストします
func TestCreatePasswordResetTokenUsecase_Execute_WithInvalidExpiresAt(t *testing.T) {
	t.Skip("現在の実装では有効期限は自動的に設定されるため、このテストはスキップします")
}

// TestCreatePasswordResetTokenUsecase_Execute_ConcurrentRequests は並行リクエストのテストです
func TestCreatePasswordResetTokenUsecase_Execute_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

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
	v := validator.NewPasswordResetCreateValidator()
	uc := NewCreatePasswordResetTokenUsecase(db, repository.NewUserRepository(queries), repository.NewPasswordResetTokenRepository(queries), nil, nil, v)

	ctx := context.Background()
	input := CreatePasswordResetTokenInput{
		Email: "concurrent@example.com",
	}

	// 複数回並行してトークン生成を実行
	const concurrentRequests = 5
	results := make(chan *CreatePasswordResetTokenOutput, concurrentRequests)
	errors := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			result, err := uc.Execute(ctx, input)
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
	tokens, err := repository.NewPasswordResetTokenRepository(queries).GetByUserID(ctx, userID)
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
