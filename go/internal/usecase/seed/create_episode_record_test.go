package seed

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/internal/testutil"
)

// TestCreateEpisodeRecordUsecase_ExecuteBatch はExecuteBatchメソッドのテスト
func TestCreateEpisodeRecordUsecase_ExecuteBatch(t *testing.T) {
	// テストDBをセットアップ
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	uc := NewCreateEpisodeRecordUsecase(db)

	// テストケース
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, tx *sql.Tx) ([]CreateEpisodeRecordParams, int64, int64, int64, int64) // ユーザーID、作品ID、エピソードID1、エピソードID2を返す
		wantCount     int
		wantErr       bool
		checkCounters bool
	}{
		{
			name: "正常系: 1件の視聴記録を作成",
			setupFunc: func(t *testing.T, tx *sql.Tx) ([]CreateEpisodeRecordParams, int64, int64, int64, int64) {
				// ユーザーを作成
				userID := testutil.NewUserBuilder(t, tx).
					WithUsername("test_user_single").
					WithEmail("test_single@example.com").
					Build()

				// 作品を作成
				workID := testutil.NewWorkBuilder(t, tx).
					WithTitle("テストアニメ").
					Build()

				// エピソードを作成
				episodeID1 := testutil.NewEpisodeBuilder(t, tx, workID).
					WithNumber("1").
					WithTitle("第1話").
					Build()

				// 視聴記録パラメータ
				rating := 4.5
				body := "面白かった！"
				records := []CreateEpisodeRecordParams{
					{
						UserID:    userID,
						EpisodeID: episodeID1,
						WorkID:    workID,
						Rating:    &rating,
						Body:      &body,
						WatchedAt: time.Now(),
					},
				}

				return records, userID, workID, episodeID1, 0
			},
			wantCount:     1,
			wantErr:       false,
			checkCounters: true,
		},
		{
			name: "正常系: 複数の視聴記録を作成",
			setupFunc: func(t *testing.T, tx *sql.Tx) ([]CreateEpisodeRecordParams, int64, int64, int64, int64) {
				// ユーザーを作成
				userID := testutil.NewUserBuilder(t, tx).
					WithUsername("test_user_multi").
					WithEmail("multi@example.com").
					Build()

				// 作品を作成
				workID := testutil.NewWorkBuilder(t, tx).
					WithTitle("テストアニメ2").
					Build()

				// エピソードを作成
				episodeID1 := testutil.NewEpisodeBuilder(t, tx, workID).
					WithNumber("1").
					WithTitle("第1話").
					Build()

				episodeID2 := testutil.NewEpisodeBuilder(t, tx, workID).
					WithNumber("2").
					WithTitle("第2話").
					Build()

				// 視聴記録パラメータ（2件）
				rating1 := 4.0
				body1 := "良かった"
				rating2 := 5.0
				body2 := "最高！"
				records := []CreateEpisodeRecordParams{
					{
						UserID:    userID,
						EpisodeID: episodeID1,
						WorkID:    workID,
						Rating:    &rating1,
						Body:      &body1,
						WatchedAt: time.Now(),
					},
					{
						UserID:    userID,
						EpisodeID: episodeID2,
						WorkID:    workID,
						Rating:    &rating2,
						Body:      &body2,
						WatchedAt: time.Now(),
					},
				}

				return records, userID, workID, episodeID1, episodeID2
			},
			wantCount:     2,
			wantErr:       false,
			checkCounters: true,
		},
		{
			name: "正常系: 評価なし・コメントなし",
			setupFunc: func(t *testing.T, tx *sql.Tx) ([]CreateEpisodeRecordParams, int64, int64, int64, int64) {
				// ユーザーを作成
				userID := testutil.NewUserBuilder(t, tx).
					WithUsername("test_user_no_rating").
					WithEmail("norating@example.com").
					Build()

				// 作品を作成
				workID := testutil.NewWorkBuilder(t, tx).
					WithTitle("テストアニメ3").
					Build()

				// エピソードを作成
				episodeID1 := testutil.NewEpisodeBuilder(t, tx, workID).
					WithNumber("1").
					WithTitle("第1話").
					Build()

				// 視聴記録パラメータ（評価なし・コメントなし）
				records := []CreateEpisodeRecordParams{
					{
						UserID:    userID,
						EpisodeID: episodeID1,
						WorkID:    workID,
						Rating:    nil,
						Body:      nil,
						WatchedAt: time.Now(),
					},
				}

				return records, userID, workID, episodeID1, 0
			},
			wantCount:     1,
			wantErr:       false,
			checkCounters: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// テストデータをセットアップ
			records, userID, workID, episodeID1, episodeID2 := tt.setupFunc(t, tx)

			// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
			results, err := uc.ExecuteBatchWithTx(ctx, tx, records, nil)

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

			// 結果の内容をチェック
			for i, result := range results {
				if result.RecordID == 0 {
					t.Errorf("result[%d].RecordID is 0, want non-zero", i)
				}
				if result.EpisodeRecordID == 0 {
					t.Errorf("result[%d].EpisodeRecordID is 0, want non-zero", i)
				}
				if result.ActivityID == 0 {
					t.Errorf("result[%d].ActivityID is 0, want non-zero", i)
				}
				if result.ActivityGroupID == 0 {
					t.Errorf("result[%d].ActivityGroupID is 0, want non-zero", i)
				}
			}

			// カウンターをチェック
			if tt.checkCounters {
				// users.episode_records_countをチェック
				var episodeRecordsCount int32
				err = tx.QueryRowContext(ctx, "SELECT episode_records_count FROM users WHERE id = $1", userID).Scan(&episodeRecordsCount)
				if err != nil {
					t.Errorf("Failed to get episode_records_count: %v", err)
				}
				if episodeRecordsCount != int32(tt.wantCount) {
					t.Errorf("users.episode_records_count = %d, want %d", episodeRecordsCount, tt.wantCount)
				}

				// works.records_countをチェック
				var recordsCount int32
				err = tx.QueryRowContext(ctx, "SELECT records_count FROM works WHERE id = $1", workID).Scan(&recordsCount)
				if err != nil {
					t.Errorf("Failed to get records_count: %v", err)
				}
				// 作品に対する視聴記録数は、すべての視聴記録が同じ作品に対するものなので、tt.wantCount と一致
				if recordsCount != int32(tt.wantCount) {
					t.Errorf("works.records_count = %d, want %d", recordsCount, tt.wantCount)
				}

				// episodes.episode_records_countをチェック
				var episode1RecordsCount int32
				err = tx.QueryRowContext(ctx, "SELECT episode_records_count FROM episodes WHERE id = $1", episodeID1).Scan(&episode1RecordsCount)
				if err != nil {
					t.Errorf("Failed to get episode_records_count: %v", err)
				}
				// 最初のエピソードには必ず1件の視聴記録がある
				if episode1RecordsCount < 1 {
					t.Errorf("episodes.episode_records_count = %d, want at least 1", episode1RecordsCount)
				}

				// 2件目のエピソードがある場合はそちらもチェック
				if episodeID2 != 0 {
					var episode2RecordsCount int32
					err = tx.QueryRowContext(ctx, "SELECT episode_records_count FROM episodes WHERE id = $1", episodeID2).Scan(&episode2RecordsCount)
					if err != nil {
						t.Errorf("Failed to get episode_records_count for episode2: %v", err)
					}
					if episode2RecordsCount < 1 {
						t.Errorf("episodes[2].episode_records_count = %d, want at least 1", episode2RecordsCount)
					}
				}

				// activity_groups.activities_countをチェック
				var activitiesCount int32
				err = tx.QueryRowContext(ctx, `
					SELECT activities_count FROM activity_groups
					WHERE user_id = $1 AND itemable_type = 'EpisodeRecord' AND single = false
					ORDER BY created_at DESC
					LIMIT 1
				`, userID).Scan(&activitiesCount)
				if err != nil {
					t.Errorf("Failed to get activities_count: %v", err)
				}
				if activitiesCount != int32(tt.wantCount) {
					t.Errorf("activity_groups.activities_count = %d, want %d", activitiesCount, tt.wantCount)
				}
			}
		})
	}
}

