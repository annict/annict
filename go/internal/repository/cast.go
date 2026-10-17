package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CastRepositoryはCast関連のデータアクセスを担当します
type CastRepository struct {
	queries *query.Queries
}

// NewCastRepositoryはCastRepositoryを作成します
func NewCastRepository(queries *query.Queries) *CastRepository {
	return &CastRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *CastRepository) WithTx(tx *sql.Tx) *CastRepository {
	return &CastRepository{queries: r.queries.WithTx(tx)}
}

// GetByWorkIDsは作品IDのリストに紐づくキャストを取得します (キャスト・人物が非公開または削除済みのものは除外)
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

// castFromRowはquery.GetCastsByWorkIDsRowを *model.Castに変換します
func castFromRow(row query.GetCastsByWorkIDsRow) *model.Cast {
	cast := &model.Cast{
		ID:     model.CastID(row.ID),
		WorkID: model.WorkID(row.WorkID),
		Name:   row.Name,
		NameEn: row.NameEn,
		Person: &model.Person{
			ID:     model.PersonID(row.PersonID),
			Name:   row.PersonName,
			NameEn: row.PersonNameEn,
		},
	}
	// charactersはLEFT JOINのため、キャラクターのレコードが見つからない行ではNULLになる
	if row.CharacterID.Valid {
		cast.Character = &model.Character{
			ID:     model.CharacterID(row.CharacterID.Int64),
			Name:   row.CharacterName.String,
			NameEn: row.CharacterNameEn.String,
		}
	}
	return cast
}
