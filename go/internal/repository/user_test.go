package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestUserRepository_Createはユーザーを正常に作成できることをテスト
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
		t.Errorf("Username = %v、期待値 = %v", user.Username, "newuser")
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("Email = %v、期待値 = %v", user.Email, "newuser@example.com")
	}
	if user.Locale != "ja" {
		t.Errorf("Locale = %v、期待値 = %v", user.Locale, "ja")
	}
	if user.Role != 0 {
		t.Errorf("Role = %v、期待値 = %v", user.Role, 0)
	}
}

// TestUserRepository_GetByUsername_Existsはユーザー名が存在する場合にnilを返すことをテスト
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

// TestUserRepository_GetByUsername_NotFoundはユーザー名が存在しない場合にエラーを返すことをテスト
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
		t.Errorf("エラー = %v、期待値 = %v", err, sql.ErrNoRows)
	}
}

// TestUserRepository_GetByUsername_CaseInsensitiveはユーザー名の大文字小文字を区別しないことをテスト
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

// TestUserRepository_FindActiveIDByEmailは論理削除されていないユーザーのIDだけを返すことをテスト
func TestUserRepository_FindActiveIDByEmail(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	activeUser := testutil.NewUserBuilder(t, tx).BuildWithResult()
	deletedUser := testutil.NewUserBuilder(t, tx).
		WithDeletedAt(time.Now()).
		BuildWithResult()

	t.Run("有効なユーザーはIDを返す (大文字小文字を区別しない)", func(t *testing.T) {
		got, err := repo.FindActiveIDByEmail(ctx, strings.ToUpper(activeUser.Email))
		if err != nil {
			t.Fatalf("FindActiveIDByEmailのエラー = %v", err)
		}
		if got == nil || *got != activeUser.ID {
			t.Errorf("ID = %v、期待値 = %v", got, activeUser.ID)
		}
	})

	t.Run("削除済みユーザーはnilを返す", func(t *testing.T) {
		got, err := repo.FindActiveIDByEmail(ctx, deletedUser.Email)
		if err != nil {
			t.Fatalf("FindActiveIDByEmailのエラー = %v", err)
		}
		if got != nil {
			t.Errorf("ID = %v、期待値 = nil", *got)
		}
	})

	t.Run("存在しないユーザーはnilを返す", func(t *testing.T) {
		got, err := repo.FindActiveIDByEmail(ctx, "find-active-id-nonexistent@example.com")
		if err != nil {
			t.Fatalf("FindActiveIDByEmailのエラー = %v", err)
		}
		if got != nil {
			t.Errorf("ID = %v、期待値 = nil", *got)
		}
	})
}

func TestUserRepository_FindDeletionTarget(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	deletedUserID := testutil.NewUserBuilder(t, tx).
		WithStripeSubscriberID(&subscriberID).
		WithDeletedAt(time.Now()).
		Build()
	activeUserID := testutil.NewUserBuilder(t, tx).Build()

	t.Run("論理削除済みのユーザーも取得できる", func(t *testing.T) {
		got, err := repo.FindDeletionTarget(ctx, deletedUserID)
		if err != nil {
			t.Fatalf("FindDeletionTargetのエラー = %v", err)
		}
		if got == nil {
			t.Fatal("FindDeletionTarget() = nil、ユーザーを返すべき")
		}
		if got.ID != deletedUserID {
			t.Errorf("ID = %v、期待値 = %v", got.ID, deletedUserID)
		}
		if !got.Deleted {
			t.Error("Deleted = false、期待値 = true")
		}
		if got.StripeSubscriberID == nil || *got.StripeSubscriberID != subscriberID {
			t.Errorf("StripeSubscriberID = %v、期待値 = %v", got.StripeSubscriberID, subscriberID)
		}
	})

	t.Run("論理削除されていないユーザーはDeletedがfalseになる", func(t *testing.T) {
		got, err := repo.FindDeletionTarget(ctx, activeUserID)
		if err != nil {
			t.Fatalf("FindDeletionTargetのエラー = %v", err)
		}
		if got == nil {
			t.Fatal("FindDeletionTarget() = nil、ユーザーを返すべき")
		}
		if got.Deleted {
			t.Error("Deleted = true、期待値 = false")
		}
		if got.StripeSubscriberID != nil {
			t.Errorf("StripeSubscriberID = %v、期待値 = nil", *got.StripeSubscriberID)
		}
	})

	t.Run("存在しないユーザーはnilを返す", func(t *testing.T) {
		got, err := repo.FindDeletionTarget(ctx, model.UserID(-1))
		if err != nil {
			t.Fatalf("FindDeletionTargetのエラー = %v", err)
		}
		if got != nil {
			t.Errorf("FindDeletionTarget() = %+v、期待値 = nil", got)
		}
	})
}

