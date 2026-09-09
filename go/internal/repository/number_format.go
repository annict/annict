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

// ExistsByID reports whether the number format is registered. works.number_format_id and
// anime_classifications.number_format_id are both foreign keys to number_formats(id), so a
// value that names no row fails the INSERT rather than reaching the column.
//
// [Ja] ExistsByID は指定の話数フォーマットが登録されているかを返す。
// works.number_format_id と anime_classifications.number_format_id はいずれも
// number_formats(id) への外部キーのため、どの行も指さない値はカラムに届かず INSERT が失敗する。
func (r *NumberFormatRepository) ExistsByID(ctx context.Context, id model.NumberFormatID) (bool, error) {
	return r.queries.ExistsNumberFormatByID(ctx, int64(id))
}
