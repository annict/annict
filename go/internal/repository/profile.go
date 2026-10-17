// Package repositoryはデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ProfileRepositoryはProfile関連のデータアクセスを担当します
type ProfileRepository struct {
	queries *query.Queries
}

// NewProfileRepositoryはProfileRepositoryを作成します
func NewProfileRepository(queries *query.Queries) *ProfileRepository {
	return &ProfileRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ProfileRepository) WithTx(tx *sql.Tx) *ProfileRepository {
	return &ProfileRepository{queries: r.queries.WithTx(tx)}
}

// Createはプロフィールを作成します
func (r *ProfileRepository) Create(ctx context.Context, userID model.UserID, name string) (*model.Profile, error) {
	row, err := r.queries.CreateProfile(ctx, query.CreateProfileParams{
		UserID: int64(userID),
		Name:   name,
	})
	if err != nil {
		return nil, err
	}

	return &model.Profile{
		ID:          model.ProfileID(row.ID),
		UserID:      model.UserID(row.UserID),
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

// ListImageDataByUserIDはユーザーのプロフィール画像と背景画像のShrineの添付データのうち、
// 空でないものを返します。
func (r *ProfileRepository) ListImageDataByUserID(ctx context.Context, userID model.UserID) ([]string, error) {
	rows, err := r.queries.ListProfileImageDataByUserID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	var imageData []string
	for _, row := range rows {
		for _, data := range []sql.NullString{row.ImageData, row.BackgroundImageData} {
			if data.Valid && data.String != "" {
				imageData = append(imageData, data.String)
			}
		}
	}
	return imageData, nil
}
