package seed

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/schollz/progressbar/v3"
)

// CreateOAuthTokenParams OAuth トークン生成のパラメータ
type CreateOAuthTokenParams struct {
	// OAuthアプリケーション情報
	ApplicationName string
	ApplicationUID  string
	RedirectURI     string
	Scopes          string

	// アクセストークン生成数
	TokenCount int

	// トークンを付与するユーザーIDのリスト
	UserIDs []int64
}

// CreateOAuthTokenResult OAuth トークン生成の結果
type CreateOAuthTokenResult struct {
	ApplicationID int64
	TokenIDs      []int64
}

// CreateOAuthTokenUsecase OAuth トークン生成 Usecase（シード専用、バルクインサート対応）
type CreateOAuthTokenUsecase struct {
	db      *sql.DB
	queries *query.Queries
}

// NewCreateOAuthTokenUsecase 新しい CreateOAuthTokenUsecase を作成
func NewCreateOAuthTokenUsecase(db *sql.DB, queries *query.Queries) *CreateOAuthTokenUsecase {
	return &CreateOAuthTokenUsecase{
		db:      db,
		queries: queries,
	}
}

// Execute OAuth アプリケーションとアクセストークンを作成します
func (uc *CreateOAuthTokenUsecase) Execute(ctx context.Context, params CreateOAuthTokenParams, progressBar *progressbar.ProgressBar) (*CreateOAuthTokenResult, error) {
	return uc.executeWithTx(ctx, nil, params, progressBar)
}

// ExecuteWithTx OAuth アプリケーションとアクセストークンを作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateOAuthTokenUsecase) ExecuteWithTx(ctx context.Context, tx *sql.Tx, params CreateOAuthTokenParams, progressBar *progressbar.ProgressBar) (*CreateOAuthTokenResult, error) {
	return uc.executeWithTx(ctx, tx, params, progressBar)
}

// executeWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateOAuthTokenUsecase) executeWithTx(ctx context.Context, existingTx *sql.Tx, params CreateOAuthTokenParams, progressBar *progressbar.ProgressBar) (*CreateOAuthTokenResult, error) {
	// トランザクション処理
	var tx *sql.Tx
	var err error
	shouldCommit := false

	if existingTx != nil {
		tx = existingTx
	} else {
		tx, err = uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()
		shouldCommit = true
	}

	// 1. OAuth アプリケーションを作成
	applicationID, err := uc.createOAuthApplication(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("OAuth アプリケーション作成エラー: %w", err)
	}

	// 進捗バーを更新（アプリケーション作成完了）
	if progressBar != nil {
		progressBar.Add(1)
	}

	// 2. OAuth アクセストークンをバッチで作成
	tokenIDs, err := uc.createOAuthAccessTokensBatch(ctx, tx, applicationID, params.UserIDs, progressBar)
	if err != nil {
		return nil, fmt.Errorf("OAuth アクセストークン作成エラー: %w", err)
	}

	// トランザクションをコミット（既存トランザクションでない場合のみ）
	if shouldCommit {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("トランザクションコミットエラー: %w", err)
		}
	}

	return &CreateOAuthTokenResult{
		ApplicationID: applicationID,
		TokenIDs:      tokenIDs,
	}, nil
}

