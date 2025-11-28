package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/seed"
	seedusecase "github.com/annict/annict/internal/usecase/seed"
	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/lib/pq"
	"github.com/schollz/progressbar/v3"
)

func main() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("シードデータを生成します", "env", cfg.Env)

	// データベース接続
	db, err := sql.Open("postgres", cfg.DatabaseDSN())
	if err != nil {
		slog.Error("データベースへの接続に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("データベース接続のクローズに失敗しました", "error", err)
		}
	}()

	// データベース接続確認
	if err := db.Ping(); err != nil {
		slog.Error("データベースへの疎通確認に失敗しました", "error", err)
		os.Exit(1)
	}

	// 接続先データベースの確認
	var dbName string
	if err := db.QueryRow("SELECT current_database()").Scan(&dbName); err != nil {
		slog.Error("データベースクエリに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("接続先データベース", "database", dbName)

	ctx := context.Background()

	// 既存データのクリーンアップ
	slog.Info("既存データをクリーンアップしています...")
	if err := cleanupExistingData(ctx, db); err != nil {
		slog.Error("既存データのクリーンアップに失敗しました", "error", err)
		os.Exit(1)
	}
	slog.Info("既存データのクリーンアップが完了しました")

	// シードデータ生成開始
	startTime := time.Now()
	fmt.Println()
	slog.Info("=== シードデータ生成を開始します ===")
	fmt.Println()

	// 乱数生成器の初期化（再現性のためにシード値を固定）
	// テストデータ生成用のため、暗号学的に安全な乱数は不要
	rnd := rand.New(rand.NewSource(42)) // #nosec G404
	gofakeit.SetGlobalFaker(gofakeit.New(42))

	// フェーズ1: ユーザー生成
	userIDs, err := generateUsers(ctx, db, rnd, 30000)
	if err != nil {
		slog.Error("ユーザー生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ1-1: プロフィール画像クリーンアップ（S3バケット内の孤立した画像を削除）
	if err := cleanupProfileImages(ctx, cfg); err != nil {
		slog.Error("プロフィール画像クリーンアップに失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ1-2: プロフィール画像生成
	if err := generateProfileImages(ctx, db, cfg, userIDs); err != nil {
		slog.Error("プロフィール画像生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ2: 作品生成
	workIDs, err := generateWorks(ctx, db, rnd, 10000)
	if err != nil {
		slog.Error("作品生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ3: エピソード生成
	episodeIDs, err := generateEpisodes(ctx, db, rnd, workIDs)
	if err != nil {
		slog.Error("エピソード生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ4: 視聴記録生成
	if err := generateEpisodeRecords(ctx, db, rnd, userIDs, episodeIDs, workIDs, 1000000); err != nil {
		slog.Error("視聴記録生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ5-1: 作品画像クリーンアップ（S3バケット内の孤立した画像を削除）
	if err := cleanupWorkImages(ctx, cfg); err != nil {
		slog.Error("作品画像クリーンアップに失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ5-2: 作品画像生成
	if err := generateWorkImages(ctx, db, cfg, workIDs, userIDs, rnd); err != nil {
		slog.Error("作品画像生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ6: フォロー関係生成
	if err := generateFollows(ctx, db, rnd, userIDs, 100000); err != nil {
		slog.Error("フォロー関係生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ7: OAuthトークン生成
	if err := generateOAuthTokens(ctx, db, userIDs, 100); err != nil {
		slog.Error("OAuthトークン生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// フェーズ8: ヘビーユーザー生成
	if err := generateHeavyUser(ctx, db); err != nil {
		slog.Error("ヘビーユーザー生成に失敗しました", "error", err)
		os.Exit(1)
	}

	// 完了
	elapsed := time.Since(startTime)
	fmt.Println()
	slog.Info("=== シードデータ生成が完了しました ===")
	slog.Info("実行時間", "elapsed", elapsed)
	fmt.Println()
}

// cleanupExistingData は既存データを削除します
func cleanupExistingData(ctx context.Context, db *sql.DB) error {
	// シードデータ生成対象のテーブル一覧
	// 外部キー制約があるため、外部キー制約を一時的に無効化してから削除する
	tables := []string{
		"activities",
		"oauth_access_tokens",
		"oauth_access_grants",
		"oauth_applications",
		"episode_records",
		"records",
		"work_images",
		"follows",
		"episodes",
		"works",
		"profiles",
		"users",
	}

	// トランザクション開始
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() {
		// トランザクションがコミット済みの場合、Rollbackはエラーを返すが問題ない
		_ = tx.Rollback()
	}()

	// 外部キー制約を一時的に無効化（レプリケーションロールを使用）
	// session_replication_role = 'replica' に設定すると、トリガーと外部キー制約が無効化される
	if _, err := tx.ExecContext(ctx, "SET session_replication_role = 'replica'"); err != nil {
		return fmt.Errorf("外部キー制約の無効化に失敗: %w", err)
	}

	// 進捗表示
	bar := progressbar.NewOptions(len(tables),
		progressbar.OptionSetDescription("テーブルのクリーンアップ"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// 各テーブルからデータを削除
	for _, table := range tables {
		// テーブル名は固定リストから取得しているため安全
		query := fmt.Sprintf("DELETE FROM %s", table) // #nosec G201
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("テーブル %s のクリーンアップに失敗: %w", table, err)
		}
		_ = bar.Add(1) // 進捗表示のエラーは無視
	}

	// 外部キー制約を再有効化
	if _, err := tx.ExecContext(ctx, "SET session_replication_role = 'origin'"); err != nil {
		return fmt.Errorf("外部キー制約の再有効化に失敗: %w", err)
	}

	// コミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行

	return nil
}

// generateUsers はユーザーを生成します
func generateUsers(ctx context.Context, db *sql.DB, rnd *rand.Rand, count int) ([]int64, error) {
	slog.Info("フェーズ1: ユーザー生成", "count", count)

	queries := query.New(db)
	uc := seedusecase.NewCreateUserUsecase(db, queries)

	// ユーザーパラメータを生成
	users := make([]seedusecase.CreateUserParams, count)
	for i := 0; i < count; i++ {
		users[i] = seedusecase.CreateUserParams{
			Username: fmt.Sprintf("user%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Password: "password",
			Locale:   "ja",
		}
	}

	// 進捗バー
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("ユーザー生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	results, err := uc.ExecuteBatch(ctx, users, bar)
	if err != nil {
		return nil, fmt.Errorf("ユーザー生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("ユーザー生成完了", "count", len(results))

	// UserIDのリストを返す
	userIDs := make([]int64, len(results))
	for i, result := range results {
		userIDs[i] = result.UserID
	}

	return userIDs, nil
}

// generateWorks は作品を生成します
func generateWorks(ctx context.Context, db *sql.DB, rnd *rand.Rand, count int) ([]int64, error) {
	slog.Info("フェーズ2: 作品生成", "count", count)

	uc := seedusecase.NewCreateWorkUsecase(db)

	// 作品パラメータを生成
	works := make([]seedusecase.CreateWorkParams, count)
	for i := 0; i < count; i++ {
		works[i] = seedusecase.GenerateRandomWorkParams(rnd)
	}

	// 進捗バー
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("作品生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	results, err := uc.ExecuteBatch(ctx, works, bar)
	if err != nil {
		return nil, fmt.Errorf("作品生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("作品生成完了", "count", len(results))

	// WorkIDのリストを返す
	workIDs := make([]int64, len(results))
	for i, result := range results {
		workIDs[i] = result.WorkID
	}

	return workIDs, nil
}

// generateEpisodes はエピソードを生成します（作品あたり平均12話）
func generateEpisodes(ctx context.Context, db *sql.DB, rnd *rand.Rand, workIDs []int64) (map[int64][]int64, error) {
	slog.Info("フェーズ3: エピソード生成", "workCount", len(workIDs), "avgEpisodes", 12)

	uc := seedusecase.NewCreateEpisodeUsecase(db)

	// 各作品に対してエピソードを生成
	allEpisodes := []seedusecase.CreateEpisodeParams{}
	for _, workID := range workIDs {
		// ランダムに1〜24話（平均12話）
		episodeCount := rnd.Intn(24) + 1
		episodes := seedusecase.GenerateEpisodeParamsForWork(rnd, workID, episodeCount)
		allEpisodes = append(allEpisodes, episodes...)
	}

	slog.Info("エピソード総数", "count", len(allEpisodes))

	// 進捗バー
	bar := progressbar.NewOptions(len(allEpisodes),
		progressbar.OptionSetDescription("エピソード生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	results, err := uc.ExecuteBatch(ctx, allEpisodes, bar)
	if err != nil {
		return nil, fmt.Errorf("エピソード生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("エピソード生成完了", "count", len(results))

	// Work ID -> Episode IDs のマップを作成
	episodesByWork := make(map[int64][]int64)
	for i, result := range results {
		workID := allEpisodes[i].WorkID
		episodesByWork[workID] = append(episodesByWork[workID], result.EpisodeID)
	}

	return episodesByWork, nil
}

// generateEpisodeRecords は視聴記録を生成します
func generateEpisodeRecords(ctx context.Context, db *sql.DB, rnd *rand.Rand, userIDs []int64, episodesByWork map[int64][]int64, workIDs []int64, count int) error {
	slog.Info("フェーズ4: 視聴記録生成", "count", count)

	uc := seedusecase.NewCreateEpisodeRecordUsecase(db)

	// 視聴記録パラメータを生成
	records := make([]seedusecase.CreateEpisodeRecordParams, count)
	for i := 0; i < count; i++ {
		// ランダムなユーザー
		userID := userIDs[rnd.Intn(len(userIDs))]

		// ランダムな作品
		workID := workIDs[rnd.Intn(len(workIDs))]

		// その作品のエピソードからランダムに選択
		episodes := episodesByWork[workID]
		if len(episodes) == 0 {
			continue
		}
		episodeID := episodes[rnd.Intn(len(episodes))]

		// 評価とコメント（確率的に設定）
		var rating *float64
		if rnd.Float64() < 0.7 { // 70%の確率で評価をつける
			r := float64(rnd.Intn(50))/10.0 + 1.0 // 1.0〜5.0
			rating = &r
		}

		var body *string
		if rnd.Float64() < 0.3 { // 30%の確率でコメントをつける
			b := seed.GenerateJapaneseEpisodeRecordBody(rnd)
			body = &b
		}

		// 視聴日時（過去1年間のランダムな日時）
		watchedAt := time.Now().AddDate(0, 0, -rnd.Intn(365))

		records[i] = seedusecase.CreateEpisodeRecordParams{
			UserID:    userID,
			EpisodeID: episodeID,
			WorkID:    workID,
			Rating:    rating,
			Body:      body,
			WatchedAt: watchedAt,
		}
	}

	// 進捗バー
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("視聴記録生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	_, err := uc.ExecuteBatch(ctx, records, bar)
	if err != nil {
		return fmt.Errorf("視聴記録生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("視聴記録生成完了", "count", count)

	return nil
}

// cleanupWorkImages はCloudflare R2上の作品画像を削除します
func cleanupWorkImages(ctx context.Context, cfg *config.Config) error {
	slog.Info("フェーズ5-1: 作品画像クリーンアップ")

	uc := seedusecase.NewCleanupWorkImagesUsecase(
		cfg.S3Endpoint,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.S3Region,
		cfg.S3BucketName,
	)

	if err := uc.Execute(ctx); err != nil {
		return fmt.Errorf("作品画像クリーンアップエラー: %w", err)
	}

	slog.Info("作品画像クリーンアップ完了")
	fmt.Println()
	return nil
}

// generateWorkImages は作品画像を生成します
func generateWorkImages(ctx context.Context, db *sql.DB, cfg *config.Config, workIDs []int64, userIDs []int64, rnd *rand.Rand) error {
	slog.Info("フェーズ5-2: 作品画像生成", "count", len(workIDs))

	queries := query.New(db)
	uc := seedusecase.NewCreateWorkImageUsecase(
		db,
		queries,
		cfg.S3Endpoint,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.S3Region,
		cfg.S3BucketName,
	)

	// 作品画像パラメータを生成（全作品に画像を付与）
	params := make([]seedusecase.CreateWorkImageParams, len(workIDs))
	for i, workID := range workIDs {
		// ランダムなユーザーを選択（作成者として）
		userID := userIDs[rnd.Intn(len(userIDs))]
		params[i] = seedusecase.CreateWorkImageParams{
			WorkID: workID,
			UserID: userID,
		}
	}

	// 進捗バー
	bar := progressbar.NewOptions(len(params),
		progressbar.OptionSetDescription("作品画像生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	_, err := uc.ExecuteBatch(ctx, params, bar)
	if err != nil {
		return fmt.Errorf("作品画像生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("作品画像生成完了", "count", len(params))

	return nil
}

// generateFollows はフォロー関係を生成します
func generateFollows(ctx context.Context, db *sql.DB, rnd *rand.Rand, userIDs []int64, count int) error {
	slog.Info("フェーズ6: フォロー関係生成", "count", count)

	uc := seedusecase.NewCreateFollowUsecase(db)

	// フォロー関係パラメータを生成（重複チェック用のマップ）
	follows := make([]seedusecase.CreateFollowParams, 0, count)
	followMap := make(map[string]bool) // "follower_id:following_id" をキーとする

	for len(follows) < count {
		followerID := userIDs[rnd.Intn(len(userIDs))]
		followingID := userIDs[rnd.Intn(len(userIDs))]

		// 自分自身はフォローできない
		if followerID == followingID {
			continue
		}

		// 重複チェック
		key := fmt.Sprintf("%d:%d", followerID, followingID)
		if followMap[key] {
			continue
		}

		follows = append(follows, seedusecase.CreateFollowParams{
			FollowerID:  followerID,
			FollowingID: followingID,
		})
		followMap[key] = true
	}

	// 進捗バー
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("フォロー関係生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	_, err := uc.ExecuteBatch(ctx, follows, bar)
	if err != nil {
		return fmt.Errorf("フォロー関係生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("フォロー関係生成完了", "count", count)

	return nil
}

// generateOAuthTokens はOAuthトークンを生成します
func generateOAuthTokens(ctx context.Context, db *sql.DB, userIDs []int64, count int) error {
	slog.Info("フェーズ7: OAuthトークン生成", "count", count)

	queries := query.New(db)
	uc := seedusecase.NewCreateOAuthTokenUsecase(db, queries)

	// トークンを付与するユーザーをランダムに選択
	selectedUserIDs := make([]int64, count)
	for i := 0; i < count; i++ {
		selectedUserIDs[i] = userIDs[i%len(userIDs)]
	}

	params := seedusecase.CreateOAuthTokenParams{
		ApplicationName: "Test Application",
		ApplicationUID:  "test-app-uid",
		RedirectURI:     "http://localhost:3000/callback",
		Scopes:          "read write",
		TokenCount:      count,
		UserIDs:         selectedUserIDs,
	}

	// 進捗バー（アプリケーション作成 + トークン数）
	bar := progressbar.NewOptions(count+1,
		progressbar.OptionSetDescription("OAuthトークン生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// 実行
	_, err := uc.Execute(ctx, params, bar)
	if err != nil {
		return fmt.Errorf("OAuthトークン生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("OAuthトークン生成完了", "count", count)

	return nil
}

// generateHeavyUser はヘビーユーザーを生成します
func generateHeavyUser(ctx context.Context, db *sql.DB) error {
	slog.Info("フェーズ8: ヘビーユーザー生成")

	queries := query.New(db)
	uc := seedusecase.NewCreateHeavyUserUsecase(db, queries)

	params := seedusecase.CreateHeavyUserParams{
		Username:          "heavy_user",
		Password:          "password",
		EpisodeRecords:    10000,
		FollowersCount:    1000,
		FollowingCount:    500,
		FolloweeRecords:   100,
		RatingProbability: 0.7,
		BodyProbability:   0.3,
	}

	// 実行
	result, err := uc.Execute(ctx, params)
	if err != nil {
		return fmt.Errorf("ヘビーユーザー生成エラー: %w", err)
	}

	slog.Info("ヘビーユーザー生成完了",
		"userID", result.HeavyUserID,
		"episodeRecordCount", result.EpisodeRecordCount,
		"followCount", result.FollowCount)

	return nil
}

// cleanupProfileImages はCloudflare R2上のプロフィール画像を削除します
func cleanupProfileImages(ctx context.Context, cfg *config.Config) error {
	slog.Info("プロフィール画像クリーンアップ")

	uc := seedusecase.NewCleanupProfileImagesUsecase(
		cfg.S3Endpoint,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.S3Region,
		cfg.S3BucketName,
	)

	if err := uc.Execute(ctx); err != nil {
		return fmt.Errorf("プロフィール画像クリーンアップエラー: %w", err)
	}

	slog.Info("プロフィール画像クリーンアップ完了")
	fmt.Println()
	return nil
}

// generateProfileImages はプロフィール画像を生成します
func generateProfileImages(ctx context.Context, db *sql.DB, cfg *config.Config, userIDs []int64) error {
	slog.Info("プロフィール画像生成", "count", len(userIDs))

	queries := query.New(db)

	// 全ユーザーのプロフィールIDを取得
	profiles, err := queries.ListAllProfiles(ctx)
	if err != nil {
		return fmt.Errorf("プロフィール一覧取得エラー: %w", err)
	}

	slog.Info("プロフィール総数", "count", len(profiles))

	uc := seedusecase.NewCreateProfileImageUsecase(
		db,
		queries,
		cfg.S3Endpoint,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.S3Region,
		cfg.S3BucketName,
	)

	// プロフィール画像パラメータを生成（全プロフィールに画像を付与）
	params := make([]seedusecase.CreateProfileImageParams, len(profiles))
	for i, profile := range profiles {
		params[i] = seedusecase.CreateProfileImageParams{
			ProfileID: profile.ID,
			UserID:    profile.UserID,
		}
	}

	// 進捗バー
	bar := progressbar.NewOptions(len(params),
		progressbar.OptionSetDescription("プロフィール画像生成"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// バッチ実行
	_, err = uc.ExecuteBatch(ctx, params, bar)
	if err != nil {
		return fmt.Errorf("プロフィール画像生成エラー: %w", err)
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("プロフィール画像生成完了", "count", len(params))

	return nil
}
