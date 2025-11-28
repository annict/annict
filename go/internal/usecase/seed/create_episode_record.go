package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// CreateEpisodeRecordParams 視聴記録作成のパラメータ
type CreateEpisodeRecordParams struct {
	UserID    int64
	EpisodeID int64
	WorkID    int64
	Rating    *float64 // nilの場合は評価なし
	Body      *string  // nilの場合はコメントなし
	WatchedAt time.Time
}

// CreateEpisodeRecordResult 視聴記録作成の結果
type CreateEpisodeRecordResult struct {
	RecordID        int64
	EpisodeRecordID int64
	ActivityID      int64
	ActivityGroupID int64
}

// CreateEpisodeRecordUsecase 視聴記録生成Usecase（シード専用、バルクインサート対応）
type CreateEpisodeRecordUsecase struct {
	db *sql.DB
}

// NewCreateEpisodeRecordUsecase 新しいCreateEpisodeRecordUsecaseを作成
func NewCreateEpisodeRecordUsecase(db *sql.DB) *CreateEpisodeRecordUsecase {
	return &CreateEpisodeRecordUsecase{
		db: db,
	}
}

// ExecuteBatch 複数の視聴記録をバッチで作成します
// 1000件ごとにコミットしてパフォーマンスを最適化します
func (uc *CreateEpisodeRecordUsecase) ExecuteBatch(ctx context.Context, records []CreateEpisodeRecordParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeRecordResult, error) {
	return uc.executeBatchWithTx(ctx, nil, records, progressBar)
}

