package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// FeatureFlagRepository はフィーチャーフラグ関連のデータアクセスを担当
type FeatureFlagRepository struct {
	queries *query.Queries
}

// NewFeatureFlagRepository はFeatureFlagRepositoryを作成
func NewFeatureFlagRepository(queries *query.Queries) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返す
func (r *FeatureFlagRepository) WithTx(tx *sql.Tx) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: r.queries.WithTx(tx)}
}

// IsEnabledByDeviceOrUser はデバイストークンまたはユーザーIDでフラグが有効かどうかを返す
// deviceTokenが空文字列の場合はデバイストークンによるマッチはスキップされる
// userIDが0の場合はユーザーIDによるマッチはスキップされる
func (r *FeatureFlagRepository) IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID int64, name model.FeatureFlagName) (bool, error) {
	dtParam := sql.NullString{}
	if deviceToken != "" {
		dtParam = sql.NullString{String: deviceToken, Valid: true}
	}

	uidParam := sql.NullInt64{}
	if userID != 0 {
		uidParam = sql.NullInt64{Int64: userID, Valid: true}
	}

	return r.queries.IsFeatureFlagEnabled(ctx, query.IsFeatureFlagEnabledParams{
		DeviceToken: dtParam,
		UserID:      uidParam,
		Name:        string(name),
	})
}

// IsEnabled は指定ユーザーに対してフラグが有効かどうかを返す
func (r *FeatureFlagRepository) IsEnabled(ctx context.Context, userID int64, name model.FeatureFlagName) (bool, error) {
	return r.IsEnabledByDeviceOrUser(ctx, "", userID, name)
}
