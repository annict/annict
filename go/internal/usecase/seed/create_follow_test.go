package seed

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/internal/testutil"
)

// TestCreateFollowUsecase_ExecuteBatch はExecuteBatchメソッドのテスト
func TestCreateFollowUsecase_ExecuteBatch(t *testing.T) {
	// テストDBをセットアップ（トランザクションは各サブテストで作成）
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	uc := NewCreateFollowUsecase(db)

	// テストケース
	tests := []struct {
		name         string
		setupUsers   func(t *testing.T, tx *sql.Tx) []int64 // テストユーザーを作成する関数
		follows      func(userIDs []int64) []CreateFollowParams
		wantCount    int
		wantErr      bool
		validateFunc func(t *testing.T, tx *sql.Tx, userIDs []int64) // カスタム検証関数
	}{
		{
			name: "正常系: 3つのフォロー関係を作成",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsers(t, tx, 4) // 4人のユーザーを作成
			},
			follows: func(userIDs []int64) []CreateFollowParams {
				return []CreateFollowParams{
					{FollowerID: userIDs[0], FollowingID: userIDs[1]}, // user0がuser1をフォロー
					{FollowerID: userIDs[0], FollowingID: userIDs[2]}, // user0がuser2をフォロー
					{FollowerID: userIDs[1], FollowingID: userIDs[3]}, // user1がuser3をフォロー
				}
			},
			wantCount: 3,
			wantErr:   false,
			validateFunc: func(t *testing.T, tx *sql.Tx, userIDs []int64) {
				// user0: following_count=2, followers_count=0
				assertUserCounts(t, tx, userIDs[0], 2, 0)
				// user1: following_count=1, followers_count=1
				assertUserCounts(t, tx, userIDs[1], 1, 1)
				// user2: following_count=0, followers_count=1
				assertUserCounts(t, tx, userIDs[2], 0, 1)
				// user3: following_count=0, followers_count=1
				assertUserCounts(t, tx, userIDs[3], 0, 1)
			},
		},
		{
			name: "正常系: 1つのフォロー関係を作成",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsers(t, tx, 2) // 2人のユーザーを作成
			},
			follows: func(userIDs []int64) []CreateFollowParams {
				return []CreateFollowParams{
					{FollowerID: userIDs[0], FollowingID: userIDs[1]},
				}
			},
			wantCount: 1,
			wantErr:   false,
			validateFunc: func(t *testing.T, tx *sql.Tx, userIDs []int64) {
				// user0: following_count=1, followers_count=0
				assertUserCounts(t, tx, userIDs[0], 1, 0)
				// user1: following_count=0, followers_count=1
				assertUserCounts(t, tx, userIDs[1], 0, 1)
			},
		},
		{
			name: "正常系: 相互フォロー",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsers(t, tx, 2)
			},
			follows: func(userIDs []int64) []CreateFollowParams {
				return []CreateFollowParams{
					{FollowerID: userIDs[0], FollowingID: userIDs[1]}, // user0がuser1をフォロー
					{FollowerID: userIDs[1], FollowingID: userIDs[0]}, // user1がuser0をフォロー
				}
			},
			wantCount: 2,
			wantErr:   false,
			validateFunc: func(t *testing.T, tx *sql.Tx, userIDs []int64) {
				// user0: following_count=1, followers_count=1
				assertUserCounts(t, tx, userIDs[0], 1, 1)
				// user1: following_count=1, followers_count=1
				assertUserCounts(t, tx, userIDs[1], 1, 1)
			},
		},
		{
			name: "正常系: 1人が複数人をフォロー",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsers(t, tx, 5)
			},
			follows: func(userIDs []int64) []CreateFollowParams {
				return []CreateFollowParams{
					{FollowerID: userIDs[0], FollowingID: userIDs[1]},
					{FollowerID: userIDs[0], FollowingID: userIDs[2]},
					{FollowerID: userIDs[0], FollowingID: userIDs[3]},
					{FollowerID: userIDs[0], FollowingID: userIDs[4]},
				}
			},
			wantCount: 4,
			wantErr:   false,
			validateFunc: func(t *testing.T, tx *sql.Tx, userIDs []int64) {
				// user0: following_count=4, followers_count=0
				assertUserCounts(t, tx, userIDs[0], 4, 0)
				// user1-4: following_count=0, followers_count=1
				for i := 1; i <= 4; i++ {
					assertUserCounts(t, tx, userIDs[i], 0, 1)
				}
			},
		},
		{
			name: "正常系: 1人が複数人からフォローされる",
			setupUsers: func(t *testing.T, tx *sql.Tx) []int64 {
				return createTestUsers(t, tx, 5)
			},
			follows: func(userIDs []int64) []CreateFollowParams {
				return []CreateFollowParams{
					{FollowerID: userIDs[1], FollowingID: userIDs[0]},
					{FollowerID: userIDs[2], FollowingID: userIDs[0]},
					{FollowerID: userIDs[3], FollowingID: userIDs[0]},
					{FollowerID: userIDs[4], FollowingID: userIDs[0]},
				}
			},
			wantCount: 4,
			wantErr:   false,
			validateFunc: func(t *testing.T, tx *sql.Tx, userIDs []int64) {
				// user0: following_count=0, followers_count=4
				assertUserCounts(t, tx, userIDs[0], 0, 4)
				// user1-4: following_count=1, followers_count=0
				for i := 1; i <= 4; i++ {
					assertUserCounts(t, tx, userIDs[i], 1, 0)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// テストユーザーを作成
			userIDs := tt.setupUsers(t, tx)

			// フォロー関係のパラメータを生成
			followParams := tt.follows(userIDs)

			// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
			results, err := uc.ExecuteBatchWithTx(ctx, tx, followParams, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 結果の数をチェック
			if len(results) != tt.wantCount {
				t.Errorf("ExecuteBatch() returned %d results, want %d", len(results), tt.wantCount)
				return
			}

			// すべてのフォロー関係が作成されたことを確認
			for _, result := range results {
				if result.FollowID == 0 {
					t.Error("ExecuteBatch() returned FollowID = 0, want non-zero")
				}
			}

			// カスタム検証関数を実行
			if tt.validateFunc != nil {
				tt.validateFunc(t, tx, userIDs)
			}

			// フォロー関係がDBに保存されているか確認
			for _, params := range followParams {
				assertFollowExists(t, tx, params.FollowerID, params.FollowingID)
			}
		})
	}
}

