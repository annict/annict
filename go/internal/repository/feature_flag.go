package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// FeatureFlagRepositoryはfeature_flagsテーブルへのデータアクセスを担う。
type FeatureFlagRepository struct {
	queries *query.Queries
}

func NewFeatureFlagRepository(queries *query.Queries) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: queries}
}

func (r *FeatureFlagRepository) WithTx(tx *sql.Tx) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: r.queries.WithTx(tx)}
}

// IsEnabledByDeviceOrUserはフラグがデバイストークン・ユーザーID・両方のいずれかでマッチして有効になるかを返す。
// deviceTokenが空文字列ならデバイストークンによるマッチをスキップし、
// userIDが0ならユーザーIDによるマッチをスキップする。
// 呼び出し側は少なくとも一方を渡すこと。
func (r *FeatureFlagRepository) IsEnabledByDeviceOrUser(ctx context.Context, deviceToken string, userID model.UserID, name model.FeatureFlagName) (bool, error) {
	dtParam := sql.NullString{}
	if deviceToken != "" {
		dtParam = sql.NullString{String: deviceToken, Valid: true}
	}

	uidParam := sql.NullInt64{}
	if userID != 0 {
		uidParam = sql.NullInt64{Int64: int64(userID), Valid: true}
	}

	return r.queries.IsFeatureFlagEnabled(ctx, query.IsFeatureFlagEnabledParams{
		DeviceToken: dtParam,
		UserID:      uidParam,
		Name:        string(name),
	})
}

func (r *FeatureFlagRepository) IsEnabled(ctx context.Context, userID model.UserID, name model.FeatureFlagName) (bool, error) {
	return r.IsEnabledByDeviceOrUser(ctx, "", userID, name)
}
