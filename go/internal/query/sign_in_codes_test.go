package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateSignInCodeはSignInCodeの作成をテスト
func TestCreateSignInCode(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_create_user").
		WithEmail("test_create@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "test_digest_123",
		ExpiresAt:  expiresAt,
	}

	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// 基本的なアサーション
	if code.UserID != int64(userID) {
		t.Errorf("ユーザーIDの期待値 = %d、実測値 = %d", userID, code.UserID)
	}
	if code.CodeDigest != "test_digest_123" {
		t.Errorf("コードのdigest = %s、期待値 = test_digest_123", code.CodeDigest)
	}
	if code.Attempts != 0 {
		t.Errorf("attempts = %d、期待値 = 0", code.Attempts)
	}
	if code.UsedAt.Valid {
		t.Error("used_atがNULLでなかった")
	}
}

// TestGetValidSignInCodeは有効なSignInCodeの取得をテスト
func TestGetValidSignInCode(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_get_valid_user").
		WithEmail("test_get_valid@example.com").
		Build()

	// 有効なSignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "valid_code_digest",
		ExpiresAt:  expiresAt,
	}
	createdCode, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// 有効なコードを取得
	code, err := queries.GetValidSignInCode(context.Background(), int64(userID))
	if err != nil {
		t.Fatalf("有効なサインインコードの取得エラー = %v", err)
	}

	if code.ID != createdCode.ID {
		t.Errorf("コードのIDの期待値 = %d、実測値 = %d", createdCode.ID, code.ID)
	}
	if code.CodeDigest != "valid_code_digest" {
		t.Errorf("コードのdigest = %s、期待値 = valid_code_digest", code.CodeDigest)
	}
}

// TestGetValidSignInCode_Expiredは期限切れのコードが取得されないことをテスト
func TestGetValidSignInCode_Expired(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_expired_user").
		WithEmail("test_expired@example.com").
		Build()

	// 期限切れのSignInCodeを作成
	expiresAt := time.Now().Add(-1 * time.Minute) // 1分前に期限切れ
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "expired_code_digest",
		ExpiresAt:  expiresAt,
	}
	_, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err == nil {
		t.Error("期限切れのコードでエラーを期待したが、nilだった")
	}
}

// TestGetValidSignInCode_Usedは使用済みのコードが取得されないことをテスト
func TestGetValidSignInCode_Used(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_used_user").
		WithEmail("test_used@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "used_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignInCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("コードを使用済みにする処理のエラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err == nil {
		t.Error("使用済みのコードでエラーを期待したが、nilだった")
	}
}

// TestIncrementSignInCodeAttemptsは試行回数のインクリメントをテスト
func TestIncrementSignInCodeAttempts(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_attempts_user").
		WithEmail("test_attempts@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "test_attempts",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// 試行回数をインクリメント
	err = queries.IncrementSignInCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("attemptsの加算エラー = %v", err)
	}

	// コードを再取得して確認
	updatedCode, err := queries.GetValidSignInCode(context.Background(), int64(userID))
	if err != nil {
		t.Fatalf("加算後のコードの取得エラー = %v", err)
	}

	if updatedCode.Attempts != 1 {
		t.Errorf("attempts = %d、期待値 = 1", updatedCode.Attempts)
	}

	// さらにインクリメント
	err = queries.IncrementSignInCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("attemptsの再加算エラー = %v", err)
	}

	updatedCode, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err != nil {
		t.Fatalf("2回目の加算後のコードの取得エラー = %v", err)
	}

	if updatedCode.Attempts != 2 {
		t.Errorf("attempts = %d、期待値 = 2", updatedCode.Attempts)
	}
}

// TestMarkSignInCodeAsUsedはコードを使用済みにするテスト
func TestMarkSignInCodeAsUsed(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_mark_used_user").
		WithEmail("test_mark_used@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "test_used",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインインコードの作成エラー = %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignInCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("コードを使用済みにする処理のエラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err == nil {
		t.Error("使用済みにしたコードでエラーを期待したが、nilだった")
	}
}

// TestDeleteExpiredSignInCodesは期限切れコードの削除をテスト
func TestDeleteExpiredSignInCodes(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_delete_expired_user").
		WithEmail("test_delete_expired@example.com").
		Build()

	// 期限切れのSignInCodeを作成
	expiredTime := time.Now().Add(-2 * time.Hour)
	params1 := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "expired_code_1",
		ExpiresAt:  expiredTime,
	}
	_, err := queries.CreateSignInCode(context.Background(), params1)
	if err != nil {
		t.Fatalf("期限切れコードの作成エラー = %v", err)
	}

	// 有効なSignInCodeを作成
	validTime := time.Now().Add(15 * time.Minute)
	params2 := query.CreateSignInCodeParams{
		UserID:     int64(userID),
		CodeDigest: "valid_code",
		ExpiresAt:  validTime,
	}
	_, err = queries.CreateSignInCode(context.Background(), params2)
	if err != nil {
		t.Fatalf("有効なコードの作成エラー = %v", err)
	}

	// 期限切れコードを削除
	cutoffTime := time.Now().Add(-1 * time.Hour)
	err = queries.DeleteExpiredSignInCodes(context.Background(), cutoffTime)
	if err != nil {
		t.Fatalf("期限切れコードの削除エラー = %v", err)
	}

	// 有効なコードが残っていることを確認
	_, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err != nil {
		t.Error("期限切れコードの削除後も有効なコードが残っているべきだが、消えていた")
	}
}

// TestInvalidateUserSignInCodesはユーザーの全コード無効化をテスト
func TestInvalidateUserSignInCodes(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_invalidate_user").
		WithEmail("test_invalidate@example.com").
		Build()

	// 複数のSignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	for i := 0; i < 3; i++ {
		params := query.CreateSignInCodeParams{
			UserID:     int64(userID),
			CodeDigest: "code_" + string(rune('0'+i)),
			ExpiresAt:  expiresAt,
		}
		_, err := queries.CreateSignInCode(context.Background(), params)
		if err != nil {
			t.Fatalf("サインインコード[%d]の作成エラー = %v", i, err)
		}
	}

	// ユーザーのすべてのコードを無効化
	err := queries.InvalidateUserSignInCodes(context.Background(), int64(userID))
	if err != nil {
		t.Fatalf("ユーザーのコードの無効化エラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignInCode(context.Background(), int64(userID))
	if err == nil {
		t.Error("ユーザーの全コードを無効化した後にエラーを期待したが、nilだった")
	}
}
