package seed

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/schollz/progressbar/v3"
)

// CreateUserParams ユーザー作成のパラメータ
type CreateUserParams struct {
	Username string
	Email    string
	Password string
	Locale   string // デフォルトは "ja"
}

// CreateUserResult ユーザー作成の結果
type CreateUserResult struct {
	UserID    int64
	ProfileID int64
}

// CreateUserUsecase ユーザー生成Usecase（シード専用、バルクインサート対応）
type CreateUserUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewCreateUserUsecase 新しいCreateUserUsecaseを作成
func NewCreateUserUsecase(db *sql.DB, queries *query.Queries) *CreateUserUsecase {
	return &CreateUserUsecase{
		db:      db,
		queries: queries,
	}
}

// ExecuteBatch 複数のユーザーをバッチで作成します
// 1000件ごとにコミットしてパフォーマンスを最適化します
func (uc *CreateUserUsecase) ExecuteBatch(ctx context.Context, users []CreateUserParams, progressBar *progressbar.ProgressBar) ([]CreateUserResult, error) {
	return uc.executeBatchWithTx(ctx, nil, users, progressBar)
}

// ExecuteBatchWithTx 複数のユーザーをバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateUserUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, users []CreateUserParams, progressBar *progressbar.ProgressBar) ([]CreateUserResult, error) {
	return uc.executeBatchWithTx(ctx, tx, users, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateUserUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, users []CreateUserParams, progressBar *progressbar.ProgressBar) ([]CreateUserResult, error) {
	results := make([]CreateUserResult, 0, len(users))

	// 既存トランザクションがある場合は、バッチサイズを無視して全件処理
	if existingTx != nil {
		// マルチ行INSERTのチャンクサイズ（100件ずつ）
		multiInsertChunkSize := 100
		for i := 0; i < len(users); i += multiInsertChunkSize {
			end := i + multiInsertChunkSize
			if end > len(users) {
				end = len(users)
			}
			chunk := users[i:end]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleUsers(ctx, existingTx, chunk)
			if err != nil {
				return nil, fmt.Errorf("ユーザーマルチ行INSERT エラー: %w", err)
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
	multiInsertChunkSize := 500

	for i := 0; i < len(users); i += commitBatchSize {
		end := i + commitBatchSize
		if end > len(users) {
			end = len(users)
		}
		batch := users[i:end]

		// トランザクション開始
		tx, err := uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()

		// バッチ内のユーザーを100件ずつマルチ行INSERTで作成
		for j := 0; j < len(batch); j += multiInsertChunkSize {
			chunkEnd := j + multiInsertChunkSize
			if chunkEnd > len(batch) {
				chunkEnd = len(batch)
			}
			chunk := batch[j:chunkEnd]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleUsers(ctx, tx, chunk)
			if err != nil {
				return nil, fmt.Errorf("ユーザーマルチ行INSERT エラー: %w", err)
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

// createMultipleUsers 複数のユーザーとプロフィールをマルチ行INSERTで作成します（トランザクション内）
func (uc *CreateUserUsecase) createMultipleUsers(ctx context.Context, tx *sql.Tx, usersList []CreateUserParams) ([]CreateUserResult, error) {
	if len(usersList) == 0 {
		return []CreateUserResult{}, nil
	}

	// 1. パスワードをハッシュ化（並列化で高速化）
	encryptedPasswords := make([]string, len(usersList))

	// goroutineの並列度を制限するセマフォ（CPUコア数 x 2）
	// bcryptは計算コストが高いため、並列化することで大幅な高速化が期待できる
	maxConcurrency := runtime.NumCPU() * 2
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	errChan := make(chan error, len(usersList))

	for i, params := range usersList {
		wg.Add(1)
		go func(idx int, p CreateUserParams) {
			defer wg.Done()

			// セマフォを取得（並列度を制限）
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			encrypted, err := auth.HashPassword(p.Password)
			if err != nil {
				errChan <- fmt.Errorf("パスワードハッシュ化エラー（username: %s）: %w", p.Username, err)
				return
			}
			encryptedPasswords[idx] = encrypted
		}(i, params)
	}

	// 全てのgoroutineの完了を待つ
	wg.Wait()
	close(errChan)

	// エラーチェック（最初のエラーのみ返す）
	if err := <-errChan; err != nil {
		return nil, err
	}

	// 2. usersテーブルにマルチ行INSERTでユーザーを一括挿入
	userIDs, err := uc.createMultipleUsersInDB(ctx, tx, usersList, encryptedPasswords)
	if err != nil {
		return nil, fmt.Errorf("usersテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 3. profilesテーブルにマルチ行INSERTでプロフィールを一括挿入
	profileIDs, err := uc.createMultipleProfilesInDB(ctx, tx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("profilesテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 4. settingsテーブルにマルチ行INSERTで設定を一括挿入
	if err := uc.createMultipleSettingsInDB(ctx, tx, userIDs); err != nil {
		return nil, fmt.Errorf("settingsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 5. email_notificationsテーブルにマルチ行INSERTでメール通知設定を一括挿入
	if err := uc.createMultipleEmailNotificationsInDB(ctx, tx, userIDs); err != nil {
		return nil, fmt.Errorf("email_notificationsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	// 6. 結果を返す
	results := make([]CreateUserResult, len(userIDs))
	for i := range userIDs {
		results[i] = CreateUserResult{
			UserID:    userIDs[i],
			ProfileID: profileIDs[i],
		}
	}

	return results, nil
}

// createMultipleUsersInDB usersテーブルに複数のユーザーをマルチ行INSERTで挿入します
func (uc *CreateUserUsecase) createMultipleUsersInDB(ctx context.Context, tx *sql.Tx, usersList []CreateUserParams, encryptedPasswords []string) ([]int64, error) {
	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO users (
		username, email, role, locale,
		created_at, updated_at,
		encrypted_password, sign_in_count,
		time_zone, allowed_locales
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, params := range usersList {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は10個のパラメータ）
		offset := i * 10
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10)

		// デフォルト値の設定
		locale := params.Locale
		if locale == "" {
			locale = "ja"
		}

		// パラメータを追加
		values = append(values,
			params.Username,
			params.Email,
			0,                              // role (0 = user, 1 = admin, 2 = editor)
			locale,                         // locale
			now,                            // created_at
			now,                            // updated_at
			encryptedPasswords[i],          // encrypted_password
			0,                              // sign_in_count
			"Asia/Tokyo",                   // time_zone
			pq.Array([]string{"ja", "en"}), // allowed_locales (PostgreSQL配列)
		)
	}

	queryBuilder += " RETURNING id"

	// マルチ行INSERTを実行
	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("usersテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	// 挿入されたIDを取得
	userIDs := make([]int64, 0, len(usersList))
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return userIDs, nil
}

// createMultipleProfilesInDB profilesテーブルに複数のプロフィールをマルチ行INSERTで挿入します
func (uc *CreateUserUsecase) createMultipleProfilesInDB(ctx context.Context, tx *sql.Tx, userIDs []int64) ([]int64, error) {
	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO profiles (
		user_id, name, description,
		created_at, updated_at,
		background_image_animated
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, userID := range userIDs {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は6個のパラメータ）
		offset := i * 6
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6)

		// パラメータを追加
		values = append(values,
			userID, // user_id
			"",     // name (デフォルト空文字)
			"",     // description (デフォルト空文字)
			now,    // created_at
			now,    // updated_at
			false,  // background_image_animated
		)
	}

	queryBuilder += " RETURNING id"

	// マルチ行INSERTを実行
	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("profilesテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	// 挿入されたIDを取得
	profileIDs := make([]int64, 0, len(userIDs))
	for rows.Next() {
		var profileID int64
		if err := rows.Scan(&profileID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		profileIDs = append(profileIDs, profileID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return profileIDs, nil
}

// createMultipleSettingsInDB settingsテーブルに複数の設定をマルチ行INSERTで挿入します
func (uc *CreateUserUsecase) createMultipleSettingsInDB(ctx context.Context, tx *sql.Tx, userIDs []int64) error {
	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO settings (
		user_id,
		privacy_policy_agreed,
		hide_record_body,
		slots_sort_type,
		display_option_work_list,
		display_option_user_work_list,
		records_sort_type,
		display_option_record_list,
		share_record_to_twitter,
		share_record_to_facebook,
		share_review_to_twitter,
		share_review_to_facebook,
		hide_supporter_badge,
		share_status_to_twitter,
		share_status_to_facebook,
		timeline_mode,
		created_at,
		updated_at
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, userID := range userIDs {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は18個のパラメータ）
		offset := i * 18
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9,
			offset+10, offset+11, offset+12, offset+13, offset+14, offset+15, offset+16, offset+17, offset+18)

		// パラメータを追加（CompleteSignUpUsecaseと同じデフォルト値）
		values = append(values,
			userID,            // user_id
			true,              // privacy_policy_agreed
			true,              // hide_record_body
			"",                // slots_sort_type
			"list_detailed",   // display_option_work_list
			"grid_detailed",   // display_option_user_work_list
			"created_at_desc", // records_sort_type
			"all_comments",    // display_option_record_list
			false,             // share_record_to_twitter
			false,             // share_record_to_facebook
			false,             // share_review_to_twitter
			false,             // share_review_to_facebook
			false,             // hide_supporter_badge
			false,             // share_status_to_twitter
			false,             // share_status_to_facebook
			"following",       // timeline_mode
			now,               // created_at
			now,               // updated_at
		)
	}

	// マルチ行INSERTを実行
	_, err := tx.ExecContext(ctx, queryBuilder, values...)
	if err != nil {
		return fmt.Errorf("settingsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	return nil
}

// createMultipleEmailNotificationsInDB email_notificationsテーブルに複数のメール通知設定をマルチ行INSERTで挿入します
func (uc *CreateUserUsecase) createMultipleEmailNotificationsInDB(ctx context.Context, tx *sql.Tx, userIDs []int64) error {
	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO email_notifications (
		user_id,
		unsubscription_key,
		event_followed_user,
		event_liked_episode_record,
		event_friends_joined,
		event_next_season_came,
		event_favorite_works_added,
		event_related_works_added,
		created_at,
		updated_at
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, userID := range userIDs {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は10個のパラメータ）
		offset := i * 10
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10)

		// UUIDを生成（CompleteSignUpUsecaseと同じ形式）
		unsubscriptionKey := fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String())

		// パラメータを追加（CompleteSignUpUsecaseと同じデフォルト値）
		values = append(values,
			userID,            // user_id
			unsubscriptionKey, // unsubscription_key
			true,              // event_followed_user
			true,              // event_liked_episode_record
			true,              // event_friends_joined
			true,              // event_next_season_came
			true,              // event_favorite_works_added
			true,              // event_related_works_added
			now,               // created_at
			now,               // updated_at
		)
	}

	// マルチ行INSERTを実行
	_, err := tx.ExecContext(ctx, queryBuilder, values...)
	if err != nil {
		return fmt.Errorf("email_notificationsテーブルへのマルチ行INSERT エラー: %w", err)
	}

	return nil
}

// createSingleUser 単一のユーザーとプロフィールを作成します（トランザクション内）
// 注意: この関数は後方互換性のために残していますが、
// パフォーマンスのためにcreateMultipleUsersの使用を推奨します
//
//lint:ignore U1000 後方互換性のために保持
func (uc *CreateUserUsecase) createSingleUser(ctx context.Context, tx *sql.Tx, params CreateUserParams) (*CreateUserResult, error) {
	// パスワードをbcryptでハッシュ化（Rails互換）
	encryptedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		return nil, fmt.Errorf("パスワードハッシュ化エラー: %w", err)
	}

	// デフォルト値の設定
	locale := params.Locale
	if locale == "" {
		locale = "ja"
	}

	// ユーザーを作成
	userID, err := uc.createUser(ctx, tx, params.Username, params.Email, encryptedPassword, locale)
	if err != nil {
		return nil, fmt.Errorf("ユーザーレコード作成エラー: %w", err)
	}

	// プロフィールを作成
	profileID, err := uc.createProfile(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("プロフィールレコード作成エラー: %w", err)
	}

	return &CreateUserResult{
		UserID:    userID,
		ProfileID: profileID,
	}, nil
}

// createUser usersテーブルにレコードを作成します
//
//lint:ignore U1000 後方互換性のために保持
func (uc *CreateUserUsecase) createUser(ctx context.Context, tx *sql.Tx, username, email, encryptedPassword, locale string) (int64, error) {
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

	var userID int64
	err := tx.QueryRowContext(
		ctx,
		query,
		username,
		email,
		0,      // role (0 = user, 1 = admin, 2 = editor)
		locale, // locale
		time.Now(),
		time.Now(),
		encryptedPassword,              // encrypted_password
		0,                              // sign_in_count
		"Asia/Tokyo",                   // time_zone
		pq.Array([]string{"ja", "en"}), // allowed_locales (PostgreSQL配列)
	).Scan(&userID)

	if err != nil {
		return 0, fmt.Errorf("usersテーブルへの挿入エラー: %w", err)
	}

	return userID, nil
}

// createProfile profilesテーブルにレコードを作成します
//
//lint:ignore U1000 後方互換性のために保持
func (uc *CreateUserUsecase) createProfile(ctx context.Context, tx *sql.Tx, userID int64) (int64, error) {
	query := `
		INSERT INTO profiles (
			user_id, name, description,
			created_at, updated_at,
			background_image_animated
		) VALUES (
			$1, $2, $3,
			$4, $5,
			$6
		) RETURNING id
	`

	var profileID int64
	err := tx.QueryRowContext(
		ctx,
		query,
		userID,
		"", // name (デフォルト空文字)
		"", // description (デフォルト空文字)
		time.Now(),
		time.Now(),
		false, // background_image_animated
	).Scan(&profileID)

	if err != nil {
		return 0, fmt.Errorf("profilesテーブルへの挿入エラー: %w", err)
	}

	return profileID, nil
}
