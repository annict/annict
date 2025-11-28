package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/schollz/progressbar/v3"
)

// CreateHeavyUserParams ヘビーユーザー作成のパラメータ
type CreateHeavyUserParams struct {
	Username          string  // デフォルト: "heavy_user"
	Password          string  // デフォルト: "password"
	EpisodeRecords    int     // ヘビーユーザーの視聴記録数（デフォルト: 10,000）
	FollowersCount    int     // フォロワー数（デフォルト: 1,000）
	FollowingCount    int     // フォロー数（デフォルト: 500）
	FolloweeRecords   int     // 各フォロイー（フォロワー）の視聴記録数（デフォルト: 100）
	RatingProbability float64 // 視聴記録に評価をつける確率（デフォルト: 0.7）
	BodyProbability   float64 // 視聴記録にコメントをつける確率（デフォルト: 0.3）
}

// CreateHeavyUserResult ヘビーユーザー作成の結果
type CreateHeavyUserResult struct {
	HeavyUserID        int64
	FollowerUserIDs    []int64
	FollowingUserIDs   []int64
	EpisodeRecordCount int
	FollowCount        int
}

// CreateHeavyUserUsecase ヘビーユーザー生成Usecase（シード専用）
// heavy_userという名前のユーザーを作成し、大量の視聴記録とフォロー関係を設定します
type CreateHeavyUserUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewCreateHeavyUserUsecase 新しいCreateHeavyUserUsecaseを作成
func NewCreateHeavyUserUsecase(db *sql.DB, queries *query.Queries) *CreateHeavyUserUsecase {
	return &CreateHeavyUserUsecase{
		db:      db,
		queries: queries,
	}
}

// Execute ヘビーユーザーを作成します
// 既存の作品とエピソードデータを使用して視聴記録を生成します
func (uc *CreateHeavyUserUsecase) Execute(ctx context.Context, params CreateHeavyUserParams) (*CreateHeavyUserResult, error) {
	// デフォルト値の設定
	if params.Username == "" {
		params.Username = "heavy_user"
	}
	if params.Password == "" {
		params.Password = "password"
	}
	if params.EpisodeRecords == 0 {
		params.EpisodeRecords = 10000
	}
	if params.FollowersCount == 0 {
		params.FollowersCount = 1000
	}
	if params.FollowingCount == 0 {
		params.FollowingCount = 500
	}
	if params.FolloweeRecords == 0 {
		params.FolloweeRecords = 100
	}
	if params.RatingProbability == 0 {
		params.RatingProbability = 0.7
	}
	if params.BodyProbability == 0 {
		params.BodyProbability = 0.3
	}

	// 1. heavy_userを作成
	fmt.Println("ヘビーユーザーを作成しています...")
	heavyUserID, err := uc.createHeavyUser(ctx, params.Username, params.Password)
	if err != nil {
		return nil, fmt.Errorf("ヘビーユーザー作成エラー: %w", err)
	}
	fmt.Printf("ヘビーユーザー作成完了（user_id: %d）\n", heavyUserID)

	// 2. フォロワーユーザー（heavy_userをフォローする人）を作成
	fmt.Printf("%d人のフォロワーユーザーを作成しています...\n", params.FollowersCount)
	followerUserIDs, err := uc.createFollowerUsers(ctx, params.FollowersCount)
	if err != nil {
		return nil, fmt.Errorf("フォロワーユーザー作成エラー: %w", err)
	}
	fmt.Printf("フォロワーユーザー作成完了（%d人）\n", len(followerUserIDs))

	// 3. フォローユーザー（heavy_userがフォローする人）を作成
	fmt.Printf("%d人のフォローユーザーを作成しています...\n", params.FollowingCount)
	followingUserIDs, err := uc.createFollowingUsers(ctx, params.FollowingCount)
	if err != nil {
		return nil, fmt.Errorf("フォローユーザー作成エラー: %w", err)
	}
	fmt.Printf("フォローユーザー作成完了（%d人）\n", len(followingUserIDs))

	// 4. heavy_userの視聴記録を作成
	fmt.Printf("ヘビーユーザーの視聴記録を%d件作成しています...\n", params.EpisodeRecords)
	heavyUserRecordCount, err := uc.createHeavyUserRecords(ctx, heavyUserID, params.EpisodeRecords, params.RatingProbability, params.BodyProbability)
	if err != nil {
		return nil, fmt.Errorf("ヘビーユーザー視聴記録作成エラー: %w", err)
	}
	fmt.Printf("ヘビーユーザー視聴記録作成完了（%d件）\n", heavyUserRecordCount)

	// 5. フォロー関係を作成（フォロワー → heavy_user、heavy_user → フォロー）
	fmt.Println("フォロー関係を作成しています...")
	followCount, err := uc.createFollowRelationships(ctx, heavyUserID, followerUserIDs, followingUserIDs)
	if err != nil {
		return nil, fmt.Errorf("フォロー関係作成エラー: %w", err)
	}
	fmt.Printf("フォロー関係作成完了（%d件）\n", followCount)

	// 6. 各フォロイー（フォロワー）の視聴記録を作成
	fmt.Printf("各フォロイーの視聴記録を作成しています（%d人 × %d件）...\n", params.FollowersCount, params.FolloweeRecords)
	if err := uc.createFolloweeRecords(ctx, followerUserIDs, params.FolloweeRecords, params.RatingProbability, params.BodyProbability); err != nil {
		return nil, fmt.Errorf("フォロイー視聴記録作成エラー: %w", err)
	}
	fmt.Printf("フォロイー視聴記録作成完了\n")

	return &CreateHeavyUserResult{
		HeavyUserID:        heavyUserID,
		FollowerUserIDs:    followerUserIDs,
		FollowingUserIDs:   followingUserIDs,
		EpisodeRecordCount: heavyUserRecordCount,
		FollowCount:        followCount,
	}, nil
}