// createOAuthApplication OAuth アプリケーションを作成します
func (uc *CreateOAuthTokenUsecase) createOAuthApplication(ctx context.Context, tx *sql.Tx, params CreateOAuthTokenParams) (int64, error) {
	// アプリケーション秘密鍵を生成（64文字のランダム文字列）
	secret, err := generateRandomToken(32) // 32バイト = 64文字（hex）
	if err != nil {
		return 0, fmt.Errorf("秘密鍵生成エラー: %w", err)
	}

	// デフォルト値の設定
	applicationName := params.ApplicationName
	if applicationName == "" {
		applicationName = "Test Application"
	}

	applicationUID := params.ApplicationUID
	if applicationUID == "" {
		// UIDを生成（20文字のランダム文字列）
		uid, err := generateRandomToken(10) // 10バイト = 20文字（hex）
		if err != nil {
			return 0, fmt.Errorf("UID生成エラー: %w", err)
		}
		applicationUID = uid
	}

	redirectURI := params.RedirectURI
	if redirectURI == "" {
		redirectURI = "urn:ietf:wg:oauth:2.0:oob" // OAuth 2.0のデフォルトリダイレクトURI（コピペ用）
	}

	scopes := params.Scopes
	if scopes == "" {
		scopes = "" // デフォルトは空文字（全スコープ）
	}

	// OAuth アプリケーションを作成
	query := `
		INSERT INTO oauth_applications (
			name, uid, secret, redirect_uri, scopes,
			aasm_state, created_at, updated_at,
			owner_id, owner_type, confidential,
			hide_social_login
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12
		) RETURNING id
	`

	var applicationID int64
	err = tx.QueryRowContext(
		ctx,
		query,
		applicationName,
		applicationUID,
		secret,
		redirectURI,
		scopes,
		"published", // aasm_state（公開状態）
		time.Now(),  // created_at
		time.Now(),  // updated_at
		nil,         // owner_id（nullの場合は管理者用アプリケーション）
		nil,         // owner_type（nullの場合は管理者用アプリケーション）
		true,        // confidential（機密アプリケーション）
		false,       // hide_social_login
	).Scan(&applicationID)

	if err != nil {
		return 0, fmt.Errorf("oauth_applications テーブルへの挿入エラー: %w", err)
	}

	return applicationID, nil
}

// createOAuthAccessTokensBatch 複数の OAuth アクセストークンをバッチで作成します
func (uc *CreateOAuthTokenUsecase) createOAuthAccessTokensBatch(ctx context.Context, tx *sql.Tx, applicationID int64, userIDs []int64, progressBar *progressbar.ProgressBar) ([]int64, error) {
	if len(userIDs) == 0 {
		return []int64{}, nil
	}

	// マルチ行INSERTのチャンクサイズ（100件ずつ）
	multiInsertChunkSize := 100
	tokenIDs := make([]int64, 0, len(userIDs))

	for i := 0; i < len(userIDs); i += multiInsertChunkSize {
		end := i + multiInsertChunkSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		chunk := userIDs[i:end]

		// マルチ行INSERTで作成
		chunkTokenIDs, err := uc.createMultipleOAuthAccessTokens(ctx, tx, applicationID, chunk)
		if err != nil {
			return nil, fmt.Errorf("OAuth アクセストークン マルチ行INSERT エラー: %w", err)
		}
		tokenIDs = append(tokenIDs, chunkTokenIDs...)

		// 進捗表示を更新
		if progressBar != nil {
			progressBar.Add(len(chunk))
		}
	}

	return tokenIDs, nil
}

// createMultipleOAuthAccessTokens 複数の OAuth アクセストークンをマルチ行INSERTで作成します
func (uc *CreateOAuthTokenUsecase) createMultipleOAuthAccessTokens(ctx context.Context, tx *sql.Tx, applicationID int64, userIDs []int64) ([]int64, error) {
	if len(userIDs) == 0 {
		return []int64{}, nil
	}

	// 1. トークンを事前生成
	tokens := make([]string, len(userIDs))
	for i := range userIDs {
		token, err := generateRandomToken(32) // 32バイト = 64文字（hex）
		if err != nil {
			return nil, fmt.Errorf("トークン生成エラー: %w", err)
		}
		tokens[i] = token
	}

	// 2. マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO oauth_access_tokens (
		resource_owner_id, application_id, token,
		refresh_token, expires_in, revoked_at,
		created_at, scopes, previous_refresh_token,
		description
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

		// パラメータを追加
		values = append(values,
			userID,        // resource_owner_id
			applicationID, // application_id
			tokens[i],     // token
			nil,           // refresh_token（null）
			nil,           // expires_in（null = 無期限）
			nil,           // revoked_at（null = 有効）
			now,           // created_at
			"",            // scopes（空文字 = 全スコープ）
			"",            // previous_refresh_token
			"Test Token",  // description
		)
	}

	queryBuilder += " RETURNING id"

	// マルチ行INSERTを実行
	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("oauth_access_tokens テーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	// 挿入されたIDを取得
	tokenIDs := make([]int64, 0, len(userIDs))
	for rows.Next() {
		var tokenID int64
		if err := rows.Scan(&tokenID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		tokenIDs = append(tokenIDs, tokenID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return tokenIDs, nil
}

// generateRandomToken ランダムなトークンを生成します（hex文字列）
func generateRandomToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("ランダムバイト生成エラー: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