// TestCreateEpisodeRecordUsecase_RatingState は rating_state の設定をテスト
func TestCreateEpisodeRecordUsecase_RatingState(t *testing.T) {
	// テストDBをセットアップ
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	uc := NewCreateEpisodeRecordUsecase(db)

	// テストケース
	tests := []struct {
		name            string
		rating          *float64
		wantRatingState string
	}{
		{
			name:            "評価なし",
			rating:          nil,
			wantRatingState: "",
		},
		{
			name:            "bad (rating < 2.5)",
			rating:          floatPtr(2.0),
			wantRatingState: "bad",
		},
		{
			name:            "average (2.5 <= rating < 3.5)",
			rating:          floatPtr(3.0),
			wantRatingState: "average",
		},
		{
			name:            "good (3.5 <= rating < 4.5)",
			rating:          floatPtr(4.0),
			wantRatingState: "good",
		},
		{
			name:            "great (4.5 <= rating)",
			rating:          floatPtr(5.0),
			wantRatingState: "great",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// ユーザーを作成（一意のメールアドレスを使用）
			userID := testutil.NewUserBuilder(t, tx).
				WithUsername(fmt.Sprintf("rating_test_user_%d", i)).
				WithEmail(fmt.Sprintf("rating_%d@example.com", i)).
				Build()

			// 作品を作成
			workID := testutil.NewWorkBuilder(t, tx).
				WithTitle(fmt.Sprintf("評価テストアニメ_%d", i)).
				Build()

			// エピソードを作成
			episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
				WithNumber("1").
				WithTitle("第1話").
				Build()

			// 視聴記録パラメータ
			records := []CreateEpisodeRecordParams{
				{
					UserID:    userID,
					EpisodeID: episodeID,
					WorkID:    workID,
					Rating:    tt.rating,
					Body:      nil,
					WatchedAt: time.Now(),
				},
			}

			// ExecuteBatchWithTxを実行
			results, err := uc.ExecuteBatchWithTx(ctx, tx, records, nil)
			if err != nil {
				t.Fatalf("ExecuteBatchWithTx() error = %v", err)
			}

			// rating_stateをチェック
			var ratingState string
			err = tx.QueryRowContext(ctx, "SELECT rating_state FROM episode_records WHERE id = $1", results[0].EpisodeRecordID).Scan(&ratingState)
			if err != nil {
				t.Fatalf("Failed to get rating_state: %v", err)
			}

			if ratingState != tt.wantRatingState {
				t.Errorf("rating_state = %q, want %q", ratingState, tt.wantRatingState)
			}
		})
	}
}

// floatPtr はfloat64のポインタを返すヘルパー関数
func floatPtr(f float64) *float64 {
	return &f
}
