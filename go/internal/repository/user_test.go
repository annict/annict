package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

// TestUserRepository_IsSupporter_StripeActive はStripeサポーター（アクティブ）の場合trueを返すことをテスト
func TestUserRepository_IsSupporter_StripeActive(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// アクティブなStripeサブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	// ユーザーをquery.User型で構築
	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{Int64: subscriberID, Valid: true},
		GumroadSubscriberID: sql.NullInt64{},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("アクティブなStripeサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_StripePastDue はStripeサポーター（past_due）の場合trueを返すことをテスト
func TestUserRepository_IsSupporter_StripePastDue(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// past_due状態のStripeサブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("past_due").
		Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{Int64: subscriberID, Valid: true},
		GumroadSubscriberID: sql.NullInt64{},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("past_due状態のStripeサポーターはtrueを返すべきです（猶予期間）")
	}
}

// TestUserRepository_IsSupporter_StripeCanceled はStripeサポーター（キャンセル済み）の場合falseを返すことをテスト
func TestUserRepository_IsSupporter_StripeCanceled(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// キャンセル済みのStripeサブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("canceled").
		Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{Int64: subscriberID, Valid: true},
		GumroadSubscriberID: sql.NullInt64{},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("キャンセル済みのStripeサポーターはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_GumroadActive はGumroadサポーター（アクティブ）の場合trueを返すことをテスト
func TestUserRepository_IsSupporter_GumroadActive(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// アクティブなGumroadサブスクライバーを作成（cancelled_atとended_atがnull）
	subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{},
		GumroadSubscriberID: sql.NullInt64{Int64: subscriberID, Valid: true},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("アクティブなGumroadサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_GumroadEnded はGumroadサポーター（終了済み）の場合falseを返すことをテスト
func TestUserRepository_IsSupporter_GumroadEnded(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// 終了済みのGumroadサブスクライバーを作成
	pastTime := time.Now().AddDate(0, 0, -1)
	subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).
		WithGumroadEndedAt(pastTime).
		Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{},
		GumroadSubscriberID: sql.NullInt64{Int64: subscriberID, Valid: true},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("終了済みのGumroadサポーターはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_BothActive はStripeとGumroad両方アクティブの場合trueを返すことをテスト
func TestUserRepository_IsSupporter_BothActive(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// 両方のサブスクライバーを作成
	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{Int64: stripeSubscriberID, Valid: true},
		GumroadSubscriberID: sql.NullInt64{Int64: gumroadSubscriberID, Valid: true},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("両方アクティブなサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_NoSubscription は非サポーターの場合falseを返すことをテスト
func TestUserRepository_IsSupporter_NoSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// サブスクリプションを持たないユーザー
	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{},
		GumroadSubscriberID: sql.NullInt64{},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("サブスクリプションを持たないユーザーはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_NoDependencies はリポジトリ依存がない場合falseを返すことをテスト
func TestUserRepository_IsSupporter_NoDependencies(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// 依存を設定しないUserRepository
	userRepo := repository.NewUserRepository(queries)

	// サブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	user := &query.User{
		StripeSubscriberID:  sql.NullInt64{Int64: subscriberID, Valid: true},
		GumroadSubscriberID: sql.NullInt64{},
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("リポジトリ依存がない場合はfalseを返すべきです")
	}
}