// TestUserRepository_UpdateStripeSubscriberID_SetはStripeサブスクライバーIDを設定できることをテスト
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
		t.Errorf("ユーザーID = %d、期待値 = %d", user.ID, userID)
	}
	if !user.StripeSubscriberID.Valid || model.StripeSubscriberID(user.StripeSubscriberID.Int64) != subscriberID {
		t.Errorf("StripeSubscriberID = %v、期待値 = %d", user.StripeSubscriberID, subscriberID)
	}
}

// TestUserRepository_UpdateStripeSubscriberID_ClearはStripeサブスクライバーIDをクリアできることをテスト
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

	// クリアされたことを確認 (GetByStripeSubscriberIDでは見つからないはず)
	_, err = repo.GetByStripeSubscriberID(context.Background(), subscriberID)
	if err == nil {
		t.Error("クリア後のユーザーが見つかるべきではありません")
	}
	if err != sql.ErrNoRows {
		t.Errorf("エラー = %v、期待値 = %v", err, sql.ErrNoRows)
	}
}

// TestUserRepository_GetByStripeSubscriberIDは正常にユーザーを取得できることをテスト
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
		t.Errorf("ユーザーID = %d、期待値 = %d", user.ID, userID)
	}
	if user.Username != "stripe_user" {
		t.Errorf("ユーザー名 = %s、期待値 = %s", user.Username, "stripe_user")
	}
	if user.Email != "stripe@example.com" {
		t.Errorf("メールアドレス = %s、期待値 = %s", user.Email, "stripe@example.com")
	}
}

// TestUserRepository_GetByStripeSubscriberID_NotFoundは存在しないIDの場合エラーが返ることをテスト
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
		t.Errorf("エラー = %v、期待値 = %v", err, sql.ErrNoRows)
	}
}

