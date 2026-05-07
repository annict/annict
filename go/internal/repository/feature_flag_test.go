package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

const testFlagName model.FeatureFlagName = "go_test_feature"

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_DeviceToken はデバイストークンでフラグが有効な場合にtrueを返すことをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_DeviceToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// デバイストークンによるフラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithDeviceToken("test_device_token_123").
		WithName(string(testFlagName)).
		Build()

	// デバイストークンでフラグが有効であることを確認
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "test_device_token_123", 0, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if !enabled {
		t.Error("デバイストークンで設定されたフラグが有効と判定されるべきです")
	}
}

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_UserID はユーザーIDでフラグが有効な場合にtrueを返すことをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_UserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// テスト用ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	// ユーザーIDによるフラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithUserID(userID).
		WithName(string(testFlagName)).
		Build()

	// ユーザーIDでフラグが有効であることを確認
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "", userID, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if !enabled {
		t.Error("ユーザーIDで設定されたフラグが有効と判定されるべきです")
	}
}

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_BothMatch はデバイストークンとユーザーIDの両方が一致する場合にtrueを返すことをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_BothMatch(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// テスト用ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	// デバイストークンによるフラグのみ作成（ユーザーIDのフラグはなし）
	testutil.NewFeatureFlagBuilder(t, tx).
		WithDeviceToken("device_both_test").
		WithName(string(testFlagName)).
		Build()

	// デバイストークンが一致するのでtrueを返す
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "device_both_test", userID, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if !enabled {
		t.Error("デバイストークンが一致する場合にフラグが有効と判定されるべきです")
	}
}

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_NotEnabled はフラグが存在しない場合にfalseを返すことをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_NotEnabled(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// フラグを作成しない状態でチェック
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "unknown_token", 99999, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if enabled {
		t.Error("フラグが存在しない場合はfalseを返すべきです")
	}
}

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_DifferentFlag は異なるフラグ名では無効であることをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_DifferentFlag(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// あるフラグ名でデバイストークンによるフラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithDeviceToken("device_diff_flag").
		WithName("go_feature_a").
		Build()

	// 異なるフラグ名ではfalseを返す
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "device_diff_flag", 0, "go_feature_b")
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if enabled {
		t.Error("異なるフラグ名ではfalseを返すべきです")
	}
}

// TestFeatureFlagRepository_IsEnabledByDeviceOrUser_EmptyParams は空のパラメータでfalseを返すことをテスト
func TestFeatureFlagRepository_IsEnabledByDeviceOrUser_EmptyParams(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// フラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithDeviceToken("some_token").
		WithName(string(testFlagName)).
		Build()

	// 空のデバイストークンと0のユーザーIDではfalseを返す
	enabled, err := repo.IsEnabledByDeviceOrUser(context.Background(), "", 0, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if enabled {
		t.Error("空のパラメータではfalseを返すべきです")
	}
}

// TestFeatureFlagRepository_IsEnabled はユーザーIDでフラグを判定できることをテスト
func TestFeatureFlagRepository_IsEnabled(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewFeatureFlagRepository(queries)

	// テスト用ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	// ユーザーIDによるフラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithUserID(userID).
		WithName(string(testFlagName)).
		Build()

	// IsEnabledでフラグが有効であることを確認
	enabled, err := repo.IsEnabled(context.Background(), userID, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledに失敗: %v", err)
	}
	if !enabled {
		t.Error("ユーザーIDで設定されたフラグが有効と判定されるべきです")
	}

	// 異なるユーザーIDではfalseを返す
	enabled, err = repo.IsEnabled(context.Background(), 99999, testFlagName)
	if err != nil {
		t.Fatalf("IsEnabledに失敗: %v", err)
	}
	if enabled {
		t.Error("異なるユーザーIDではfalseを返すべきです")
	}
}

// TestFeatureFlagRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestFeatureFlagRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewFeatureFlagRepository(queries)

	// WithTxでトランザクション内のRepositoryを取得
	repoWithTx := repo.WithTx(tx)

	// トランザクション内でフラグを作成
	testutil.NewFeatureFlagBuilder(t, tx).
		WithDeviceToken("withtx_device").
		WithName(string(testFlagName)).
		Build()

	// WithTxで取得したRepositoryからフラグを確認できる
	enabled, err := repoWithTx.IsEnabledByDeviceOrUser(context.Background(), "withtx_device", 0, testFlagName)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでIsEnabledByDeviceOrUserに失敗: %v", err)
	}
	if !enabled {
		t.Error("WithTxで取得したRepositoryでフラグが有効と判定されるべきです")
	}
}
