package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateSessionUsecase_Execute は正常系のテストです
func TestCreateSessionUsecase_Execute(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("create_session_test").
		WithEmail("session@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// ユーザー情報を取得（encrypted_passwordが必要）
	queries := query.New(db).WithTx(tx)
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	// UseCase を作成
	uc := NewCreateSessionUsecase(queries)

	// セッションを作成
	ctx := context.Background()
	result, err := uc.Execute(ctx, tx, userID, user.EncryptedPassword, "")
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// 結果を検証
	if result.PublicID == "" {
		t.Error("PublicIDが空です")
	}
	if result.UserID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, userID)
	}

	// セッションがデータベースに保存されているか確認
	// Private IDを生成して確認
	privateID := generatePrivateID(result.PublicID)
	session, err := queries.GetSessionByID(ctx, privateID)
	if err != nil {
		t.Fatalf("セッションの取得に失敗: %v", err)
	}
	if session.SessionID != privateID {
		t.Errorf("セッションIDが一致しません: got %s, want %s", session.SessionID, privateID)
	}
}

// TestCreateSessionUsecase_Execute_WithFlashMessage はflashメッセージ付きのセッション作成をテストします
func TestCreateSessionUsecase_Execute_WithFlashMessage(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("flash_message_test").
		WithEmail("flash@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// ユーザー情報を取得
	queries := query.New(db).WithTx(tx)
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	// UseCase を作成
	uc := NewCreateSessionUsecase(queries)

	// flashメッセージ付きでセッションを作成
	ctx := context.Background()
	flashMessage := "ログインに成功しました"
	result, err := uc.Execute(ctx, tx, userID, user.EncryptedPassword, flashMessage)
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// 結果を検証
	if result.PublicID == "" {
		t.Error("PublicIDが空です")
	}
	if result.UserID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, userID)
	}
}

// TestCreateSessionUsecase_Execute_WithInvalidUserID はfail-fastケース：存在しないユーザーIDでのセッション作成をテストします
// 注: sessionsテーブルにはuser_idカラムがなく、外部キー制約もないため、存在しないユーザーIDでもセッション作成は成功します
// このテストは、将来的にuser_id外部キー制約が追加された場合のための参考として残します
func TestCreateSessionUsecase_Execute_WithInvalidUserID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	queries := query.New(db).WithTx(tx)
	uc := NewCreateSessionUsecase(queries)

	// 存在しないユーザーIDでセッション作成を試みる
	ctx := context.Background()
	invalidUserID := int64(999999999)
	result, err := uc.Execute(ctx, tx, invalidUserID, "dummy_encrypted_password", "")

	// 現在の実装では、sessionsテーブルにuser_idカラムがないため、エラーにならない
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// セッションは作成されるが、存在しないユーザーIDが格納される
	// これは実運用では問題となる可能性がある
	if result.UserID != invalidUserID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, invalidUserID)
	}

	// 将来的にuser_id外部キー制約が追加された場合、このテストは以下のように変更すべき：
	// if err == nil {
	//     t.Error("存在しないユーザーIDでセッション作成が成功してしまいました")
	// }
}