func TestUserRepository_ClearActiveUserStripeSubscriberID_AccountDeletionRace(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = setupTx.Rollback() })
	subscriberID := testutil.NewStripeSubscriberBuilder(t, setupTx).Build()
	userID := testutil.NewUserBuilder(t, setupTx).WithStripeSubscriberID(&subscriberID).Build()
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	repo := repository.NewUserRepository(query.New(db))
	t.Cleanup(func() {
		cleanupTx, err := db.Begin()
		if err != nil {
			t.Errorf("後片付けのトランザクションの開始に失敗: %v", err)
			return
		}
		defer func() { _ = cleanupTx.Rollback() }()
		if _, err := repo.WithTx(cleanupTx).DeleteByID(context.Background(), userID); err != nil {
			t.Errorf("後片付けでユーザーを削除できない: %v", err)
			return
		}
		if err := cleanupTx.Commit(); err != nil {
			t.Errorf("後片付けのコミットに失敗: %v", err)
		}
	})

	deletionTx, err := db.Begin()
	if err != nil {
		t.Fatalf("退会受付トランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = deletionTx.Rollback() })
	webhookTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Webhookトランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = webhookTx.Rollback() })
	var webhookPID int
	if err := webhookTx.QueryRow("SELECT pg_backend_pid()").Scan(&webhookPID); err != nil {
		t.Fatalf("Webhook接続のPID取得に失敗: %v", err)
	}
	if _, err := deletionTx.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(userID)); err != nil {
		t.Fatalf("退会受付の更新に失敗: %v", err)
	}

	type clearResult struct {
		userID *model.UserID
		err    error
	}
	resultCh := make(chan clearResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func() {
		id, err := repo.WithTx(webhookTx).ClearActiveUserStripeSubscriberID(ctx, subscriberID)
		resultCh <- clearResult{userID: id, err: err}
	}()

	// 解除の更新が退会受付の行ロックを待つまで確認してから、退会をコミットする。
	deadline := time.Now().Add(5 * time.Second)
	for {
		select {
		case result := <-resultCh:
			t.Fatalf("退会受付のコミット前に紐付け解除が終わった: %v", result.err)
		default:
		}
		var waitEvent sql.NullString
		if err := db.QueryRow("SELECT wait_event_type FROM pg_stat_activity WHERE pid = $1", webhookPID).Scan(&waitEvent); err != nil {
			t.Fatalf("Webhook接続の待機状態の取得に失敗: %v", err)
		}
		if waitEvent.Valid && waitEvent.String == "Lock" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Webhookの紐付け解除が退会受付の行ロックを待たなかった")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := deletionTx.Commit(); err != nil {
		t.Fatalf("退会受付のコミットに失敗: %v", err)
	}
	var result clearResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("退会受付のコミット後も紐付け解除が終了しなかった: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("紐付け解除のエラー = %v", result.err)
	}
	if result.userID != nil {
		t.Errorf("解除されたユーザーID = %d、期待値 = nil", *result.userID)
	}
	if err := webhookTx.Commit(); err != nil {
		t.Fatalf("Webhookのコミットに失敗: %v", err)
	}
	var retainedSubscriberID sql.NullInt64
	if err := db.QueryRow("SELECT stripe_subscriber_id FROM users WHERE id = $1", int64(userID)).Scan(&retainedSubscriberID); err != nil {
		t.Fatalf("退会後のユーザー取得に失敗: %v", err)
	}
	if !retainedSubscriberID.Valid || retainedSubscriberID.Int64 != int64(subscriberID) {
		t.Errorf("退会後の購読情報ID = %v、期待値 = %d", retainedSubscriberID, subscriberID)
	}
}

// TestUserRepository_IsSupporter_StripeActiveはStripeサポーター (アクティブ) の場合trueを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("アクティブなStripeサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_StripePastDueはStripeサポーター (past_due) の場合trueを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("past_due状態のStripeサポーターはtrueを返すべきです (猶予期間)")
	}
}

// TestUserRepository_IsSupporter_StripeCanceledはStripeサポーター (キャンセル済み) の場合falseを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("キャンセル済みのStripeサポーターはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_GumroadActiveはGumroadサポーター (アクティブ) の場合trueを返すことをテスト
func TestUserRepository_IsSupporter_GumroadActive(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	stripeRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadRepo := repository.NewGumroadSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries).
		WithStripeSubscriberRepo(stripeRepo).
		WithGumroadSubscriberRepo(gumroadRepo)

	// アクティブなGumroadサブスクライバーを作成 (cancelled_atとended_atがnull)
	subscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	user := &model.User{
		GumroadSubscriberID: &subscriberID,
	}

	isSupporter, err := userRepo.IsSupporter(context.Background(), user)
	if err != nil {
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("アクティブなGumroadサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_GumroadEndedはGumroadサポーター (終了済み) の場合falseを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("終了済みのGumroadサポーターはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_BothActiveはStripeとGumroad両方アクティブの場合trueを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if !isSupporter {
		t.Error("両方アクティブなサポーターはtrueを返すべきです")
	}
}

// TestUserRepository_IsSupporter_NoSubscriptionは非サポーターの場合falseを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("サブスクリプションを持たないユーザーはfalseを返すべきです")
	}
}