// ExecuteBatchWithTx 複数の視聴記録をバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateEpisodeRecordUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, records []CreateEpisodeRecordParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeRecordResult, error) {
	return uc.executeBatchWithTx(ctx, tx, records, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateEpisodeRecordUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, records []CreateEpisodeRecordParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeRecordResult, error) {
	results := make([]CreateEpisodeRecordResult, 0, len(records))

	// 既存トランザクションがある場合は、バッチサイズを無視して全件処理
	if existingTx != nil {
		// マルチ行INSERTのチャンクサイズ（100件ずつ）
		multiInsertChunkSize := 100
		for i := 0; i < len(records); i += multiInsertChunkSize {
			end := i + multiInsertChunkSize
			if end > len(records) {
				end = len(records)
			}
			chunk := records[i:end]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleEpisodeRecords(ctx, existingTx, chunk)
			if err != nil {
				return nil, fmt.Errorf("視聴記録マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}
		return results, nil
	}

	// 既存トランザクションがない場合は、5000件ごとにコミット
	commitBatchSize := 5000
	multiInsertChunkSize := 300

	for i := 0; i < len(records); i += commitBatchSize {
		end := i + commitBatchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		// トランザクション開始
		tx, err := uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()

		// バッチ内の視聴記録を100件ずつマルチ行INSERTで作成
		for j := 0; j < len(batch); j += multiInsertChunkSize {
			chunkEnd := j + multiInsertChunkSize
			if chunkEnd > len(batch) {
				chunkEnd = len(batch)
			}
			chunk := batch[j:chunkEnd]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleEpisodeRecords(ctx, tx, chunk)
			if err != nil {
				return nil, fmt.Errorf("視聴記録マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}

		// コミット
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("トランザクションコミットエラー: %w", err)
		}
	}

	return results, nil
}

// createMultipleEpisodeRecords 複数の視聴記録をマルチ行INSERTで作成します（トランザクション内）
// Record（親）→ EpisodeRecord（子）→ ActivityGroup → Activity の順で作成します
func (uc *CreateEpisodeRecordUsecase) createMultipleEpisodeRecords(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams) ([]CreateEpisodeRecordResult, error) {
	if len(recordsList) == 0 {
		return []CreateEpisodeRecordResult{}, nil
	}

	// 1. Record（親レコード）をマルチ行INSERTで作成
	recordIDs, err := uc.createMultipleRecordsInDB(ctx, tx, recordsList)
	if err != nil {
		return nil, fmt.Errorf("recordsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 2. EpisodeRecord（子レコード）をマルチ行INSERTで作成
	episodeRecordIDs, err := uc.createMultipleEpisodeRecordsInDB(ctx, tx, recordsList, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("episode_recordsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 3. ActivityGroup（アクティビティグループ）を作成（ユーザーごとに1つ）
	activityGroupIDs, err := uc.createActivityGroupsForUsers(ctx, tx, recordsList)
	if err != nil {
		return nil, fmt.Errorf("activity_groupsテーブルへのINSERT エラー: %w", err)
	}

	// 4. Activity（アクティビティ）をマルチ行INSERTで作成
	activityIDs, err := uc.createMultipleActivitiesInDB(ctx, tx, recordsList, episodeRecordIDs, activityGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("activitiesテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 5. カウンターを更新（users.episode_records_count, works.records_count, episodes.episode_records_count）
	if err := uc.updateCounters(ctx, tx, recordsList); err != nil {
		return nil, fmt.Errorf("カウンター更新エラー: %w", err)
	}

	// 6. 結果を返す
	results := make([]CreateEpisodeRecordResult, len(recordsList))
	for i := range recordsList {
		results[i] = CreateEpisodeRecordResult{
			RecordID:        recordIDs[i],
			EpisodeRecordID: episodeRecordIDs[i],
			ActivityID:      activityIDs[i],
			ActivityGroupID: activityGroupIDs[recordsList[i].UserID],
		}
	}

	return results, nil
}

// createMultipleRecordsInDB recordsテーブルに複数のレコードをマルチ行INSERTで挿入します
func (uc *CreateEpisodeRecordUsecase) createMultipleRecordsInDB(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams) ([]int64, error) {
	queryBuilder := `INSERT INTO records (
		user_id, work_id, aasm_state, impressions_count,
		created_at, updated_at, watched_at
	) VALUES `

	values := []interface{}{}
	now := time.Now()

	for i, params := range recordsList {
		if i > 0 {
			queryBuilder += ", "
		}

		offset := i * 7
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7)

		values = append(values,
			params.UserID,
			params.WorkID,
			"published",      // aasm_state
			0,                // impressions_count
			now,              // created_at
			now,              // updated_at
			params.WatchedAt, // watched_at
		)
	}

	queryBuilder += " RETURNING id"

	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("recordsテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	recordIDs := make([]int64, 0, len(recordsList))
	for rows.Next() {
		var recordID int64
		if err := rows.Scan(&recordID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		recordIDs = append(recordIDs, recordID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return recordIDs, nil
}

// createMultipleEpisodeRecordsInDB episode_recordsテーブルに複数のレコードをマルチ行INSERTで挿入します
func (uc *CreateEpisodeRecordUsecase) createMultipleEpisodeRecordsInDB(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams, recordIDs []int64) ([]int64, error) {
	queryBuilder := `INSERT INTO episode_records (
		user_id, episode_id, work_id, record_id,
		body, rating, rating_state,
		aasm_state, locale,
		comments_count, likes_count,
		created_at, updated_at,
		modify_body, shared_twitter, shared_facebook,
		twitter_click_count, facebook_click_count
	) VALUES `

	values := []interface{}{}
	now := time.Now()

	for i, params := range recordsList {
		if i > 0 {
			queryBuilder += ", "
		}

		offset := i * 18
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9,
			offset+10, offset+11, offset+12, offset+13, offset+14, offset+15, offset+16, offset+17, offset+18)

		// rating_stateを決定（ratingがnilの場合は空文字、それ以外は評価に応じて設定）
		var ratingState string
		if params.Rating != nil {
			rating := *params.Rating
			if rating < 2.5 {
				ratingState = "bad"
			} else if rating < 3.5 {
				ratingState = "average"
			} else if rating < 4.5 {
				ratingState = "good"
			} else {
				ratingState = "great"
			}
		} else {
			ratingState = ""
		}

		values = append(values,
			params.UserID,
			params.EpisodeID,
			params.WorkID,
			recordIDs[i],  // record_id
			params.Body,   // body (nil可)
			params.Rating, // rating (nil可)
			ratingState,   // rating_state
			"published",   // aasm_state
			"other",       // locale
			0,             // comments_count
			0,             // likes_count
			now,           // created_at
			now,           // updated_at
			false,         // modify_body
			false,         // shared_twitter
			false,         // shared_facebook
			0,             // twitter_click_count
			0,             // facebook_click_count
		)
	}

	queryBuilder += " RETURNING id"

	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("episode_recordsテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	episodeRecordIDs := make([]int64, 0, len(recordsList))
	for rows.Next() {
		var episodeRecordID int64
		if err := rows.Scan(&episodeRecordID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		episodeRecordIDs = append(episodeRecordIDs, episodeRecordID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return episodeRecordIDs, nil
}

// createActivityGroupsForUsers ユーザーごとにActivityGroupを作成します
// 既存のActivityGroupがあればそれを返し、なければ新規作成します
func (uc *CreateEpisodeRecordUsecase) createActivityGroupsForUsers(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams) (map[int64]int64, error) {
	// ユニークなユーザーIDを抽出
	userIDSet := make(map[int64]bool)
	for _, params := range recordsList {
		userIDSet[params.UserID] = true
	}

	activityGroupIDs := make(map[int64]int64)
	now := time.Now()

	for userID := range userIDSet {
		// 既存のActivityGroupを確認
		var existingID sql.NullInt64
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM activity_groups
			WHERE user_id = $1 AND itemable_type = 'EpisodeRecord' AND single = false
			ORDER BY created_at DESC
			LIMIT 1
		`, userID).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("activity_groupsの検索エラー: %w", err)
		}

		if existingID.Valid {
			// 既存のActivityGroupがある場合は再利用
			activityGroupIDs[userID] = existingID.Int64
		} else {
			// 新規作成
			var activityGroupID int64
			err := tx.QueryRowContext(ctx, `
				INSERT INTO activity_groups (
					user_id, itemable_type, single, activities_count,
					created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id
			`, userID, "EpisodeRecord", false, 0, now, now).Scan(&activityGroupID)

			if err != nil {
				return nil, fmt.Errorf("activity_groupsテーブルへのINSERT エラー: %w", err)
			}
			activityGroupIDs[userID] = activityGroupID
		}
	}

	return activityGroupIDs, nil
}

// createMultipleActivitiesInDB activitiesテーブルに複数のレコードをマルチ行INSERTで挿入します
func (uc *CreateEpisodeRecordUsecase) createMultipleActivitiesInDB(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams, episodeRecordIDs []int64, activityGroupIDs map[int64]int64) ([]int64, error) {
	queryBuilder := `INSERT INTO activities (
		user_id, trackable_id, trackable_type,
		episode_record_id, work_id, episode_id,
		activity_group_id, action,
		created_at, updated_at
	) VALUES `

	values := []interface{}{}
	now := time.Now()

	for i, params := range recordsList {
		if i > 0 {
			queryBuilder += ", "
		}

		offset := i * 10
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10)

		values = append(values,
			params.UserID,
			episodeRecordIDs[i],             // trackable_id
			"EpisodeRecord",                 // trackable_type
			episodeRecordIDs[i],             // episode_record_id
			params.WorkID,                   // work_id
			params.EpisodeID,                // episode_id
			activityGroupIDs[params.UserID], // activity_group_id
			"create",                        // action
			now,                             // created_at
			now,                             // updated_at
		)
	}

	queryBuilder += " RETURNING id"

	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("activitiesテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	activityIDs := make([]int64, 0, len(recordsList))
	for rows.Next() {
		var activityID int64
		if err := rows.Scan(&activityID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		activityIDs = append(activityIDs, activityID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return activityIDs, nil
}

// updateCounters カウンターを更新します
// - users.episode_records_count: ユーザーの視聴記録数
// - works.records_count: 作品の視聴記録数
// - episodes.episode_records_count: エピソードの視聴記録数
// - activity_groups.activities_count: アクティビティグループのアクティビティ数
func (uc *CreateEpisodeRecordUsecase) updateCounters(ctx context.Context, tx *sql.Tx, recordsList []CreateEpisodeRecordParams) error {
	// ユーザーごとの視聴記録数をカウント
	userCounts := make(map[int64]int)
	for _, params := range recordsList {
		userCounts[params.UserID]++
	}

	// 作品ごとの視聴記録数をカウント
	workCounts := make(map[int64]int)
	for _, params := range recordsList {
		workCounts[params.WorkID]++
	}

	// エピソードごとの視聴記録数をカウント
	episodeCounts := make(map[int64]int)
	for _, params := range recordsList {
		episodeCounts[params.EpisodeID]++
	}

	// users.episode_records_countを更新
	for userID, count := range userCounts {
		_, err := tx.ExecContext(ctx, `
			UPDATE users
			SET episode_records_count = episode_records_count + $1
			WHERE id = $2
		`, count, userID)
		if err != nil {
			return fmt.Errorf("users.episode_records_count更新エラー（user_id: %d）: %w", userID, err)
		}
	}

	// works.records_countを更新
	for workID, count := range workCounts {
		_, err := tx.ExecContext(ctx, `
			UPDATE works
			SET records_count = records_count + $1
			WHERE id = $2
		`, count, workID)
		if err != nil {
			return fmt.Errorf("works.records_count更新エラー（work_id: %d）: %w", workID, err)
		}
	}

	// episodes.episode_records_countを更新
	for episodeID, count := range episodeCounts {
		_, err := tx.ExecContext(ctx, `
			UPDATE episodes
			SET episode_records_count = episode_records_count + $1
			WHERE id = $2
		`, count, episodeID)
		if err != nil {
			return fmt.Errorf("episodes.episode_records_count更新エラー（episode_id: %d）: %w", episodeID, err)
		}
	}

	// activity_groups.activities_countを更新
	for userID, count := range userCounts {
		_, err := tx.ExecContext(ctx, `
			UPDATE activity_groups
			SET activities_count = activities_count + $1
			WHERE user_id = $2 AND itemable_type = 'EpisodeRecord' AND single = false
		`, count, userID)
		if err != nil {
			return fmt.Errorf("activity_groups.activities_count更新エラー（user_id: %d）: %w", userID, err)
		}
	}

	return nil
}
