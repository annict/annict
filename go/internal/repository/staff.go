package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// StaffRepository はStaff関連のデータアクセスを担当します
type StaffRepository struct {
	queries *query.Queries
}

// NewStaffRepository はStaffRepositoryを作成します
func NewStaffRepository(queries *query.Queries) *StaffRepository {
	return &StaffRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *StaffRepository) WithTx(tx *sql.Tx) *StaffRepository {
	return &StaffRepository{queries: r.queries.WithTx(tx)}
}

// GetByWorkIDs は作品IDのリストに紐づくスタッフを取得します（role='other' は除外）
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

// staffFromRow は query.GetStaffsByWorkIDsRow を *model.Staff に変換します
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