// TestUserRepository_IsSupporter_NoDependenciesはリポジトリ依存がない場合falseを返すことをテスト
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
		t.Fatalf("IsSupporter()の実行に失敗: %v", err)
	}

	if isSupporter {
		t.Error("リポジトリ依存がない場合はfalseを返すべきです")
	}
}

// TestUserDeletion_ReferencingRowsは、ユーザー行を削除したときに参照元の行が外部キーの
// ON DELETEに従って扱われ、削除が外部キー違反で失敗しないことをテスト。
// 退会の削除処理はこれらのテーブルを明示的に削除せず、外部キーに任せる前提で作られる。
func TestUserDeletion_ReferencingRows(t *testing.T) {
	t.Parallel()

	_, tx := testutil.SetupTx(t)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("deleting_user").
		WithEmail("deleting_user@example.com").
		Build()
	otherUserID := testutil.NewUserBuilder(t, tx).
		WithUsername("other_user").
		WithEmail("other_user@example.com").
		Build()

	workID := testutil.CreateTestWork(t, tx, "外部キーのテスト作品")
	workImageID := testutil.NewWorkImageBuilder(t, tx, workID).Build()
	if _, err := tx.ExecContext(ctx, `UPDATE work_images SET user_id = $1 WHERE id = $2`, int64(userID), workImageID); err != nil {
		t.Fatalf("作品画像の投稿者の更新に失敗: %v", err)
	}

	var characterID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO characters (name, created_at, updated_at) VALUES ('外部キーのテストキャラクター', NOW(), NOW())
		RETURNING id
	`).Scan(&characterID); err != nil {
		t.Fatalf("キャラクターの作成に失敗: %v", err)
	}
	var characterImageID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO character_images (
			character_id, user_id, attachment_file_name, attachment_file_size,
			attachment_content_type, attachment_updated_at, created_at, updated_at
		) VALUES ($1, $2, 'test.jpg', 1, 'image/jpeg', NOW(), NOW(), NOW())
		RETURNING id
	`, characterID, int64(userID)).Scan(&characterImageID); err != nil {
		t.Fatalf("キャラクター画像の作成に失敗: %v", err)
	}

	featureFlagID := testutil.NewFeatureFlagBuilder(t, tx).
		WithUserID(userID).
		WithName("go_foreign_key_test").
		Build()

	var collectionItemID int64
	if err := tx.QueryRowContext(ctx, `
		WITH collection AS (
			INSERT INTO collections (user_id, name, created_at, updated_at)
			VALUES ($1, '外部キーのテストコレクション', NOW(), NOW())
			RETURNING id
		)
		INSERT INTO collection_items (user_id, collection_id, work_id, created_at, updated_at)
		SELECT $1, collection.id, $2, NOW(), NOW() FROM collection
		RETURNING id
	`, int64(otherUserID), int64(workID)).Scan(&collectionItemID); err != nil {
		t.Fatalf("コレクション項目の作成に失敗: %v", err)
	}

	insertReaction := func(t *testing.T, reactorID, targetUserID model.UserID, collectionItemID *int64) int64 {
		t.Helper()

		var id int64
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO reactions (user_id, target_user_id, kind, collection_item_id, created_at, updated_at)
			VALUES ($1, $2, 'thumbs_up', $3, NOW(), NOW())
			RETURNING id
		`, int64(reactorID), int64(targetUserID), collectionItemID).Scan(&id); err != nil {
			t.Fatalf("リアクションの作成に失敗: %v", err)
		}
		return id
	}
	reactionByUserID := insertReaction(t, userID, otherUserID, nil)
	reactionToUserID := insertReaction(t, otherUserID, userID, nil)
	reactionToCollectionItemID := insertReaction(t, otherUserID, otherUserID, &collectionItemID)

	countRows := func(t *testing.T, table string, id int64) int {
		t.Helper()

		var count int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE id = $1", id).Scan(&count); err != nil {
			t.Fatalf("%sの件数の取得に失敗: %v", table, err)
		}
		return count
	}

	// NewUserBuilderが併せて作る行は外部キーにON DELETEが無いため、先に消しておく。
	for _, statement := range []string{
		`DELETE FROM email_notifications WHERE user_id = $1`,
		`DELETE FROM settings WHERE user_id = $1`,
		`DELETE FROM profiles WHERE user_id = $1`,
		`DELETE FROM users WHERE id = $1`,
	} {
		if _, err := tx.ExecContext(ctx, statement, int64(userID)); err != nil {
			t.Fatalf("ユーザーの削除に失敗 (%s): %v", statement, err)
		}
	}

	t.Run("作品画像は残り、投稿者がNULLになる", func(t *testing.T) {
		var uploaderID sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT user_id FROM work_images WHERE id = $1`, workImageID).Scan(&uploaderID); err != nil {
			t.Fatalf("作品画像の取得に失敗: %v", err)
		}
		if uploaderID.Valid {
			t.Errorf("作品画像のuser_id = %d、期待値 = NULL", uploaderID.Int64)
		}
	})

	t.Run("キャラクター画像は残り、投稿者がNULLになる", func(t *testing.T) {
		var uploaderID sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT user_id FROM character_images WHERE id = $1`, characterImageID).Scan(&uploaderID); err != nil {
			t.Fatalf("キャラクター画像の取得に失敗: %v", err)
		}
		if uploaderID.Valid {
			t.Errorf("キャラクター画像のuser_id = %d、期待値 = NULL", uploaderID.Int64)
		}
	})

	t.Run("ユーザー単位のフィーチャーフラグが消える", func(t *testing.T) {
		if got := countRows(t, "feature_flags", int64(featureFlagID)); got != 0 {
			t.Errorf("フィーチャーフラグの件数 = %d、期待値 = 0", got)
		}
	})

	t.Run("本人が付けたリアクションと本人へのリアクションが消える", func(t *testing.T) {
		if got := countRows(t, "reactions", reactionByUserID); got != 0 {
			t.Errorf("本人が付けたリアクションの件数 = %d、期待値 = 0", got)
		}
		if got := countRows(t, "reactions", reactionToUserID); got != 0 {
			t.Errorf("本人へのリアクションの件数 = %d、期待値 = 0", got)
		}
		if got := countRows(t, "reactions", reactionToCollectionItemID); got != 1 {
			t.Errorf("無関係なリアクションの件数 = %d、期待値 = 1", got)
		}
	})

	t.Run("コレクション項目を削除するとそのリアクションが消える", func(t *testing.T) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM collection_items WHERE id = $1`, collectionItemID); err != nil {
			t.Fatalf("コレクション項目の削除に失敗: %v", err)
		}
		if got := countRows(t, "reactions", reactionToCollectionItemID); got != 0 {
			t.Errorf("コレクション項目へのリアクションの件数 = %d、期待値 = 0", got)
		}
	})
}

