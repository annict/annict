package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestUserRepository_UpdateStripeSubscriberID_Set はStripeサブスクライバーIDを設定できることをテスト
func TestUserRepository_UpdateStripeSubscriberID_Set(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	// Stripeサブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	// StripeサブスクライバーIDを設定
	err := repo.UpdateStripeSubscriberID(context.Background(), userID, &subscriberID)
	if err != nil {
		t.Fatalf("StripeサブスクライバーIDの設定に失敗: %v", err)
	}

	// 設定されたことを確認
	user, err := repo.GetByStripeSubscriberID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}

	if user.ID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", user.ID, userID)
	}
	if !user.StripeSubscriberID.Valid || user.StripeSubscriberID.Int64 != subscriberID {
		t.Errorf("StripeSubscriberIDが一致しません: got %v, want %d", user.StripeSubscriberID, subscriberID)
	}
}

// TestUserRepository_UpdateStripeSubscriberID_Clear はStripeサブスクライバーIDをクリアできることをテスト
func TestUserRepository_UpdateStripeSubscriberID_Clear(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	// Stripeサブスクライバーを作成して関連付け
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	err := repo.UpdateStripeSubscriberID(context.Background(), userID, &subscriberID)
	if err != nil {
		t.Fatalf("StripeサブスクライバーIDの設定に失敗: %v", err)
	}

	// StripeサブスクライバーIDをクリア
	err = repo.UpdateStripeSubscriberID(context.Background(), userID, nil)
	if err != nil {
		t.Fatalf("StripeサブスクライバーIDのクリアに失敗: %v", err)
	}

	// クリアされたことを確認（GetByStripeSubscriberIDでは見つからないはず）
	_, err = repo.GetByStripeSubscriberID(context.Background(), subscriberID)
	if err == nil {
		t.Error("クリア後のユーザーが見つかるべきではありません")
	}
	if err != sql.ErrNoRows {
		t.Errorf("期待するエラーではありません: got %v, want %v", err, sql.ErrNoRows)
	}
}

// TestUserRepository_GetByStripeSubscriberID は正常にユーザーを取得できることをテスト
func TestUserRepository_GetByStripeSubscriberID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("stripe_user").
		WithEmail("stripe@example.com").
		Build()

	// Stripeサブスクライバーを作成して関連付け
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	err := repo.UpdateStripeSubscriberID(context.Background(), userID, &subscriberID)
	if err != nil {
		t.Fatalf("StripeサブスクライバーIDの設定に失敗: %v", err)
	}

	// StripeサブスクライバーIDでユーザーを取得
	user, err := repo.GetByStripeSubscriberID(context.Background(), subscriberID)
	if err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}

	if user.ID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", user.ID, userID)
	}
	if user.Username != "stripe_user" {
		t.Errorf("ユーザー名が一致しません: got %s, want %s", user.Username, "stripe_user")
	}
	if user.Email != "stripe@example.com" {
		t.Errorf("メールアドレスが一致しません: got %s, want %s", user.Email, "stripe@example.com")
	}
}

// TestUserRepository_GetByStripeSubscriberID_NotFound は存在しないIDの場合エラーが返ることをテスト
func TestUserRepository_GetByStripeSubscriberID_NotFound(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	_, err := repo.GetByStripeSubscriberID(context.Background(), 99999)
	if err == nil {
		t.Error("存在しないIDでエラーが返されるべきです")
	}
	if err != sql.ErrNoRows {
		t.Errorf("期待するエラーではありません: got %v, want %v", err, sql.ErrNoRows)
	}
}
