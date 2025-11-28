package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// CreateFollowParams フォロー関係作成のパラメータ
type CreateFollowParams struct {
	FollowerID  int64 // フォローする人（user_id）
	FollowingID int64 // フォローされる人（following_id）
}

// CreateFollowResult フォロー関係作成の結果
type CreateFollowResult struct {
	FollowID int64
}

// CreateFollowUsecase フォロー関係生成Usecase（シード専用、バルクインサート対応）
type CreateFollowUsecase struct {
	db *sql.DB
}

// NewCreateFollowUsecase 新しいCreateFollowUsecaseを作成
func NewCreateFollowUsecase(db *sql.DB) *CreateFollowUsecase {
	return &CreateFollowUsecase{
		db: db,
	}
}

// ExecuteBatch 複数のフォロー関係をバッチで作成します
// 1000件ごとにコミットしてパフォーマンスを最適化します
func (uc *CreateFollowUsecase) ExecuteBatch(ctx context.Context, follows []CreateFollowParams, progressBar *progressbar.ProgressBar) ([]CreateFollowResult, error) {
	return uc.executeBatchWithTx(ctx, nil, follows, progressBar)
}

// ExecuteBatchWithTx 複数のフォロー関係をバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateFollowUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, follows []CreateFollowParams, progressBar *progressbar.ProgressBar) ([]CreateFollowResult, error) {
	return uc.executeBatchWithTx(ctx, tx, follows, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateFollowUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, follows []CreateFollowParams, progressBar *progressbar.ProgressBar) ([]CreateFollowResult, error) {
	results := make([]CreateFollowResult, 0, len(follows))

	// 既存トランザクションがある場合は、バッチサイズを無視して全件処理
	if existingTx != nil {
		// マルチ行INSERTのチャンクサイズ（100件ずつ）
		multiInsertChunkSize := 100
		for i := 0; i < len(follows); i += multiInsertChunkSize {
			end := i + multiInsertChunkSize
			if end > len(follows) {
				end = len(follows)
			}
			chunk := follows[i:end]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleFollows(ctx, existingTx, chunk)
			if err != nil {
				return nil, fmt.Errorf("フォロー関係マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}

		// カウンター更新（既存トランザクション内）
		if err := uc.updateCounters(ctx, existingTx, follows); err != nil {
			return nil, fmt.Errorf("カウンター更新エラー: %w", err)
		}

		return results, nil
	}

	// 既存トランザクションがない場合は、5000件ごとにコミット
	commitBatchSize := 5000
	multiInsertChunkSize := 500

	for i := 0; i < len(follows); i += commitBatchSize {
		end := i + commitBatchSize
		if end > len(follows) {
			end = len(follows)
		}
		batch := follows[i:end]

		// トランザクション開始
		tx, err := uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()

		// バッチ内のフォロー関係を100件ずつマルチ行INSERTで作成
		for j := 0; j < len(batch); j += multiInsertChunkSize {
			chunkEnd := j + multiInsertChunkSize
			if chunkEnd > len(batch) {
				chunkEnd = len(batch)
			}
			chunk := batch[j:chunkEnd]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleFollows(ctx, tx, chunk)
			if err != nil {
				return nil, fmt.Errorf("フォロー関係マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}

		// カウンター更新
		if err := uc.updateCounters(ctx, tx, batch); err != nil {
			return nil, fmt.Errorf("カウンター更新エラー: %w", err)
		}

		// コミット
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("トランザクションコミットエラー: %w", err)
		}
	}

	return results, nil
}

// createMultipleFollows 複数のフォロー関係をマルチ行INSERTで作成します（トランザクション内）
func (uc *CreateFollowUsecase) createMultipleFollows(ctx context.Context, tx *sql.Tx, followsList []CreateFollowParams) ([]CreateFollowResult, error) {
	if len(followsList) == 0 {
		return []CreateFollowResult{}, nil
	}

	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO follows (
		user_id, following_id,
		created_at, updated_at
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, params := range followsList {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は4個のパラメータ）
		offset := i * 4
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4)

		// パラメータを追加
		values = append(values,
			params.FollowerID,  // user_id (フォローする人)
			params.FollowingID, // following_id (フォローされる人)
			now,                // created_at
			now,                // updated_at
		)
	}

	queryBuilder += " RETURNING id"

	// マルチ行INSERTを実行
	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("followsテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	// 挿入されたIDを取得
	followIDs := make([]CreateFollowResult, 0, len(followsList))
	for rows.Next() {
		var followID int64
		if err := rows.Scan(&followID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		followIDs = append(followIDs, CreateFollowResult{FollowID: followID})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return followIDs, nil
}

// updateCounters users.followers_count と users.following_count を更新します
func (uc *CreateFollowUsecase) updateCounters(ctx context.Context, tx *sql.Tx, follows []CreateFollowParams) error {
	// フォロワー数とフォロー数をカウント
	followerCounts := make(map[int64]int)  // following_id -> count (フォローされた回数)
	followingCounts := make(map[int64]int) // follower_id -> count (フォローした回数)

	for _, follow := range follows {
		followerCounts[follow.FollowingID]++
		followingCounts[follow.FollowerID]++
	}

	// followers_countを更新（フォローされた人）
	for userID, count := range followerCounts {
		query := `UPDATE users SET followers_count = followers_count + $1 WHERE id = $2`
		if _, err := tx.ExecContext(ctx, query, count, userID); err != nil {
			return fmt.Errorf("followers_count更新エラー (user_id=%d): %w", userID, err)
		}
	}

	// following_countを更新（フォローした人）
	for userID, count := range followingCounts {
		query := `UPDATE users SET following_count = following_count + $1 WHERE id = $2`
		if _, err := tx.ExecContext(ctx, query, count, userID); err != nil {
			return fmt.Errorf("following_count更新エラー (user_id=%d): %w", userID, err)
		}
	}

	return nil
}