func TestUserRepository_DeleteByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	sharedStripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
	userID := testutil.NewUserBuilder(t, tx).WithStripeSubscriberID(&stripeSubscriberID).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).WithStripeSubscriberID(&sharedStripeSubscriberID).Build()
	// ほかのユーザーからも参照されている購読情報は、外部キーを満たすため残す。
	sharingUserID := testutil.NewUserBuilder(t, tx).WithStripeSubscriberID(&sharedStripeSubscriberID).Build()
	if _, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, int64(gumroadSubscriberID), int64(userID)); err != nil {
		t.Fatalf("Gumroadの購読情報の設定に失敗: %v", err)
	}
	emailConfirmationID := testutil.CreateTestEmailConfirmation(t, tx, userID)

	got, err := repo.DeleteByID(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteByIDのエラー = %v", err)
	}
	if got != 1 {
		t.Errorf("削除件数 = %d、期待値 = 1", got)
	}

	// 共有している購読情報を持つユーザーを削除しても、購読情報は残る。
	if _, err := repo.DeleteByID(ctx, otherUserID); err != nil {
		t.Fatalf("購読情報を共有するユーザーのDeleteByIDのエラー = %v", err)
	}

	for _, c := range []struct {
		name  string
		table string
		where string
		arg   int64
		want  int
	}{
		{name: "ユーザー", table: "users", where: "id = $1", arg: int64(userID), want: 0},
		{name: "設定", table: "settings", where: "user_id = $1", arg: int64(userID), want: 0},
		{name: "メール通知設定", table: "email_notifications", where: "user_id = $1", arg: int64(userID), want: 0},
		{name: "メールアドレスの確認", table: "email_confirmations", where: "id = $1", arg: emailConfirmationID, want: 0},
		{name: "プロフィール", table: "profiles", where: "user_id = $1", arg: int64(userID), want: 0},
		{name: "Stripeの購読情報", table: "stripe_subscribers", where: "id = $1", arg: int64(stripeSubscriberID), want: 0},
		{name: "Gumroadの購読情報", table: "gumroad_subscribers", where: "id = $1", arg: int64(gumroadSubscriberID), want: 0},
		{name: "共有しているStripeの購読情報", table: "stripe_subscribers", where: "id = $1", arg: int64(sharedStripeSubscriberID), want: 1},
		{name: "購読情報を共有するユーザー", table: "users", where: "id = $1", arg: int64(sharingUserID), want: 1},
	} {
		if got := testutil.CountRows(t, tx, c.table, c.where, c.arg); got != c.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", c.name, got, c.want)
		}
	}

	// すでに削除したユーザーでは何もしない。
	got, err = repo.DeleteByID(ctx, userID)
	if err != nil {
		t.Fatalf("削除済みのユーザーのDeleteByIDのエラー = %v", err)
	}
	if got != 0 {
		t.Errorf("削除済みのユーザーの削除件数 = %d、期待値 = 0", got)
	}
}

