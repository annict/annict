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

// ReconcileAnimeIDsAfterは、IDがafterIDより大きいキャストの1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *CastRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileCastAnimeIDsAfter(ctx, query.ReconcileCastAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
