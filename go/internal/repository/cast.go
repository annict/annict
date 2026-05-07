package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CastRepository はCast関連のデータアクセスを担当します
type CastRepository struct {
	queries *query.Queries
}

// NewCastRepository はCastRepositoryを作成します
func NewCastRepository(queries *query.Queries) *CastRepository {
	return &CastRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *CastRepository) WithTx(tx *sql.Tx) *CastRepository {
	return &CastRepository{queries: r.queries.WithTx(tx)}
}

// GetByWorkIDs は作品IDのリストに紐づくキャストを取得します
func (r *CastRepository) GetByWorkIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Cast, error) {
	if len(workIDs) == 0 {
		return []*model.Cast{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.GetCastsByWorkIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	casts := make([]*model.Cast, len(rows))
	for i, row := range rows {
		casts[i] = castFromRow(row)
	}
	return casts, nil
}

// castFromRow は query.GetCastsByWorkIDsRow を *model.Cast に変換します
func castFromRow(row query.GetCastsByWorkIDsRow) *model.Cast {
	cast := &model.Cast{
		ID:     model.CastID(row.ID),
		WorkID: model.WorkID(row.WorkID),
		Name:   row.Name,
		NameEn: row.NameEn,
	}
	if row.CharacterName.Valid {
		cast.CharacterName = row.CharacterName.String
	}
	if row.CharacterNameEn.Valid {
		cast.CharacterNameEn = row.CharacterNameEn.String
	}
	if row.PersonName.Valid {
		cast.PersonName = row.PersonName.String
	}
	if row.PersonNameEn.Valid {
		cast.PersonNameEn = row.PersonNameEn.String
	}
	return cast
}