func TestUserRepository_DeleteByID_SharedSubscribersConcurrently(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		table string
	}{
		{name: "Stripe", table: "stripe_subscribers"},
		{name: "Gumroad", table: "gumroad_subscribers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			db := testutil.GetTestDB()
			setupTx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
			}
			defer func() { _ = setupTx.Rollback() }()

			var subscriberID int64
			if tt.table == "stripe_subscribers" {
				stripeID := testutil.NewStripeSubscriberBuilder(t, setupTx).Build()
				subscriberID = int64(stripeID)
			} else {
				gumroadID := testutil.NewGumroadSubscriberBuilder(t, setupTx).Build()
				subscriberID = int64(gumroadID)
			}
			userIDs := [2]model.UserID{}
			for i := range userIDs {
				builder := testutil.NewUserBuilder(t, setupTx)
				if tt.table == "stripe_subscribers" {
					stripeID := model.StripeSubscriberID(subscriberID)
					builder.WithStripeSubscriberID(&stripeID)
				}
				userIDs[i] = builder.Build()
				if tt.table == "gumroad_subscribers" {
					if _, err := setupTx.ExecContext(ctx, "UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2", subscriberID, int64(userIDs[i])); err != nil {
						t.Fatalf("Gumroad購読情報の設定に失敗: %v", err)
					}
				}
			}
			if err := setupTx.Commit(); err != nil {
				t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
			}

			repo := repository.NewUserRepository(query.New(db))
			// beginはロックの保持をテストから観測できるよう、トランザクションを開いてバックエンドのPIDを返す。
			begin := func(ctx context.Context) (*sql.Tx, int, error) {
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					return nil, 0, err
				}
				var pid int
				if err := tx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
					_ = tx.Rollback()
					return nil, 0, err
				}
				return tx, pid, nil
			}
			deleteInTx := func(ctx context.Context, tx *sql.Tx, id model.UserID) error {
				defer func() { _ = tx.Rollback() }()
				if _, err := repo.WithTx(tx).DeleteByID(ctx, id); err != nil {
					return err
				}
				return tx.Commit()
			}
			t.Cleanup(func() {
				ctx := context.Background()
				for _, id := range userIDs {
					tx, _, err := begin(ctx)
					if err != nil {
						t.Errorf("後片付けのトランザクションの開始に失敗: %v", err)
						continue
					}
					if err := deleteInTx(ctx, tx, id); err != nil {
						t.Errorf("後片付けでユーザーを削除できない: %v", err)
					}
				}
				if _, err := db.Exec("DELETE FROM "+tt.table+" WHERE id = $1", subscriberID); err != nil {
					t.Errorf("後片付けで購読情報を削除できない: %v", err)
				}
			})

			// 両方の削除を購読情報のロックで待たせ、重なる実行を確実に作る。
			blocker, _, err := begin(ctx)
			if err != nil {
				t.Fatalf("ロック用トランザクションの開始に失敗: %v", err)
			}
			defer func() { _ = blocker.Rollback() }()
			if _, err := blocker.ExecContext(ctx, "SELECT id FROM "+tt.table+" WHERE id = $1 FOR UPDATE", subscriberID); err != nil {
				t.Fatalf("購読情報のロックに失敗: %v", err)
			}

			pids := make([]int, 0, len(userIDs))
			results := make(chan error, len(userIDs))
			for _, id := range userIDs {
				tx, pid, err := begin(ctx)
				if err != nil {
					t.Fatalf("削除用トランザクションの開始に失敗: %v", err)
				}
				pids = append(pids, pid)
				go func() { results <- deleteInTx(ctx, tx, id) }()
			}

			// 両方の削除がロック待ちに入ってから、ロックを解放する。
			for waiting := 0; waiting < len(pids); {
				select {
				case err := <-results:
					t.Fatalf("購読情報のロック中に削除が終了しました: %v", err)
				case <-ctx.Done():
					t.Fatalf("削除がロック待ちに入りません: %v", ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
				waiting = 0
				for _, pid := range pids {
					var waitEventType sql.NullString
					if err := db.QueryRowContext(ctx, "SELECT wait_event_type FROM pg_stat_activity WHERE pid = $1", pid).Scan(&waitEventType); err != nil {
						t.Fatalf("ロック待ちの確認に失敗: %v", err)
					}
					if waitEventType.String == "Lock" {
						waiting++
					}
				}
			}
			if err := blocker.Commit(); err != nil {
				t.Fatalf("ロック用トランザクションのコミットに失敗: %v", err)
			}
			for range userIDs {
				if err := <-results; err != nil {
					t.Fatalf("同時削除に失敗: %v", err)
				}
			}
			var count int
			for _, id := range userIDs {
				if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", int64(id)).Scan(&count); err != nil {
					t.Fatalf("ユーザーの確認に失敗: %v", err)
				}
				if count != 0 {
					t.Errorf("ユーザーが残っています: %d", id)
				}
			}
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tt.table+" WHERE id = $1", subscriberID).Scan(&count); err != nil {
				t.Fatalf("購読情報の確認に失敗: %v", err)
			}
			if count != 0 {
				t.Errorf("共有購読情報の件数 = %d、期待値 = 0", count)
			}
		})
	}
}

