package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// StaffRepositoryはStaff関連のデータアクセスを担当します
type StaffRepository struct {
	queries *query.Queries
}

// NewStaffRepositoryはStaffRepositoryを作成します
func NewStaffRepository(queries *query.Queries) *StaffRepository {
	return &StaffRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *StaffRepository) WithTx(tx *sql.Tx) *StaffRepository {
	return &StaffRepository{queries: r.queries.WithTx(tx)}
}

// GetByWorkIDsは作品IDのリストに紐づくスタッフを取得します
// (role='other' と、スタッフ・人物が非公開または削除済みのものは除外)
func (r *StaffRepository) GetByWorkIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Staff, error) {
	if len(workIDs) == 0 {
		return []*model.Staff{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.GetStaffsByWorkIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	staffs := make([]*model.Staff, len(rows))
	for i, row := range rows {
		staffs[i] = staffFromRow(row)
	}
	return staffs, nil
}

// staffFromRowはquery.GetStaffsByWorkIDsRowを *model.Staffに変換します
func staffFromRow(row query.GetStaffsByWorkIDsRow) *model.Staff {
	staff := &model.Staff{
		ID:          model.StaffID(row.ID),
		WorkID:      model.WorkID(row.WorkID),
		Name:        row.Name,
		NameEn:      row.NameEn,
		Role:        row.Role,
		RoleOtherEn: row.RoleOtherEn,
	}
	if row.RoleOther.Valid {
		staff.RoleOther = row.RoleOther.String
	}
	return staff
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きいスタッフの1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *StaffRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileStaffAnimeIDsAfter(ctx, query.ReconcileStaffAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
