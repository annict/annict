// Package repositoryはデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// SettingRepositoryはSetting関連のデータアクセスを担当します
type SettingRepository struct {
	queries *query.Queries
}

// NewSettingRepositoryはSettingRepositoryを作成します
func NewSettingRepository(queries *query.Queries) *SettingRepository {
	return &SettingRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *SettingRepository) WithTx(tx *sql.Tx) *SettingRepository {
	return &SettingRepository{queries: r.queries.WithTx(tx)}
}

// Createは設定を作成します
func (r *SettingRepository) Create(ctx context.Context, userID model.UserID) (*model.Setting, error) {
	row, err := r.queries.CreateSetting(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	return &model.Setting{
		ID:                  model.SettingID(row.ID),
		UserID:              model.UserID(row.UserID),
		PrivacyPolicyAgreed: row.PrivacyPolicyAgreed,
		HideRecordBody:      row.HideRecordBody,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}