// createTestUsers テスト用のユーザーを作成するヘルパー関数
func createTestUsers(t *testing.T, tx *sql.Tx, count int) []int64 {
	t.Helper()
	userIDs := make([]int64, count)
	for i := 0; i < count; i++ {
		userID := testutil.NewUserBuilder(t, tx).
			WithUsername(t.Name() + "_user_" + string(rune('0'+i))).
			WithEmail(t.Name() + "_user_" + string(rune('0'+i)) + "@example.com").
			Build()
		userIDs[i] = userID
	}
	return userIDs
}

// assertUserCounts ユーザーのフォロー/フォロワー数を検証するヘルパー関数
func assertUserCounts(t *testing.T, tx *sql.Tx, userID int64, expectedFollowingCount, expectedFollowersCount int) {
	t.Helper()

	query := `SELECT following_count, followers_count FROM users WHERE id = $1`
	var followingCount, followersCount int
	err := tx.QueryRow(query, userID).Scan(&followingCount, &followersCount)
	if err != nil {
		t.Fatalf("Failed to get user counts (user_id=%d): %v", userID, err)
	}

	if followingCount != expectedFollowingCount {
		t.Errorf("User %d: following_count = %d, want %d", userID, followingCount, expectedFollowingCount)
	}
	if followersCount != expectedFollowersCount {
		t.Errorf("User %d: followers_count = %d, want %d", userID, followersCount, expectedFollowersCount)
	}
}

// assertFollowExists フォロー関係がDBに存在することを検証するヘルパー関数
func assertFollowExists(t *testing.T, tx *sql.Tx, followerID, followingID int64) {
	t.Helper()

	query := `SELECT COUNT(*) FROM follows WHERE user_id = $1 AND following_id = $2`
	var count int
	err := tx.QueryRow(query, followerID, followingID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check follow existence (follower=%d, following=%d): %v", followerID, followingID, err)
	}

	if count == 0 {
		t.Errorf("Follow relationship does not exist: follower_id=%d, following_id=%d", followerID, followingID)
	}
}