// TestUserRepository_AnonymizeAndSoftDeleteは、ユーザー名とメールアドレスを置き換えて論理削除し、
// すでに論理削除されたユーザーは変更しないことをテスト
func TestUserRepository_AnonymizeAndSoftDelete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	var originalUsername string
	if err := tx.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&originalUsername); err != nil {
		t.Fatalf("元のユーザー名の取得に失敗: %v", err)
	}

	updated, err := repo.AnonymizeAndSoftDelete(ctx, userID, "古いユーザー名", "anonymized_user", "anonymized_user@example.com")
	if err != nil {
		t.Fatalf("ユーザー名が異なる場合のAnonymizeAndSoftDelete()のエラー = %v", err)
	}
	if updated != 0 {
		t.Errorf("ユーザー名が異なる場合の変更件数 = %d、期待値 = 0", updated)
	}

	updated, err = repo.AnonymizeAndSoftDelete(ctx, userID, originalUsername, "anonymized_user", "anonymized_user@example.com")
	if err != nil {
		t.Fatalf("AnonymizeAndSoftDelete()のエラー = %v", err)
	}
	if updated != 1 {
		t.Errorf("変更した件数 = %d、期待値 = 1", updated)
	}

	var username, email string
	var deletedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, "SELECT username, email, deleted_at FROM users WHERE id = $1", int64(userID)).Scan(&username, &email, &deletedAt); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	if username != "anonymized_user" {
		t.Errorf("username = %q、期待値 = %q", username, "anonymized_user")
	}
	if email != "anonymized_user@example.com" {
		t.Errorf("email = %q、期待値 = %q", email, "anonymized_user@example.com")
	}
	if !deletedAt.Valid {
		t.Error("deleted_atが立っていない")
	}

	updated, err = repo.AnonymizeAndSoftDelete(ctx, userID, originalUsername, "another_user", "another_user@example.com")
	if err != nil {
		t.Fatalf("2回目のAnonymizeAndSoftDelete()のエラー = %v", err)
	}
	if updated != 0 {
		t.Errorf("論理削除済みのユーザーで変更した件数 = %d、期待値 = 0", updated)
	}
	if err := tx.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	if username != "anonymized_user" {
		t.Errorf("論理削除済みのユーザーのusername = %q、期待値 = %q", username, "anonymized_user")
	}
}