// createHeavyUser heavy_userを作成します
func (uc *CreateHeavyUserUsecase) createHeavyUser(ctx context.Context, username, password string) (int64, error) {
	createUserUC := NewCreateUserUsecase(uc.db, uc.queries)
	email := fmt.Sprintf("%s@example.com", username)

	userParams := []CreateUserParams{
		{
			Username: username,
			Email:    email,
			Password: password,
			Locale:   "ja",
		},
	}

	results, err := createUserUC.ExecuteBatch(ctx, userParams, nil)
	if err != nil {
		return 0, err
	}

	return results[0].UserID, nil
}

// createFollowerUsers フォロワーユーザー（heavy_userをフォローする人）を作成します
func (uc *CreateHeavyUserUsecase) createFollowerUsers(ctx context.Context, count int) ([]int64, error) {
	createUserUC := NewCreateUserUsecase(uc.db, uc.queries)

	userParams := make([]CreateUserParams, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("follower_%d", i+1)
		email := fmt.Sprintf("%s@example.com", username)
		userParams[i] = CreateUserParams{
			Username: username,
			Email:    email,
			Password: "password",
			Locale:   "ja",
		}
	}

	// 進捗表示
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("フォロワーユーザー作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	results, err := createUserUC.ExecuteBatch(ctx, userParams, bar)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, len(results))
	for i, result := range results {
		userIDs[i] = result.UserID
	}

	return userIDs, nil
}

// createFollowingUsers フォローユーザー（heavy_userがフォローする人）を作成します
func (uc *CreateHeavyUserUsecase) createFollowingUsers(ctx context.Context, count int) ([]int64, error) {
	createUserUC := NewCreateUserUsecase(uc.db, uc.queries)

	userParams := make([]CreateUserParams, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("following_%d", i+1)
		email := fmt.Sprintf("%s@example.com", username)
		userParams[i] = CreateUserParams{
			Username: username,
			Email:    email,
			Password: "password",
			Locale:   "ja",
		}
	}

	// 進捗表示
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("フォローユーザー作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	results, err := createUserUC.ExecuteBatch(ctx, userParams, bar)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, len(results))
	for i, result := range results {
		userIDs[i] = result.UserID
	}

	return userIDs, nil
}

// createHeavyUserRecords heavy_userの視聴記録を作成します
func (uc *CreateHeavyUserUsecase) createHeavyUserRecords(ctx context.Context, userID int64, count int, ratingProbability, bodyProbability float64) (int, error) {
	// 既存のエピソードをランダムに取得
	episodes, err := uc.getRandomEpisodes(ctx, count)
	if err != nil {
		return 0, fmt.Errorf("エピソード取得エラー: %w", err)
	}

	// 視聴記録パラメータを作成
	recordParams := make([]CreateEpisodeRecordParams, len(episodes))
	for i, episode := range episodes {
		recordParams[i] = CreateEpisodeRecordParams{
			UserID:    userID,
			EpisodeID: episode.ID,
			WorkID:    episode.WorkID,
			Rating:    uc.generateRating(ratingProbability),
			Body:      uc.generateBody(bodyProbability),
			WatchedAt: uc.generateWatchedAt(),
		}
	}

	// 進捗表示
	bar := progressbar.NewOptions(len(recordParams),
		progressbar.OptionSetDescription("ヘビーユーザー視聴記録作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	createRecordUC := NewCreateEpisodeRecordUsecase(uc.db)
	_, err = createRecordUC.ExecuteBatch(ctx, recordParams, bar)
	if err != nil {
		return 0, err
	}

	return len(recordParams), nil
}

// createFollowRelationships フォロー関係を作成します
func (uc *CreateHeavyUserUsecase) createFollowRelationships(ctx context.Context, heavyUserID int64, followerUserIDs, followingUserIDs []int64) (int, error) {
	createFollowUC := NewCreateFollowUsecase(uc.db)

	// フォロワー → heavy_user のフォロー関係を作成
	followerFollows := make([]CreateFollowParams, len(followerUserIDs))
	for i, followerID := range followerUserIDs {
		followerFollows[i] = CreateFollowParams{
			FollowerID:  followerID,  // フォローする人
			FollowingID: heavyUserID, // フォローされる人（heavy_user）
		}
	}

	// heavy_user → フォロー のフォロー関係を作成
	heavyUserFollows := make([]CreateFollowParams, len(followingUserIDs))
	for i, followingID := range followingUserIDs {
		heavyUserFollows[i] = CreateFollowParams{
			FollowerID:  heavyUserID, // フォローする人（heavy_user）
			FollowingID: followingID, // フォローされる人
		}
	}

	// 全フォロー関係を結合
	allFollows := append(followerFollows, heavyUserFollows...)

	// 進捗表示
	bar := progressbar.NewOptions(len(allFollows),
		progressbar.OptionSetDescription("フォロー関係作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	_, err := createFollowUC.ExecuteBatch(ctx, allFollows, bar)
	if err != nil {
		return 0, err
	}

	return len(allFollows), nil
}

// createFolloweeRecords フォロイー（フォロワー）の視聴記録を作成します
func (uc *CreateHeavyUserUsecase) createFolloweeRecords(ctx context.Context, followerUserIDs []int64, recordsPerUser int, ratingProbability, bodyProbability float64) error {
	createRecordUC := NewCreateEpisodeRecordUsecase(uc.db)

	// 全フォロイーの視聴記録を一括で作成
	totalRecords := len(followerUserIDs) * recordsPerUser
	episodes, err := uc.getRandomEpisodes(ctx, totalRecords)
	if err != nil {
		return fmt.Errorf("エピソード取得エラー: %w", err)
	}

	// 各フォロイーに視聴記録を割り当て
	recordParams := make([]CreateEpisodeRecordParams, 0, totalRecords)
	episodeIndex := 0

	for _, userID := range followerUserIDs {
		for i := 0; i < recordsPerUser && episodeIndex < len(episodes); i++ {
			episode := episodes[episodeIndex]
			recordParams = append(recordParams, CreateEpisodeRecordParams{
				UserID:    userID,
				EpisodeID: episode.ID,
				WorkID:    episode.WorkID,
				Rating:    uc.generateRating(ratingProbability),
				Body:      uc.generateBody(bodyProbability),
				WatchedAt: uc.generateWatchedAt(),
			})
			episodeIndex++
		}
	}

	// 進捗表示
	bar := progressbar.NewOptions(len(recordParams),
		progressbar.OptionSetDescription("フォロイー視聴記録作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	_, err = createRecordUC.ExecuteBatch(ctx, recordParams, bar)
	if err != nil {
		return err
	}

	return nil
}

// episodeData エピソードデータの簡易構造体
type episodeData struct {
	ID     int64
	WorkID int64
}

// getRandomEpisodes ランダムなエピソードを取得します（重複あり）
// ORDER BY RANDOM()を使って1〜2回のクエリで大量のランダムエピソードを効率的に取得します
func (uc *CreateHeavyUserUsecase) getRandomEpisodes(ctx context.Context, count int) ([]episodeData, error) {
	// 全エピソード数を取得
	var totalEpisodes int64
	err := uc.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM episodes").Scan(&totalEpisodes)
	if err != nil {
		return nil, fmt.Errorf("エピソード数取得エラー: %w", err)
	}

	if totalEpisodes == 0 {
		return nil, fmt.Errorf("エピソードが存在しません。先に作品とエピソードを生成してください")
	}

	// 必要な数のエピソードをランダムに取得（ORDER BY RANDOM()を使用）
	// 重複を許容する仕様なので、totalEpisodesより多い場合は複数回に分けて取得
	episodes := make([]episodeData, 0, count)

	for len(episodes) < count {
		remaining := count - len(episodes)
		batchSize := remaining
		if batchSize > int(totalEpisodes) {
			batchSize = int(totalEpisodes)
		}

		// ORDER BY RANDOM()で一度に取得（大幅に高速化）
		rows, err := uc.db.QueryContext(ctx, `
			SELECT id, work_id FROM episodes ORDER BY RANDOM() LIMIT $1
		`, batchSize)
		if err != nil {
			return nil, fmt.Errorf("エピソード取得エラー: %w", err)
		}

		for rows.Next() {
			var episode episodeData
			if err := rows.Scan(&episode.ID, &episode.WorkID); err != nil {
				rows.Close()
				return nil, fmt.Errorf("エピソードスキャンエラー: %w", err)
			}
			episodes = append(episodes, episode)
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("行取得エラー: %w", err)
		}
	}

	return episodes, nil
}

// generateRating 評価を生成します（確率に基づいて nil または 1.0〜5.0 の値を返す）
// テストデータ生成用のため、暗号学的に安全な乱数は不要
func (uc *CreateHeavyUserUsecase) generateRating(probability float64) *float64 {
	// #nosec G404
	if rand.Float64() > probability {
		return nil
	}
	// #nosec G404
	rating := 1.0 + rand.Float64()*4.0 // 1.0〜5.0
	return &rating
}

// generateBody コメントを生成します（確率に基づいて nil または短いコメントを返す）
// テストデータ生成用のため、暗号学的に安全な乱数は不要
func (uc *CreateHeavyUserUsecase) generateBody(probability float64) *string {
	// #nosec G404
	if rand.Float64() > probability {
		return nil
	}
	body := gofakeit.Sentence(10) // 10単語のランダムな文章
	return &body
}

// generateWatchedAt 視聴日時を生成します（過去1年以内のランダムな日時）
// テストデータ生成用のため、暗号学的に安全な乱数は不要
func (uc *CreateHeavyUserUsecase) generateWatchedAt() time.Time {
	now := time.Now()
	daysAgo := rand.Intn(365) // #nosec G404
	return now.AddDate(0, 0, -daysAgo)
}
