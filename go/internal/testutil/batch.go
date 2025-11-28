package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// ProgressCallback は進捗を通知するコールバック関数
// current: 現在の処理数、total: 総処理数
type ProgressCallback func(current, total int)

// BatchBuildWorks は複数の作品データをバッチで作成します
// testing.T に依存しないため、seed コマンドからも使用可能です
func BatchBuildWorks(ctx context.Context, tx *sql.Tx, count int, callback ProgressCallback) ([]int64, error) {
	ids := make([]int64, count)

	query := `
		INSERT INTO works (
			title, title_kana, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10, $11
		) RETURNING id
	`

	for i := 0; i < count; i++ {
		var id int64
		err := tx.QueryRowContext(
			ctx,
			query,
			fmt.Sprintf("作品 %d", i+1), // title
			"",                        // title_kana
			0,                         // media (0 = tv)
			"",                        // official_site_url
			"",                        // wikipedia_url
			2024,                      // season_year
			SeasonSpring,              // season_name
			100,                       // watchers_count
			12,                        // episodes_count
			time.Now(),                // created_at
			time.Now(),                // updated_at
		).Scan(&id)

		if err != nil {
			return nil, fmt.Errorf("作品 %d の作成に失敗: %w", i+1, err)
		}

		ids[i] = id

		// 進捗コールバックを呼び出し
		if callback != nil {
			callback(i+1, count)
		}
	}

	return ids, nil
}

// BatchBuildUsers は複数のユーザーデータをバッチで作成します
// testing.T に依存しないため、seed コマンドからも使用可能です
func BatchBuildUsers(ctx context.Context, tx *sql.Tx, count int, callback ProgressCallback) ([]int64, error) {
	ids := make([]int64, count)

	query := `
		INSERT INTO users (
			username, email, role, locale,
			created_at, updated_at,
			encrypted_password, sign_in_count,
			time_zone, allowed_locales
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8,
			$9, $10
		) RETURNING id
	`

	for i := 0; i < count; i++ {
		var id int64
		err := tx.QueryRowContext(
			ctx,
			query,
			fmt.Sprintf("user_%d", i+1),             // username
			fmt.Sprintf("user_%d@example.com", i+1), // email
			0,                                       // role (0 = user)
			"ja",                                    // locale
			time.Now(),                              // created_at
			time.Now(),                              // updated_at
			"encrypted_password",                    // encrypted_password
			0,                                       // sign_in_count
			"Asia/Tokyo",                            // time_zone
			pq.Array([]string{"ja", "en"}),          // allowed_locales
		).Scan(&id)

		if err != nil {
			return nil, fmt.Errorf("ユーザー %d の作成に失敗: %w", i+1, err)
		}

		ids[i] = id

		// 進捗コールバックを呼び出し
		if callback != nil {
			callback(i+1, count)
		}
	}

	return ids, nil
}

// BatchBuildEpisodes は複数のエピソードデータをバッチで作成します
// testing.T に依存しないため、seed コマンドからも使用可能です
func BatchBuildEpisodes(ctx context.Context, tx *sql.Tx, workID int64, count int, callback ProgressCallback) ([]int64, error) {
	ids := make([]int64, count)

	query := `
		INSERT INTO episodes (
			work_id, number, sort_number, title,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6
		) RETURNING id
	`

	for i := 0; i < count; i++ {
		var id int64
		err := tx.QueryRowContext(
			ctx,
			query,
			workID,                   // work_id
			fmt.Sprintf("%d", i+1),   // number
			(i+1)*10,                 // sort_number
			fmt.Sprintf("第%d話", i+1), // title
			time.Now(),               // created_at
			time.Now(),               // updated_at
		).Scan(&id)

		if err != nil {
			return nil, fmt.Errorf("エピソード %d の作成に失敗: %w", i+1, err)
		}

		ids[i] = id

		// 進捗コールバックを呼び出し
		if callback != nil {
			callback(i+1, count)
		}
	}

	return ids, nil
}