// TestUserRepository_LockForAccountDeletionは、ロックしたユーザーの匿名化する前の最新の属性と
// 論理削除済みかどうかを返し、存在しないユーザーではsql.ErrNoRowsを返すことをテスト
func TestUserRepository_LockForAccountDeletion(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	stripeSubscriberID := testutil.CreateTestStripeSubscriber(t, tx)
	activeUserID := testutil.NewUserBuilder(t, tx).WithLocale("en").WithStripeSubscriberID(&stripeSubscriberID).Build()
	deletedUserID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

	tests := []struct {
		name                   string
		userID                 model.UserID
		wantLocale             string
		wantStripeSubscriberID *model.StripeSubscriberID
		wantDeleted            bool
	}{
		{name: "論理削除されていないユーザー", userID: activeUserID, wantLocale: "en", wantStripeSubscriberID: &stripeSubscriberID, wantDeleted: false},
		{name: "論理削除済みのユーザー", userID: deletedUserID, wantLocale: "ja", wantDeleted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wantUsername, wantEmail string
			if err := tx.QueryRowContext(ctx, "SELECT username, email FROM users WHERE id = $1", int64(tt.userID)).Scan(&wantUsername, &wantEmail); err != nil {
				t.Fatalf("ユーザーの取得に失敗: %v", err)
			}

			user, err := repo.LockForAccountDeletion(ctx, tt.userID)
			if err != nil {
				t.Fatalf("LockForAccountDeletion()のエラー = %v", err)
			}
			if user.Username != wantUsername {
				t.Errorf("Username = %q、期待値 = %q", user.Username, wantUsername)
			}
			if user.Email != wantEmail {
				t.Errorf("Email = %q、期待値 = %q", user.Email, wantEmail)
			}
			if user.Locale != tt.wantLocale {
				t.Errorf("Locale = %q、期待値 = %q", user.Locale, tt.wantLocale)
			}
			if (user.StripeSubscriberID == nil) != (tt.wantStripeSubscriberID == nil) ||
				(user.StripeSubscriberID != nil && *user.StripeSubscriberID != *tt.wantStripeSubscriberID) {
				t.Errorf("StripeSubscriberID = %v、期待値 = %v", user.StripeSubscriberID, tt.wantStripeSubscriberID)
			}
			if user.Deleted != tt.wantDeleted {
				t.Errorf("Deleted = %v、期待値 = %v", user.Deleted, tt.wantDeleted)
			}
		})
	}

	t.Run("存在しないユーザーはsql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := repo.LockForAccountDeletion(ctx, model.UserID(-1))
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("エラー = %v、期待値 = sql.ErrNoRows", err)
		}
	})
}
