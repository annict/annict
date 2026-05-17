package repository

import (
	"context"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// NumberFormatRepository はNumberFormat関連のデータアクセスを担当します
type NumberFormatRepository struct {
	queries *query.Queries
}

// NewNumberFormatRepository はNumberFormatRepositoryを作成します
func NewNumberFormatRepository(queries *query.Queries) *NumberFormatRepository {
	return &NumberFormatRepository{queries: queries}
}

// ListAll は全てのNumberFormatをsort_number順で取得します
func (r *NumberFormatRepository) ListAll(ctx context.Context) ([]model.NumberFormat, error) {
	rows, err := r.queries.ListNumberFormats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.NumberFormat, len(rows))
	for i, row := range rows {
		result[i] = model.NumberFormat{
			ID:         model.NumberFormatID(row.ID),
			Name:       row.Name,
			SortNumber: row.SortNumber,
		}
	}
	return result, nil
}
