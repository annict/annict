package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/schollz/progressbar/v3"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
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
	HeavyUserID        model.UserID
	FollowerUserIDs    []model.UserID
	FollowingUserIDs   []model.UserID
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
func (uc *CreateHeavyUserUsecase) createHeavyUser(ctx context.Context, username, password string) (model.UserID, error) {
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
func (uc *CreateHeavyUserUsecase) createFollowerUsers(ctx context.Context, count int) ([]model.UserID, error) {
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

	userIDs := make([]model.UserID, len(results))
	for i, result := range results {
		userIDs[i] = result.UserID
	}

	return userIDs, nil
}

// createFollowingUsers フォローユーザー（heavy_userがフォローする人）を作成します
func (uc *CreateHeavyUserUsecase) createFollowingUsers(ctx context.Context, count int) ([]model.UserID, error) {
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

	userIDs := make([]model.UserID, len(results))
	for i, result := range results {
		userIDs[i] = result.UserID
	}

	return userIDs, nil
}

// createHeavyUserRecords heavy_userの視聴記録を作成します
func (uc *CreateHeavyUserUsecase) createHeavyUserRecords(ctx context.Context, userID model.UserID, count int, ratingProbability, bodyProbability float64) (int, error) {
	// Build one owner entry per record, then pass the ordered list to the chunked writer.
	//
	// [Ja] 記録 1 件につき 1 エントリを持つ受け手の一覧にして、チャンク処理に渡す。
	recordOwners := make([]model.UserID, count)
	for i := range recordOwners {
		recordOwners[i] = userID
	}

	// 進捗表示
	bar := progressbar.NewOptions(len(recordOwners),
		progressbar.OptionSetDescription("ヘビーユーザー視聴記録作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	return uc.createEpisodeRecordsForOwners(ctx, recordOwners, ratingProbability, bodyProbability, bar)
}

// createFollowRelationships フォロー関係を作成します
func (uc *CreateHeavyUserUsecase) createFollowRelationships(ctx context.Context, heavyUserID model.UserID, followerUserIDs, followingUserIDs []model.UserID) (int, error) {
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
func (uc *CreateHeavyUserUsecase) createFolloweeRecords(ctx context.Context, followerUserIDs []model.UserID, recordsPerUser int, ratingProbability, bodyProbability float64) error {
	// Build one owner entry per record, assigning recordsPerUser entries to each followee.
	//
	// [Ja] 各フォロイーに recordsPerUser 件ずつ割り当てた、記録 1 件につき 1 エントリの
	// 受け手の一覧を作る。
	recordOwners := make([]model.UserID, 0, len(followerUserIDs)*recordsPerUser)
	for _, userID := range followerUserIDs {
		for range recordsPerUser {
			recordOwners = append(recordOwners, userID)
		}
	}

	// 進捗表示
	bar := progressbar.NewOptions(len(recordOwners),
		progressbar.OptionSetDescription("フォロイー視聴記録作成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	_, err := uc.createEpisodeRecordsForOwners(ctx, recordOwners, ratingProbability, bodyProbability, bar)

	return err
}

// episodeRecordCommitChunkSize is the number of episode records written per commit. Each
// chunk uses one transaction that selects its episodes and inserts the records.
//
// [Ja] episodeRecordCommitChunkSize は視聴記録を何件ごとにコミットするかを決める。
// 1 チャンクが 1 トランザクションになり、その中でエピソードの抽選と記録の INSERT を行う。
const episodeRecordCommitChunkSize = 5000

// createEpisodeRecordsForOwners creates one episode record for each entry in recordOwners.
// Selecting the referenced episodes and inserting the records share a transaction. If they
// used separate transactions, another test or process could delete a selected episode before
// the insert and make the activities foreign key fail.
//
// [Ja] createEpisodeRecordsForOwners は recordOwners の 1 エントリにつき 1 件の視聴記録を
// 作成する。記録が参照するエピソードの抽選と INSERT は同じトランザクションで行う。別々の
// トランザクションで行うと、抽選から INSERT までの間に他のテストや処理がそのエピソードを
// 削除でき、activities の外部キー違反になる。
func (uc *CreateHeavyUserUsecase) createEpisodeRecordsForOwners(
	ctx context.Context,
	recordOwners []model.UserID,
	ratingProbability, bodyProbability float64,
	bar *progressbar.ProgressBar,
) (int, error) {
	created := 0
	for start := 0; start < len(recordOwners); start += episodeRecordCommitChunkSize {
		end := min(start+episodeRecordCommitChunkSize, len(recordOwners))

		count, err := uc.createEpisodeRecordChunk(ctx, recordOwners[start:end], ratingProbability, bodyProbability, bar)
		if err != nil {
			return created, err
		}
		created += count
	}

	return created, nil
}

// createEpisodeRecordChunk selects episodes and creates one chunk of records in a single
// transaction.
//
// [Ja] createEpisodeRecordChunk は 1 トランザクションで、チャンク分のエピソードを抽選し、
// そのエピソードに対する視聴記録を作成する。
func (uc *CreateHeavyUserUsecase) createEpisodeRecordChunk(
	ctx context.Context,
	recordOwners []model.UserID,
	ratingProbability, bodyProbability float64,
	bar *progressbar.ProgressBar,
) (int, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("トランザクション開始エラー: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	episodes, err := uc.getRandomEpisodes(ctx, tx, len(recordOwners))
	if err != nil {
		return 0, fmt.Errorf("エピソード取得エラー: %w", err)
	}

	recordParams := make([]CreateEpisodeRecordParams, len(recordOwners))
	for i, userID := range recordOwners {
		recordParams[i] = CreateEpisodeRecordParams{
			UserID:    userID,
			EpisodeID: episodes[i].ID,
			WorkID:    episodes[i].WorkID,
			Rating:    uc.generateRating(ratingProbability),
			Body:      uc.generateBody(bodyProbability),
			WatchedAt: uc.generateWatchedAt(),
		}
	}

	createRecordUC := NewCreateEpisodeRecordUsecase(uc.db)
	if _, err := createRecordUC.ExecuteBatchWithTx(ctx, tx, recordParams, bar); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("トランザクションコミットエラー: %w", err)
	}

	return len(recordParams), nil
}

// episodeData エピソードデータの簡易構造体
type episodeData struct {
	ID     model.EpisodeID
	WorkID model.WorkID
}

var errNoEpisodesAvailable = errors.New("エピソードが存在しません。先に作品とエピソードを生成してください")

// getRandomEpisodes returns the requested number of random episodes, allowing repeats when
// the request exceeds the number of available rows. It runs inside the caller's transaction
// and locks every selected row with FOR KEY SHARE, so another transaction cannot delete a
// returned episode before the caller inserts its references and commits.
//
// [Ja] getRandomEpisodes は指定件数のエピソードをランダムに返し、利用可能な行数を超える場合は
// 重複を許可する。呼び出し元のトランザクションで実行し、取得した行を FOR KEY SHARE でロック
// する。これにより、返したエピソードは呼び出し元が参照行を INSERT してコミットするまで、他の
// トランザクションから削除されない。
func (uc *CreateHeavyUserUsecase) getRandomEpisodes(ctx context.Context, tx *sql.Tx, count int) ([]episodeData, error) {
	// 全エピソード数を取得
	var totalEpisodes int64
	err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM episodes").Scan(&totalEpisodes)
	if err != nil {
		return nil, fmt.Errorf("エピソード数取得エラー: %w", err)
	}

	if totalEpisodes == 0 {
		return nil, errNoEpisodesAvailable
	}

	return getRandomEpisodesForKnownTotal(ctx, tx, count, totalEpisodes)
}

// getRandomEpisodesForKnownTotal selects rows after the caller has counted the available
// episodes. The observed total can become stale before this query under READ COMMITTED, so a
// batch that returns no rows is an error rather than another iteration of the same loop.
//
// [Ja] getRandomEpisodesForKnownTotal は、呼び出し元が利用可能なエピソード数を数えた後で行を
// 抽選する。READ COMMITTED では観測した件数がこのクエリまでに古くなる可能性があるため、1 件も
// 返さないバッチは同じループの再試行ではなくエラーとして扱う。
func getRandomEpisodesForKnownTotal(
	ctx context.Context,
	tx *sql.Tx,
	count int,
	totalEpisodes int64,
) ([]episodeData, error) {
	// Fetch the requested episodes in random batches. When count exceeds totalEpisodes,
	// repeated batches provide the allowed duplicate entries.
	//
	// [Ja] 必要な数のエピソードをランダムなバッチで取得する。count が totalEpisodes を超える
	// 場合は複数のバッチから、許可されている重複エントリを得る。
	episodes := make([]episodeData, 0, count)

	for len(episodes) < count {
		selectedBefore := len(episodes)
		remaining := count - selectedBefore
		batchSize := remaining
		if batchSize > int(totalEpisodes) {
			batchSize = int(totalEpisodes)
		}

		// Fetch one batch with ORDER BY RANDOM().
		//
		// [Ja] ORDER BY RANDOM() で 1 バッチを取得する。
		rows, err := tx.QueryContext(ctx, `
			SELECT id, work_id FROM episodes ORDER BY RANDOM() LIMIT $1 FOR KEY SHARE
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

		if len(episodes) == selectedBefore {
			return nil, errNoEpisodesAvailable
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
