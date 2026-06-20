package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestUserRepository_Create はユーザーを正常に作成できることをテスト
func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	user, err := repo.Create(context.Background(), repository.UserCreateParams{
		Username:          "newuser",
		Email:             "newuser@example.com",
		EncryptedPassword: "",
		Locale:            "ja",
	})
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if user.ID == 0 {
		t.Error("IDが0です")
	}
	if user.Username != "newuser" {
		t.Errorf("Usernameが一致しません: got %v, want %v", user.Username, "newuser")
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("Emailが一致しません: got %v, want %v", user.Email, "newuser@example.com")
	}
	if user.Locale != "ja" {
		t.Errorf("Localeが一致しません: got %v, want %v", user.Locale, "ja")
	}
	if user.Role != 0 {
		t.Errorf("Roleが一致しません: got %v, want %v", user.Role, 0)
	}
}

// TestUserRepository_GetByUsername_Exists はユーザー名が存在する場合にnilを返すことをテスト
func TestUserRepository_GetByUsername_Exists(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	testutil.NewUserBuilder(t, tx).
		WithUsername("existinguser").
		WithEmail("existing@example.com").
		Build()

	err := repo.GetByUsername(context.Background(), "existinguser")
	if err != nil {
		t.Errorf("存在するユーザー名でエラーが返されました: %v", err)
	}
}

// TestUserRepository_GetByUsername_NotFound はユーザー名が存在しない場合にエラーを返すことをテスト
func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	err := repo.GetByUsername(context.Background(), "nonexistentuser")
	if err == nil {
		t.Error("存在しないユーザー名でエラーが返されるべきです")
	}
	if err != sql.ErrNoRows {
		t.Errorf("期待するエラーではありません: got %v, want %v", err, sql.ErrNoRows)
	}
}

// TestUserRepository_GetByUsername_CaseInsensitive はユーザー名の大文字小文字を区別しないことをテスト
func TestUserRepository_GetByUsername_CaseInsensitive(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserRepository(queries)

	testutil.NewUserBuilder(t, tx).
		WithUsername("TestUser").
		WithEmail("testcase@example.com").
		Build()

	err := repo.GetByUsername(context.Background(), "testuser")
	if err != nil {
		t.Errorf("大文字小文字を区別しない検索でエラーが返されました: %v", err)
	}
}

// TestUserRepository_UpdateStripeSubscriberID_Set はStripeサブスクライバーIDを設定できることをテスト
func TestUserRepository_UpdateStripeSubscriberID_Set(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	if model.UserID(user.ID) != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", user.ID, userID)
	}
	if !user.StripeSubscriberID.Valid || model.StripeSubscriberID(user.StripeSubscriberID.Int64) != subscriberID {
		t.Errorf("StripeSubscriberIDが一致しません: got %v, want %d", user.StripeSubscriberID, subscriberID)
	}
}

// TestUserRepository_UpdateStripeSubscriberID_Clear はStripeサブスクライバーIDをクリアできることをテスト
func TestUserRepository_UpdateStripeSubscriberID_Clear(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	if model.UserID(user.ID) != userID {
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	user := &model.User{
		StripeSubscriberID: &subscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	user := &model.User{
		StripeSubscriberID: &subscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	user := &model.User{
		StripeSubscriberID: &subscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// アクティブなGumroadサブスクライバーを作成（cancelled_atとended_atがnull）
	subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	user := &model.User{
		GumroadSubscriberID: &subscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	user := &model.User{
		GumroadSubscriberID: &subscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	user := &model.User{
		StripeSubscriberID:  &stripeSubscriberID,
		GumroadSubscriberID: &gumroadSubscriberID,
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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// サブスクリプションを持たないユーザー
	user := &model.User{}

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// 依存を設定しないUserRepository
	userRepo := repository.NewUserRepository(queries)

	// サブスクライバーを作成
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()

	user := &model.User{
		StripeSubscriberID: &subscriberID,
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporterの実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("リポジトリ依存がない場合はfalseを返すべきです")
	}
}
