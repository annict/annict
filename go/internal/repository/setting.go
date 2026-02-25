// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/query"
)

// SettingRepository はSetting関連のデータアクセスを担当します
type SettingRepository struct {
	queries *query.Queries
}

// NewSettingRepository はSettingRepositoryを作成します
func NewSettingRepository(queries *query.Queries) *SettingRepository {
	return &SettingRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *SettingRepository) WithTx(tx *sql.Tx) *SettingRepository {
	return &SettingRepository{queries: r.queries.WithTx(tx)}
}

// Create は設定を作成します
func (r *SettingRepository) Create(ctx context.Context, userID int64) error {
	_, err := r.queries.CreateSetting(ctx, userID)
	return err
}
