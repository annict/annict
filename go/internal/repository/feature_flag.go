package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// FeatureFlagRepository handles data access for the feature_flags table.
// [Ja] FeatureFlagRepository は feature_flags テーブルへのデータアクセスを担う。
type FeatureFlagRepository struct {
	queries *query.Queries
}

func NewFeatureFlagRepository(queries *query.Queries) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: queries}
}

func (r *FeatureFlagRepository) WithTx(tx *sql.Tx) *FeatureFlagRepository {
	return &FeatureFlagRepository{queries: r.queries.WithTx(tx)}
}

// IsEnabledByDeviceOrUser reports whether the flag is enabled for the caller,
// matching by device token, user id, or both. An empty deviceToken skips the
// device-token match and a zero userID skips the user-id match; at least one
// of them is expected to be supplied.
//
// [Ja] フラグがデバイストークン・ユーザー ID・両方のいずれかでマッチして有効になるかを返す。
// deviceToken が空文字列ならデバイストークンによるマッチをスキップし、
// userID が 0 ならユーザー ID によるマッチをスキップする。
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