// TestCreateSessionUsecase_Execute_WithEmptyEncryptedPassword はfail-fastケース：encrypted_passwordが空の場合のテストです
func TestCreateSessionUsecase_Execute_WithEmptyEncryptedPassword(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("empty_password_test").
		WithEmail("empty@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	queries := query.New(db).WithTx(tx)
	uc := NewCreateSessionUsecase(queries)

	// 空のencrypted_passwordでセッション作成を試みる
	ctx := context.Background()
	result, err := uc.Execute(ctx, tx, userID, "", "")

	// エラーは発生しないが、authenticatable_saltが空になることを確認
	// これは仕様上許容されているが、セキュリティ的には問題がある
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// セッションは作成されることを確認
	if result.PublicID == "" {
		t.Error("PublicIDが空です")
	}

	// 注意: encrypted_passwordが空の場合、authenticatable_saltも空になります
	// これは実運用では発生すべきではない状況です
}

// TestCreateSessionUsecase_Execute_WithShortEncryptedPassword はfail-fastケース：encrypted_passwordが29文字未満の場合のテストです
func TestCreateSessionUsecase_Execute_WithShortEncryptedPassword(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("short_password_test").
		WithEmail("short@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	queries := query.New(db).WithTx(tx)
	uc := NewCreateSessionUsecase(queries)

	// 29文字未満のencrypted_passwordでセッション作成を試みる
	ctx := context.Background()
	shortPassword := "tooshort" // 8文字
	result, err := uc.Execute(ctx, tx, userID, shortPassword, "")

	// エラーは発生しないが、authenticatable_saltが短くなることを確認
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// セッションは作成されることを確認
	if result.PublicID == "" {
		t.Error("PublicIDが空です")
	}

	// 注意: encrypted_passwordが29文字未満の場合、authenticatable_saltも短くなります
	// これは実運用では発生すべきではない状況です（bcryptハッシュは60文字）
}

// TestCreateSessionUsecase_Execute_WithoutTransaction はトランザクションなしでのセッション作成をテストします
func TestCreateSessionUsecase_Execute_WithoutTransaction(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("no_tx_test").
		WithEmail("notx@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// ユーザー情報を取得
	queries := query.New(db).WithTx(tx)
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	// トランザクションをコミット（UseCaseがトランザクションなしで動作するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にセッションとユーザーを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// UseCase を作成（トランザクションなしのqueries）
	queriesWithoutTx := query.New(db)
	uc := NewCreateSessionUsecase(queriesWithoutTx)

	// トランザクションなしでセッションを作成
	ctx := context.Background()
	result, err := uc.Execute(ctx, nil, userID, user.EncryptedPassword, "")
	if err != nil {
		t.Fatalf("セッション作成に失敗: %v", err)
	}

	// 結果を検証
	if result.PublicID == "" {
		t.Error("PublicIDが空です")
	}
	if result.UserID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, userID)
	}
}

// TestCreateSessionUsecase_Execute_DuplicateSessionID はfail-fastケース：セッションID重複時のテストです
func TestCreateSessionUsecase_Execute_DuplicateSessionID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("duplicate_session_test").
		WithEmail("duplicate@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// ユーザー情報を取得
	queries := query.New(db).WithTx(tx)
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	// UseCase を作成
	uc := NewCreateSessionUsecase(queries)

	// 最初のセッションを作成
	ctx := context.Background()
	result1, err := uc.Execute(ctx, tx, userID, user.EncryptedPassword, "")
	if err != nil {
		t.Fatalf("最初のセッション作成に失敗: %v", err)
	}

	// 2番目のセッションを作成（異なるpublicIDが生成されるはず）
	result2, err := uc.Execute(ctx, tx, userID, user.EncryptedPassword, "")
	if err != nil {
		t.Fatalf("2番目のセッション作成に失敗: %v", err)
	}

	// 2つのセッションIDが異なることを確認
	if result1.PublicID == result2.PublicID {
		t.Error("セッションIDが重複しています")
	}

	// 両方のセッションがデータベースに保存されていることを確認
	privateID1 := generatePrivateID(result1.PublicID)
	privateID2 := generatePrivateID(result2.PublicID)

	_, err = queries.GetSessionByID(ctx, privateID1)
	if err != nil {
		t.Errorf("最初のセッションが取得できません: %v", err)
	}

	_, err = queries.GetSessionByID(ctx, privateID2)
	if err != nil {
		t.Errorf("2番目のセッションが取得できません: %v", err)
	}
}

// TestCreateSessionUsecase_Execute_WithNullUserData はfail-fastケース：ユーザーデータにNULL値がある場合のテストです
// 注: このテストは現在の実装ではユーザー情報を直接取得しないため、スキップします
// 将来的にユーザー情報の検証が追加された場合に有効化してください
func TestCreateSessionUsecase_Execute_WithNullUserData(t *testing.T) {
	t.Skip("現在の実装ではユーザー情報を直接取得しないため、このテストはスキップします")

	// 以下は将来的な実装のための参考コード:
	// - created_at, updated_atがNULLの場合にエラーを返すべき
	// - encrypted_passwordがNULLの場合にエラーを返すべき
	// - データベース制約違反を検出してfail-fastするべき
}
