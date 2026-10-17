package repository

import (
	"context"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// PrefectureRepositoryは出身地 (prefectures) のデータアクセスを担当します
type PrefectureRepository struct {
	queries *query.Queries
}

// NewPrefectureRepositoryはPrefectureRepositoryを作成します
func NewPrefectureRepository(queries *query.Queries) *PrefectureRepository {
	return &PrefectureRepository{queries: queries}
}

// ListAllは全ての出身地をIDの順で取得します
func (r *PrefectureRepository) ListAll(ctx context.Context) ([]model.Prefecture, error) {
	rows, err := r.queries.ListPrefectures(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Prefecture, len(rows))
	for i, row := range rows {
		result[i] = model.Prefecture{
			ID:   model.PrefectureID(row.ID),
			Name: row.Name,
		}
	}
	return result, nil
}

// ExistsByIDは指定の出身地が登録されているかを返します。people.prefecture_idは
// prefectures(id) への外部キーのため、どの行も指さない値はカラムに届かずUPDATEが失敗します。
func (r *PrefectureRepository) ExistsByID(ctx context.Context, id model.PrefectureID) (bool, error) {
	return r.queries.ExistsPrefectureByID(ctx, int64(id))
}
