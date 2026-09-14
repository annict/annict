package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateSignUpCodeはSignUpCodeの作成をテスト
func TestCreateSignUpCode(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      "test_create@example.com",
		CodeDigest: "test_digest_123",
		ExpiresAt:  expiresAt,
	}

	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// 基本的なアサーション
	if code.Email != "test_create@example.com" {
		t.Errorf("email = %s、期待値 = test_create@example.com", code.Email)
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

// TestGetValidSignUpCodeは有効なSignUpCodeの取得をテスト
func TestGetValidSignUpCode(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_get_valid@example.com"

	// 有効なSignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "valid_code_digest",
		ExpiresAt:  expiresAt,
	}
	createdCode, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// 有効なコードを取得
	code, err := queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("有効なサインアップコードの取得エラー = %v", err)
	}

	if code.ID != createdCode.ID {
		t.Errorf("コードのIDの期待値 = %d、実測値 = %d", createdCode.ID, code.ID)
	}
	if code.CodeDigest != "valid_code_digest" {
		t.Errorf("コードのdigest = %s、期待値 = valid_code_digest", code.CodeDigest)
	}
}

// TestGetValidSignUpCode_Expiredは期限切れのコードが取得されないことをテスト
func TestGetValidSignUpCode_Expired(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_expired@example.com"

	// 期限切れのSignUpCodeを作成
	expiresAt := time.Now().Add(-1 * time.Minute) // 1分前に期限切れ
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "expired_code_digest",
		ExpiresAt:  expiresAt,
	}
	_, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("期限切れのコードでエラーを期待したが、nilだった")
	}
}

// TestGetValidSignUpCode_Usedは使用済みのコードが取得されないことをテスト
func TestGetValidSignUpCode_Used(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_used@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "used_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignUpCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("コードを使用済みにする処理のエラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("使用済みのコードでエラーを期待したが、nilだった")
	}
}

// TestGetValidSignUpCode_ExceedsAttemptsは試行回数制限を超えたコードが取得されないことをテスト
func TestGetValidSignUpCode_ExceedsAttempts(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_attempts@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "attempts_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// 試行回数を5回インクリメント
	for i := 0; i < 5; i++ {
		err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
		if err != nil {
			t.Fatalf("attemptsの加算エラー = %v", err)
		}
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("試行回数を超えたコードでエラーを期待したが、nilだった")
	}
}

// TestIncrementSignUpCodeAttemptsは試行回数のインクリメントをテスト
func TestIncrementSignUpCodeAttempts(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_increment@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "test_attempts",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// 試行回数をインクリメント
	err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("attemptsの加算エラー = %v", err)
	}

	// コードを再取得して確認
	updatedCode, err := queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("加算後のコードの取得エラー = %v", err)
	}

	if updatedCode.Attempts != 1 {
		t.Errorf("attempts = %d、期待値 = 1", updatedCode.Attempts)
	}

	// さらにインクリメント
	err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("attemptsの再加算エラー = %v", err)
	}

	updatedCode, err = queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("2回目の加算後のコードの取得エラー = %v", err)
	}

	if updatedCode.Attempts != 2 {
		t.Errorf("attempts = %d、期待値 = 2", updatedCode.Attempts)
	}
}

// TestMarkSignUpCodeAsUsedはコードを使用済みにするテスト
func TestMarkSignUpCodeAsUsed(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_mark_used@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "test_used",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("サインアップコードの作成エラー = %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignUpCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("コードを使用済みにする処理のエラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("使用済みにしたコードでエラーを期待したが、nilだった")
	}
}

// TestInvalidateSignUpCodesByEmailはメールアドレスの全コード無効化をテスト
func TestInvalidateSignUpCodesByEmail(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	email := "test_invalidate@example.com"

	// 複数のSignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	for i := 0; i < 3; i++ {
		params := query.CreateSignUpCodeParams{
			Email:      email,
			CodeDigest: "code_" + string(rune('0'+i)),
			ExpiresAt:  expiresAt,
		}
		_, err := queries.CreateSignUpCode(context.Background(), params)
		if err != nil {
			t.Fatalf("サインアップコード[%d]の作成エラー = %v", i, err)
		}
	}

	// メールアドレスのすべてのコードを無効化
	err := queries.InvalidateSignUpCodesByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("メール単位のコードの無効化エラー = %v", err)
	}

	// 有効なコードの取得を試みる (失敗するべき)
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("メール単位で全コードを無効化した後にエラーを期待したが、nilだった")
	}
}
